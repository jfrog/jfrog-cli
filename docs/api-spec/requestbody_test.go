package apispec

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// parseRawDoc is a test-only helper: unlike parseFile, it takes a literal YAML
// string instead of reading an embedded file, so these tests can exercise
// buildRequestBody against synthetic fixtures without needing new stub files.
func parseRawDoc(t *testing.T, doc string) rawDoc {
	t.Helper()
	var d rawDoc
	require.NoError(t, yaml.Unmarshal([]byte(doc), &d))
	return d
}

func TestBuildRequestBody_NilWhenNoRequestBody(t *testing.T) {
	assert.Nil(t, buildRequestBody(nil, nil))
}

func TestBuildRequestBody_NilWhenNoJSONContent(t *testing.T) {
	raw := &rawRequestBody{Required: true, Content: map[string]rawMediaTypeItem{
		"multipart/form-data": {Schema: rawSchema{Type: "object"}},
	}}
	assert.Nil(t, buildRequestBody(raw, nil))
}

func TestBuildRequestBody_RefResolution(t *testing.T) {
	d := parseRawDoc(t, `
components:
  schemas:
    Widget:
      type: object
      required: [name]
      properties:
        name:
          type: string
          description: Widget name
        active:
          type: boolean
          default: true
`)
	raw := &rawRequestBody{Required: true, Content: map[string]rawMediaTypeItem{
		"application/json": {Schema: rawSchema{Ref: "#/components/schemas/Widget"}},
	}}

	rb := buildRequestBody(raw, d.Components.Schemas)
	require.NotNil(t, rb)
	assert.True(t, rb.Required)
	require.Len(t, rb.Properties, 2)

	byName := make(map[string]Property, len(rb.Properties))
	for _, p := range rb.Properties {
		byName[p.Name] = p
	}
	assert.True(t, byName["name"].Required)
	assert.Equal(t, "string", byName["name"].Type)
	assert.Equal(t, "Widget name", byName["name"].Description)
	assert.False(t, byName["active"].Required)
	assert.Equal(t, "true", byName["active"].Default)
}

func TestBuildRequestBody_InlineSchemaNoRef(t *testing.T) {
	raw := &rawRequestBody{Required: false, Content: map[string]rawMediaTypeItem{
		"application/json": {Schema: rawSchema{
			Type:     "object",
			Required: []string{"password"},
			Properties: map[string]rawSchema{
				"password": {Type: "string"},
			},
		}},
	}}

	rb := buildRequestBody(raw, nil)
	require.NotNil(t, rb)
	assert.False(t, rb.Required)
	require.Len(t, rb.Properties, 1)
	assert.Equal(t, "password", rb.Properties[0].Name)
	assert.True(t, rb.Properties[0].Required)
}

func TestPropertyType(t *testing.T) {
	tests := []struct {
		name string
		in   rawSchema
		want string
	}{
		{"primitive string", rawSchema{Type: "string"}, "string"},
		{"untyped falls back to object", rawSchema{}, "object"},
		{"nested $ref is not flattened", rawSchema{Ref: "#/components/schemas/BuildTarget"}, "BuildTarget"},
		{"array of primitives", rawSchema{Type: "array", Items: &rawSchema{Type: "string"}}, "array<string>"},
		{"array of refs", rawSchema{Type: "array", Items: &rawSchema{Ref: "#/components/schemas/PatternFilter"}}, "array<PatternFilter>"},
		{"array with no items info", rawSchema{Type: "array"}, "array<object>"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, propertyType(tt.in))
		})
	}
}

func TestSchemaRefName(t *testing.T) {
	assert.Equal(t, "Widget", schemaRefName("#/components/schemas/Widget"))
	assert.Equal(t, "no-slash", schemaRefName("no-slash"))
}

func TestStringifyDefault(t *testing.T) {
	assert.Equal(t, "", stringifyDefault(nil))
	assert.Equal(t, "true", stringifyDefault(true))
	assert.Equal(t, "0", stringifyDefault(0))
	assert.Equal(t, "hello", stringifyDefault("hello"))
}
