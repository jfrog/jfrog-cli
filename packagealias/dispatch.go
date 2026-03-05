package packagealias

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"syscall"

	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
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

	// Filter alias dir from PATH to prevent recursion when execRealTool runs.
	// If this fails, exec.LookPath may find the alias again instead of the real tool, causing infinite recursion.
	pathFilterErr := DisableAliasesForThisProcess()
	if pathFilterErr != nil {
		log.Warn(fmt.Sprintf("Failed to filter PATH: %v", pathFilterErr))
	}

	// Check if aliasing is enabled before intercepting
	if !isEnabled() {
		log.Info(fmt.Sprintf("Package aliasing is disabled - running native '%s'", tool))
		if pathFilterErr != nil {
			return fmt.Errorf("cannot run native %s: failed to remove alias from PATH (would cause recursion): %w", tool, pathFilterErr)
		}
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
		if pathFilterErr != nil {
			return fmt.Errorf("cannot run native %s: failed to remove alias from PATH (would cause recursion): %w", tool, pathFilterErr)
		}
		return runEnvMode(tool, os.Args[1:])
	case ModePass:
		// Pass through to native tool
		if pathFilterErr != nil {
			return fmt.Errorf("cannot run native %s: failed to remove alias from PATH (would cause recursion): %w", tool, pathFilterErr)
		}
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
		return fmt.Errorf("could not determine executable path: %w", err)
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
// On Unix, uses syscall.Exec to replace the process. On Windows, syscall.Exec
// returns EWINDOWS (not supported), so we run the tool as a child and exit with its code.
func execRealTool(tool string, args []string) error {
	realPath, err := exec.LookPath(tool)
	if err != nil {
		return fmt.Errorf("could not find real %s: %w", tool, err)
	}

	log.Debug(fmt.Sprintf("Executing real tool: %s", realPath))

	argv := append([]string{tool}, args...)

	if runtime.GOOS == "windows" {
		return execRealToolWindows(realPath, argv)
	}

	// #nosec G702 -- realPath is resolved via exec.LookPath from a controlled tool name, not arbitrary user input.
	return syscall.Exec(realPath, argv, os.Environ())
}

// execRealToolWindows runs the real tool as a child process and exits with the child's exit code.
// Used because syscall.Exec is not supported on Windows (returns EWINDOWS).
// On success, exits with 0 (never returns). On failure, returns coreutils.CliError so the caller's ExitOnErr can exit with the correct code.
func execRealToolWindows(realPath string, argv []string) error {
	cmd := exec.Command(realPath, argv[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	err := cmd.Run()
	if err == nil {
		os.Exit(0)
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		return coreutils.CliError{
			ExitCode: coreutils.ExitCode{Code: exitErr.ExitCode()},
			ErrorMsg: err.Error(),
		}
	}
	return coreutils.CliError{ExitCode: coreutils.ExitCodeError, ErrorMsg: err.Error()}
}
