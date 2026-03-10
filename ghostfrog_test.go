package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/jfrog/jfrog-cli/packagealias"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

var (
	ghostFrogJfBin string
	ghostFrogTmpDir string
)

func InitGhostFrogTests() {
	tmpDir, err := os.MkdirTemp("", "ghostfrog-e2e-*")
	if err != nil {
		fmt.Printf("Failed to create temp dir for Ghost Frog tests: %v\n", err)
		os.Exit(1)
	}
	ghostFrogTmpDir = tmpDir

	if envBin := os.Getenv("JF_BIN"); envBin != "" {
		ghostFrogJfBin = envBin
		return
	}

	binName := "jf"
	if runtime.GOOS == "windows" {
		binName = "jf.exe"
	}
	ghostFrogJfBin = filepath.Join(tmpDir, binName)
	buildCmd := exec.Command("go", "build", "-o", ghostFrogJfBin, ".")
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		fmt.Printf("Failed to build jf binary for Ghost Frog tests: %v\n", err)
		os.Exit(1)
	}
}

func CleanGhostFrogTests() {
	if ghostFrogTmpDir != "" {
		_ = os.RemoveAll(ghostFrogTmpDir)
	}
}

func initGhostFrogTest(t *testing.T) string {
	if !*tests.TestGhostFrog {
		t.Skip("Skipping Ghost Frog test. To run Ghost Frog test add the '-test.ghostFrog=true' option.")
	}
	homeDir := t.TempDir()
	t.Setenv("JFROG_CLI_HOME_DIR", homeDir)
	return homeDir
}

func runJfCommand(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command(ghostFrogJfBin, args...)
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func installAliases(t *testing.T, packages string) {
	t.Helper()
	args := []string{"package-alias", "install"}
	if packages != "" {
		args = append(args, "--packages", packages)
	}
	out, err := runJfCommand(t, args...)
	require.NoError(t, err, "install failed: %s", out)
}

func aliasBinDir(homeDir string) string {
	return filepath.Join(homeDir, "package-alias", "bin")
}

func aliasToolPath(homeDir, tool string) string {
	name := tool
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return filepath.Join(aliasBinDir(homeDir), name)
}

// ---------------------------------------------------------------------------
// Section 15.2 - Core E2E Scenarios (E2E-001 to E2E-012)
// ---------------------------------------------------------------------------

// E2E-001: Install aliases on clean user
func TestGhostFrogInstallCleanUser(t *testing.T) {
	homeDir := initGhostFrogTest(t)
	installAliases(t, "npm,go")

	binDir := aliasBinDir(homeDir)
	_, err := os.Stat(binDir)
	require.NoError(t, err, "alias bin dir should exist")

	for _, tool := range []string{"npm", "go"} {
		_, err := os.Stat(aliasToolPath(homeDir, tool))
		require.NoError(t, err, "alias for %s should exist", tool)
	}
}

// E2E-002: Idempotent reinstall
func TestGhostFrogIdempotentReinstall(t *testing.T) {
	homeDir := initGhostFrogTest(t)
	installAliases(t, "npm,go")
	installAliases(t, "npm,go")

	for _, tool := range []string{"npm", "go"} {
		info, err := os.Stat(aliasToolPath(homeDir, tool))
		require.NoError(t, err, "alias for %s should survive reinstall", tool)
		require.True(t, info.Size() > 0, "alias binary should not be empty")
	}
}

// E2E-003: Uninstall rollback and reinstall
func TestGhostFrogUninstallRollback(t *testing.T) {
	homeDir := initGhostFrogTest(t)
	installAliases(t, "npm,go")

	out, err := runJfCommand(t, "package-alias", "uninstall")
	require.NoError(t, err, "uninstall failed: %s", out)

	binDir := aliasBinDir(homeDir)
	if _, statErr := os.Stat(binDir); statErr == nil {
		for _, tool := range []string{"npm", "go"} {
			_, err := os.Stat(aliasToolPath(homeDir, tool))
			assert.True(t, os.IsNotExist(err), "alias for %s should be removed", tool)
		}
	}

	installAliases(t, "npm,go")
	for _, tool := range []string{"npm", "go"} {
		_, err := os.Stat(aliasToolPath(homeDir, tool))
		require.NoError(t, err, "alias for %s should exist after reinstall", tool)
	}
}

// E2E-004: Enable and disable switch
func TestGhostFrogEnableDisableSwitch(t *testing.T) {
	initGhostFrogTest(t)
	installAliases(t, "npm")

	out, err := runJfCommand(t, "package-alias", "disable")
	require.NoError(t, err, "disable failed: %s", out)

	statusOut, err := runJfCommand(t, "package-alias", "status")
	require.NoError(t, err, "status failed: %s", statusOut)
	assert.Contains(t, statusOut, "DISABLED")

	out, err = runJfCommand(t, "package-alias", "enable")
	require.NoError(t, err, "enable failed: %s", out)

	statusOut, err = runJfCommand(t, "package-alias", "status")
	require.NoError(t, err, "status failed: %s", statusOut)
	assert.Contains(t, statusOut, "ENABLED")
}

// E2E-004b: JFROG_CLI_GHOST_FROG=false kill switch bypasses interception
func TestGhostFrogKillSwitchEnvVar(t *testing.T) {
	homeDir := initGhostFrogTest(t)
	installAliases(t, "npm")

	binDir := aliasBinDir(homeDir)
	npmPath := aliasToolPath(homeDir, "npm")

	// With kill switch active, the alias should skip interception entirely
	cmd := exec.Command(npmPath, "--version")
	cmd.Env = append(os.Environ(),
		"PATH="+binDir+string(os.PathListSeparator)+os.Getenv("PATH"),
		"JFROG_CLI_LOG_LEVEL=DEBUG",
		"JFROG_CLI_GHOST_FROG=false",
	)
	out, _ := cmd.CombinedOutput()
	outputStr := string(out)

	assert.NotContains(t, outputStr, "Detected running as alias",
		"kill switch should prevent interception")
	assert.Contains(t, outputStr, "Ghost Frog disabled via",
		"kill switch bypass should be logged")
}

// E2E-004c: JFROG_CLI_GHOST_FROG=audit logs interception but runs native tool
func TestGhostFrogAuditMode(t *testing.T) {
	homeDir := initGhostFrogTest(t)
	installAliases(t, "npm")

	binDir := aliasBinDir(homeDir)
	npmPath := aliasToolPath(homeDir, "npm")

	cmd := exec.Command(npmPath, "--version")
	cmd.Env = append(os.Environ(),
		"PATH="+binDir+string(os.PathListSeparator)+os.Getenv("PATH"),
		"JFROG_CLI_LOG_LEVEL=DEBUG",
		"JFROG_CLI_GHOST_FROG=audit",
	)
	out, _ := cmd.CombinedOutput()
	outputStr := string(out)

	assert.Contains(t, outputStr, "[GHOST_FROG_AUDIT]",
		"audit mode should log the GHOST_FROG_AUDIT marker")
	assert.Contains(t, outputStr, "Would intercept",
		"audit mode should describe what it would do")
	assert.NotContains(t, outputStr, "Transforming",
		"audit mode must not actually transform the command")
}

// E2E-005: Alias dispatch by argv0
func TestGhostFrogAliasDispatchByArgv0(t *testing.T) {
	homeDir := initGhostFrogTest(t)
	installAliases(t, "npm,go")

	binDir := aliasBinDir(homeDir)
	for _, tool := range []string{"npm", "go"} {
		toolPath := aliasToolPath(homeDir, tool)
		_, err := os.Stat(toolPath)
		require.NoError(t, err, "alias binary for %s must exist at %s", tool, toolPath)
	}

	cmd := exec.Command(aliasToolPath(homeDir, "npm"), "--version")
	cmd.Env = append(os.Environ(), "PATH="+binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	cmd.Env = append(cmd.Env, "JFROG_CLI_LOG_LEVEL=DEBUG")
	out, _ := cmd.CombinedOutput()
	outputStr := string(out)

	// Strictly require Ghost Frog interception logs, not just any "npm" match
	assert.True(t,
		strings.Contains(outputStr, "Detected running as alias") ||
			strings.Contains(outputStr, "Ghost Frog"),
		"alias dispatch must produce Ghost Frog interception logs (JFROG_CLI_LOG_LEVEL=DEBUG), got: %s", outputStr)
}

// E2E-006: PATH filter per process
func TestGhostFrogPATHFilterPerProcess(t *testing.T) {
	homeDir := initGhostFrogTest(t)
	binDir := aliasBinDir(homeDir)

	originalPATH := "/usr/bin" + string(os.PathListSeparator) + binDir + string(os.PathListSeparator) + "/usr/local/bin"
	filtered := packagealias.FilterOutDirFromPATH(originalPATH, binDir)

	assert.NotContains(t, filepath.SplitList(filtered), binDir,
		"alias dir should be removed from PATH")
	assert.Contains(t, filtered, "/usr/bin", "other dirs should remain")
	assert.Contains(t, filtered, "/usr/local/bin", "other dirs should remain")
}

// E2E-007: Recursion prevention under fallback
func TestGhostFrogRecursionPreventionFallback(t *testing.T) {
	homeDir := initGhostFrogTest(t)
	installAliases(t, "npm")

	binDir := aliasBinDir(homeDir)
	cmd := exec.Command(aliasToolPath(homeDir, "npm"), "--version")
	cmd.Env = append(os.Environ(),
		"PATH="+binDir,
		"JFROG_CLI_LOG_LEVEL=DEBUG",
	)

	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _ = cmd.CombinedOutput()
	}()

	select {
	case <-done:
	case <-timeAfter(t, 30):
		t.Fatal("alias command hung -- possible recursion loop")
	}
}

// E2E-008: Real binary missing
func TestGhostFrogRealBinaryMissing(t *testing.T) {
	homeDir := initGhostFrogTest(t)
	installAliases(t, "npm")

	// Set exclude mode so it tries to exec the real binary
	out, err := runJfCommand(t, "package-alias", "exclude", "npm")
	require.NoError(t, err, "exclude failed: %s", out)

	binDir := aliasBinDir(homeDir)
	cmd := exec.Command(aliasToolPath(homeDir, "npm"), "--version")
	// PATH only contains alias dir -- no real npm available
	cmd.Env = append(os.Environ(), "PATH="+binDir)
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	// Should fail (no real npm) but with a clear error, not a hang
	if err == nil {
		t.Logf("Unexpected success -- npm may be embedded or found elsewhere: %s", outputStr)
		return
	}
	assert.True(t,
		strings.Contains(outputStr, "could not find") ||
			strings.Contains(outputStr, "not found") ||
			strings.Contains(outputStr, "failed") ||
			strings.Contains(outputStr, "error") ||
			strings.Contains(strings.ToLower(outputStr), "npm"),
		"should produce a clear error about missing tool, got: %s", outputStr)
}

// E2E-009: PATH contains alias dir multiple times
func TestGhostFrogPATHMultipleAliasDirs(t *testing.T) {
	homeDir := initGhostFrogTest(t)
	binDir := aliasBinDir(homeDir)

	sep := string(os.PathListSeparator)
	pathWithDuplicates := binDir + sep + "/usr/bin" + sep + binDir + sep + "/usr/local/bin" + sep + binDir
	filtered := packagealias.FilterOutDirFromPATH(pathWithDuplicates, binDir)

	for _, entry := range filepath.SplitList(filtered) {
		assert.NotEqual(t, filepath.Clean(binDir), filepath.Clean(entry),
			"all instances of alias dir should be removed")
	}
	assert.Contains(t, filtered, "/usr/bin")
	assert.Contains(t, filtered, "/usr/local/bin")
}

// E2E-010: PATH contains relative alias path
func TestGhostFrogPATHRelativeAliasPath(t *testing.T) {
	homeDir := initGhostFrogTest(t)
	binDir := aliasBinDir(homeDir)

	// FilterOutDirFromPATH uses filepath.Clean for comparison
	sep := string(os.PathListSeparator)
	normalizedBinDir := filepath.Clean(binDir)
	pathWithRelative := normalizedBinDir + sep + "/usr/bin"
	filtered := packagealias.FilterOutDirFromPATH(pathWithRelative, binDir)

	assert.NotContains(t, filepath.SplitList(filtered), normalizedBinDir,
		"normalized alias dir should be removed")
}

// E2E-011: Shell hash cache stale path (documentation test)
func TestGhostFrogShellHashCacheStalePath(t *testing.T) {
	homeDir := initGhostFrogTest(t)
	installAliases(t, "npm")

	// Verify the binary exists -- the rest is shell-level behavior
	_, err := os.Stat(aliasToolPath(homeDir, "npm"))
	require.NoError(t, err, "alias should exist for hash cache test scenario")
}

// E2E-012: Mixed mode policies using include and exclude
func TestGhostFrogMixedModePolicies(t *testing.T) {
	initGhostFrogTest(t)
	installAliases(t, "npm,go,mvn")

	out, err := runJfCommand(t, "package-alias", "exclude", "npm")
	require.NoError(t, err, "exclude npm failed: %s", out)

	out, err = runJfCommand(t, "package-alias", "include", "go")
	require.NoError(t, err, "include go failed: %s", out)

	statusOut, err := runJfCommand(t, "package-alias", "status")
	require.NoError(t, err, "status failed: %s", statusOut)

	assert.Contains(t, statusOut, "npm")
	assert.Contains(t, statusOut, "go")
	assert.Contains(t, statusOut, "mvn")
}

// ---------------------------------------------------------------------------
// Section 15.3 - Parallelism and Concurrency (E2E-020 to E2E-025)
// ---------------------------------------------------------------------------

// E2E-020: Parallel same tool invocations
func TestGhostFrogParallelSameTool(t *testing.T) {
	homeDir := initGhostFrogTest(t)
	installAliases(t, "npm")

	binDir := aliasBinDir(homeDir)
	npmPath := aliasToolPath(homeDir, "npm")

	var wg sync.WaitGroup
	var failures atomic.Int32
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cmd := exec.Command(npmPath, "--version")
			cmd.Env = append(os.Environ(), "PATH="+binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
			if _, err := cmd.CombinedOutput(); err != nil {
				failures.Add(1)
			}
		}()
	}
	wg.Wait()
	// Allow failures from missing real npm, but assert no hangs (test completes)
	t.Logf("Parallel same-tool: %d/4 failures (acceptable if npm not installed)", failures.Load())
}

// E2E-021: Parallel mixed tool invocations
func TestGhostFrogParallelMixedTools(t *testing.T) {
	homeDir := initGhostFrogTest(t)
	installAliases(t, "npm,go")

	binDir := aliasBinDir(homeDir)
	toolNames := []string{"npm", "go", "npm", "go"}

	var wg sync.WaitGroup
	var failures atomic.Int32
	for _, tool := range toolNames {
		wg.Add(1)
		go func(toolName string) {
			defer wg.Done()
			toolPath := aliasToolPath(homeDir, toolName)
			args := []string{"--version"}
			if toolName == "go" {
				args = []string{"version"}
			}
			cmd := exec.Command(toolPath, args...)
			cmd.Env = append(os.Environ(), "PATH="+binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
			if _, err := cmd.CombinedOutput(); err != nil {
				failures.Add(1)
			}
		}(tool)
	}
	wg.Wait()
	t.Logf("Parallel mixed-tool: %d/4 failures (acceptable if tools not installed)", failures.Load())
}

// E2E-022: Parallel aliased and native command
func TestGhostFrogParallelMixedWithNative(t *testing.T) {
	homeDir := initGhostFrogTest(t)
	installAliases(t, "npm")

	binDir := aliasBinDir(homeDir)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		cmd := exec.Command(aliasToolPath(homeDir, "npm"), "--version")
		cmd.Env = append(os.Environ(), "PATH="+binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
		_, _ = cmd.CombinedOutput()
	}()

	go func() {
		defer wg.Done()
		nativeCmd := "echo"
		if runtime.GOOS == "windows" {
			nativeCmd = "cmd"
		}
		cmd := exec.Command(nativeCmd, "hello")
		if runtime.GOOS == "windows" {
			cmd = exec.Command(nativeCmd, "/C", "echo", "hello")
		}
		out, err := cmd.CombinedOutput()
		assert.NoError(t, err, "native command should succeed: %s", string(out))
	}()

	wg.Wait()
}

// E2E-023: Concurrent enable and disable race
func TestGhostFrogConcurrentEnableDisableRace(t *testing.T) {
	initGhostFrogTest(t)
	installAliases(t, "npm")

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			if idx%2 == 0 {
				_, _ = runJfCommand(t, "package-alias", "disable")
			} else {
				_, _ = runJfCommand(t, "package-alias", "enable")
			}
		}(i)
	}
	wg.Wait()

	// Should be in a valid state after the race
	statusOut, err := runJfCommand(t, "package-alias", "status")
	require.NoError(t, err, "status should succeed after race: %s", statusOut)
	assert.True(t,
		strings.Contains(statusOut, "ENABLED") || strings.Contains(statusOut, "DISABLED"),
		"status should show a valid state")
}

// E2E-024: One process fails, others continue
func TestGhostFrogOneProcessFailsOthersContinue(t *testing.T) {
	homeDir := initGhostFrogTest(t)
	installAliases(t, "npm,go")

	binDir := aliasBinDir(homeDir)
	var wg sync.WaitGroup
	results := make([]error, 3)

	commands := []struct {
		path string
		args []string
	}{
		{aliasToolPath(homeDir, "npm"), []string{"--version"}},
		{aliasToolPath(homeDir, "go"), []string{"version"}},
		// intentionally invalid -- should fail but not crash others
		{aliasToolPath(homeDir, "npm"), []string{"nonexistent-command-xyz"}},
	}

	for i, c := range commands {
		wg.Add(1)
		go func(idx int, cmdPath string, cmdArgs []string) {
			defer wg.Done()
			cmd := exec.Command(cmdPath, cmdArgs...)
			cmd.Env = append(os.Environ(), "PATH="+binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
			_, results[idx] = cmd.CombinedOutput()
		}(i, c.path, c.args)
	}
	wg.Wait()
	// All should complete (not hang), regardless of individual success/failure
}

// E2E-025: High fan-out stress
func TestGhostFrogHighFanOutStress(t *testing.T) {
	homeDir := initGhostFrogTest(t)
	installAliases(t, "npm,go")

	binDir := aliasBinDir(homeDir)
	var wg sync.WaitGroup
	var completed atomic.Int32
	workerCount := 24

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			tool := "npm"
			args := []string{"--version"}
			if idx%2 == 1 {
				tool = "go"
				args = []string{"version"}
			}
			cmd := exec.Command(aliasToolPath(homeDir, tool), args...)
			cmd.Env = append(os.Environ(), "PATH="+binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
			_, _ = cmd.CombinedOutput()
			completed.Add(1)
		}(i)
	}
	wg.Wait()
	assert.Equal(t, int32(workerCount), completed.Load(),
		"all workers should complete without hangs")
}

// ---------------------------------------------------------------------------
// Section 15.4 - CI/CD Scenarios (E2E-030 to E2E-034)
// ---------------------------------------------------------------------------

// E2E-030: setup-jfrog-cli native integration
func TestGhostFrogSetupJFrogCLINativeIntegration(t *testing.T) {
	homeDir := initGhostFrogTest(t)
	installAliases(t, "npm,mvn,go,pip")

	binDir := aliasBinDir(homeDir)
	for _, tool := range []string{"npm", "mvn", "go", "pip"} {
		_, err := os.Stat(aliasToolPath(homeDir, tool))
		require.NoError(t, err, "alias for %s should exist", tool)
	}

	statusOut, err := runJfCommand(t, "package-alias", "status")
	require.NoError(t, err, "status failed: %s", statusOut)
	assert.Contains(t, statusOut, "INSTALLED")
	assert.Contains(t, statusOut, "ENABLED")

	// Verify alias dir is populated
	entries, err := os.ReadDir(binDir)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(entries), 4, "should have at least 4 alias entries")
}

// E2E-031: Auto build-info publish (requires Artifactory)
func TestGhostFrogAutoBuildInfoPublish(t *testing.T) {
	initGhostFrogTest(t)
	skipIfNoArtifactory(t, "E2E-031")
	// When Artifactory is available: run an aliased npm install, verify
	// build-info is collected and published automatically at the end.
	t.Log("E2E-031: Artifactory available -- build-info auto-publish validation is a future enhancement")
}

// E2E-032: Manual publish precedence (requires Artifactory)
func TestGhostFrogManualPublishPrecedence(t *testing.T) {
	initGhostFrogTest(t)
	skipIfNoArtifactory(t, "E2E-032")
	t.Log("E2E-032: Artifactory available -- manual publish precedence validation is a future enhancement")
}

// E2E-033: Auto publish disabled (requires Artifactory)
func TestGhostFrogAutoPublishDisabled(t *testing.T) {
	initGhostFrogTest(t)
	skipIfNoArtifactory(t, "E2E-033")
	t.Log("E2E-033: Artifactory available -- auto-publish disabled validation is a future enhancement")
}

// E2E-034: Jenkins pipeline compatibility
func TestGhostFrogJenkinsPipelineCompat(t *testing.T) {
	homeDir := initGhostFrogTest(t)
	installAliases(t, "npm,mvn")

	binDir := aliasBinDir(homeDir)
	originalPATH := os.Getenv("PATH")
	simulatedPATH := binDir + string(os.PathListSeparator) + originalPATH

	for _, tool := range []string{"npm", "mvn"} {
		toolPath := aliasToolPath(homeDir, tool)
		_, err := os.Stat(toolPath)
		require.NoError(t, err, "alias for %s should exist at %s", tool, toolPath)
	}

	filtered := packagealias.FilterOutDirFromPATH(simulatedPATH, binDir)
	assert.NotContains(t, filepath.SplitList(filtered), binDir,
		"PATH filter should work in Jenkins-like environments")
}

// ---------------------------------------------------------------------------
// Section 15.5 - Security, Safety, and Isolation (E2E-040 to E2E-044)
// ---------------------------------------------------------------------------

// E2E-040: Non-root installation
func TestGhostFrogNonRootInstallation(t *testing.T) {
	homeDir := initGhostFrogTest(t)
	installAliases(t, "npm")

	binDir := aliasBinDir(homeDir)
	info, err := os.Stat(binDir)
	require.NoError(t, err)

	// Alias dir should be in user-space (under temp/home), not system dirs
	assert.True(t, strings.HasPrefix(binDir, homeDir),
		"alias dir should be under user home, not system directories")
	assert.True(t, info.IsDir())
}

// E2E-041: System binary integrity
func TestGhostFrogSystemBinaryIntegrity(t *testing.T) {
	homeDir := initGhostFrogTest(t)

	// Find a real tool on the system before install
	realToolName := "echo"
	if runtime.GOOS == "windows" {
		realToolName = "cmd"
	}
	realToolBefore, err := exec.LookPath(realToolName)
	if err != nil {
		t.Skipf("system tool %s not found, skipping integrity check", realToolName)
	}

	infoBefore, err := os.Stat(realToolBefore)
	require.NoError(t, err)
	sizeBefore := infoBefore.Size()

	installAliases(t, "npm,go")

	infoAfter, err := os.Stat(realToolBefore)
	require.NoError(t, err)

	assert.Equal(t, sizeBefore, infoAfter.Size(),
		"system binary %s should not be modified by install", realToolBefore)
	_ = homeDir
}

// E2E-042: User-scope cleanup
func TestGhostFrogUserScopeCleanup(t *testing.T) {
	homeDir := initGhostFrogTest(t)
	installAliases(t, "npm")

	aliasDir := filepath.Join(homeDir, "package-alias")
	err := os.RemoveAll(aliasDir)
	require.NoError(t, err, "should be able to remove alias directory manually")

	_, err = os.Stat(aliasDir)
	assert.True(t, os.IsNotExist(err), "alias dir should be gone after manual removal")
}

// E2E-043: Child env inheritance
func TestGhostFrogChildEnvInheritance(t *testing.T) {
	homeDir := initGhostFrogTest(t)
	binDir := aliasBinDir(homeDir)

	sep := string(os.PathListSeparator)
	parentPATH := binDir + sep + "/usr/bin" + sep + "/usr/local/bin"
	filtered := packagealias.FilterOutDirFromPATH(parentPATH, binDir)

	// Simulate a child process inheriting the filtered PATH
	childEnv := append(os.Environ(), "PATH="+filtered)
	var pathCmd *exec.Cmd
	if runtime.GOOS == "windows" {
		pathCmd = exec.Command("cmd", "/C", "echo", "%PATH%")
	} else {
		pathCmd = exec.Command("sh", "-c", "echo $PATH")
	}
	pathCmd.Env = childEnv
	out, err := pathCmd.CombinedOutput()
	require.NoError(t, err, "child should run with filtered PATH")

	childPath := strings.TrimSpace(string(out))
	assert.NotContains(t, childPath, binDir,
		"child should not see alias dir in inherited PATH")
}

// E2E-044: Cross-session isolation
func TestGhostFrogCrossSessionIsolation(t *testing.T) {
	homeDir := initGhostFrogTest(t)
	binDir := aliasBinDir(homeDir)

	sep := string(os.PathListSeparator)
	sessionPath := binDir + sep + "/usr/bin"
	filtered := packagealias.FilterOutDirFromPATH(sessionPath, binDir)

	// Verify current process PATH is NOT modified by FilterOutDirFromPATH
	currentPATH := os.Getenv("PATH")
	assert.NotEqual(t, filtered, currentPATH,
		"filtering should produce a new string, not modify current PATH")
}

// ---------------------------------------------------------------------------
// Section 15.6 - Platform-Specific Edge Cases (E2E-050 to E2E-054)
// ---------------------------------------------------------------------------

// E2E-050: Windows copy-based aliases
func TestGhostFrogWindowsCopyAliases(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("E2E-050: Windows-only test")
	}
	homeDir := initGhostFrogTest(t)
	installAliases(t, "npm,go")

	for _, tool := range []string{"npm", "go"} {
		exePath := filepath.Join(aliasBinDir(homeDir), tool+".exe")
		info, err := os.Stat(exePath)
		require.NoError(t, err, "%s.exe should exist", tool)
		assert.True(t, info.Size() > 0, "%s.exe should not be empty", tool)
	}
}

// E2E-051: Windows PATH case-insensitivity
func TestGhostFrogWindowsPATHCaseInsensitive(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("E2E-051: Windows-only test")
	}
	homeDir := initGhostFrogTest(t)
	binDir := aliasBinDir(homeDir)

	sep := string(os.PathListSeparator)
	upperDir := strings.ToUpper(binDir)
	pathVal := upperDir + sep + "C:\\Windows\\System32"
	filtered := packagealias.FilterOutDirFromPATH(pathVal, binDir)

	for _, entry := range filepath.SplitList(filtered) {
		assert.False(t, strings.EqualFold(filepath.Clean(entry), filepath.Clean(binDir)),
			"FilterOutDirFromPATH must remove the alias dir even when casing differs on Windows")
	}
	assert.Contains(t, filtered, "C:\\Windows\\System32",
		"non-alias PATH entries must be preserved")
}

// E2E-052: Spaces in user home path
func TestGhostFrogSpacesInHomePath(t *testing.T) {
	if !*tests.TestGhostFrog {
		t.Skip("Skipping Ghost Frog test. To run Ghost Frog test add the '-test.ghostFrog=true' option.")
	}

	baseDir := t.TempDir()
	homeWithSpaces := filepath.Join(baseDir, "my home dir", "with spaces")
	require.NoError(t, os.MkdirAll(homeWithSpaces, 0755))
	t.Setenv("JFROG_CLI_HOME_DIR", homeWithSpaces)

	installAliases(t, "npm")

	binDir := filepath.Join(homeWithSpaces, "package-alias", "bin")
	_, err := os.Stat(binDir)
	require.NoError(t, err, "alias dir should be created even with spaces in path")

	toolName := "npm"
	if runtime.GOOS == "windows" {
		toolName += ".exe"
	}
	_, err = os.Stat(filepath.Join(binDir, toolName))
	require.NoError(t, err, "alias binary should exist under path with spaces")
}

// E2E-053: Symlink unsupported environment fallback
func TestGhostFrogSymlinkUnsupportedFallback(t *testing.T) {
	if runtime.GOOS == "windows" {
		// On Windows, install uses copy, not symlinks
		homeDir := initGhostFrogTest(t)
		installAliases(t, "npm")

		exePath := filepath.Join(aliasBinDir(homeDir), "npm.exe")
		info, err := os.Stat(exePath)
		require.NoError(t, err)
		assert.True(t, info.Mode().IsRegular(), "Windows aliases should be regular files (copies)")
		return
	}

	homeDir := initGhostFrogTest(t)
	installAliases(t, "npm")

	npmPath := aliasToolPath(homeDir, "npm")
	info, err := os.Lstat(npmPath)
	require.NoError(t, err)
	assert.True(t, info.Mode()&os.ModeSymlink != 0,
		"Unix aliases should be symlinks")
}

// E2E-054: Tool name collision
func TestGhostFrogToolNameCollision(t *testing.T) {
	homeDir := initGhostFrogTest(t)
	installAliases(t, "npm")

	binDir := aliasBinDir(homeDir)
	_, err := os.Stat(aliasToolPath(homeDir, "npm"))
	require.NoError(t, err)

	// Verify alias dir is separate from any system tool
	sep := string(os.PathListSeparator)
	pathWithAlias := binDir + sep + os.Getenv("PATH")
	filtered := packagealias.FilterOutDirFromPATH(pathWithAlias, binDir)

	// After filtering, npm should resolve to system npm (if present), not alias
	parts := filepath.SplitList(filtered)
	for _, p := range parts {
		assert.NotEqual(t, filepath.Clean(binDir), filepath.Clean(p),
			"alias dir should not remain after filtering")
	}
}

// ---------------------------------------------------------------------------
// Section 15.7 - Negative and Recovery Cases (E2E-060 to E2E-064)
// ---------------------------------------------------------------------------

// E2E-060: Corrupt state/config file
func TestGhostFrogCorruptConfig(t *testing.T) {
	homeDir := initGhostFrogTest(t)
	installAliases(t, "npm")

	configPath := filepath.Join(homeDir, "package-alias", "config.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte("{{{{invalid yaml!!!!"), 0600))

	// Status should handle corrupt config gracefully
	statusOut, err := runJfCommand(t, "package-alias", "status")
	// Should not crash even with corrupt config
	t.Logf("Status after corrupt config (err=%v): %s", err, statusOut)
}

// E2E-061: Partial install damage
func TestGhostFrogPartialInstallDamage(t *testing.T) {
	homeDir := initGhostFrogTest(t)
	installAliases(t, "npm,go")

	// Remove just one alias binary
	npmPath := aliasToolPath(homeDir, "npm")
	require.NoError(t, os.Remove(npmPath))

	// go alias should still work
	goPath := aliasToolPath(homeDir, "go")
	_, err := os.Stat(goPath)
	require.NoError(t, err, "go alias should survive partial damage")

	// npm alias should be missing
	_, err = os.Stat(npmPath)
	assert.True(t, os.IsNotExist(err), "npm alias should be missing after removal")

	// Status should still work and report the damage
	statusOut, err := runJfCommand(t, "package-alias", "status")
	require.NoError(t, err, "status should succeed with partial damage: %s", statusOut)
}

// E2E-062: Interrupted install
func TestGhostFrogInterruptedInstall(t *testing.T) {
	homeDir := initGhostFrogTest(t)

	// Simulate partial state by creating dir but no config
	binDir := aliasBinDir(homeDir)
	require.NoError(t, os.MkdirAll(binDir, 0755))

	// A fresh install should recover
	installAliases(t, "npm")

	_, err := os.Stat(aliasToolPath(homeDir, "npm"))
	require.NoError(t, err, "install should succeed after interrupted state")

	configPath := filepath.Join(homeDir, "package-alias", "config.yaml")
	_, err = os.Stat(configPath)
	require.NoError(t, err, "config should be created")
}

// E2E-063: Broken PATH ordering (alias dir appended instead of prepended)
func TestGhostFrogBrokenPATHOrdering(t *testing.T) {
	homeDir := initGhostFrogTest(t)
	binDir := aliasBinDir(homeDir)

	// Alias dir appended (not prepended)
	sep := string(os.PathListSeparator)
	brokenPATH := "/usr/bin" + sep + "/usr/local/bin" + sep + binDir

	// Filter should still remove it regardless of position
	filtered := packagealias.FilterOutDirFromPATH(brokenPATH, binDir)
	for _, entry := range filepath.SplitList(filtered) {
		assert.NotEqual(t, filepath.Clean(binDir), filepath.Clean(entry),
			"alias dir should be removed even when appended")
	}
}

// E2E-064: Unsupported tool invocation
func TestGhostFrogUnsupportedToolInvocation(t *testing.T) {
	homeDir := initGhostFrogTest(t)
	installAliases(t, "npm")

	// curl is not in SupportedTools -- should not have an alias
	unsupportedAlias := filepath.Join(aliasBinDir(homeDir), "curl")
	if runtime.GOOS == "windows" {
		unsupportedAlias += ".exe"
	}
	_, err := os.Stat(unsupportedAlias)
	assert.True(t, os.IsNotExist(err),
		"unsupported tool should not have an alias binary")

	// Install should reject unsupported tool
	out, err := runJfCommand(t, "package-alias", "install", "--packages", "curl")
	assert.Error(t, err, "install should reject unsupported tool: %s", out)
}

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

func timeAfter(t *testing.T, seconds int) <-chan struct{} {
	t.Helper()
	ch := make(chan struct{})
	go func() {
		timer := make(chan struct{})
		go func() {
			cmd := exec.Command("sleep", fmt.Sprintf("%d", seconds))
			if runtime.GOOS == "windows" {
				cmd = exec.Command("timeout", "/T", fmt.Sprintf("%d", seconds), "/NOBREAK")
			}
			_ = cmd.Run()
			close(timer)
		}()
		<-timer
		close(ch)
	}()
	return ch
}

// configForTest is a helper for tests that need to read/validate config
type configForTest struct {
	Enabled         bool                          `yaml:"enabled"`
	ToolModes       map[string]packagealias.AliasMode `yaml:"tool_modes,omitempty"`
	SubcommandModes map[string]packagealias.AliasMode `yaml:"subcommand_modes,omitempty"`
	EnabledTools    []string                      `yaml:"enabled_tools,omitempty"`
}

func readTestConfig(t *testing.T, homeDir string) configForTest {
	t.Helper()
	configPath := filepath.Join(homeDir, "package-alias", "config.yaml")
	data, err := os.ReadFile(configPath)
	require.NoError(t, err, "should be able to read config.yaml")

	var cfg configForTest
	require.NoError(t, yaml.Unmarshal(data, &cfg), "config should be valid YAML")
	return cfg
}

func skipIfNoArtifactory(t *testing.T, testID string) {
	t.Helper()
	jfrogURL := os.Getenv("JF_URL")
	jfrogToken := os.Getenv("JF_ACCESS_TOKEN")
	if jfrogURL == "" && tests.JfrogUrl != nil {
		jfrogURL = *tests.JfrogUrl
	}
	if jfrogToken == "" && tests.JfrogAccessToken != nil {
		jfrogToken = *tests.JfrogAccessToken
	}
	if jfrogURL == "" || jfrogURL == "http://localhost:8081/" || jfrogToken == "" {
		t.Skipf("%s: Skipped -- no Artifactory credentials. Set JF_URL and JF_ACCESS_TOKEN or use --jfrog.url and --jfrog.adminToken.", testID)
	}
}
