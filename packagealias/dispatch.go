package packagealias

import (
	"fmt"
	"os"
	"os/exec"
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

	// Filter alias dir from PATH to prevent recursion.
	if err := DisableAliasesForThisProcess(); err != nil {
		log.Warn(fmt.Sprintf("Failed to filter PATH: %v", err))
	}

	// Check if aliasing is enabled before intercepting
	if !isEnabled() {
		log.Info(fmt.Sprintf("Package aliasing is disabled - running native '%s'", tool))
		return execRealTool(tool, os.Args[1:])
	}

	log.Info(fmt.Sprintf("👻 Ghost Frog: Intercepting '%s' command", tool))

	// Load tool configuration
	mode := getToolMode(tool, os.Args[1:])

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
	return getEnabledState(aliasDir)
}

// getToolMode returns the effective mode for a tool.
func getToolMode(tool string, args []string) AliasMode {
	aliasDir, err := GetAliasHomeDir()
	if err != nil {
		return ModeJF
	}

	config, err := loadConfig(aliasDir)
	if err != nil {
		log.Warn(fmt.Sprintf("Failed to read package-alias config: %v. Falling back to default mode.", err))
		return ModeJF
	}
	return getModeForTool(config, tool, args)
}

// runJFMode rewrites invocation to `jf <tool> <args>`.
func runJFMode(tool string, args []string) error {
	execPath, err := os.Executable()
	if err != nil {
		execPath = os.Args[0]
	}

	newArgs := make([]string, 0, len(os.Args)+1)
	newArgs = append(newArgs, execPath) // Use actual executable path
	newArgs = append(newArgs, tool)     // Add tool name as first argument
	newArgs = append(newArgs, args...)  // Add remaining arguments

	os.Args = newArgs

	log.Debug(fmt.Sprintf("Running in JF mode: %v", os.Args))
	log.Info(fmt.Sprintf("👻 Ghost Frog: Transforming '%s' to 'jf %s'", tool, tool))

	return nil
}

// runEnvMode runs the tool with injected environment variables
func runEnvMode(tool string, args []string) error {
	// Environment injection mode is reserved for future use
	// Currently, this mode acts as a pass-through
	return execRealTool(tool, args)
}

// execRealTool replaces current process with real tool binary.
func execRealTool(tool string, args []string) error {
	realPath, err := exec.LookPath(tool)
	if err != nil {
		return fmt.Errorf("could not find real %s: %w", tool, err)
	}

	log.Debug(fmt.Sprintf("Executing real tool: %s", realPath))

	argv := append([]string{tool}, args...)
	return syscall.Exec(realPath, argv, os.Environ())
}
