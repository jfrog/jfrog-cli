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

// Config holds per-tool policies
type Config struct {
	Enabled         bool                 `json:"enabled" yaml:"enabled"`
	ToolModes       map[string]AliasMode `json:"tool_modes,omitempty" yaml:"tool_modes,omitempty"`
	SubcommandModes map[string]AliasMode `json:"subcommand_modes,omitempty" yaml:"subcommand_modes,omitempty"`
	EnabledTools    []string             `json:"enabled_tools,omitempty" yaml:"enabled_tools,omitempty"`
	JfBinarySHA256  string               `json:"jf_binary_sha256,omitempty" yaml:"jf_binary_sha256,omitempty"`
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

// IsRunningAsAlias returns whether the current process was invoked through
// a package-alias entry and, if so, the detected tool name.
func IsRunningAsAlias() (bool, string) {
	if len(os.Args) == 0 {
		return false, ""
	}

	invokeName := filepath.Base(os.Args[0])

	// Remove .exe extension on Windows
	if runtime.GOOS == "windows" {
		invokeName = strings.TrimSuffix(invokeName, ".exe")
	}

	for _, tool := range SupportedTools {
		if invokeName == tool {
			aliasDir, _ := GetAliasBinDir()
			currentExec, _ := os.Executable()

			if aliasDir != "" && isPathWithinDir(currentExec, aliasDir) {
				return true, tool
			}

			if aliasDir != "" && isPathWithinDir(os.Args[0], aliasDir) {
				return true, tool
			}

			if aliasDir != "" && !filepath.IsAbs(os.Args[0]) {
				aliasPath := filepath.Join(aliasDir, tool)
				if runtime.GOOS == "windows" {
					aliasPath += ".exe"
				}

				if linkTarget, err := os.Readlink(aliasPath); err == nil {
					if absTarget, err := filepath.Abs(linkTarget); err == nil {
						resolvedExec, err := filepath.EvalSymlinks(currentExec)
						if err != nil {
							resolvedExec = currentExec
						}

						if resolvedExec == absTarget || filepath.Clean(resolvedExec) == filepath.Clean(absTarget) {
							return true, tool
						}
					}
				}
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

// DisableAliasesForThisProcess removes the alias directory from PATH for the
// current process and its future child processes.
func DisableAliasesForThisProcess() error {
	aliasDir, err := GetAliasBinDir()
	if err != nil {
		return err
	}

	oldPath := os.Getenv("PATH")
	newPath := FilterOutDirFromPATH(oldPath, aliasDir)

	return os.Setenv("PATH", newPath)
}

func isPathWithinDir(pathValue, parentDir string) bool {
	if pathValue == "" || parentDir == "" {
		return false
	}
	absolutePath, pathErr := filepath.Abs(pathValue)
	if pathErr != nil {
		return false
	}
	absoluteParentDir, parentErr := filepath.Abs(parentDir)
	if parentErr != nil {
		return false
	}
	absolutePath = filepath.Clean(absolutePath)
	absoluteParentDir = filepath.Clean(absoluteParentDir)
	relativePath, relErr := filepath.Rel(absoluteParentDir, absolutePath)
	if relErr != nil {
		return false
	}
	if relativePath == "." {
		return true
	}
	return relativePath != ".." && !strings.HasPrefix(relativePath, ".."+string(filepath.Separator))
}
