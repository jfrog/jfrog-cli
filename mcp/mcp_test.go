package mcp

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	coreconfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveMcpURL(t *testing.T) {
	t.Run("flag override wins", func(t *testing.T) {
		got, err := ResolveMcpURL("https://override.example/mcp/", &coreconfig.ServerDetails{Url: "https://platform.example/"})
		require.NoError(t, err)
		assert.Equal(t, "https://override.example/mcp", got)
	})

	t.Run("env override when no flag", func(t *testing.T) {
		t.Setenv(cliutils.JfrogCliMcpUrl, "https://env.example/mcp/")
		got, err := ResolveMcpURL("", &coreconfig.ServerDetails{Url: "https://platform.example/"})
		require.NoError(t, err)
		assert.Equal(t, "https://env.example/mcp", got)
	})

	t.Run("derived from platform url", func(t *testing.T) {
		got, err := ResolveMcpURL("", &coreconfig.ServerDetails{Url: "https://platform.example/"})
		require.NoError(t, err)
		assert.Equal(t, "https://platform.example/mcp", got)
	})

	t.Run("error when no url available", func(t *testing.T) {
		_, err := ResolveMcpURL("", &coreconfig.ServerDetails{})
		assert.Error(t, err)
	})
}

func TestCheckAvailability(t *testing.T) {
	cases := []struct {
		status    int
		available bool
	}{
		{http.StatusOK, true},
		{http.StatusUnauthorized, true},
		{http.StatusForbidden, true},
		{http.StatusMethodNotAllowed, true},
		{http.StatusNotFound, false},
	}
	for _, tc := range cases {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(tc.status)
		}))
		err := CheckAvailability(&coreconfig.ServerDetails{}, srv.URL)
		if tc.available {
			assert.NoError(t, err, "status %d should be available", tc.status)
		} else {
			assert.Error(t, err, "status %d should not be available", tc.status)
		}
		srv.Close()
	}

	t.Run("network error is not available", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
		url := srv.URL
		srv.Close() // ensure the address is no longer listening
		assert.Error(t, CheckAvailability(&coreconfig.ServerDetails{}, url))
	})
}

func readServers(t *testing.T, path string) map[string]interface{} {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	root := map[string]interface{}{}
	require.NoError(t, json.Unmarshal(data, &root))
	servers, ok := root[mcpServersKey].(map[string]interface{})
	require.True(t, ok, "expected mcpServers object")
	return servers
}

func TestInstall_Cursor_WritesEntryWithoutType(t *testing.T) {
	dir := t.TempDir()
	var out bytes.Buffer
	err := Install(InstallParams{
		Agent:      "cursor",
		ServerName: DefaultServerName,
		McpURL:     "https://platform.example/mcp",
		ProjectDir: dir,
		SkipCheck:  true,
	}, &out)
	require.NoError(t, err)

	servers := readServers(t, filepath.Join(dir, ".cursor", "mcp.json"))
	entry, ok := servers["jfrog"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "https://platform.example/mcp", entry["url"])
	_, hasType := entry["type"]
	assert.False(t, hasType, "cursor entries must not include a type field")
}

func TestInstall_Claude_WritesEntryWithType(t *testing.T) {
	dir := t.TempDir()
	var out bytes.Buffer
	err := Install(InstallParams{
		Agent:      "claude",
		ServerName: DefaultServerName,
		McpURL:     "https://platform.example/mcp",
		ProjectDir: dir,
		SkipCheck:  true,
	}, &out)
	require.NoError(t, err)

	servers := readServers(t, filepath.Join(dir, ".mcp.json"))
	entry, ok := servers["jfrog"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "http", entry["type"])
	assert.Equal(t, "https://platform.example/mcp", entry["url"])
}

func TestInstall_PreservesExistingEntries(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".cursor", "mcp.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0700))
	existing := `{"mcpServers":{"other":{"url":"https://other.example/mcp"}},"someOtherKey":true}`
	require.NoError(t, os.WriteFile(path, []byte(existing), 0600))

	var out bytes.Buffer
	require.NoError(t, Install(InstallParams{
		Agent:      "cursor",
		ServerName: DefaultServerName,
		McpURL:     "https://platform.example/mcp",
		ProjectDir: dir,
		SkipCheck:  true,
	}, &out))

	servers := readServers(t, path)
	assert.Contains(t, servers, "other", "existing entry must be preserved")
	assert.Contains(t, servers, "jfrog", "new entry must be added")

	// Unrelated top-level keys must be preserved.
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	root := map[string]interface{}{}
	require.NoError(t, json.Unmarshal(data, &root))
	assert.Equal(t, true, root["someOtherKey"])
}

func TestInstall_Idempotent(t *testing.T) {
	dir := t.TempDir()
	params := InstallParams{
		Agent:      "cursor",
		ServerName: DefaultServerName,
		McpURL:     "https://platform.example/mcp",
		ProjectDir: dir,
		SkipCheck:  true,
	}
	var out bytes.Buffer
	require.NoError(t, Install(params, &out))
	require.NoError(t, Install(params, &out))

	servers := readServers(t, filepath.Join(dir, ".cursor", "mcp.json"))
	assert.Len(t, servers, 1)
}

func TestInstall_DryRunWritesNothing(t *testing.T) {
	dir := t.TempDir()
	var out bytes.Buffer
	require.NoError(t, Install(InstallParams{
		Agent:      "cursor",
		ServerName: DefaultServerName,
		McpURL:     "https://platform.example/mcp",
		ProjectDir: dir,
		DryRun:     true,
		SkipCheck:  true,
	}, &out))

	_, statErr := os.Stat(filepath.Join(dir, ".cursor", "mcp.json"))
	assert.True(t, os.IsNotExist(statErr), "dry-run must not write a file")
	assert.Contains(t, out.String(), "Dry run")
}

func TestInstall_ReadinessFailureAbortsWithoutWriting(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	dir := t.TempDir()
	var out bytes.Buffer
	err := Install(InstallParams{
		Agent:         "cursor",
		ServerName:    DefaultServerName,
		McpURL:        srv.URL,
		ProjectDir:    dir,
		ServerDetails: &coreconfig.ServerDetails{},
	}, &out)
	require.Error(t, err)
	_, statErr := os.Stat(filepath.Join(dir, ".cursor", "mcp.json"))
	assert.True(t, os.IsNotExist(statErr), "no config should be written when readiness fails")
}

func TestInstall_UnsupportedAgent(t *testing.T) {
	var out bytes.Buffer
	err := Install(InstallParams{Agent: "windsurf", ServerName: DefaultServerName, SkipCheck: true}, &out)
	require.Error(t, err)
}

func TestUninstall_RemovesOnlyJfrogEntry(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".cursor", "mcp.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0700))
	existing := `{"mcpServers":{"jfrog":{"url":"https://platform.example/mcp"},"other":{"url":"https://other.example/mcp"}}}`
	require.NoError(t, os.WriteFile(path, []byte(existing), 0600))

	var out bytes.Buffer
	require.NoError(t, Uninstall(UninstallParams{
		Agent:      "cursor",
		ServerName: DefaultServerName,
		ProjectDir: dir,
	}, &out))

	servers := readServers(t, path)
	assert.NotContains(t, servers, "jfrog")
	assert.Contains(t, servers, "other")
}

func TestUninstall_MissingFileIsNoOp(t *testing.T) {
	dir := t.TempDir()
	var out bytes.Buffer
	require.NoError(t, Uninstall(UninstallParams{
		Agent:      "cursor",
		ServerName: DefaultServerName,
		ProjectDir: dir,
	}, &out))
	assert.Contains(t, out.String(), "nothing to remove")
}
