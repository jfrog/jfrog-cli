package packagealias

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"

	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

const GhostFrogEnvVar = "JFROG_CLI_GHOST_FROG"

const ghostFrogLogPrefix = "[GHOST_FROG]"

// DispatchIfAlias checks if we were invoked as an alias and handles it.
// This should be called very early in main() before any other logic.
//
// JFROG_CLI_GHOST_FROG values:
//
//	"false" - bypass alias interception entirely
//	"audit" - log what would happen but run the native tool unchanged
//	any other / unset - normal interception
func DispatchIfAlias() error {
	envVal := strings.ToLower(strings.TrimSpace(os.Getenv(GhostFrogEnvVar)))
	if envVal == "false" {
		log.Debug(ghostFrogLogPrefix + " Ghost Frog disabled via " + GhostFrogEnvVar + "=false")
		return nil
	}
	auditMode := envVal == "audit"

	isAlias, tool := IsRunningAsAlias()
	if !isAlias {
		return nil
	}

	log.Debug(fmt.Sprintf("%s Detected running as alias: %s", ghostFrogLogPrefix, tool))

	// Filter alias dir from PATH to prevent recursion when execRealTool runs.
	// If this fails, exec.LookPath may find the alias again instead of the real tool, causing infinite recursion.
	pathFilterErr := DisableAliasesForThisProcess()
	if pathFilterErr != nil {
		log.Warn(fmt.Sprintf("%s Failed to filter PATH: %v", ghostFrogLogPrefix, pathFilterErr))
	}

	if auditMode {
		mode := getToolMode(tool, os.Args[1:])
		log.Info(fmt.Sprintf("[GHOST_FROG_AUDIT] Would intercept '%s' (mode=%s) -- passing to native tool instead", tool, mode))
		if pathFilterErr != nil {
			return fmt.Errorf("%s cannot run native %s in audit mode: failed to remove alias from PATH (would cause recursion): %w", ghostFrogLogPrefix, tool, pathFilterErr)
		}
		return execRealTool(tool, os.Args[1:])
	}

	if !isEnabled() {
		log.Info(fmt.Sprintf("%s Package aliasing is disabled -- running native '%s'", ghostFrogLogPrefix, tool))
		if pathFilterErr != nil {
			return fmt.Errorf("%s cannot run native %s: failed to remove alias from PATH (would cause recursion): %w", ghostFrogLogPrefix, tool, pathFilterErr)
		}
		return execRealTool(tool, os.Args[1:])
	}

	log.Info(fmt.Sprintf("%s Intercepting '%s' command", ghostFrogLogPrefix, tool))

	mode := getToolMode(tool, os.Args[1:])

	switch mode {
	case ModeJF:
		return runJFMode(tool, os.Args[1:])
	case ModeEnv:
		if pathFilterErr != nil {
			return fmt.Errorf("%s cannot run native %s: failed to remove alias from PATH (would cause recursion): %w", ghostFrogLogPrefix, tool, pathFilterErr)
		}
		return runEnvMode(tool, os.Args[1:])
	case ModePass:
		if pathFilterErr != nil {
			return fmt.Errorf("%s cannot run native %s: failed to remove alias from PATH (would cause recursion): %w", ghostFrogLogPrefix, tool, pathFilterErr)
		}
		return execRealTool(tool, os.Args[1:])
	default:
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
		return fmt.Errorf("%s could not determine executable path: %w", ghostFrogLogPrefix, err)
	}

	newArgs := make([]string, 0, len(os.Args)+1)
	newArgs = append(newArgs, execPath) // Use actual executable path
	newArgs = append(newArgs, tool)     // Add tool name as first argument
	newArgs = append(newArgs, args...)  // Add remaining arguments

	os.Args = newArgs

	log.Debug(fmt.Sprintf("%s Running in JF mode: %v", ghostFrogLogPrefix, os.Args))
	log.Info(fmt.Sprintf("%s Transforming '%s' to 'jf %s'", ghostFrogLogPrefix, tool, tool))

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
		return fmt.Errorf("%s could not find real '%s' binary on PATH (Ghost Frog shim cannot dispatch): %w", ghostFrogLogPrefix, tool, err)
	}

	log.Debug(fmt.Sprintf("%s Executing real tool: %s", ghostFrogLogPrefix, realPath))

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
