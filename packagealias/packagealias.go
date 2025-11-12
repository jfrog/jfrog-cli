// Package packagealias implements the Ghost Frog technical specification for
// transparent package manager command interception
package packagealias

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
)

const (
	stateFile  = "state.json"
	configFile = "config.yaml"
)

// SupportedTools lists all package managers we create aliases for
var SupportedTools = []string{
	"mvn",
	"gradle",
	"npm",
	"yarn",
	"pnpm",
	"go",
	"pip",
	"pipenv",
	"poetry",
	"dotnet",
	"nuget",
	"docker",
	"gem",
	"bundle",
}

// AliasMode represents how a tool should be handled
type AliasMode string

const (
	// ModeJF runs through JFrog CLI integration flow (default)
	ModeJF AliasMode = "jf"
	// ModeEnv injects environment variables then runs native tool
	ModeEnv AliasMode = "env"
	// ModePass runs native tool directly without modification
	ModePass AliasMode = "pass"
)

// State tracks enable/disable status
type State struct {
	Enabled bool `json:"enabled"`
}

// Config holds per-tool policies
type Config struct {
	ToolModes map[string]AliasMode `json:"tool_modes,omitempty"`
	Enabled   bool                 `json:"enabled"`
}

// GetAliasHomeDir returns the base package-alias directory
func GetAliasHomeDir() (string, error) {
	homeDir, err := coreutils.GetJfrogHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, "package-alias"), nil
}

// GetAliasBinDir returns the bin directory where symlinks are created
func GetAliasBinDir() (string, error) {
	homeDir, err := coreutils.GetJfrogHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, "package-alias", "bin"), nil
}

// IsRunningAsAlias checks if the current process was invoked via an alias
func IsRunningAsAlias() (bool, string) {
	if len(os.Args) == 0 {
		return false, ""
	}

	// Get the name we were invoked as
	invokeName := filepath.Base(os.Args[0])

	// Remove .exe extension on Windows
	if runtime.GOOS == "windows" {
		invokeName = strings.TrimSuffix(invokeName, ".exe")
	}

	// Check if it's one of our supported tools
	for _, tool := range SupportedTools {
		if invokeName == tool {
			// For symlinks, os.Executable() resolves to the target, not the symlink
			// So we need to check if Args[0] contains our alias directory
			aliasDir, _ := GetAliasBinDir()
			if aliasDir != "" && strings.Contains(os.Args[0], aliasDir) {
				return true, tool
			}
		}
	}

	return false, ""
}

// FilterOutDirFromPATH removes a directory from PATH
func FilterOutDirFromPATH(pathVal, rmDir string) string {
	rmDir = filepath.Clean(rmDir)
	parts := filepath.SplitList(pathVal)
	keep := make([]string, 0, len(parts))

	for _, dir := range parts {
		if dir == "" {
			continue
		}
		if filepath.Clean(dir) == rmDir {
			continue
		}
		keep = append(keep, dir)
	}

	return strings.Join(keep, string(os.PathListSeparator))
}

// DisableAliasesForThisProcess removes the alias directory from PATH
// This prevents recursion when we try to execute the real tool
func DisableAliasesForThisProcess() error {
	aliasDir, err := GetAliasBinDir()
	if err != nil {
		return err
	}

	oldPath := os.Getenv("PATH")
	newPath := FilterOutDirFromPATH(oldPath, aliasDir)

	return os.Setenv("PATH", newPath)
}
