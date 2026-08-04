package apispec

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildResponses_NilWhenEmpty(t *testing.T) {
	assert.Nil(t, buildResponses(nil))
	assert.Nil(t, buildResponses(map[string]rawResponse{}))
}

func TestBuildResponses_SortedByCode(t *testing.T) {
	raw := map[string]rawResponse{
		"404": {Description: "Not Found"},
		"200": {Description: "Success"},
		"400": {Description: "Bad Request"},
	}
	responses := buildResponses(raw)
	require.Len(t, responses, 3)
	assert.Equal(t, []Response{
		{Code: "200", Description: "Success"},
		{Code: "400", Description: "Bad Request"},
		{Code: "404", Description: "Not Found"},
	}, responses)
}

func TestBuildResponses_MissingDescriptionIsEmpty(t *testing.T) {
	responses := buildResponses(map[string]rawResponse{"204": {}})
	require.Len(t, responses, 1)
	assert.Equal(t, "204", responses[0].Code)
	assert.Empty(t, responses[0].Description)
}

func TestBuildExample_NilWhenAbsent(t *testing.T) {
	assert.Nil(t, buildExample(nil))
}

func TestBuildExample_RoundTripsNestedMapping(t *testing.T) {
	raw := &rawRequestBody{Required: true, Content: map[string]rawMediaTypeItem{
		"application/json": {
			Schema: rawSchema{Type: "object"},
			Example: map[string]any{
				"username": "newuser",
				"active":   true,
				"groups":   []any{"readers", "writers"},
				"meta":     map[string]any{"count": 2},
			},
		},
	}}

	rb := buildRequestBody(raw, nil)
	require.NotNil(t, rb)
	require.NotNil(t, rb.Example)
	assert.JSONEq(t, `{"username":"newuser","active":true,"groups":["readers","writers"],"meta":{"count":2}}`, string(rb.Example))
}

func TestBuildRequestBody_NoExampleLeavesFieldNil(t *testing.T) {
	raw := &rawRequestBody{Required: true, Content: map[string]rawMediaTypeItem{
		"application/json": {Schema: rawSchema{Type: "object"}},
	}}
	rb := buildRequestBody(raw, nil)
	require.NotNil(t, rb)
	assert.Nil(t, rb.Example)
}

func TestFindOperation(t *testing.T) {
	op, ok := FindOperation("get", "/access/api/v2/users")
	require.True(t, ok, "case-insensitive method match should find getUserList")
	assert.Equal(t, "getUserList", op.OperationId)

	_, ok = FindOperation("DELETE", "/access/api/v2/users")
	assert.False(t, ok, "wrong method for an existing path should not match")

	_, ok = FindOperation("GET", "/access/api/v2/user")
	assert.False(t, ok, "a path that's a prefix/substring of a real one should not match")

	_, ok = FindOperation("GET", "/not/a/real/path")
	assert.False(t, ok)
}
