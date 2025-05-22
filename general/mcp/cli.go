package mcp

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"time"

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

// getMCPServerArgs extracts and sets command arguments from CLI flags or environment variables
func (mcp *Command) getMCPServerArgs(c *cli.Context) {
	// Accept --toolset and --tool-access from flags or env vars (flags win)
	mcp.toolSets = c.String(cliutils.McpToolsets)
	if mcp.toolSets == "" {
		mcp.toolSets = os.Getenv(mcpToolSetsEnvVar)
	}
	mcp.toolAccess = c.String(cliutils.McpToolAccess)
	if mcp.toolAccess == "" {
		mcp.toolAccess = os.Getenv(mcpToolAccessEnvVar)
	}
	// Add a flag to allow specifying a specific version of the MCP server
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

	log.Debug("Starting MCP server with toolset:", displayToolset, "and tools access:", displayToolsAccess)
}

// Cmd handles the CLI command execution and argument parsing
func Cmd(c *cli.Context) error {
	// Show help if needed
	if show, err := cliutils.ShowCmdHelpIfNeeded(c, c.Args()); show || err != nil {
		return err
	}
	// Validate arguments
	cmdArg := c.Args().Get(0)
	switch cmdArg {
	case "update":
		return updateMcpServerExecutable()
	case "start":
		cmd := createAndConfigureCommand(c)
		return commands.Exec(cmd)
	default:
		return cliutils.PrintHelpAndReturnError(fmt.Sprintf("Unknown subcommand: %s", cmdArg), c)
	}
}

// updateMcpServerExecutable forces an update of the MCP server binary
func updateMcpServerExecutable() error {
	log.Info("Updating MCP server binary...")
	osName, arch, binaryName := getOsArchBinaryInfo()
	fullPath, exists, err := getLocalBinaryPath(binaryName)
	if err != nil {
		return err
	}

	var currentVersion string
	// Check current version if binary exists
	if exists {
		currentVersion, err = getMcpServerVersion(fullPath)
		if err != nil {
			log.Warn("Could not determine current MCP server version:", err)
		} else {
			log.Info("Current MCP server version:", currentVersion)
		}
	}

	// Check if we already have the latest version
	if exists && currentVersion != "" {
		latestVersion, err := getLatestMcpServerVersion(osName, arch)
		switch {
		case err != nil:
			log.Warn("Could not determine latest MCP server version:", err)
		case currentVersion == latestVersion:
			log.Info("MCP server is already at the latest version:", currentVersion)
			return nil
		default:
			log.Info("A newer version is available:", latestVersion)
		}
	}

	// Download the latest version
	_, err = downloadBinary(fullPath, defaultServerVersion, osName, arch)
	if err != nil {
		return err
	}

	// Check new version after update
	newVersion, err := getMcpServerVersion(fullPath)
	if err != nil {
		log.Warn("Could not determine new MCP server version:", err)
	} else {
		log.Info("Updated MCP server to version:", newVersion)
	}

	log.Info("MCP server binary updated successfully.")
	return nil
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

	fullPath, exists, err := getLocalBinaryPath(binaryName)
	if err != nil {
		return "", err
	}

	if exists {
		return fullPath, nil
	}

	return downloadBinary(fullPath, version, osName, arch)
}

// getLocalBinaryPath determines the path to the binary and checks if it exists
func getLocalBinaryPath(binaryName string) (fullPath string, exists bool, err error) {
	jfrogHomeDir, err := coreutils.GetJfrogHomeDir()
	if err != nil {
		return "", false, fmt.Errorf("failed to get JFrog home directory: %w", err)
	}

	targetDir := path.Join(jfrogHomeDir, cliMcpDirName)
	if err := os.MkdirAll(targetDir, 0777); err != nil {
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
func downloadBinary(fullPath, version, osName, arch string) (string, error) {
	// Build the download URL
	url := fmt.Sprintf("%s/%s/%s-%s/%s", mcpDownloadBaseURL, version, osName, arch, mcpServerBinaryName)
	log.Debug("Downloading MCP server from:", url)

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download MCP server: %w", err)
	}
	defer func() {
		err = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download MCP server: received status %s", resp.Status)
	}

	return saveAndMakeExecutable(fullPath, resp.Body)
}

// saveAndMakeExecutable saves the binary to disk and makes it executable
func saveAndMakeExecutable(fullPath string, content io.Reader) (string, error) {
	out, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file '%s': %w", fullPath, err)
	}
	defer func() {
		err = out.Close()
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

// getLatestMcpServerVersion determines the latest available version for the given OS and architecture
func getLatestMcpServerVersion(osName, arch string) (string, error) {
	// Build the URL for the latest version (same as download URL but we'll make a HEAD request)
	url := fmt.Sprintf("%s/%s/%s-%s/%s", mcpDownloadBaseURL, defaultServerVersion, osName, arch, mcpServerBinaryName)

	// Make a HEAD request to get information without downloading the binary
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	req, err := http.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to check latest version: received status %s", resp.Status)
	}

	// Try to get version from headers
	// Server might include version information in headers like X-Version, X-Artifact-Version, etc.
	// If not available, we'll try to parse it from the ETag or Last-Modified headers

	// For simplicity, we'll fall back to a generic version check
	// In a real implementation, you would parse the version from appropriate headers

	// Check if we can determine version from Content-Disposition header
	contentDisposition := resp.Header.Get("Content-Disposition")
	if contentDisposition != "" {
		// Try to extract version from filename, if present
		if versionStr := extractVersionFromHeader(contentDisposition); versionStr != "" {
			return versionStr, nil
		}
	}

	// Check if we can get it from an X-Version or similar header
	// This is hypothetical - your actual server may use different headers
	if version := resp.Header.Get("X-Version"); version != "" {
		return version, nil
	}

	// If we can't determine the version from headers, we'll make another request
	// to the binary with --version flag after downloading

	// As a fallback, download the binary to a temporary location and check its version
	tempDir, err := os.MkdirTemp("", "mcp-version-check")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tempDir)

	tempBinaryPath := path.Join(tempDir, mcpServerBinaryName)
	if osName == "windows" {
		tempBinaryPath += ".exe"
	}

	// Download to temporary location
	_, err = downloadBinary(tempBinaryPath, defaultServerVersion, osName, arch)
	if err != nil {
		return "", err
	}

	// Check the version of the downloaded binary
	version, err := getMcpServerVersion(tempBinaryPath)
	if err != nil {
		return "", err
	}

	return version, nil
}

// extractVersionFromHeader attempts to extract version information from a header value
func extractVersionFromHeader(headerValue string) string {
	// This is a simple implementation - you might need to adjust based on your header format
	// Example: attachment; filename="cli-mcp-server-0.1.0"
	if strings.Contains(headerValue, "filename=") {
		parts := strings.Split(headerValue, "filename=")
		if len(parts) > 1 {
			filename := strings.Trim(parts[1], "\"' ")
			// Try to extract version from filename
			versionParts := strings.Split(filename, "-")
			if len(versionParts) > 0 {
				lastPart := versionParts[len(versionParts)-1]
				// Check if lastPart looks like a version number
				if strings.Contains(lastPart, ".") {
					return lastPart
				}
			}
		}
	}
	return ""
}
