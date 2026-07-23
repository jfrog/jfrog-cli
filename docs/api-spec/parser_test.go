//go:build !full

// These tests assert exact values from the stub fixtures and don't apply to a
// full build (whose docs/api-spec/full/ content is populated at release time).

package apispec

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOperations_Stub(t *testing.T) {
	ops, err := Operations()
	require.NoError(t, err)
	require.Len(t, ops, 10, "expected exactly the 10 trimmed real operations across the 7 stub files")

	byOperationId := make(map[string]Operation, len(ops))
	for _, op := range ops {
		byOperationId[op.OperationId] = op
	}

	getUserList, ok := byOperationId["getUserList"]
	require.True(t, ok, "getUserList should be present")
	assert.Equal(t, "GET", getUserList.Method)
	assert.Equal(t, "/access/api/v2/users", getUserList.Path)
	assert.Equal(t, "Get User List", getUserList.Summary)
	assert.Equal(t, []string{"Users"}, getUserList.Tags)
	assert.Len(t, getUserList.Parameters, 10)

	assert.Nil(t, getUserList.RequestBody, "a GET operation should have no request body")

	createUser, ok := byOperationId["createUser"]
	require.True(t, ok, "createUser should be present")
	assert.Equal(t, "POST", createUser.Method)
	assert.Equal(t, "/access/api/v2/users", createUser.Path)

	require.NotNil(t, createUser.RequestBody, "createUser's requestBody ($ref: UserCreateRequest) should resolve")
	assert.True(t, createUser.RequestBody.Required)
	propsByName := make(map[string]Property, len(createUser.RequestBody.Properties))
	for _, p := range createUser.RequestBody.Properties {
		propsByName[p.Name] = p
	}
	username, ok := propsByName["username"]
	require.True(t, ok, "username should be a flattened property of UserCreateRequest")
	assert.Equal(t, "string", username.Type)
	assert.True(t, username.Required, "username is in UserCreateRequest's required list")

	password, ok := propsByName["password"]
	require.True(t, ok)
	assert.False(t, password.Required, "password is not in UserCreateRequest's required list")

	admin, ok := propsByName["admin"]
	require.True(t, ok)
	assert.Equal(t, "boolean", admin.Type)
	assert.Equal(t, "false", admin.Default)

	groups, ok := propsByName["groups"]
	require.True(t, ok)
	assert.Equal(t, "array<string>", groups.Type)

	deleteWorker, ok := byOperationId["deleteWorker"]
	require.True(t, ok, "deleteWorker should be present")
	assert.Equal(t, "DELETE", deleteWorker.Method)
	assert.Equal(t, "/worker/api/v1/workers/{workerKey}", deleteWorker.Path)
	require.Len(t, deleteWorker.Parameters, 1)
	assert.Equal(t, "workerKey", deleteWorker.Parameters[0].Name)
	assert.Equal(t, "path", deleteWorker.Parameters[0].In)
	assert.True(t, deleteWorker.Parameters[0].Required)

	ping, ok := byOperationId["artifactoryPing"]
	require.True(t, ok, "artifactoryPing should be present")
	assert.Equal(t, []string{"Artifactory System"}, ping.Tags)
	assert.Empty(t, ping.Parameters)
}

func TestOperations_SortedByPathThenMethod(t *testing.T) {
	ops, err := Operations()
	require.NoError(t, err)

	for i := 1; i < len(ops); i++ {
		prev, cur := ops[i-1], ops[i]
		if prev.Path == cur.Path {
			assert.LessOrEqual(t, prev.Method, cur.Method, "same path %q should be sorted by method", prev.Path)
			continue
		}
		assert.Less(t, prev.Path, cur.Path, "operations should be sorted by path")
	}
}

func TestOperations_CachedAcrossCalls(t *testing.T) {
	first, err := Operations()
	require.NoError(t, err)
	second, err := Operations()
	require.NoError(t, err)
	assert.Same(t, &first[0], &second[0], "Operations should return the same cached backing array on repeat calls")
}

func TestInfo_Stub(t *testing.T) {
	info := Info()
	assert.Equal(t, "stub", info.SpecBundle)
	assert.Empty(t, info.SpecVersion, "stub builds have no rdme-admin version")
}

func TestIsSpecFile(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"users-api.yaml", true},
		{"artifactory-security_openapi.yaml", true},
		{"_order.yaml", false},
		{".placeholder.yaml", false},
		{"VERSION", false},
		{"ReadMe.md", false},
		{"notes.txt", false},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, isSpecFile(tt.name), "isSpecFile(%q)", tt.name)
	}
}
