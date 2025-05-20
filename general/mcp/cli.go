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

	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/urfave/cli"
)

const (
	mcpToolSetsEnvVar   = "JFROG_MCP_TOOLSETS"
	mcpToolAccessEnvVar = "JFROG_MCP_TOOL_ACCESS"
	mcpServerBinaryName = "cli-mcp-server"
)

func McpCmd(c *cli.Context) error {
	// Show help if needed
	if show, err := cliutils.ShowCmdHelpIfNeeded(c, c.Args()); show || err != nil {
		return err
	}

	// Require at least one argument (the subcommand, e.g. "start")
	if c.NArg() < 1 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}
	cmdArg := c.Args().Get(0)
	if cmdArg != "start" {
		return cliutils.PrintHelpAndReturnError(fmt.Sprintf("Unknown subcommand: %s", cmdArg), c)
	}

	// Accept --toolset and --tool-access from flags or env vars (flags win)
	toolset := c.String(cliutils.McpToolsets)
	if toolset == "" {
		toolset = os.Getenv(mcpToolSetsEnvVar)
	}
	toolsAccess := c.String(cliutils.McpToolAccess)
	if toolsAccess == "" {
		toolsAccess = os.Getenv(mcpToolAccessEnvVar)
	}

	// Add a flag to allow specifying a specific version of the MCP server
	mcpVersion := c.String(cliutils.McpServerVersion)
	if mcpVersion == "" {
		mcpVersion = "[RELEASE]"
	}

	executablePath, err := downloadServerExecutable(mcpVersion)
	if err != nil {
		return err
	}

	cmd := exec.Command(executablePath, "--toolsets="+toolset, "--tools-access="+toolsAccess)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	displayToolset := toolset
	if displayToolset == "" {
		displayToolset = "--tools-access=read"
	}
	displayToolsAccess := toolsAccess
	if displayToolsAccess == "" {
		displayToolsAccess = "all-toolsets"
	}
	log.Debug("Starting MCP server with toolset:", displayToolset, "and tools access:", displayToolsAccess)
	return cmd.Run()
}

func downloadServerExecutable(version string) (string, error) {
	osName, arch, binaryName, err := getOsArchBinaryInfo()
	if err != nil {
		return "", err
	}

	jfrogHomeDir, err := coreutils.GetJfrogHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get JFrog home directory: %w", err)
	}
	targetDir := path.Join(jfrogHomeDir, "cli-mcp")
	if err := os.MkdirAll(targetDir, 0777); err != nil {
		return "", fmt.Errorf("failed to create directory '%s': %w", targetDir, err)
	}
	fullPath := path.Join(targetDir, binaryName)
	fileInfo, err := os.Stat(fullPath)
	if err == nil {
		// On Unix, check if the file is executable
		if runtime.GOOS != "windows" && fileInfo.Mode()&0111 == 0 {
			log.Debug("File exists but is not executable, will re-download:", fullPath)
		} else {
			log.Debug("MCP server binary already present at:", fullPath)
			return fullPath, nil
		}
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to stat '%s': %w", fullPath, err)
	}

	// Build the download URL (update as needed for your actual release location)
	url := fmt.Sprintf("https://releases.jfrog.io/artifactory/cli-mcp-server/v0/%s/%s-%s/%s", version, osName, arch, "cli-mcp-server")
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
	out, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file '%s': %w", fullPath, err)
	}
	defer func() {
		err = out.Close()
	}()
	if _, err := io.Copy(out, resp.Body); err != nil {
		return "", fmt.Errorf("failed to write binary: %w", err)
	}
	if err := os.Chmod(fullPath, 0755); err != nil && !strings.HasSuffix(binaryName, ".exe") {
		return "", fmt.Errorf("failed to make binary executable: %w", err)
	}
	log.Debug("MCP server binary downloaded to:", fullPath)
	return fullPath, nil
}

func getOsArchBinaryInfo() (osName, arch, binaryName string, err error) {
	osName = runtime.GOOS
	arch = runtime.GOARCH
	binaryName = mcpServerBinaryName
	if osName == "windows" {
		binaryName += ".exe"
	}
	return
}
