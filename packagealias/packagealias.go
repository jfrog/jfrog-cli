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

// IsRunningAsAlias checks if the current process was invoked via a Ghost Frog alias.
//
// An "alias" is a symlink created by `jf package-alias install` that maps a package manager
// tool name (e.g., "mvn", "npm", "pip") to the jf binary. For example:
//
//	~/.jfrog/package-alias/bin/mvn → /path/to/jf
//	~/.jfrog/package-alias/bin/npm → /path/to/jf
//
// When a user runs `mvn clean install`, the shell resolves `mvn` to the alias symlink,
// which executes the jf binary. This function detects that scenario by checking:
//  1. If the command name (os.Args[0]) matches a supported tool name
//  2. If the current executable path is within the alias directory, OR
//  3. If there's an alias symlink pointing to the current executable
//
// Returns:
//   - bool: true if running as an alias, false otherwise
//   - string: the tool name if running as alias (e.g., "mvn", "npm"), empty string otherwise
//
// Example:
//
//	When user runs: mvn clean install
//	Shell resolves to: ~/.jfrog/package-alias/bin/mvn (alias symlink)
//	jf binary detects: IsRunningAsAlias() returns (true, "mvn")
//	Result: Command is transformed to "jf mvn clean install"
func IsRunningAsAlias() (bool, string) {
	if len(os.Args) == 0 {
		return false, ""
	}

	// Extract the command name from how we were invoked
	// Examples:
	//   - If invoked as "mvn": invokeName = "mvn"
	//   - If invoked as "/path/to/mvn": invokeName = "mvn"
	//   - If invoked as "npm": invokeName = "npm"
	invokeName := filepath.Base(os.Args[0])

	// Remove .exe extension on Windows
	if runtime.GOOS == "windows" {
		invokeName = strings.TrimSuffix(invokeName, ".exe")
	}

	// Check if the command name matches one of our supported package manager tools
	// Supported tools: mvn, gradle, npm, yarn, pnpm, go, pip, pipenv, poetry, etc.
	for _, tool := range SupportedTools {
		if invokeName == tool {
			aliasDir, _ := GetAliasBinDir()   // ~/.jfrog/package-alias/bin
			currentExec, _ := os.Executable() // Actual path to jf binary

			// Detection Method 1: Check if current executable is in alias directory
			// This handles the case: user runs "mvn" → shell resolves to alias symlink
			// → os.Executable() returns ~/.jfrog/package-alias/bin/mvn
			if aliasDir != "" && strings.Contains(currentExec, aliasDir) {
				return true, tool
			}

			// Detection Method 2: Check if os.Args[0] contains alias directory (full path case)
			// This handles: user runs "/path/to/.jfrog/package-alias/bin/mvn" directly
			if aliasDir != "" && strings.Contains(os.Args[0], aliasDir) {
				return true, tool
			}

			// Detection Method 3: If invoked by name only (not full path), verify alias exists
			// This handles: user runs "mvn" (just the name, not a path)
			// We check if there's an alias symlink that points to the current jf binary
			if aliasDir != "" && !filepath.IsAbs(os.Args[0]) {
				aliasPath := filepath.Join(aliasDir, tool) // ~/.jfrog/package-alias/bin/mvn
				if runtime.GOOS == "windows" {
					aliasPath += ".exe"
				}

				// Read the symlink to see what it points to
				if linkTarget, err := os.Readlink(aliasPath); err == nil {
					// Resolve symlink target to absolute path
					if absTarget, err := filepath.Abs(linkTarget); err == nil {
						// Resolve current executable (might itself be a symlink) to get actual target
						resolvedExec, err := filepath.EvalSymlinks(currentExec)
						if err != nil {
							resolvedExec = currentExec // Fallback to original if resolution fails
						}

						// If the alias symlink points to the same binary we're running as, we're an alias!
						// Example: alias ~/.jfrog/package-alias/bin/mvn → /usr/local/bin/jf
						//          currentExec = /usr/local/bin/jf
						//          Match! We're running as the mvn alias
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

// DisableAliasesForThisProcess removes the alias directory from PATH
// This prevents recursion when we try to execute the real tool.
//
// When jf <tool> internally needs to execute the real tool (e.g., jf mvn calling mvn),
// it uses exec.LookPath() or exec.Command() which search PATH. By removing the alias
// directory from PATH in the current process, we ensure these lookups find the real
// tool binary, not our alias symlink.
//
// This PATH modification affects:
// - The current process (all subsequent PATH lookups)
// - All subprocesses spawned by this process (they inherit the modified PATH)
func DisableAliasesForThisProcess() error {
	aliasDir, err := GetAliasBinDir()
	if err != nil {
		return err
	}

	oldPath := os.Getenv("PATH")
	newPath := FilterOutDirFromPATH(oldPath, aliasDir)

	return os.Setenv("PATH", newPath)
}
