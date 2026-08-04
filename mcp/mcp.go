package mcp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	coreconfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-client-go/http/httpclient"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/httputils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

// mcpServersKey is the top-level key used by both Cursor and Claude to hold MCP
// server definitions.
const mcpServersKey = "mcpServers"

// DefaultServerName is the entry name written to the agent configuration.
const DefaultServerName = "jfrog"

// agentSpec describes where and how to write the MCP server entry for a given
// AI agent. The config schema is agent-specific: Cursor omits the "type" field
// for remote servers, while Claude requires "type": "http".
type agentSpec struct {
	// projectFile is the config file path relative to the project directory.
	projectFile string
	// globalFile is the user-level config file path (may start with "~").
	globalFile string
	// includeType controls whether the entry includes "type": "http" (Claude).
	includeType bool
}

// supportedAgents maps the user-facing agent name to its configuration spec.
// Scope is intentionally limited to cursor and claude.
var supportedAgents = map[string]agentSpec{
	"cursor": {projectFile: filepath.Join(".cursor", "mcp.json"), globalFile: filepath.Join("~", ".cursor", "mcp.json"), includeType: false},
	"claude": {projectFile: ".mcp.json", globalFile: filepath.Join("~", ".claude.json"), includeType: true},
}

// SupportedAgentNames returns the supported agent names, sorted, for help and
// error messages.
func SupportedAgentNames() string {
	names := make([]string, 0, len(supportedAgents))
	for name := range supportedAgents {
		names = append(names, name)
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}

func resolveAgent(name string) (agentSpec, string, error) {
	normalized := strings.ToLower(strings.TrimSpace(name))
	if normalized == "" {
		return agentSpec{}, "", errorutils.CheckErrorf("the --agent flag is required. Supported agents: %s", SupportedAgentNames())
	}
	spec, ok := supportedAgents[normalized]
	if !ok {
		return agentSpec{}, "", errorutils.CheckErrorf("unsupported agent %q. Supported agents: %s", name, SupportedAgentNames())
	}
	return spec, normalized, nil
}

// configFilePath resolves the absolute config file path for the agent at the
// requested scope.
func (a agentSpec) configFilePath(projectDir string, global bool) (string, error) {
	if global {
		return expandHome(a.globalFile)
	}
	if projectDir == "" {
		projectDir = "."
	}
	return filepath.Abs(filepath.Join(projectDir, a.projectFile))
}

func expandHome(path string) (string, error) {
	if path == "~" || strings.HasPrefix(path, "~"+string(os.PathSeparator)) {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", errorutils.CheckError(err)
		}
		path = filepath.Join(home, strings.TrimPrefix(path, "~"))
	}
	return filepath.Abs(path)
}

// ResolveMcpURL determines the remote MCP endpoint. Precedence: explicit override
// (flag value, passed in as urlOverride) > JFROG_CLI_MCP_URL env var > derived
// from the configured platform URL as <platform-url>/mcp.
func ResolveMcpURL(urlOverride string, serverDetails *coreconfig.ServerDetails) (string, error) {
	if trimmed := strings.TrimSpace(urlOverride); trimmed != "" {
		return strings.TrimRight(trimmed, "/"), nil
	}
	if env := strings.TrimSpace(os.Getenv(cliutils.JfrogCliMcpUrl)); env != "" {
		return strings.TrimRight(env, "/"), nil
	}
	base := strings.TrimSpace(serverDetails.GetUrl())
	if base == "" {
		return "", errorutils.CheckErrorf("no JFrog Platform URL is configured. Set one with 'jf config' or pass --url / --mcp-url / the %s environment variable", cliutils.JfrogCliMcpUrl)
	}
	return strings.TrimRight(base, "/") + "/mcp", nil
}

// EndpointInfo is the stable JSON schema printed by 'jf mcp show --format json'.
type EndpointInfo struct {
	ServerId    string `json:"serverId,omitempty"`
	PlatformUrl string `json:"platformUrl,omitempty"`
	McpUrl      string `json:"mcpUrl"`
	Transport   string `json:"transport"`
}

// CheckAvailability verifies the MCP endpoint is reachable. Any HTTP response —
// including an auth-required 401/403 or a 405 — proves the endpoint exists and
// counts as available, mirroring how MCP clients detect OAuth-protected servers.
// A network error or a 404 is treated as not available.
func CheckAvailability(serverDetails *coreconfig.ServerDetails, mcpURL string) error {
	client, err := httpclient.ClientBuilder().
		SetInsecureTls(serverDetails.InsecureTls).
		SetClientCertPath(serverDetails.ClientCertPath).
		SetClientCertKeyPath(serverDetails.ClientCertKeyPath).
		Build()
	if err != nil {
		return err
	}
	resp, _, _, err := client.SendGet(mcpURL, true, httputils.HttpClientDetails{}, "")
	if err != nil {
		return errorutils.CheckErrorf("the MCP server at %s is not reachable: %s", mcpURL, err.Error())
	}
	defer func() {
		if resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()
	if resp.StatusCode == http.StatusNotFound {
		return errorutils.CheckErrorf("the MCP server at %s is not available (HTTP 404). Verify the platform URL and that the MCP server is enabled", mcpURL)
	}
	log.Debug(fmt.Sprintf("MCP readiness check for %s returned HTTP %d", mcpURL, resp.StatusCode))
	return nil
}

// readConfig loads the agent config file into a generic JSON object. A missing
// file yields an empty object.
func readConfig(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]interface{}{}, nil
		}
		return nil, errorutils.CheckError(err)
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return map[string]interface{}{}, nil
	}
	root := map[string]interface{}{}
	if err := json.Unmarshal(data, &root); err != nil {
		return nil, errorutils.CheckErrorf("failed to parse existing config %s: %s", path, err.Error())
	}
	return root, nil
}

func mcpServersMap(root map[string]interface{}) map[string]interface{} {
	if existing, ok := root[mcpServersKey].(map[string]interface{}); ok {
		return existing
	}
	servers := map[string]interface{}{}
	root[mcpServersKey] = servers
	return servers
}

func marshalConfig(root map[string]interface{}) ([]byte, error) {
	data, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return nil, errorutils.CheckError(err)
	}
	return append(data, '\n'), nil
}

func writeConfigFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return errorutils.CheckError(err)
	}
	return errorutils.CheckError(os.WriteFile(path, data, 0600))
}

// fprintf writes a formatted message to out and returns a checked error, so
// callers report write failures instead of silently ignoring them.
func fprintf(out io.Writer, format string, a ...interface{}) error {
	_, err := fmt.Fprintf(out, format, a...)
	return errorutils.CheckError(err)
}

// InstallParams holds the resolved inputs for an install operation.
type InstallParams struct {
	Agent         string
	ServerName    string
	McpURL        string
	ProjectDir    string
	Global        bool
	DryRun        bool
	ServerDetails *coreconfig.ServerDetails
	SkipCheck     bool
}

// Install writes (or previews) the JFrog MCP server entry into the agent config.
func Install(params InstallParams, out io.Writer) error {
	spec, agentName, err := resolveAgent(params.Agent)
	if err != nil {
		return err
	}
	if !params.SkipCheck {
		if err := CheckAvailability(params.ServerDetails, params.McpURL); err != nil {
			return err
		}
	}
	path, err := spec.configFilePath(params.ProjectDir, params.Global)
	if err != nil {
		return err
	}

	root, err := readConfig(path)
	if err != nil {
		return err
	}
	servers := mcpServersMap(root)
	entry := map[string]interface{}{"url": params.McpURL}
	if spec.includeType {
		entry["type"] = "http"
	}
	servers[params.ServerName] = entry

	data, err := marshalConfig(root)
	if err != nil {
		return err
	}

	if params.DryRun {
		return fprintf(out, "[Dry run] %s would be updated to:\n%s", path, data)
	}
	if err := writeConfigFile(path, data); err != nil {
		return err
	}

	scope := "project"
	if params.Global {
		scope = "global"
	}
	msg := fmt.Sprintf("Configured the '%s' MCP server for %s (%s scope): %s\n\n", params.ServerName, agentName, scope, path)
	msg += "Next step: the connection is not active until you complete OAuth authorization.\n"
	switch agentName {
	case "claude":
		msg += "  Run /mcp inside Claude Code and complete the browser login.\n"
	case "cursor":
		msg += "  Reload Cursor and approve the JFrog MCP server when prompted.\n"
	}
	return fprintf(out, "%s", msg)
}

// UninstallParams holds the resolved inputs for an uninstall operation.
type UninstallParams struct {
	Agent      string
	ServerName string
	ProjectDir string
	Global     bool
	DryRun     bool
}

// Uninstall removes the JFrog MCP server entry from the agent config, leaving
// other entries untouched.
func Uninstall(params UninstallParams, out io.Writer) error {
	spec, agentName, err := resolveAgent(params.Agent)
	if err != nil {
		return err
	}
	path, err := spec.configFilePath(params.ProjectDir, params.Global)
	if err != nil {
		return err
	}

	if _, statErr := os.Stat(path); os.IsNotExist(statErr) {
		return fprintf(out, "No %s configuration found at %s; nothing to remove.\n", agentName, path)
	}

	root, err := readConfig(path)
	if err != nil {
		return err
	}
	servers, ok := root[mcpServersKey].(map[string]interface{})
	if !ok {
		return fprintf(out, "No MCP servers configured in %s; nothing to remove.\n", path)
	}
	if _, present := servers[params.ServerName]; !present {
		return fprintf(out, "No '%s' MCP server entry in %s; nothing to remove.\n", params.ServerName, path)
	}
	delete(servers, params.ServerName)

	data, err := marshalConfig(root)
	if err != nil {
		return err
	}
	if params.DryRun {
		return fprintf(out, "[Dry run] %s would be updated to:\n%s", path, data)
	}
	if err := writeConfigFile(path, data); err != nil {
		return err
	}
	return fprintf(out, "Removed the '%s' MCP server entry from %s (%s).\n", params.ServerName, path, agentName)
}
