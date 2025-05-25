package mcp

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"

	"github.com/jfrog/jfrog-cli-core/v2/common/commands"

	"github.com/jfrog/jfrog-cli-core/v2/utils/config"

	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/urfave/cli"
)

const (
	mcpToolSetsEnvVar    = "JFROG_MCP_TOOLSETS"
	mcpToolAccessEnvVar  = "JFROG_MCP_TOOL_ACCESS"
	mcpServerBinaryName  = "cli-mcp-server"
	defaultServerVersion = "[RELEASE]"
	cliMcpDirName        = "cli-mcp"
	defaultToolsets      = "read"
	defaultToolAccess    = "all-toolsets"
	mcpDownloadBaseURL   = "https://releases.jfrog.io/artifactory/cli-mcp-server/v0"
)

type Command struct {
	serverDetails *config.ServerDetails
	toolSets      string
	toolAccess    string
	serverVersion string
}

// NewMcpCommand returns a new MCP command instance
func NewMcpCommand() *Command {
	return &Command{}
}

// SetServerDetails sets the Artifactory server details for the command
func (mcp *Command) SetServerDetails(serverDetails *config.ServerDetails) {
	mcp.serverDetails = serverDetails
}

// ServerDetails returns the Artifactory server details associated with the command
func (mcp *Command) ServerDetails() (*config.ServerDetails, error) {
	return mcp.serverDetails, nil
}

// CommandName returns the name of the command for usage reporting
func (mcp *Command) CommandName() string {
	return "jf_mcp_start"
}

// McpToolset - specifies the toolsets to use (e.g artifactory, distribution, etc.)
// toolAccess - specifies the tool access level (e.g read, write, etc.)
// McpServerVersion - specifies the version of the MCP server to run
func (mcp *Command) getMCPServerArgs(c *cli.Context) {
	mcp.toolSets = c.String(cliutils.McpToolsets)
	if mcp.toolSets == "" {
		mcp.toolSets = os.Getenv(mcpToolSetsEnvVar)
	}
	mcp.toolAccess = c.String(cliutils.McpToolAccess)
	if mcp.toolAccess == "" {
		mcp.toolAccess = os.Getenv(mcpToolAccessEnvVar)
	}
	mcp.serverVersion = c.String(cliutils.McpServerVersion)
	if mcp.serverVersion == "" {
		mcp.serverVersion = defaultServerVersion
	}
}

// Run executes the MCP command, downloading the server binary if needed and starting it
func (mcp *Command) Run() error {
	executablePath, err := downloadServerExecutable(mcp.serverVersion)
	if err != nil {
		return err
	}

	// Create command to execute the MCP server
	cmd := createMcpServerCommand(executablePath, mcp.toolSets, mcp.toolAccess)

	// Log startup information
	logStartupInfo(mcp.toolSets, mcp.toolAccess)

	// Execute the command
	return cmd.Run()
}

// createMcpServerCommand creates the exec.Command for the MCP server
func createMcpServerCommand(executablePath, toolSets, toolAccess string) *exec.Cmd {
	cmd := exec.Command(
		executablePath,
		cliutils.McpToolsets+toolSets,
		cliutils.McpToolAccess+toolAccess,
	)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	return cmd
}

// logStartupInfo logs the MCP server startup parameters
func logStartupInfo(toolSets, toolAccess string) {
	displayToolset := toolSets
	if displayToolset == "" {
		displayToolset = defaultToolsets
	}

	displayToolsAccess := toolAccess
	if displayToolsAccess == "" {
		displayToolsAccess = defaultToolAccess
	}
	log.Info("Starting MCP server with toolset:", displayToolset, "and tools access:", displayToolsAccess)
}

// Cmd handles the CLI command execution and argument parsing
func Cmd(c *cli.Context) error {
	// Show help if needed
	if show, err := cliutils.ShowCmdHelpIfNeeded(c, c.Args()); show || err != nil {
		return err
	}
	cmd := createAndConfigureCommand(c)
	return commands.Exec(cmd)

}

// getMcpServerVersion runs the MCP server binary with --version flag to get its version
func getMcpServerVersion(binaryPath string) (string, error) {
	cmd := exec.Command(binaryPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	// Trim whitespace and return the output
	return strings.TrimSpace(string(output)), nil
}

// createAndConfigureCommand creates and configures the MCP command
func createAndConfigureCommand(c *cli.Context) *Command {
	serverDetails, err := cliutils.CreateArtifactoryDetailsByFlags(c)
	if err != nil {
		log.Error("Failed to create Artifactory details:", err)
		return nil
	}

	cmd := NewMcpCommand()
	cmd.SetServerDetails(serverDetails)
	cmd.getMCPServerArgs(c)

	return cmd
}

// downloadServerExecutable downloads the MCP server binary if it doesn't exist locally
func downloadServerExecutable(version string) (string, error) {
	osName, arch, binaryName := getOsArchBinaryInfo()
	targetPath, exists, err := getLocalBinaryPath(binaryName)
	if err != nil {
		return "", err
	}

	if exists {
		return targetPath, nil
	}

	return downloadBinary(targetPath, version, osName, arch)
}

// getLocalBinaryPath determines the path to the binary and checks if it exists
func getLocalBinaryPath(binaryName string) (fullPath string, exists bool, err error) {
	jfrogHomeDir, err := coreutils.GetJfrogHomeDir()
	if err != nil {
		return "", false, fmt.Errorf("failed to get JFrog home directory: %w", err)
	}

	targetDir := path.Join(jfrogHomeDir, cliMcpDirName)
	if err = os.MkdirAll(targetDir, 0777); err != nil {
		return "", false, fmt.Errorf("failed to create directory '%s': %w", targetDir, err)
	}

	fullPath = path.Join(targetDir, binaryName)
	fileInfo, err := os.Stat(fullPath)
	if err == nil {
		// On Unix, check if the file is executable
		if runtime.GOOS != "windows" && fileInfo.Mode()&0111 == 0 {
			log.Debug("File exists but is not executable, will re-download:", fullPath)
			return fullPath, false, nil
		}
		log.Debug("MCP server binary already present at:", fullPath)
		return fullPath, true, nil
	}

	if !os.IsNotExist(err) {
		return "", false, fmt.Errorf("failed to stat '%s': %w", fullPath, err)
	}

	return fullPath, false, nil
}

// downloadBinary downloads the binary from the remote server
func downloadBinary(targetPath, version, osName, arch string) (string, error) {
	// Build the download URL
	urlStr := fmt.Sprintf("%s/%s/%s-%s/%s", mcpDownloadBaseURL, version, osName, arch, mcpServerBinaryName)
	log.Debug("Downloading MCP server from:", urlStr)

	// Validate URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	resp, err := http.Get(parsedURL.String())
	if err != nil {
		return "", fmt.Errorf("failed to download MCP server: %w", err)
	}
	defer func() {
		err = errors.Join(err, resp.Body.Close())
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download MCP server: received status %s", resp.Status)
	}

	return saveAndMakeExecutable(targetPath, resp.Body)
}

// saveAndMakeExecutable saves the binary to disk and makes it executable
func saveAndMakeExecutable(fullPath string, content io.Reader) (string, error) {
	out, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file '%s': %w", fullPath, err)
	}
	defer func() {
		err = errors.Join(err, out.Close())
	}()

	if _, err = io.Copy(out, content); err != nil {
		return "", fmt.Errorf("failed to write binary: %w", err)
	}

	if err = os.Chmod(fullPath, 0755); err != nil && !strings.HasSuffix(fullPath, ".exe") {
		return "", fmt.Errorf("failed to make binary executable: %w", err)
	}

	log.Debug("MCP server binary downloaded to:", fullPath)
	return fullPath, nil
}

// getOsArchBinaryInfo returns the current OS, architecture, and appropriate binary name
func getOsArchBinaryInfo() (osName, arch, binaryName string) {
	osName = runtime.GOOS
	arch = runtime.GOARCH
	binaryName = mcpServerBinaryName
	if osName == "windows" {
		binaryName += ".exe"
	}
	return
}
