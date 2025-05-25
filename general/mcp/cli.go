package mcp

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"

	"github.com/jfrog/build-info-go/utils"

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

// resolveMCPServerArgs resolves the MCP server arguments (toolSets, toolAccess, serverVersion)
// in the following order for each value:
//  1. CLI flag
//  2. Environment variable
//  3. Default constant
func (mcp *Command) resolveMCPServerArgs(c *cli.Context) {
	// Resolve toolSets: CLI flag -> Env var -> Default
	mcp.toolSets = c.String(cliutils.McpToolsets)
	if mcp.toolSets == "" {
		mcp.toolSets = os.Getenv(mcpToolSetsEnvVar)
	}
	if mcp.toolSets == "" {
		mcp.toolSets = defaultToolsets
	}

	// Resolve toolAccess: CLI flag -> Env var -> Default
	mcp.toolAccess = c.String(cliutils.McpToolAccess)
	if mcp.toolAccess == "" {
		mcp.toolAccess = os.Getenv(mcpToolAccessEnvVar)
	}
	if mcp.toolAccess == "" {
		mcp.toolAccess = defaultToolAccess
	}

	// Resolve serverVersion: CLI flag -> Default
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

	log.Info(fmt.Sprintf("Starting MCP server | toolset: %s | tools access: %s", mcp.toolSets, mcp.toolAccess))
	// Execute the command
	return cmd.Run()
}

// createMcpServerCommand creates the exec.Command for the MCP server
func createMcpServerCommand(executablePath, toolSets, toolAccess string) *exec.Cmd {
	cmd := exec.Command(
		executablePath,
		fmt.Sprintf("--%s=%s", cliutils.McpToolsets, toolSets),
		fmt.Sprintf("--%s=%s", cliutils.McpToolAccess, toolAccess),
	)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	return cmd
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
	cmd.resolveMCPServerArgs(c)

	return cmd
}

// downloadServerExecutable downloads the MCP server binary if it doesn't exist locally
func downloadServerExecutable(version string) (string, error) {
	osName, arch, binaryName := getOsArchBinaryInfo()
	targetPath, err := getLocalBinaryPath(binaryName)
	if err != nil {
		return "", err
	}
	urlStr := fmt.Sprintf("%s/%s/%s-%s/%s", mcpDownloadBaseURL, version, osName, arch, mcpServerBinaryName)
	log.Info("Downloading MCP server from:", urlStr)
	return targetPath, utils.DownloadFile(targetPath, urlStr)

}

// getLocalBinaryPath determines the path to the binary and checks if it exists
func getLocalBinaryPath(binaryName string) (fullPath string, err error) {
	jfrogHomeDir, err := coreutils.GetJfrogHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get JFrog home directory: %w", err)
	}

	targetDir := path.Join(jfrogHomeDir, cliMcpDirName)
	if err = os.MkdirAll(targetDir, 0777); err != nil {
		return "", fmt.Errorf("failed to create directory '%s': %w", targetDir, err)
	}

	fullPath = path.Join(targetDir, binaryName)
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
