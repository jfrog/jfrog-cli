package packagealias

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/jfrog/jfrog-client-go/utils/log"
)

// DispatchIfAlias checks if we were invoked as an alias and handles it
// This should be called very early in main() before any other logic
func DispatchIfAlias() error {
	isAlias, tool := IsRunningAsAlias()
	if !isAlias {
		// Not running as alias, continue normal jf execution
		return nil
	}

	log.Debug(fmt.Sprintf("Detected running as alias: %s", tool))

	// CRITICAL: Remove alias directory from PATH to prevent recursion
	if err := DisableAliasesForThisProcess(); err != nil {
		log.Warn(fmt.Sprintf("Failed to filter PATH: %v", err))
	}

	// Check if aliasing is enabled
	if !isEnabled() {
		log.Debug("Package aliasing is disabled, running native tool")
		return execRealTool(tool, os.Args[1:])
	}

	// Load tool configuration
	mode := getToolMode(tool)

	switch mode {
	case ModeJF:
		// Run through JFrog CLI integration
		return runJFMode(tool, os.Args[1:])
	case ModeEnv:
		// Inject environment variables then run native
		return runEnvMode(tool, os.Args[1:])
	case ModePass:
		// Pass through to native tool
		return execRealTool(tool, os.Args[1:])
	default:
		// Default to JF mode
		return runJFMode(tool, os.Args[1:])
	}
}

// isEnabled checks if package aliasing is enabled
func isEnabled() bool {
	aliasDir, err := GetAliasHomeDir()
	if err != nil {
		return false
	}

	statePath := filepath.Join(aliasDir, stateFile)
	data, err := os.ReadFile(statePath)
	if err != nil {
		// If state file doesn't exist, assume enabled
		return true
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return true
	}

	return state.Enabled
}

// getToolMode returns the configured mode for a tool
func getToolMode(tool string) AliasMode {
	aliasDir, err := GetAliasHomeDir()
	if err != nil {
		return ModeJF
	}

	configPath := filepath.Join(aliasDir, configFile)
	data, err := os.ReadFile(configPath)
	if err != nil {
		// Default to JF mode if no config
		return ModeJF
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return ModeJF
	}

	if mode, ok := config.ToolModes[tool]; ok {
		return mode
	}

	return ModeJF
}

// runJFMode runs the tool through JFrog CLI integration
func runJFMode(tool string, args []string) error {
	// Simply adjust os.Args to look like "jf <tool> <args>"
	// and return to continue normal jf execution
	newArgs := []string{"jf", tool}
	newArgs = append(newArgs, args...)
	os.Args = newArgs

	log.Debug(fmt.Sprintf("Running in JF mode: %v", os.Args))

	// Return nil to continue with normal jf command processing
	return nil
}

// runEnvMode runs the tool with injected environment variables
func runEnvMode(tool string, args []string) error {
	// Environment injection mode is reserved for future use
	// Currently, this mode acts as a pass-through
	return execRealTool(tool, args)
}

// execRealTool executes the real binary, replacing the current process
func execRealTool(tool string, args []string) error {
	// Find the real tool (PATH has been filtered)
	realPath, err := exec.LookPath(tool)
	if err != nil {
		return fmt.Errorf("could not find real %s: %w", tool, err)
	}

	log.Debug(fmt.Sprintf("Executing real tool: %s", realPath))

	// Prepare arguments - first arg should be the tool name
	argv := append([]string{tool}, args...)

	// On Unix, use syscall.Exec to replace the process
	// This is the cleanest way - no subprocess, just exec
	return syscall.Exec(realPath, argv, os.Environ())
}
