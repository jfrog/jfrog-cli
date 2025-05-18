package mcp

import (
	"fmt"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/urfave/cli"
	"os"
	"os/exec"
)

func McpCmd(c *cli.Context) error {
	// Show help if needed
	if show, err := cliutils.ShowCmdHelpIfNeeded(c, c.Args()); show || err != nil {
		return err
	}
	//log.Output(coreutils.PrintBoldTitle("Welcome to the MCP-Start Command! 🚀"))
	// TODO this should downloaded from releases and version should be a variable
	//log.Output("Download MCP server binary version : v0.0.1 ... ")
	// TODO need to decide where the executable is being downloaded..maybe the current dir is okay.
	exePath, err := downloadServerExecutable()
	if err != nil {
		return err
	}
	//log.Output(fmt.Sprintf("✅ Successfully downloaded the MCP server binary to: %s", exePath))

	//binaryPath, err := resolveOrDownloadMcpBinary()
	//if err != nil {
	//	return fmt.Errorf("failed to get MCP binary: %w", err)
	//}

	cmd := exec.Command(exePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	return cmd.Run()
	//
	//client, err := promptClientSelection()
	//if err != nil {
	//	return err
	//}
	//log.Output(fmt.Sprintf("You selected: %s", client))
	//
	//// Output the corresponding template
	//if err = outputClientTemplate(client, exePath); err != nil {
	//	return err
	//}
	//
	//log.Output("✅ Successfully completed the `mcp-start` process!\n")
	//log.Output("ℹ️ For further assistance, questions, or issues, please visit the repository: https://github.com/jfrog/mcp-jfrog-go")
	//return nil
}

func downloadServerExecutable() (string, error) {
	//binaryName := "mcp-jfrog-go"
	// TODO this has to point to latest
	repoPath := "v0/0.0.1"

	targetDir, err := coreutils.GetJfrogHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get JFrog home directory: %w", err)
	}

	// Create the target directory if it doesn't exist
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		if err := os.Mkdir(targetDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create directory '%s': %w", targetDir, err)
		}
	}

	// Change into the target directory
	if err := os.Chdir(targetDir); err != nil {
		return "", fmt.Errorf("failed to cd into directory '%s': %w", targetDir, err)
	}

	// Construct the full path for the binary
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %w", err)
	}

	// Run the JFrog CLI download command
	cmd := exec.Command("jf", "rt", "dl", targetDir, repoPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to download binary: %w", err)
	}
	fullPath := fmt.Sprintf("%s/%s", wd, repoPath)
	// Make the binary executable
	if err := os.Chmod(fullPath, 0755); err != nil {
		return "", fmt.Errorf("failed to make binary executable: %w", err)
	}

	// Return the full path of the binary
	return fullPath, nil
}
