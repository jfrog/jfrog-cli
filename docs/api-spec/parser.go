// Package apispec exposes the operations declared in jfrog-cli's embedded
// OpenAPI spec bundle (a small "stub" set by default, or the real "full" set
// in JFrog's internal release build — see docs/api-spec/).
package apispec

import (
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
	Name        string `yaml:"name"`
	In          string `yaml:"in"`
	Required    bool   `yaml:"required"`
	Description string `yaml:"description"`
}

// Operation describes a single OpenAPI path+method operation.
type Operation struct {
	Method      string
	Path        string
	Summary     string
	Tags        []string
	OperationId string
	Parameters  []Parameter
}

// Metadata describes which spec bundle is embedded in this binary.
type Metadata struct {
	SpecBundle  string
	SpecVersion string
}

type rawDoc struct {
	Paths map[string]map[string]yaml.Node `yaml:"paths"`
}

type rawOperation struct {
	Summary     string      `yaml:"summary"`
	OperationId string      `yaml:"operationId"`
	Tags        []string    `yaml:"tags"`
	Parameters  []Parameter `yaml:"parameters"`
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
			})
		}
	}
	return ops, nil
}
