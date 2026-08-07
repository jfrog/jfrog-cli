// Package apispec exposes the operations declared in jfrog-cli's embedded
// OpenAPI spec bundle (a small "stub" set by default, or the real "full" set
// in JFrog's internal release build — see docs/api-spec/).
package apispec

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

var httpMethods = map[string]bool{
	"get": true, "put": true, "post": true, "delete": true,
	"options": true, "head": true, "patch": true, "trace": true,
}

// Parameter describes a single OpenAPI operation parameter.
type Parameter struct {
	Name        string `yaml:"name" json:"name"`
	In          string `yaml:"in" json:"in"`
	Required    bool   `yaml:"required" json:"required"`
	Description string `yaml:"description" json:"description,omitempty"`
}

// Property describes a single top-level field of an operation's JSON request
// body. Type is either a JSON-schema primitive ("string", "boolean",
// "integer", "object", ...), "array<item-type>" for arrays, or the name of a
// referenced component schema (e.g. "BuildTarget") when the property itself
// is a $ref -- that nested schema's own fields are not flattened further.
type Property struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description,omitempty"`
	Default     string `json:"default,omitempty"`
}

// RequestBody describes an operation's JSON request body payload.
type RequestBody struct {
	Required   bool       `json:"required"`
	Properties []Property `json:"properties,omitempty"`
	// Example is the operation's declared requestBody.content.application/json.example,
	// verbatim, when the spec provides one. yaml.v3 decodes YAML mappings into
	// map[string]interface{} (not v2's map[interface{}]interface{}), so this
	// round-trips through encoding/json without a custom converter.
	Example json.RawMessage `json:"example,omitempty"`
}

// Response describes a single declared HTTP response for an operation.
type Response struct {
	Code        string `json:"code"`
	Description string `json:"description,omitempty"`
}

// Operation describes a single OpenAPI path+method operation.
type Operation struct {
	Method      string
	Path        string
	Summary     string
	Tags        []string
	OperationId string
	Parameters  []Parameter
	// RequestBody is nil for operations with no application/json request body
	// (typically GET/DELETE).
	RequestBody *RequestBody
	// Responses is sorted by status code ascending; empty when the spec
	// declares no responses object for this operation.
	Responses []Response
}

// Metadata describes which spec bundle is embedded in this binary.
type Metadata struct {
	SpecBundle  string
	SpecVersion string
}

type rawDoc struct {
	Paths      map[string]map[string]yaml.Node `yaml:"paths"`
	Components struct {
		Schemas map[string]rawSchema `yaml:"schemas"`
	} `yaml:"components"`
}

type rawOperation struct {
	Summary     string                 `yaml:"summary"`
	OperationId string                 `yaml:"operationId"`
	Tags        []string               `yaml:"tags"`
	Parameters  []Parameter            `yaml:"parameters"`
	RequestBody *rawRequestBody        `yaml:"requestBody"`
	Responses   map[string]rawResponse `yaml:"responses"`
}

type rawResponse struct {
	Description string `yaml:"description"`
}

type rawRequestBody struct {
	Required bool                        `yaml:"required"`
	Content  map[string]rawMediaTypeItem `yaml:"content"`
}

type rawMediaTypeItem struct {
	Schema  rawSchema `yaml:"schema"`
	Example any       `yaml:"example"`
}

// rawSchema is a deliberately narrow subset of OpenAPI's Schema Object: only
// what's needed to flatten a request body's top-level properties into
// Property. Nested object properties are identified by referenced schema name
// (see Property) rather than recursively flattened, to keep output compact
// and sidestep any $ref cycles in the full (non-stub) bundle.
type rawSchema struct {
	Ref         string               `yaml:"$ref"`
	Type        string               `yaml:"type"`
	Description string               `yaml:"description"`
	Default     any                  `yaml:"default"`
	Required    []string             `yaml:"required"`
	Properties  map[string]rawSchema `yaml:"properties"`
	Items       *rawSchema           `yaml:"items"`
}

var (
	once       sync.Once
	operations []Operation
	parseErr   error
)

// Operations returns every operation across the embedded OpenAPI spec bundle.
// Parsing happens once per process and the result is cached.
func Operations() ([]Operation, error) {
	once.Do(func() {
		operations, parseErr = parseAll()
	})
	return operations, parseErr
}

// Info reports which spec bundle is embedded and, for full builds, the
// rdme-admin commit it was fetched from.
func Info() Metadata {
	return Metadata{
		SpecBundle:  Bundle,
		SpecVersion: specVersion(),
	}
}

// isSpecFile reports whether name is a top-level OpenAPI YAML file that should
// be parsed. Excludes rdme-admin's per-endpoint _order.yaml nav files and any
// dotfile (including this package's own full/.placeholder.yaml).
func isSpecFile(name string) bool {
	return strings.HasSuffix(name, ".yaml") && !strings.HasPrefix(name, ".") && !strings.HasPrefix(name, "_")
}

func parseAll() ([]Operation, error) {
	entries, err := specFS.ReadDir(rootDir)
	if err != nil {
		return nil, fmt.Errorf("apispec: reading %s: %w", rootDir, err)
	}

	var ops []Operation
	for _, entry := range entries {
		if entry.IsDir() || !isSpecFile(entry.Name()) {
			continue
		}
		fileOps, err := parseFile(rootDir + "/" + entry.Name())
		if err != nil {
			return nil, fmt.Errorf("apispec: parsing %s: %w", entry.Name(), err)
		}
		ops = append(ops, fileOps...)
	}

	sort.Slice(ops, func(i, j int) bool {
		if ops[i].Path != ops[j].Path {
			return ops[i].Path < ops[j].Path
		}
		return ops[i].Method < ops[j].Method
	})
	return ops, nil
}

func parseFile(name string) ([]Operation, error) {
	data, err := specFS.ReadFile(name)
	if err != nil {
		return nil, err
	}

	var doc rawDoc
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}

	var ops []Operation
	for p, methods := range doc.Paths {
		for method, node := range methods {
			lower := strings.ToLower(method)
			if !httpMethods[lower] {
				continue
			}
			var op rawOperation
			if err := node.Decode(&op); err != nil {
				return nil, fmt.Errorf("path %s method %s: %w", p, method, err)
			}
			ops = append(ops, Operation{
				Method:      strings.ToUpper(method),
				Path:        p,
				Summary:     op.Summary,
				Tags:        op.Tags,
				OperationId: op.OperationId,
				Parameters:  op.Parameters,
				RequestBody: buildRequestBody(op.RequestBody, doc.Components.Schemas),
				Responses:   buildResponses(op.Responses),
			})
		}
	}
	return ops, nil
}

// buildRequestBody flattens a JSON request body's top-level schema properties
// into a RequestBody. Returns nil when there's no application/json content
// (the only content type used across today's rdme-admin reference bundle).
func buildRequestBody(raw *rawRequestBody, schemas map[string]rawSchema) *RequestBody {
	if raw == nil {
		return nil
	}
	media, ok := raw.Content["application/json"]
	if !ok {
		return nil
	}

	schema := media.Schema
	if schema.Ref != "" {
		if resolved, ok := schemas[schemaRefName(schema.Ref)]; ok {
			schema = resolved
		}
	}

	required := make(map[string]bool, len(schema.Required))
	for _, name := range schema.Required {
		required[name] = true
	}

	properties := make([]Property, 0, len(schema.Properties))
	for name, prop := range schema.Properties {
		properties = append(properties, Property{
			Name:        name,
			Type:        propertyType(prop),
			Required:    required[name],
			Description: prop.Description,
			Default:     stringifyDefault(prop.Default),
		})
	}
	sort.Slice(properties, func(i, j int) bool { return properties[i].Name < properties[j].Name })

	return &RequestBody{Required: raw.Required, Properties: properties, Example: buildExample(media.Example)}
}

// buildExample marshals a requestBody media-type's declared example (decoded by
// yaml.v3 into map[string]interface{}/[]interface{}/primitives) into JSON.
// Returns nil when there's no example to marshal or marshaling somehow fails --
// an example is a nice-to-have, not worth failing operation parsing over.
func buildExample(v any) json.RawMessage {
	if v == nil {
		return nil
	}
	data, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return data
}

// buildResponses flattens an operation's responses object into a slice sorted
// by status code ascending. String sort is sufficient for the 2/3/4/5-digit
// numeric codes present in both the stub and full bundles today; neither uses
// wildcard forms like "2XX".
func buildResponses(raw map[string]rawResponse) []Response {
	if len(raw) == 0 {
		return nil
	}
	responses := make([]Response, 0, len(raw))
	for code, r := range raw {
		responses = append(responses, Response{Code: code, Description: r.Description})
	}
	sort.Slice(responses, func(i, j int) bool { return responses[i].Code < responses[j].Code })
	return responses
}

// FindOperation returns the operation matching method (case-insensitive) and
// path (exact, case-sensitive -- paths may contain literal {param} segments,
// e.g. "/worker/api/v1/workers/{workerKey}", matched verbatim against the
// catalog rather than against a concrete instantiated path) from the embedded
// bundle, or ok=false if none matches.
func FindOperation(method, path string) (op Operation, ok bool) {
	ops, err := Operations()
	if err != nil {
		return Operation{}, false
	}
	method = strings.ToUpper(strings.TrimSpace(method))
	for _, o := range ops {
		if o.Method == method && o.Path == path {
			return o, true
		}
	}
	return Operation{}, false
}

// schemaRefName extracts "Foo" from a local-document ref like
// "#/components/schemas/Foo".
func schemaRefName(ref string) string {
	if idx := strings.LastIndex(ref, "/"); idx != -1 {
		return ref[idx+1:]
	}
	return ref
}

// propertyType renders a rawSchema property as a compact type hint: a $ref
// becomes the referenced schema's name (not flattened further), an array
// becomes "array<item-type>", and anything else falls back to its declared
// type or "object" when untyped.
func propertyType(p rawSchema) string {
	if p.Ref != "" {
		return schemaRefName(p.Ref)
	}
	if p.Type == "array" {
		itemType := "object"
		if p.Items != nil {
			switch {
			case p.Items.Ref != "":
				itemType = schemaRefName(p.Items.Ref)
			case p.Items.Type != "":
				itemType = p.Items.Type
			}
		}
		return "array<" + itemType + ">"
	}
	if p.Type == "" {
		return "object"
	}
	return p.Type
}

func stringifyDefault(v any) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}
