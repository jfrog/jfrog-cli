package main

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	biutils "github.com/jfrog/build-info-go/utils"
	agentTestutil "github.com/jfrog/jfrog-cli-artifactory/agent/common/testutil"
	plugincommon "github.com/jfrog/jfrog-cli-artifactory/agent/plugins/common"
	"github.com/jfrog/jfrog-cli-artifactory/artifactory/commands/generic"
	artUtils "github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	coreBuild "github.com/jfrog/jfrog-cli-core/v2/common/build"
	"github.com/jfrog/jfrog-cli-core/v2/common/spec"
	coretests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	"github.com/jfrog/jfrog-cli-evidence/evidence/cryptox"
	"github.com/jfrog/jfrog-cli-evidence/evidence/generate"
	"github.com/jfrog/jfrog-client-go/utils/log"
	clientTestUtils "github.com/jfrog/jfrog-client-go/utils/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jfrog/jfrog-cli/inttestutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
)

// ---------------------------------------------------------------------------
// Init / cleanup
// ---------------------------------------------------------------------------

// InitAgentPluginsTests sets up the e2e test environment. .github/workflows/agentPluginsTests.yml
// installs real claude-code and codex binaries onto PATH, so plugincommon's default
// LookPath*/*Exec hooks (which shell out to "claude"/"codex" on PATH) work as-is in CI. Until
// that workflow change is merged - and for local runs where neither CLI is installed -
// ensureNativeAgentCLIs falls back to a stub binary per missing agent.
func InitAgentPluginsTests() {
	initArtifactoryCli()
	cleanUpOldRepositories()
	tests.AddTimestampToGlobalVars()
	createRequiredRepos()
}

func CleanAgentPluginsTests() {
	deleteCreatedRepos()
}

func initAgentPluginsTest(t *testing.T) {
	if !*tests.TestAgentPlugins {
		t.Skip("Skipping Agent Plugins test. To run Agent Plugins test add the '-test.agentPlugins=true' option.")
	}
	createJfrogHomeConfig(t, false)
	require.True(t, isRepoExist(tests.AgentPluginsLocalRepo), "agent plugins local repo does not exist: "+tests.AgentPluginsLocalRepo)
	// The test Artifactory instance has no evidence/One-Model service configured.
	// Disable the quiet-failure evidence gate so install commands don't block on 403.
	t.Setenv("JFROG_AGENT_PLUGINS_DISABLE_QUIET_FAILURE", "true")
	ensureNativeAgentCLIs(t)
}

func cleanAgentPluginsTest() {
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.AgentPluginsBuildName, artHttpDetails)
	tests.CleanFileSystem()
}

// runAgentPluginsCmd executes `jf agent plugins <args...>`.
func runAgentPluginsCmd(t *testing.T, args ...string) error {
	t.Helper()
	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	return jfrogCli.Exec(append([]string{"agent", "plugins"}, args...)...)
}

// runAgentPluginsCmdWithOutput executes `jf agent plugins <args...>` and returns captured stdout.
func runAgentPluginsCmdWithOutput(t *testing.T, args ...string) (string, error) {
	t.Helper()
	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	return jfrogCli.RunCliCmdWithOutputs(t, append([]string{"agent", "plugins"}, args...)...)
}

// assertErrorContainsAll requires a non-nil error whose message contains every substring.
// Prefer this over loose OR-chains (e.g. "repo" || "405") that pass on unrelated failures.
func assertErrorContainsAll(t *testing.T, err error, substrings ...string) {
	t.Helper()
	require.Error(t, err)
	msg := err.Error()
	for _, sub := range substrings {
		assert.Contains(t, msg, sub, "error %q should contain %q", msg, sub)
	}
}

// recreateAgentPluginsLocalRepo recreates the e2e agentplugins repository after a temporary delete.
func recreateAgentPluginsLocalRepo(t *testing.T) {
	t.Helper()
	repoConfig := tests.GetTestResourcesPath() + tests.AgentPluginsLocalRepositoryConfig
	repoConfig, err := tests.ReplaceTemplateVariables(repoConfig, "")
	require.NoError(t, err)
	execCreateRepoRest(repoConfig, tests.AgentPluginsLocalRepo)
	require.True(t, isRepoExist(tests.AgentPluginsLocalRepo),
		"agent plugins local repo must exist after recreate: "+tests.AgentPluginsLocalRepo)
}

// createTestPlugin copies the test-plugin fixture to a fresh temp dir and patches
// plugin.json with the given slug and version so tests don't conflict.
func createTestPlugin(t *testing.T, slug, version string) string {
	t.Helper()
	pluginSrc := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "agent_plugins", "test-plugin")
	pluginPath, cleanup := coretests.CreateTempDirWithCallbackAndAssert(t)
	t.Cleanup(cleanup)

	require.NoError(t, biutils.CopyDir(pluginSrc, pluginPath, true, nil))

	manifest := map[string]any{
		"name":        slug,
		"version":     version,
		"description": "Integration test plugin",
		"skills":      []string{},
	}
	data, err := json.Marshal(manifest)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(pluginPath, "plugin.json"), data, 0644)) // #nosec G306 -- test fixture
	return pluginPath
}

// assertPluginExists verifies the zip for slug/version is present in the local repo.
func assertPluginExists(t *testing.T, slug, version string) {
	t.Helper()
	sm, err := artUtils.CreateServiceManager(serverDetails, -1, 0, false)
	require.NoError(t, err)
	_, err = sm.GetItemProps(pluginArtifactPath(tests.AgentPluginsLocalRepo, slug, version))
	require.NoError(t, err, "artifact should exist: %s v%s", slug, version)
}

// assertPluginAbsent verifies the zip for slug/version is gone from the local repo.
// A search-based check gives a cleaner absence assertion without relying on error message text.
func assertPluginAbsent(t *testing.T, slug, version string) {
	t.Helper()
	path := pluginArtifactPath(tests.AgentPluginsLocalRepo, slug, version)
	searchSpec := spec.NewBuilder().Pattern(path).BuildSpec()
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetServerDetails(serverDetails).SetSpec(searchSpec)
	reader, err := searchCmd.Search()
	require.NoError(t, err, "search for absent artifact failed")
	// Read-side Close: search reader has no write flush; nothing actionable on failure.
	defer func() { _ = reader.Close() }()
	var found bool
	for item := new(artUtils.SearchResult); reader.NextRecord(item) == nil; item = new(artUtils.SearchResult) {
		found = true
		break
	}
	assert.False(t, found, "artifact should not exist: %s v%s", slug, version)
}

// pluginArtifactPath returns the Artifactory path for a published plugin zip:
// <repo>/<slug>/<version>/<slug>-<version>.zip
func pluginArtifactPath(repo, slug, version string) string {
	return repo + "/" + slug + "/" + version + "/" + slug + "-" + version + ".zip"
}

// ---------------------------------------------------------------------------
// Harness helpers
// ---------------------------------------------------------------------------

// agentPluginHarnessCase is one of the required harness conditions:
// claude, codex, cursor, vscode, and combined claude,codex,cursor,vscode.
type agentPluginHarnessCase struct {
	name      string
	harnesses []string
}

// allAgentHarnesses represents all 4 supported agent harnesses.
var allAgentHarnesses = []string{"claude", "codex", "cursor", "vscode"}

func agentPluginHarnessCases() []agentPluginHarnessCase {
	return []agentPluginHarnessCase{
		{name: "claude", harnesses: []string{"claude"}},
		{name: "codex", harnesses: []string{"codex"}},
		{name: "cursor", harnesses: []string{"cursor"}},
		{name: "vscode", harnesses: []string{"vscode"}},
		{name: "claude,codex,cursor,vscode", harnesses: allAgentHarnesses},
	}
}

func harnessFlag(harnesses []string) string {
	return strings.Join(harnesses, ",")
}

// globalPluginInstallDir returns the current global install destination for a built-in harness.
// claude/codex use repo-keyed layout under .../local/jfrog/<repo>/<slug>;
// vscode uses repo-keyed layout under ~/.copilot/installed-plugins/<repo>/<slug>;
// cursor installs under ~/.cursor/plugins/local/<slug>.
func globalPluginInstallDir(homeDir, harness, repoKey, slug string) string {
	switch strings.ToLower(harness) {
	case "claude":
		return filepath.Join(homeDir, ".claude", "plugins", "local", "jfrog", repoKey, slug)
	case "codex":
		return filepath.Join(homeDir, ".agents", "plugins", "local", "jfrog", repoKey, slug)
	case "cursor":
		return filepath.Join(homeDir, ".cursor", "plugins", "local", slug)
	case "vscode":
		return filepath.Join(homeDir, ".copilot", "installed-plugins", repoKey, slug)
	default:
		return filepath.Join(homeDir, "."+harness, "plugins", slug)
	}
}

// assertPluginsInstalledGlobally checks each harness install directory and plugin-info.json.
// When wantVersion is set, also asserts installedVersion, slug, agent, and repo fields.
func assertPluginsInstalledGlobally(t *testing.T, homeDir string, harnesses []string, slug string, wantVersion ...string) {
	t.Helper()
	version := ""
	if len(wantVersion) > 0 {
		version = wantVersion[0]
	}
	for _, harness := range harnesses {
		path := globalPluginInstallDir(homeDir, harness, tests.AgentPluginsLocalRepo, slug)
		assert.DirExists(t, path, "plugin %q should be installed for harness %q at %s", slug, harness, path)
		manifestPath := filepath.Join(path, ".jfrog", "plugin-info.json")
		assert.FileExists(t, manifestPath, "plugin-info.json should exist for harness %q", harness)
		if version == "" {
			continue
		}
		data, err := os.ReadFile(manifestPath) // #nosec G304 -- path under t.TempDir
		require.NoError(t, err)
		var manifest map[string]any
		require.NoError(t, json.Unmarshal(data, &manifest))
		assert.Equal(t, version, manifest["installedVersion"], "installedVersion for harness %q", harness)
		assert.Equal(t, slug, manifest["slug"], "slug for harness %q", harness)
		assert.Equal(t, harness, manifest["agent"], "agent for harness %q", harness)
		assert.Equal(t, tests.AgentPluginsLocalRepo, manifest["repo"], "repo for harness %q", harness)
	}
}

// assertPluginsInstalledNatively verifies plugins via each harness's native CLI commands.
// Only applies to Claude/Codex which have native CLI support for plugin listing.
// Cursor/VSCode don't have CLI support, so they skip this check (filesystem check via assertPluginsInstalledGlobally is sufficient).
func assertPluginsInstalledNatively(t *testing.T, harnesses []string, slug string, wantVersion string) {
	t.Helper()
	// Validate inputs to catch test bugs early
	require.NotEmpty(t, harnesses, "harnesses list must not be empty")
	require.NotEmpty(t, slug, "plugin slug must not be empty")
	require.NotEmpty(t, wantVersion, "plugin version must not be empty")

	for _, harness := range harnesses {
		switch strings.ToLower(harness) {
		case "claude":
			assertClaudePluginInstalled(t, slug, wantVersion)
		case "codex":
			assertCodexPluginInstalled(t, slug, wantVersion)
		case "cursor", "vscode":
			// Cursor and VSCode don't have native CLI plugin commands
			// Filesystem validation via assertPluginsInstalledGlobally is sufficient
			continue
		default:
			t.Fatalf("unknown harness: %s", harness)
		}
	}
}

// assertClaudePluginInstalled calls `claude plugin list --json` and verifies plugin presence.
func assertClaudePluginInstalled(t *testing.T, slug string, wantVersion string) {
	t.Helper()
	// Use timeout to prevent test hangs if claude CLI is unresponsive
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "claude", "plugin", "list", "--json") // #nosec G204 -- fixed command
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	require.NoError(t, err, "claude plugin list --json failed: %s", stderr.String())

	var plugins []map[string]any
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &plugins),
		"failed to parse claude plugin list output (stderr: %s)", stderr.String())

	found := false
	for _, p := range plugins {
		id, ok := p["id"].(string)
		if !ok {
			// Log type mismatch for debugging if id is missing/wrong type
			continue
		}
		// id format: "<slug>@<repo>"
		if strings.HasPrefix(id, slug+"@") {
			version, ok := p["version"].(string)
			require.True(t, ok, "plugin %s missing version field or wrong type; got: %T", id, p["version"])
			assert.Equal(t, wantVersion, version, "claude plugin %s has wrong version", id)
			found = true
			break
		}
	}
	require.True(t, found, "claude plugin %q not found in `claude plugin list`", slug)
}

// assertCodexPluginInstalled calls `codex plugin list --json` and verifies plugin in installed[].
func assertCodexPluginInstalled(t *testing.T, slug string, wantVersion string) {
	t.Helper()
	// Use timeout to prevent test hangs if codex CLI is unresponsive
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "codex", "plugin", "list", "--json") // #nosec G204 -- fixed command
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	require.NoError(t, err, "codex plugin list --json failed: %s", stderr.String())

	var result map[string]any
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &result),
		"failed to parse codex plugin list output (stderr: %s)", stderr.String())

	installed, ok := result["installed"].([]any)
	require.True(t, ok, "codex plugin list missing 'installed' array")

	found := false
	for _, p := range installed {
		plugin, ok := p.(map[string]any)
		if !ok {
			// Array element is not a map; log for debugging
			continue
		}
		pluginID, ok := plugin["pluginId"].(string)
		if !ok {
			// pluginId missing or wrong type; log for debugging
			continue
		}
		// pluginId format: "<slug>@<marketplace>"
		if strings.HasPrefix(pluginID, slug+"@") {
			version, ok := plugin["version"].(string)
			require.True(t, ok, "codex plugin %s missing version field or wrong type; got: %T", pluginID, plugin["version"])
			assert.Equal(t, wantVersion, version, "codex plugin %s has wrong version", pluginID)
			found = true
			break
		}
	}
	require.True(t, found, "codex plugin %q not found in `codex plugin list`", slug)
}

// verifyIsolatedHome verifies that HOME was actually changed to an isolated directory.
func verifyIsolatedHome(t *testing.T, homeDir string) {
	t.Helper()
	// Verify directory exists
	require.DirExists(t, homeDir, "isolated HOME directory should exist")
	// Verify HOME env var is set to isolated dir
	require.Equal(t, homeDir, os.Getenv("HOME"), "HOME env var should point to isolated directory")
	// Verify directory is under system temp (cross-platform: os.TempDir() returns system temp path)
	// t.TempDir() creates subdirectories under os.TempDir(), so we check that homeDir starts with temp root
	tempDir := os.TempDir()
	tempRoot := filepath.Dir(tempDir) // Get parent of temp dir to allow for t.TempDir() subdirectories
	require.True(t, strings.HasPrefix(filepath.Clean(homeDir), filepath.Clean(tempRoot)),
		"isolated HOME should be in system temp directory, got %s (temp root: %s)", homeDir, tempRoot)
}

func setIsolatedHome(t *testing.T) string {
	t.Helper()
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("USERPROFILE", homeDir)
	return homeDir
}

// pluginManifestDir returns the directory holding a harness's plugin.json inside a plugin.
// Harnesses use .<harness>-plugin, except vscode, which reads plugin.json at the plugin root.
func pluginManifestDir(pluginPath, harness string) string {
	if strings.EqualFold(harness, "vscode") {
		return pluginPath
	}
	return filepath.Join(pluginPath, "."+harness+"-plugin")
}

// createTestHarnessPlugin creates a plugin fixture with a plugin.json for each harness.
func createTestHarnessPlugin(t *testing.T, slug, version string, harnesses []string) string {
	t.Helper()
	pluginPath, cleanup := coretests.CreateTempDirWithCallbackAndAssert(t)
	t.Cleanup(cleanup)

	for _, harness := range harnesses {
		harnessDir := pluginManifestDir(pluginPath, harness)
		require.NoError(t, os.MkdirAll(harnessDir, 0755)) // #nosec G301 -- test directory
		manifest := map[string]any{
			"name":        slug,
			"version":     version,
			"description": fmt.Sprintf("Integration test plugin for %s", harness),
			"skills":      []string{},
		}
		data, err := json.Marshal(manifest)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(filepath.Join(harnessDir, "plugin.json"), data, 0644)) // #nosec G306 -- test fixture
	}
	return pluginPath
}

// ensureNativeAgentCLIs makes sure both "claude" and "codex" are runnable for the CLI under
// test. It prefers the real CLIs - installed onto PATH by .github/workflows/agentPluginsTests.yml
// in CI - and only falls back to a stub binary, per agent, when that agent's real CLI isn't
// found (e.g. locally, or before that workflow change has been merged). Leaving the real CLI
// alone when present means CI still exercises actual claude/codex behavior.
func ensureNativeAgentCLIs(t *testing.T) {
	t.Helper()
	ensureNativeAgentCLI(t, "claude", plugincommon.LookPathClaude, func(bin string) func() {
		prevLook, prevExec := plugincommon.LookPathClaude, plugincommon.ClaudeExec
		plugincommon.LookPathClaude = func() (string, error) { return bin, nil }
		plugincommon.ClaudeExec = func(args ...string) error {
			return exec.Command(bin, args...).Run() // #nosec G204 -- test stub binary
		}
		return func() {
			plugincommon.LookPathClaude = prevLook
			plugincommon.ClaudeExec = prevExec
		}
	})
	ensureNativeAgentCLI(t, "codex", plugincommon.LookPathCodex, func(bin string) func() {
		prevLook, prevExec := plugincommon.LookPathCodex, plugincommon.CodexExec
		plugincommon.LookPathCodex = func() (string, error) { return bin, nil }
		plugincommon.CodexExec = func(args ...string) error {
			return exec.Command(bin, args...).Run() // #nosec G204 -- test stub binary
		}
		return func() {
			plugincommon.LookPathCodex = prevLook
			plugincommon.CodexExec = prevExec
		}
	})
}

// ensureNativeAgentCLI checks lookPath (the agent's current LookPath hook); if it already
// resolves, the real CLI is left untouched. Otherwise it installs the shared stub binary under
// the given name and calls wire to point the agent's LookPath/Exec hooks at it, restoring the
// previous hooks via t.Cleanup.
func ensureNativeAgentCLI(t *testing.T, name string, lookPath func() (string, error), wire func(bin string) (restore func())) {
	t.Helper()
	if _, err := lookPath(); err == nil {
		return // real CLI already on PATH (installed by CI, or present locally) - nothing to do
	}

	data, err := nativeAgentStubBinary()
	require.NoError(t, err)

	binDir := t.TempDir()
	bin := filepath.Join(binDir, name)
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	require.NoError(t, os.WriteFile(bin, data, 0755)) // #nosec G306 -- test stub binary path is under t.TempDir

	// jfrog-cli-artifactory's native-registry lookups (claudePluginListJSON/codexPluginListJSON,
	// used by "list --check-updates" and uninstall-detection) shell out to the literal
	// "claude"/"codex" command via OS PATH resolution, not through the exported LookPath/Exec
	// hooks wired below - so the stub must also be reachable on PATH under its bare name.
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	t.Cleanup(wire(bin))
}

// nativeAgentStubOnce builds the shared claude/codex stub binary once per process.
// ensureNativeAgentCLI copies that binary into each test's fallback bin dir so we avoid
// recompiling on every initAgentPluginsTest call.
var (
	nativeAgentStubOnce sync.Once
	nativeAgentStubData []byte
	nativeAgentStubErr  error
)

func nativeAgentStubBinary() ([]byte, error) {
	nativeAgentStubOnce.Do(func() {
		srcDir, err := os.MkdirTemp("", "jf-agent-plugins-native-stub-src-*")
		if err != nil {
			nativeAgentStubErr = err
			return
		}
		defer func() { _ = os.RemoveAll(srcDir) }() // best-effort teardown of compile workspace

		if err := os.WriteFile(filepath.Join(srcDir, "go.mod"), []byte("module nativeagentstub\n\ngo 1.22\n"), 0644); err != nil { // #nosec G306 -- test fixture
			nativeAgentStubErr = err
			return
		}
		if err := os.WriteFile(filepath.Join(srcDir, "main.go"), []byte(nativeAgentCLIStubSource), 0644); err != nil { // #nosec G306 -- test fixture
			nativeAgentStubErr = err
			return
		}

		outBin := filepath.Join(srcDir, "claude")
		if runtime.GOOS == "windows" {
			outBin += ".exe"
		}
		build := exec.Command("go", "build", "-o", outBin, ".") // #nosec G204 -- fixed go build of local stub sources
		build.Dir = srcDir
		out, err := build.CombinedOutput()
		if err != nil {
			nativeAgentStubErr = fmt.Errorf("building claude stub failed: %w: %s", err, string(out))
			return
		}
		nativeAgentStubData, nativeAgentStubErr = os.ReadFile(outBin) // #nosec G304 -- path under MkdirTemp
	})
	return nativeAgentStubData, nativeAgentStubErr
}

// nativeAgentCLIStubSource is a tiny stdlib-only program that pretends to be claude/codex.
// `plugin list --json` scans jf's global install layout under $HOME/$USERPROFILE.
const nativeAgentCLIStubSource = `package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	name := filepath.Base(os.Args[0])
	name = strings.TrimSuffix(name, ".exe")
	if len(os.Args) >= 3 && os.Args[1] == "plugin" && os.Args[2] == "list" {
		home := os.Getenv("HOME")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		switch name {
		case "claude":
			_ = json.NewEncoder(os.Stdout).Encode(scanClaude(home))
		case "codex":
			_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"installed": scanCodex(home)})
		default:
			_, _ = os.Stdout.Write([]byte("[]\n"))
		}
		return
	}
}

type pluginInfo struct {
	InstalledVersion string ` + "`json:\"installedVersion\"`" + `
}

func scanClaude(home string) []map[string]string {
	root := filepath.Join(home, ".claude", "plugins", "local", "jfrog")
	var out []map[string]string
	repos, _ := os.ReadDir(root)
	for _, repo := range repos {
		if !repo.IsDir() || strings.HasPrefix(repo.Name(), ".") {
			continue
		}
		slugs, _ := os.ReadDir(filepath.Join(root, repo.Name()))
		for _, slug := range slugs {
			if !slug.IsDir() || strings.HasPrefix(slug.Name(), ".") {
				continue
			}
			dir := filepath.Join(root, repo.Name(), slug.Name())
			out = append(out, map[string]string{
				"id":          slug.Name() + "@" + repo.Name(),
				"version":     installedVersion(dir),
				"installPath": dir,
			})
		}
	}
	if out == nil {
		out = []map[string]string{}
	}
	return out
}

func scanCodex(home string) []map[string]any {
	root := filepath.Join(home, ".agents", "plugins", "local", "jfrog")
	var out []map[string]any
	repos, _ := os.ReadDir(root)
	for _, repo := range repos {
		if !repo.IsDir() || strings.HasPrefix(repo.Name(), ".") {
			continue
		}
		slugs, _ := os.ReadDir(filepath.Join(root, repo.Name()))
		for _, slug := range slugs {
			if !slug.IsDir() || strings.HasPrefix(slug.Name(), ".") {
				continue
			}
			dir := filepath.Join(root, repo.Name(), slug.Name())
			out = append(out, map[string]any{
				"pluginId":        slug.Name() + "@" + repo.Name(),
				"name":            slug.Name(),
				"marketplaceName": repo.Name(),
				"version":         installedVersion(dir),
				"source":          map[string]string{"path": dir},
			})
		}
	}
	if out == nil {
		out = []map[string]any{}
	}
	return out
}

func installedVersion(dir string) string {
	data, err := os.ReadFile(filepath.Join(dir, ".jfrog", "plugin-info.json"))
	if err != nil {
		return ""
	}
	var info pluginInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return ""
	}
	return info.InstalledVersion
}
`

// ---------------------------------------------------------------------------
// Publish
// ---------------------------------------------------------------------------

// TestAgentPluginsPublish verifies that publishing a plugin directory uploads
// the zip to the correct path in the agentplugins local repository.
func TestAgentPluginsPublish(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "publish-plugin"
	version := "1.0.0"
	pluginPath := createTestPlugin(t, slug, version)

	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	))

	assertPluginExists(t, slug, version)
}

// TestAgentPluginsVersionCollisionCI verifies that publishing the same version
// twice in CI/non-interactive mode fails with a clear "already exists" error.
func TestAgentPluginsVersionCollisionCI(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "collision-plugin"
	version := "1.0.0"
	pluginPath := createTestPlugin(t, slug, version)

	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	))

	// Force non-interactive mode so the collision check fails immediately.
	t.Setenv("CI", "true")
	err := runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	)
	require.Error(t, err, "second publish of the same version in CI mode should fail")
	assertErrorContainsAll(t, err,
		fmt.Sprintf("version %s of plugin '%s' already exists", version, slug),
		"Use a different version or remove the existing one")
}

// TestAgentPluginsPublishWithVersion verifies that --version overrides the
// manifest version on publish.
func TestAgentPluginsPublishWithVersion(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "version-override-plugin"
	overrideVersion := "2.0.0"
	pluginPath := createTestPlugin(t, slug, "1.0.0")

	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--version="+overrideVersion,
	))

	assertPluginExists(t, slug, overrideVersion)
}

// TestAgentPluginsPublishMissingPluginJson verifies that publishing a directory
// without plugin.json returns a clear error.
func TestAgentPluginsPublishMissingPluginJson(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	emptyDir := t.TempDir()
	err := runAgentPluginsCmd(t,
		"publish", emptyDir,
		"--repo="+tests.AgentPluginsLocalRepo,
	)
	assertErrorContainsAll(t, err, "no plugin.json")
}

// TestAgentPluginsPublishToNonExistentRepo verifies that publishing to a
// nonexistent repository returns a clear error.
func TestAgentPluginsPublishToNonExistentRepo(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	pluginPath := createTestPlugin(t, "invalid-repo-plugin", "1.0.0")
	err := runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo=nonexistent-agent-plugins-repo-xyz",
	)
	// Publish wraps the Artifactory upload failure (see publish.go: "upload failed: %w").
	assertErrorContainsAll(t, err, "upload failed")
}

// TestAgentPluginsChecksumIntegrity verifies that after publish the artifact
// in build info has non-empty, non-"untrusted" MD5, SHA1, and SHA256 checksums,
// confirming Artifactory computed all three correctly on upload.
func TestAgentPluginsChecksumIntegrity(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "checksum-plugin"
	version := "1.0.0"
	buildNumber := t.Name()
	// Best-effort teardown of local build-info dir; leftover dirs do not affect assertions.
	t.Cleanup(func() { _ = coreBuild.RemoveBuildDir(tests.AgentPluginsBuildName, buildNumber, "") })
	pluginPath := createTestPlugin(t, slug, version)

	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--build-name="+tests.AgentPluginsBuildName,
		"--build-number="+buildNumber,
	))
	require.NoError(t, artifactoryCli.Exec("bp", tests.AgentPluginsBuildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.AgentPluginsBuildName, buildNumber)
	require.NoError(t, err, "GetBuildInfo failed")
	require.True(t, found, "build info not found — was jf rt bp successful?")
	require.Len(t, publishedBuildInfo.BuildInfo.Modules, 1)
	require.NotEmpty(t, publishedBuildInfo.BuildInfo.Modules[0].Artifacts,
		"expected at least one artifact in build info")

	for _, artifact := range publishedBuildInfo.BuildInfo.Modules[0].Artifacts {
		assert.NotEmpty(t, artifact.Md5, "artifact %s: md5 must not be empty", artifact.Name)
		assert.NotEmpty(t, artifact.Sha1, "artifact %s: sha1 must not be empty", artifact.Name)
		assert.NotEmpty(t, artifact.Sha256, "artifact %s: sha256 must not be empty", artifact.Name)
		assert.NotEqual(t, "untrusted", strings.ToLower(artifact.Md5),
			"artifact %s: md5 must not be 'untrusted'", artifact.Name)
		assert.NotEqual(t, "untrusted", strings.ToLower(artifact.Sha1),
			"artifact %s: sha1 must not be 'untrusted'", artifact.Name)
		assert.NotEqual(t, "untrusted", strings.ToLower(artifact.Sha256),
			"artifact %s: sha256 must not be 'untrusted'", artifact.Name)
	}
}

// TestAgentPluginsPublishWithBuildInfo verifies that --build-name and
// --build-number cause the published zip to appear as an artifact in build info
// with a valid SHA256 checksum.
func TestAgentPluginsPublishWithBuildInfo(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "buildinfo-plugin"
	version := "1.0.0"
	buildNumber := t.Name()
	// Best-effort teardown of local build-info dir; leftover dirs do not affect assertions.
	t.Cleanup(func() { _ = coreBuild.RemoveBuildDir(tests.AgentPluginsBuildName, buildNumber, "") })
	pluginPath := createTestPlugin(t, slug, version)

	assert.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--build-name="+tests.AgentPluginsBuildName,
		"--build-number="+buildNumber,
	))
	require.NoError(t, artifactoryCli.Exec("bp", tests.AgentPluginsBuildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.AgentPluginsBuildName, buildNumber)
	require.NoError(t, err, "GetBuildInfo failed")
	require.True(t, found, "build info not found — was 'jf rt bp' successful?")
	require.Len(t, publishedBuildInfo.BuildInfo.Modules, 1, "expected 1 build info module")

	module := publishedBuildInfo.BuildInfo.Modules[0]
	require.NotEmpty(t, module.Artifacts, "published zip should appear as an artifact in build info")
	assert.NotEmpty(t, module.Artifacts[0].Sha256, "artifact sha256 should be non-empty in build info")
}

// TestAgentPluginsPublishWithProjectFlag verifies that --project=<key> stores the
// build-info partials under the project-key-aware local directory (the build dir
// hash includes the project key), and that nothing is stored under the
// empty-project directory.
func TestAgentPluginsPublishWithProjectFlag(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "project-flag-plugin"
	buildName := tests.AgentPluginsBuildName + "-project"
	buildNumber := "1"
	projectKey := "testprj"
	t.Cleanup(func() {
		_ = coreBuild.RemoveBuildDir(buildName, buildNumber, projectKey)
		_ = coreBuild.RemoveBuildDir(buildName, buildNumber, "")
	})

	require.NoError(t, runAgentPluginsCmd(t,
		"publish", createTestPlugin(t, slug, "1.0.0"),
		"--repo="+tests.AgentPluginsLocalRepo,
		"--build-name="+buildName,
		"--build-number="+buildNumber,
		"--project="+projectKey,
	))

	partials, err := coreBuild.ReadPartialBuildInfoFiles(buildName, buildNumber, projectKey)
	require.NoError(t, err)
	require.Len(t, partials, 1, "expected 1 build-info partial stored with project key %q", projectKey)
	assert.Equal(t, slug, partials[0].ModuleId)
	assert.NotEmpty(t, partials[0].Artifacts, "published zip should be recorded as a build-info artifact")

	partialsNoProject, err := coreBuild.ReadPartialBuildInfoFiles(buildName, buildNumber, "")
	assert.NoError(t, err)
	assert.Empty(t, partialsNoProject, "build info should NOT be stored under the empty project key directory")
}

// TestAgentPluginsNoBuildInfoWithoutFlags verifies that publishing without
// --build-name and --build-number does not create a build info entry.
func TestAgentPluginsNoBuildInfoWithoutFlags(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "no-buildinfo-plugin"
	pluginPath := createTestPlugin(t, slug, "1.0.0")

	assert.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	))

	localBuilds, err := coreBuild.GetGeneratedBuildsInfo(tests.AgentPluginsBuildName, "1", "")
	assert.NoError(t, err)
	assert.Empty(t, localBuilds, "no local build info should be stored when --build-name/--build-number are absent")
}

// TestAgentPluginsPublishBuildNameWithoutNumber verifies that providing only
// one of --build-name / --build-number (scenarios #61 and #62) returns an
// error requiring both flags together.
func TestAgentPluginsPublishBuildNameWithoutNumber(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	cases := []struct {
		name        string
		extraArgs   []string
		description string
	}{
		{
			name:        "name-only",
			extraArgs:   []string{"--build-name=" + tests.AgentPluginsBuildName},
			description: "--build-name without --build-number must return an error",
		},
		{
			name:        "number-only",
			extraArgs:   []string{"--build-number=42"},
			description: "--build-number without --build-name must return an error",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			slug := "build-flag-validation-plugin-" + tc.name
			pluginPath := createTestPlugin(t, slug, "1.0.0")
			args := append([]string{"publish", pluginPath, "--repo=" + tests.AgentPluginsLocalRepo}, tc.extraArgs...)
			err := runAgentPluginsCmd(t, args...)
			assertErrorContainsAll(t, err, "the build-name and build-number options cannot be provided separately")
		})
	}
}

// TestAgentPluginsModuleOverride verifies that --module overrides the default
// module ID (slug) in build info.
func TestAgentPluginsModuleOverride(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "module-override-plugin"
	buildNumber := t.Name()
	// Best-effort teardown of local build-info dir; leftover dirs do not affect assertions.
	t.Cleanup(func() { _ = coreBuild.RemoveBuildDir(tests.AgentPluginsBuildName, buildNumber, "") })
	customModule := "my-custom-agent-module"
	pluginPath := createTestPlugin(t, slug, "1.0.0")

	assert.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--build-name="+tests.AgentPluginsBuildName,
		"--build-number="+buildNumber,
		"--module="+customModule,
	))
	require.NoError(t, artifactoryCli.Exec("bp", tests.AgentPluginsBuildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.AgentPluginsBuildName, buildNumber)
	require.NoError(t, err, "GetBuildInfo failed")
	require.True(t, found, "build info not found")
	require.NotEmpty(t, publishedBuildInfo.BuildInfo.Modules)
	assert.Equal(t, customModule, publishedBuildInfo.BuildInfo.Modules[0].Id,
		"--module flag should override the default module ID in build info")
}

// TestAgentPluginsPublishInvalidSemver verifies that a manifest with a
// non-semver version string is rejected before any upload attempt.
func TestAgentPluginsPublishInvalidSemver(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	cases := []struct {
		name        string
		version     string
		errContains []string
	}{
		{
			name:        "non-numeric-patch",
			version:     "1.9.e",
			errContains: []string{`invalid version "1.9.e"`, `patch must be a number (got "e")`},
		},
		{
			name:        "missing-patch-segment",
			version:     "1.9",
			errContains: []string{`invalid version "1.9"`, "expected format major.minor.patch"},
		},
		{
			name:        "non-numeric-major",
			version:     "x.1.0",
			errContains: []string{`invalid version "x.1.0"`, `major must be a number (got "x")`},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pluginPath := createMinimalPlugin(t, "semver-plugin", tc.version)
			err := runAgentPluginsCmd(t,
				"publish", pluginPath,
				"--repo="+tests.AgentPluginsLocalRepo,
			)
			assertErrorContainsAll(t, err, tc.errContains...)
		})
	}
}

// TestAgentPluginsPublishInvalidSlug verifies that a manifest whose name field
// contains invalid characters is rejected with a ValidateSlug error.
func TestAgentPluginsPublishInvalidSlug(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	pluginPath := createMinimalPlugin(t, "Invalid Slug With Spaces!", "1.0.0")
	err := runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	)
	assertErrorContainsAll(t, err, "invalid slug")
}

// TestAgentPluginsPublishMissingPathArg verifies that omitting the required
// <path> argument returns a usage error.
func TestAgentPluginsPublishMissingPathArg(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	err := runAgentPluginsCmd(t, "publish", "--repo="+tests.AgentPluginsLocalRepo)
	assertErrorContainsAll(t, err, "usage: jf agent plugins publish")
}

// TestAgentPluginsPublishToWrongRepoType verifies that publishing to a
// repository of the wrong package type (e.g. a generic local repo) returns an
// error from Artifactory.
func TestAgentPluginsPublishToWrongRepoType(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	// Use a known non-agentplugins repo (generic local) to trigger a type mismatch.
	wrongTypeRepo := tests.RtRepo1
	pluginPath := createTestPlugin(t, "wrong-repo-type-plugin", "1.0.0")
	err := runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+wrongTypeRepo,
	)
	assertErrorContainsAll(t, err, "upload failed")
}

// TestAgentPluginsPublishPrebuiltZip verifies that a prebuilt <slug>-<version>.zip
// inside a zip/ sub-directory is uploaded as-is without being re-zipped.
func TestAgentPluginsPublishPrebuiltZip(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "prebuilt-zip-plugin"
	version := "1.0.0"

	// Create the plugin directory with a plugin.json and a prebuilt zip.
	pluginDir := t.TempDir()
	manifest := map[string]string{"name": slug, "version": version, "description": "prebuilt test"}
	data, err := json.Marshal(manifest)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(pluginDir, "plugin.json"), data, 0644)) // #nosec G306 -- test fixture

	// The prebuilt zip lives at <pluginDir>/zip/<slug>-<version>.zip.
	zipSubDir := filepath.Join(pluginDir, "zip")
	require.NoError(t, os.MkdirAll(zipSubDir, 0755)) // #nosec G301 -- test directory
	var zipBuf bytes.Buffer
	zw := zip.NewWriter(&zipBuf)
	f, err := zw.Create("placeholder.txt")
	require.NoError(t, err)
	_, err = f.Write([]byte("placeholder"))
	require.NoError(t, err)
	require.NoError(t, zw.Close())
	zipContent := zipBuf.Bytes()
	require.NoError(t, os.WriteFile(
		filepath.Join(zipSubDir, slug+"-"+version+".zip"), zipContent, 0644, // #nosec G306 -- test fixture
	))

	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginDir,
		"--repo="+tests.AgentPluginsLocalRepo,
	), "publish with prebuilt zip should succeed without re-zipping")

	assertPluginExists(t, slug, version)
}

// TestAgentPluginsBuildPropertiesOnArtifact verifies that after publish with
// build info collection, build.name / build.number / build.timestamp are
// stamped on the artifact in Artifactory.
func TestAgentPluginsBuildPropertiesOnArtifact(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "build-props-plugin"
	version := "1.0.0"
	buildNumber := t.Name()
	// Best-effort teardown of local build-info dir; leftover dirs do not affect assertions.
	t.Cleanup(func() { _ = coreBuild.RemoveBuildDir(tests.AgentPluginsBuildName, buildNumber, "") })
	pluginPath := createTestPlugin(t, slug, version)

	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--build-name="+tests.AgentPluginsBuildName,
		"--build-number="+buildNumber,
	))
	require.NoError(t, artifactoryCli.Exec("bp", tests.AgentPluginsBuildName, buildNumber))

	sm, err := artUtils.CreateServiceManager(serverDetails, -1, 0, false)
	require.NoError(t, err)

	artifactPath := pluginArtifactPath(tests.AgentPluginsLocalRepo, slug, version)
	props, err := sm.GetItemProps(artifactPath)
	require.NoError(t, err, "GetItemProps should succeed for %s", artifactPath)
	require.NotNil(t, props)

	assert.Contains(t, props.Properties, "build.name",
		"build.name property must be stamped on the published zip")
	assert.Contains(t, props.Properties, "build.number",
		"build.number property must be stamped on the published zip")
	assert.Contains(t, props.Properties, "build.timestamp",
		"build.timestamp property must be stamped on the published zip")
}

// TestAgentPluginsBuildInfoFromEnvVars verifies that JFROG_CLI_BUILD_NAME and
// JFROG_CLI_BUILD_NUMBER environment variables trigger build info collection
// even when the --build-name/--build-number flags are not passed explicitly.
func TestAgentPluginsBuildInfoFromEnvVars(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	envBuildName := tests.AgentPluginsBuildName + "-envvar"
	envBuildNumber := "42"

	t.Setenv("JFROG_CLI_BUILD_NAME", envBuildName)
	t.Setenv("JFROG_CLI_BUILD_NUMBER", envBuildNumber)

	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, envBuildName, artHttpDetails)
	defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, envBuildName, artHttpDetails)

	slug := "envvar-build-plugin"
	pluginPath := createTestPlugin(t, slug, "1.0.0")

	// No --build-name / --build-number flags; env vars should be picked up.
	assert.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	))

	require.NoError(t, artifactoryCli.Exec("bp", envBuildName, envBuildNumber))

	_, found, err := tests.GetBuildInfo(serverDetails, envBuildName, envBuildNumber)
	require.NoError(t, err, "GetBuildInfo failed")
	assert.True(t, found,
		"build info should be captured from JFROG_CLI_BUILD_NAME/NUMBER env vars")
}

// TestAgentPluginsBuildPublishRetrievable verifies the full build info flow:
// publish plugin → publish build info → retrieve build info from Artifactory.
func TestAgentPluginsBuildPublishRetrievable(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "bp-retrievable-plugin"
	version := "1.0.0"
	buildNumber := t.Name()
	// Best-effort teardown of local build-info dir; leftover dirs do not affect assertions.
	t.Cleanup(func() { _ = coreBuild.RemoveBuildDir(tests.AgentPluginsBuildName, buildNumber, "") })
	pluginPath := createTestPlugin(t, slug, version)

	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--build-name="+tests.AgentPluginsBuildName,
		"--build-number="+buildNumber,
	))

	require.NoError(t, artifactoryCli.Exec("bp", tests.AgentPluginsBuildName, buildNumber),
		"jf rt bp should succeed")

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.AgentPluginsBuildName, buildNumber)
	require.NoError(t, err, "GetBuildInfo failed")
	require.True(t, found, "build info must be retrievable from Artifactory after jf rt bp")
	assert.Equal(t, tests.AgentPluginsBuildName, publishedBuildInfo.BuildInfo.Name,
		"retrieved build info name must match")
}

// TestAgentPluginsChecksumStoredByArtifactory publishes a plugin and verifies
// that Artifactory stores non-empty MD5, SHA1, and SHA256 for the artifact.
func TestAgentPluginsChecksumStoredByArtifactory(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "checksum-rt-plugin"
	version := "1.0.0"
	pluginPath := createTestPlugin(t, slug, version)

	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	))

	// Retrieve checksums Artifactory stored for the zip via AQL search.
	artifactPath := pluginArtifactPath(tests.AgentPluginsLocalRepo, slug, version)
	searchSpec := spec.NewBuilder().Pattern(artifactPath).BuildSpec()
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetServerDetails(serverDetails).SetSpec(searchSpec)
	reader, err := searchCmd.Search()
	require.NoError(t, err, "search for artifact checksum failed")
	// Read-side Close: search reader has no write flush; nothing actionable on failure.
	defer func() { _ = reader.Close() }()
	item := new(artUtils.SearchResult)
	require.NoError(t, reader.NextRecord(item), "artifact must be found in Artifactory")
	assert.NotEmpty(t, item.Md5, "Artifactory must store an md5 for the artifact")
	assert.NotEmpty(t, item.Sha1, "Artifactory must store a sha1 for the artifact")
	assert.NotEmpty(t, item.Sha256, "Artifactory must store a sha256 for the artifact")
}

// TestAgentPluginsPublishWithSigningKey generates a real ECDSA key pair,
// uploads the public key to Artifactory trusted keys, and publishes a plugin
// with --signing-key. Asserts the artifact exists; evidence attachment depends
// on the Artifactory evidence service and is not queried here.
func TestAgentPluginsPublishWithSigningKey(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	// The trusted keys API rejects duplicate aliases, and the suite may run more than
	// once against the same Artifactory instance, so keep the alias unique per run.
	keyAlias := fmt.Sprintf("agent-plugins-test-key-%d", time.Now().UnixNano())

	// Generate an ECDSA P-256 key pair without network access.
	privateKeyPEM, publicKeyPEM, err := cryptox.GenerateECDSAKeyPair()
	require.NoError(t, err, "key generation must succeed")

	privateKeyPath := filepath.Join(t.TempDir(), "evidence.key")
	require.NoError(t, os.WriteFile(privateKeyPath, []byte(privateKeyPEM), 0600)) // #nosec G306 -- test private key under t.TempDir

	// Upload the public key to Artifactory trusted keys so the evidence service can verify
	// signatures made with the corresponding private key. KeyPairCommand.Run is not used here:
	// it refuses to overwrite an existing key file and only warns (never returns) on upload
	// failure, so uploading directly is what surfaces a real error to skip on.
	serviceManager, err := artUtils.CreateUploadServiceManager(serverDetails, 1, 0, 0, false, nil)
	require.NoError(t, err)
	keyPairCmd := generate.NewGenerateKeyPairCommand(serverDetails, true, keyAlias, "", "")
	if _, err := keyPairCmd.UploadTrustedKey(&serviceManager, keyAlias, publicKeyPEM); err != nil {
		t.Skipf("skipping: could not upload public key to trusted keys (evidence service may not be configured): %v", err)
	}

	slug := "signed-plugin"
	version := "1.0.0"
	pluginPath := createTestPlugin(t, slug, version)

	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--signing-key="+privateKeyPath,
		"--key-alias="+keyAlias,
	), "publish with --signing-key must succeed")

	assertPluginExists(t, slug, version)
}

// TestAgentPluginsPublishWithoutSigningKey confirms that omitting --signing-key
// and clearing key env vars still results in a successful publish (evidence is
// skipped with an info log, not a failure).
func TestAgentPluginsPublishWithoutSigningKey(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	// Ensure no signing key is picked up from the environment.
	t.Setenv("EVD_SIGNING_KEY_PATH", "")
	t.Setenv("JFROG_CLI_SIGNING_KEY", "")

	slug := "no-signing-plugin"
	version := "1.0.0"
	pluginPath := createTestPlugin(t, slug, version)

	assert.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	), "publish without signing key must succeed; evidence should be silently skipped")

	assertPluginExists(t, slug, version)
}

// ---------------------------------------------------------------------------
// Install
// ---------------------------------------------------------------------------

// TestAgentPluginsInstallLatest verifies that installing a plugin without
// --version picks up the latest published version and places files at
// <installPath>/<slug>/. Uses --path to bypass harness resolution.
func TestAgentPluginsInstallLatest(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "install-latest-plugin"
	version := "1.0.0"
	pluginPath := createTestPlugin(t, slug, version)

	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	))

	installDir := t.TempDir()
	assert.NoError(t, runAgentPluginsCmd(t,
		"install", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--path="+installDir,
	))

	// Plugin files should be at <installDir>/<slug>/
	assert.FileExists(t, filepath.Join(installDir, slug, "plugin.json"),
		"plugin.json should exist after install")
}

// TestAgentPluginsInstallSpecificVersion verifies that --version installs the
// requested version rather than latest.
func TestAgentPluginsInstallSpecificVersion(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "install-version-plugin"

	// Publish two versions.
	v1Path := createTestPlugin(t, slug, "1.0.0")
	require.NoError(t, runAgentPluginsCmd(t, "publish", v1Path, "--repo="+tests.AgentPluginsLocalRepo))
	v2Path := createTestPlugin(t, slug, "2.0.0")
	require.NoError(t, runAgentPluginsCmd(t, "publish", v2Path, "--repo="+tests.AgentPluginsLocalRepo))

	installDir := t.TempDir()
	assert.NoError(t, runAgentPluginsCmd(t,
		"install", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--path="+installDir,
		"--version=1.0.0",
	))

	installedManifest := filepath.Join(installDir, slug, "plugin.json")
	require.FileExists(t, installedManifest)
	data, err := os.ReadFile(installedManifest) // #nosec G304 -- path from t.TempDir
	require.NoError(t, err)
	var manifest map[string]any
	require.NoError(t, json.Unmarshal(data, &manifest))
	assert.Equal(t, "1.0.0", manifest["version"], "installed version should be 1.0.0, not latest 2.0.0")
}

// TestAgentPluginsInstallNotFound verifies that installing an unknown slug
// returns a clear not-found error.
func TestAgentPluginsInstallNotFound(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	installDir := t.TempDir()
	err := runAgentPluginsCmd(t,
		"install", "nonexistent-slug-xyzzy",
		"--repo="+tests.AgentPluginsLocalRepo,
		"--path="+installDir,
	)
	assertErrorContainsAll(t, err, "not found in repository")
}

// TestAgentPluginsInstallProjectScopeRejectedForBuiltIns verifies that built-in
// harnesses reject --project-dir (global-only) for each harness condition.
func TestAgentPluginsInstallProjectScopeRejectedForBuiltIns(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "project-dir-plugin"
	pluginPath := createTestHarnessPlugin(t, slug, "1.0.0", allAgentHarnesses)
	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	))

	projectDir := t.TempDir()
	for _, tc := range agentPluginHarnessCases() {
		t.Run(tc.name, func(t *testing.T) {
			err := runAgentPluginsCmd(t,
				"install", slug,
				"--repo="+tests.AgentPluginsLocalRepo,
				"--harness="+harnessFlag(tc.harnesses),
				"--project-dir="+projectDir,
			)
			require.Error(t, err, "built-in harnesses must reject project-scoped install")
			// Production message from RejectUnsupportedProjectScope (exact phrases).
			assertErrorContainsAll(t, err,
				"does not support project-scoped plugin installs",
				"Use --global instead",
			)
		})
	}
}

// TestAgentPluginsInstallGlobal verifies that --global installs the plugin into
// each built-in harness destination for all harness conditions.
// Version is resolved from the harness marketplace (no --version) after publish indexing.
func TestAgentPluginsInstallGlobal(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "global-install-plugin"
	pluginPath := createTestHarnessPlugin(t, slug, "1.0.0", allAgentHarnesses)
	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	))

	for _, tc := range agentPluginHarnessCases() {
		t.Run(tc.name, func(t *testing.T) {
			homeDir := setIsolatedHome(t)
			require.NoError(t, installViaMarketplaceWithRetry(t, slug, harnessFlag(tc.harnesses)))
			assertPluginsInstalledGlobally(t, homeDir, tc.harnesses, slug, "1.0.0")
			assertPluginsInstalledNatively(t, tc.harnesses, slug, "1.0.0")
		})
	}
}

// TestAgentPluginsInstallMarketplace verifies that a published multi-harness plugin can be
// installed without --version. When --version is omitted, install downloads each harness's
// <harness>-marketplace.json from the repo root (generated by Artifactory indexing), resolves
// the version, then discards the temp download. There is no slug@marketplace CLI syntax —
// ValidateSlug rejects '@'. Retrying covers Artifactory's async marketplace index generation.
func TestAgentPluginsInstallMarketplace(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "marketplace-plugin"
	pluginPath := createTestHarnessPlugin(t, slug, "1.0.0", allAgentHarnesses)
	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	))

	for _, tc := range agentPluginHarnessCases() {
		t.Run(tc.name, func(t *testing.T) {
			homeDir := setIsolatedHome(t)
			require.NoError(t, installViaMarketplaceWithRetry(t, slug, harnessFlag(tc.harnesses)),
				"install without --version should resolve through the generated marketplace")
			assertPluginsInstalledGlobally(t, homeDir, tc.harnesses, slug, "1.0.0")
			assertPluginsInstalledNatively(t, tc.harnesses, slug, "1.0.0")
		})
	}
}

// TestAgentPluginsInstallMarketplaceSlugNotListed verifies install without --version fails with
// InstallBypassMarketplaceHint when the slug is absent from the harness marketplace index.
func TestAgentPluginsInstallMarketplaceSlugNotListed(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	// Ensure at least one marketplace.json exists by publishing a different plugin first.
	seed := createTestHarnessPlugin(t, "marketplace-seed-plugin", "1.0.0", []string{"cursor"})
	require.NoError(t, runAgentPluginsCmd(t, "publish", seed, "--repo="+tests.AgentPluginsLocalRepo))
	assertMarketplaceContainsPlugin(t, "cursor", "marketplace-seed-plugin", "1.0.0")

	// Isolate HOME for global install; returned path is unused in this assertion.
	_ = setIsolatedHome(t)
	err := runAgentPluginsCmd(t,
		"install", "slug-absent-from-marketplace",
		"--repo="+tests.AgentPluginsLocalRepo,
		"--harness=cursor",
		"--global",
	)
	assertErrorContainsAll(t, err,
		"plugin 'slug-absent-from-marketplace' is not listed in cursor-marketplace.json",
		"--version",
		"without using the marketplace",
	)
}

// installViaMarketplaceWithRetry retries `install` without --version to accommodate Artifactory
// generating the per-harness marketplace.json index asynchronously after publish. Mirrors the same
// wait-for-async-indexing pattern used for terraform's module.json in verifyModuleInArtifactoryWithRetry.
func installViaMarketplaceWithRetry(t *testing.T, slug, harnesses string) error {
	t.Helper()
	return retryWithBackoff(t, "install "+slug+" via the "+harnesses+" marketplace index", func() error {
		return runAgentPluginsCmd(t,
			"install", slug,
			"--repo="+tests.AgentPluginsLocalRepo,
			"--harness="+harnesses,
			"--global",
		)
	})
}

// TestAgentPluginsInstallAgentConfigOverride verifies that a custom agent entry
// defined in agent-config.json under "plugins-agents" is respected for both
// --global (globalDir) and --project-dir (projectDir) installs.
// Full install → list → update coverage for a custom agent is in
// TestAgentPluginsCustomAgentLifecycle.
func TestAgentPluginsInstallAgentConfigOverride(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "agent-config-plugin"
	pluginPath := createTestPlugin(t, slug, "1.0.0")
	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	))

	jfrogHome := os.Getenv("JFROG_CLI_HOME_DIR")

	customGlobalDir := t.TempDir()
	agentTestutil.WriteAgentConfig(t, jfrogHome, `{
		"plugins-agents": {
			"my-custom-agent": {
				"globalDir": "`+filepath.ToSlash(customGlobalDir)+`",
				"projectDir": ".my-custom-agent/plugins"
			}
		}
	}`)

	// Custom agents are not indexed into <harness>-marketplace.json, so --version is
	// required (marketplace resolution only applies to built-in harness indexes).
	require.NoError(t, runAgentPluginsCmd(t,
		"install", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--harness=my-custom-agent",
		"--global",
		"--version=1.0.0",
	))
	assert.DirExists(t, filepath.Join(customGlobalDir, slug),
		"plugin should be installed into the globalDir from agent-config.json")

	// Verify projectDir override with explicit --project-dir.
	projectDir := t.TempDir()
	require.NoError(t, runAgentPluginsCmd(t,
		"install", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--harness=my-custom-agent",
		"--project-dir="+projectDir,
		"--version=1.0.0",
	))
	assert.DirExists(t, filepath.Join(projectDir, ".my-custom-agent", "plugins", slug),
		"plugin should be installed into projectDir/.my-custom-agent/plugins/<slug> from agent-config.json")

	// When neither --global nor --project-dir is set, install defaults to global scope
	// and still honors globalDir from agent-config.json.
	defaultScopeSlug := "agent-config-default-scope-plugin"
	require.NoError(t, runAgentPluginsCmd(t,
		"publish", createTestPlugin(t, defaultScopeSlug, "1.0.0"),
		"--repo="+tests.AgentPluginsLocalRepo,
	))
	require.NoError(t, runAgentPluginsCmd(t,
		"install", defaultScopeSlug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--harness=my-custom-agent",
		"--version=1.0.0",
	))
	assert.DirExists(t, filepath.Join(customGlobalDir, defaultScopeSlug),
		"omitting --global/--project-dir must default to globalDir from agent-config.json")
	assert.NoDirExists(t, filepath.Join(projectDir, ".my-custom-agent", "plugins", defaultScopeSlug),
		"default-scope install must not land under projectDir")
}

// TestAgentPluginsCustomAgentLifecycle is the custom-agent counterpart to the
// built-in agentPluginHarnessCases matrix. It registers "my-agent" in
// agent-config.json and exercises install, list, and update end-to-end:
//
//	install (pinned older version) → list → list --check-updates (behind)
//	→ update → list → list --check-updates (current)
func TestAgentPluginsCustomAgentLifecycle(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	const (
		agentName  = "my-agent"
		slug       = "my-agent-lifecycle-plugin"
		oldVersion = "1.0.0"
		newVersion = "2.0.0"
	)

	require.NoError(t, runAgentPluginsCmd(t, "publish", createTestPlugin(t, slug, oldVersion),
		"--repo="+tests.AgentPluginsLocalRepo))
	require.NoError(t, runAgentPluginsCmd(t, "publish", createTestPlugin(t, slug, newVersion),
		"--repo="+tests.AgentPluginsLocalRepo))

	customGlobalDir := t.TempDir()
	agentTestutil.WriteAgentConfig(t, os.Getenv("JFROG_CLI_HOME_DIR"), `{
		"plugins-agents": {
			"my-agent": {
				"globalDir": "`+filepath.ToSlash(customGlobalDir)+`",
				"projectDir": ".my-agent/plugins"
			}
		}
	}`)

	harnesses := []string{agentName}

	// --- install ---
	// Pin oldVersion: v2 is already published, and custom agents have no marketplace index.
	require.NoError(t, runAgentPluginsCmd(t,
		"install", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--harness="+agentName,
		"--global",
		"--version="+oldVersion,
	), "custom agent install must succeed")
	assertCustomAgentPluginInstalled(t, customGlobalDir, agentName, slug, oldVersion)

	// --- list (after install) ---
	listOut, err := runAgentPluginsCmdWithOutput(t,
		"list",
		"--harness="+agentName,
		"--global",
		"--format=json",
	)
	require.NoError(t, err, "custom agent list after install must succeed")
	assertListContainsInstalledPlugin(t, listOut, harnesses, slug, oldVersion)

	// --- list --check-updates (behind) ---
	checkOut, err := runAgentPluginsCmdWithOutput(t,
		"list",
		"--harness="+agentName,
		"--global",
		"--check-updates",
		"--format=json",
	)
	require.NoError(t, err, "custom agent list --check-updates must succeed")
	assertListContainsPluginStatus(t, checkOut, harnesses, slug, "behind", newVersion)

	// --- update ---
	require.NoError(t, runAgentPluginsCmd(t,
		"update",
		"--slug="+slug,
		"--harness="+agentName,
		"--global",
		"--repo="+tests.AgentPluginsLocalRepo,
	), "custom agent update must upgrade to latest")
	assertCustomAgentPluginInstalled(t, customGlobalDir, agentName, slug, newVersion)

	// --- list (after update) ---
	listOut, err = runAgentPluginsCmdWithOutput(t,
		"list",
		"--harness="+agentName,
		"--global",
		"--format=json",
	)
	require.NoError(t, err, "custom agent list after update must succeed")
	assertListContainsInstalledPlugin(t, listOut, harnesses, slug, newVersion)

	// --- list --check-updates (current) ---
	checkOut, err = runAgentPluginsCmdWithOutput(t,
		"list",
		"--harness="+agentName,
		"--global",
		"--check-updates",
		"--format=json",
	)
	require.NoError(t, err, "custom agent list --check-updates after update must succeed")
	assertListContainsPluginStatus(t, checkOut, harnesses, slug, "current", "")
	assertListContainsInstalledPlugin(t, checkOut, harnesses, slug, newVersion)
}

// assertCustomAgentPluginInstalled checks install dir + plugin-info.json for a
// custom agent whose globalDir comes from agent-config.json (not a built-in layout).
func assertCustomAgentPluginInstalled(t *testing.T, globalDir, agentName, slug, version string) {
	t.Helper()
	pluginDir := filepath.Join(globalDir, slug)
	assert.DirExists(t, pluginDir, "plugin %q should be installed under custom agent globalDir", slug)
	manifestPath := filepath.Join(pluginDir, ".jfrog", "plugin-info.json")
	require.FileExists(t, manifestPath)
	data, err := os.ReadFile(manifestPath) // #nosec G304 -- path under t.TempDir
	require.NoError(t, err)
	var manifest map[string]any
	require.NoError(t, json.Unmarshal(data, &manifest))
	assert.Equal(t, version, manifest["installedVersion"])
	assert.Equal(t, slug, manifest["slug"])
	assert.Equal(t, agentName, manifest["agent"])
	assert.Equal(t, tests.AgentPluginsLocalRepo, manifest["repo"])
}

// TestAgentPluginsInstallMissingSlugArg verifies that omitting the required
// <slug> argument returns a clear usage error.
func TestAgentPluginsInstallMissingSlugArg(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	err := runAgentPluginsCmd(t, "install",
		"--repo="+tests.AgentPluginsLocalRepo,
		"--path="+t.TempDir(),
	)
	assertErrorContainsAll(t, err, "usage: jf agent plugins install")
}

// TestAgentPluginsInstallUnknownHarness verifies that specifying an unknown
// harness name returns a clear error.
func TestAgentPluginsInstallUnknownHarness(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "unknown-harness-plugin"
	pluginPath := createTestPlugin(t, slug, "1.0.0")

	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	))

	err := runAgentPluginsCmd(t,
		"install", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--harness=totally-unknown-harness-xyz",
		"--global",
	)
	assertErrorContainsAll(t, err, `unknown agent "totally-unknown-harness-xyz"`)
}

// TestAgentPluginsInstallEmptyHarness verifies that --harness with an empty
// or duplicate name is rejected.
func TestAgentPluginsInstallEmptyHarness(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	err := runAgentPluginsCmd(t,
		"install", "some-plugin",
		"--repo="+tests.AgentPluginsLocalRepo,
		"--harness=",
		"--global",
	)
	assertErrorContainsAll(t, err, "--harness is required unless --path is set")
}

// TestAgentPluginsInstallGlobalProjectDirMutuallyExclusive verifies that passing
// both --global and --project-dir to install returns a clear error.
func TestAgentPluginsInstallGlobalProjectDirMutuallyExclusive(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	err := runAgentPluginsCmd(t,
		"install", "some-plugin",
		"--repo="+tests.AgentPluginsLocalRepo,
		"--harness=claude",
		"--global",
		"--project-dir="+t.TempDir(),
	)
	assertErrorContainsAll(t, err,
		"--global and --project-dir are mutually exclusive",
		"please choose either --global or --project-dir",
	)
}

// TestAgentPluginsInstallHarnessPathMutuallyExclusive verifies that passing
// both --harness and --path to install returns a clear error (they are
// mutually exclusive install target modes).
func TestAgentPluginsInstallHarnessPathMutuallyExclusive(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	err := runAgentPluginsCmd(t,
		"install", "some-plugin",
		"--repo="+tests.AgentPluginsLocalRepo,
		"--harness=claude",
		"--path="+t.TempDir(),
	)
	assertErrorContainsAll(t, err, "--path cannot be combined with --harness")
}

// TestAgentPluginsInstallWritesPluginInfoManifest verifies that after a
// successful install, plugin-info.json is written with the correct slug and
// installed version.
func TestAgentPluginsInstallWritesPluginInfoManifest(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "manifest-plugin"
	version := "1.0.0"
	pluginPath := createTestPlugin(t, slug, version)

	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	))

	installDir := t.TempDir()
	require.NoError(t, runAgentPluginsCmd(t,
		"install", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--path="+installDir,
	))

	// plugin-info.json lives under .jfrog/ inside the install destination dir.
	manifestPath := filepath.Join(installDir, slug, ".jfrog", "plugin-info.json")
	require.FileExists(t, manifestPath, "plugin-info.json should be written after install")

	data, err := os.ReadFile(manifestPath) // #nosec G304 -- path from t.TempDir
	require.NoError(t, err)

	var manifest map[string]any
	require.NoError(t, json.Unmarshal(data, &manifest))
	assert.Equal(t, version, manifest["installedVersion"],
		"plugin-info.json installedVersion should match the published version")
	assert.Equal(t, slug, manifest["slug"],
		"plugin-info.json slug should match the installed plugin")
}

// TestAgentPluginsInstallEvidenceGateCI verifies that installing in CI/quiet
// mode when evidence is absent fails with a hint about
// JFROG_AGENT_PLUGINS_DISABLE_QUIET_FAILURE.
func TestAgentPluginsInstallEvidenceGateCI(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	// Publish a plugin without a signing key so no evidence is attached.
	slug := "evidence-gate-plugin"
	pluginPath := createTestPlugin(t, slug, "1.0.0")
	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	))

	// Enable quiet mode and ensure the evidence-disable env var is NOT set.
	t.Setenv("CI", "true")
	t.Setenv("JFROG_AGENT_PLUGINS_DISABLE_QUIET_FAILURE", "")

	installDir := t.TempDir()
	err := runAgentPluginsCmd(t,
		"install", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--path="+installDir,
		"--quiet",
	)
	// The command may succeed or fail depending on whether evidence enforcement
	// is active (Enterprise+). If it fails, the error must reference the disable env var.
	if err != nil {
		assertErrorContainsAll(t, err,
			"evidence verification failed",
			"JFROG_AGENT_PLUGINS_DISABLE_QUIET_FAILURE",
		)
	} else {
		t.Log("evidence gate not enforced on this Artifactory instance; failure path not exercised")
	}
}

// TestAgentPluginsInstallEvidenceGateDisabled verifies that setting
// JFROG_AGENT_PLUGINS_DISABLE_QUIET_FAILURE=true allows install in CI mode
// to succeed even without evidence.
func TestAgentPluginsInstallEvidenceGateDisabled(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "evidence-disabled-plugin"
	pluginPath := createTestPlugin(t, slug, "1.0.0")
	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	))

	t.Setenv("CI", "true")
	t.Setenv("JFROG_AGENT_PLUGINS_DISABLE_QUIET_FAILURE", "true")

	installDir := t.TempDir()
	assert.NoError(t, runAgentPluginsCmd(t,
		"install", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--path="+installDir,
		"--quiet",
	), "install should succeed in CI mode when JFROG_AGENT_PLUGINS_DISABLE_QUIET_FAILURE=true")
}

// TestAgentPluginsUpdateAllNothingInstalled verifies that update --all succeeds
// when no plugins are installed for each harness condition.
func TestAgentPluginsUpdateAllNothingInstalled(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	for _, tc := range agentPluginHarnessCases() {
		t.Run(tc.name, func(t *testing.T) {
			// Isolate HOME so --global update does not scan the real user profile.
			_ = setIsolatedHome(t)
			assert.NoError(t, runAgentPluginsCmd(t,
				"update",
				"--all",
				"--harness="+harnessFlag(tc.harnesses),
				"--global",
				"--repo="+tests.AgentPluginsLocalRepo,
				"--quiet",
			), "update --all with no installed plugins should succeed without error")
		})
	}
}

// TestAgentPluginsInstallWithPath publishes a plugin then installs it using
// --path <dir>, which writes the plugin directly to <dir>/<slug> without any
// harness lookup.
func TestAgentPluginsInstallWithPath(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "path-mode-plugin"
	version := "1.0.0"
	pluginPath := createTestPlugin(t, slug, version)
	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	))

	installBase := t.TempDir()
	require.NoError(t, runAgentPluginsCmd(t,
		"install", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--path="+installBase,
	), "install --path should bypass harness and write to the given directory")

	expectedPluginDir := filepath.Join(installBase, slug)
	info, err := os.Stat(expectedPluginDir)
	require.NoError(t, err, "plugin directory must exist under --path target")
	assert.True(t, info.IsDir(), "install --path target must be a directory")
}

// TestAgentPluginsInstallPathWithVersion verifies that --path combined with
// --version installs the exact requested version into the given directory.
func TestAgentPluginsInstallPathWithVersion(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "path-version-plugin"
	v1Path := createTestPlugin(t, slug, "1.0.0")
	require.NoError(t, runAgentPluginsCmd(t, "publish", v1Path, "--repo="+tests.AgentPluginsLocalRepo))
	v2Path := createTestPlugin(t, slug, "2.0.0")
	require.NoError(t, runAgentPluginsCmd(t, "publish", v2Path, "--repo="+tests.AgentPluginsLocalRepo))

	installBase := t.TempDir()
	require.NoError(t, runAgentPluginsCmd(t,
		"install", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--path="+installBase,
		"--version=1.0.0",
	), "install --path --version should install the specific version")

	manifestPath := filepath.Join(installBase, slug, ".jfrog", "plugin-info.json")
	require.FileExists(t, manifestPath)
	data, err := os.ReadFile(manifestPath) // #nosec G304 -- path from t.TempDir
	require.NoError(t, err)
	var manifest map[string]any
	require.NoError(t, json.Unmarshal(data, &manifest))
	assert.Equal(t, "1.0.0", manifest["installedVersion"],
		"--path --version=1.0.0 should install v1 even though v2 exists")
}

// TestAgentPluginsInstallFormatJSON verifies that install with --format json
// produces parseable JSON output rather than a human-readable table.
func TestAgentPluginsInstallFormatJSON(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "format-json-install-plugin"
	version := "1.0.0"
	pluginPath := createTestPlugin(t, slug, version)
	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	))

	installBase := t.TempDir()
	out, err := runAgentPluginsCmdWithOutput(t,
		"install", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--path="+installBase,
		"--format=json",
	)
	require.NoError(t, err, "install --format json should succeed without error")
	assertInstallSummaryJSON(t, out, slug, version)
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

// TestAgentPluginsUpdateSlug verifies that `update --slug` installs a newer
// version when one is available in the repository.
func TestAgentPluginsUpdateSlug(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "update-slug-plugin"
	oldVersion := "1.0.0"
	newVersion := "2.0.0"

	// Publish both versions.
	v1Path := createTestPlugin(t, slug, oldVersion)
	require.NoError(t, runAgentPluginsCmd(t, "publish", v1Path, "--repo="+tests.AgentPluginsLocalRepo))
	v2Path := createTestPlugin(t, slug, newVersion)
	require.NoError(t, runAgentPluginsCmd(t, "publish", v2Path, "--repo="+tests.AgentPluginsLocalRepo))

	// Install v1 first.
	installDir := t.TempDir()
	require.NoError(t, runAgentPluginsCmd(t,
		"install", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--path="+installDir,
		"--version="+oldVersion,
	))

	// Run update — should upgrade to v2.
	assert.NoError(t, runAgentPluginsCmd(t,
		"update",
		"--slug="+slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--path="+installDir,
	))

	// Verify the installed version changed to the latest.
	manifestPath := filepath.Join(installDir, slug, ".jfrog", "plugin-info.json")
	require.FileExists(t, manifestPath)
	data, err := os.ReadFile(manifestPath) // #nosec G304 -- path from t.TempDir
	require.NoError(t, err)
	var manifest map[string]any
	require.NoError(t, json.Unmarshal(data, &manifest))
	assert.Equal(t, newVersion, manifest["installedVersion"],
		"update should upgrade installed version from %s to %s", oldVersion, newVersion)
}

// TestAgentPluginsUpdateDryRun verifies that --dry-run reports the plan without
// changing any files on the filesystem.
func TestAgentPluginsUpdateDryRun(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "update-dryrun-plugin"
	oldVersion := "1.0.0"
	newVersion := "2.0.0"

	v1Path := createTestPlugin(t, slug, oldVersion)
	require.NoError(t, runAgentPluginsCmd(t, "publish", v1Path, "--repo="+tests.AgentPluginsLocalRepo))
	v2Path := createTestPlugin(t, slug, newVersion)
	require.NoError(t, runAgentPluginsCmd(t, "publish", v2Path, "--repo="+tests.AgentPluginsLocalRepo))

	installDir := t.TempDir()
	require.NoError(t, runAgentPluginsCmd(t,
		"install", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--path="+installDir,
		"--version="+oldVersion,
	))

	assert.NoError(t, runAgentPluginsCmd(t,
		"update",
		"--slug="+slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--path="+installDir,
		"--dry-run",
	))

	// Filesystem must be unchanged after dry-run: plugin-info.json should still
	// report the old version.
	manifestPath := filepath.Join(installDir, slug, ".jfrog", "plugin-info.json")
	require.FileExists(t, manifestPath)
	data, err := os.ReadFile(manifestPath) // #nosec G304 -- path from t.TempDir
	require.NoError(t, err)
	var manifest map[string]any
	require.NoError(t, json.Unmarshal(data, &manifest))
	assert.Equal(t, oldVersion, manifest["installedVersion"],
		"--dry-run must not change installed version on disk")
}

// TestAgentPluginsUpdateForce verifies that --force overwrites an already
// up-to-date install without reporting it as skipped.
func TestAgentPluginsUpdateForce(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "update-force-plugin"
	version := "1.0.0"

	pluginPath := createTestPlugin(t, slug, version)
	require.NoError(t, runAgentPluginsCmd(t, "publish", pluginPath, "--repo="+tests.AgentPluginsLocalRepo))

	installDir := t.TempDir()
	require.NoError(t, runAgentPluginsCmd(t,
		"install", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--path="+installDir,
	))

	// Update with --force: already at latest but --force should still re-install cleanly.
	require.NoError(t, runAgentPluginsCmd(t,
		"update",
		"--slug="+slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--path="+installDir,
		"--force",
	), "--force should succeed even when plugin is already at the latest version")

	manifestPath := filepath.Join(installDir, slug, ".jfrog", "plugin-info.json")
	require.FileExists(t, manifestPath)
	data, err := os.ReadFile(manifestPath) // #nosec G304 -- path from t.TempDir
	require.NoError(t, err)
	var manifest map[string]any
	require.NoError(t, json.Unmarshal(data, &manifest))
	assert.Equal(t, version, manifest["installedVersion"],
		"--force should leave the plugin installed at the same version")
}

// TestAgentPluginsUpdateAll verifies that `update --all` discovers and updates
// every installed plugin under each harness condition.
func TestAgentPluginsUpdateAll(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slugA := "update-all-plugin-a"
	slugB := "update-all-plugin-b"

	for _, entry := range []struct{ slug, oldVer, newVer string }{
		{slugA, "1.0.0", "2.0.0"},
		{slugB, "1.0.0", "2.0.0"},
	} {
		// Real codex requires the manifest under .codex-plugin/plugin.json (mirroring
		// claude's .claude-plugin/ convention); a flat root plugin.json fails
		// "codex plugin add" with "missing plugin.json". Use the harness-aware fixture
		// since these cases install with harness=codex (via agentPluginHarnessCases()).
		v1Path := createTestHarnessPlugin(t, entry.slug, entry.oldVer, allAgentHarnesses)
		require.NoError(t, runAgentPluginsCmd(t, "publish", v1Path, "--repo="+tests.AgentPluginsLocalRepo))
		v2Path := createTestHarnessPlugin(t, entry.slug, entry.newVer, allAgentHarnesses)
		require.NoError(t, runAgentPluginsCmd(t, "publish", v2Path, "--repo="+tests.AgentPluginsLocalRepo))
	}

	for _, tc := range agentPluginHarnessCases() {
		t.Run(tc.name, func(t *testing.T) {
			homeDir := setIsolatedHome(t)
			for _, slug := range []string{slugA, slugB} {
				require.NoError(t, runAgentPluginsCmd(t,
					"install", slug,
					"--repo="+tests.AgentPluginsLocalRepo,
					"--harness="+harnessFlag(tc.harnesses),
					"--global",
					"--version=1.0.0",
				))
			}

			require.NoError(t, runAgentPluginsCmd(t,
				"update",
				"--all",
				"--harness="+harnessFlag(tc.harnesses),
				"--global",
				"--repo="+tests.AgentPluginsLocalRepo,
				"--quiet",
			))

			for _, slug := range []string{slugA, slugB} {
				for _, harness := range tc.harnesses {
					manifestPath := filepath.Join(globalPluginInstallDir(homeDir, harness, tests.AgentPluginsLocalRepo, slug), ".jfrog", "plugin-info.json")
					require.FileExists(t, manifestPath, "plugin-info.json should exist for %s/%s after update --all", harness, slug)
					data, err := os.ReadFile(manifestPath) // #nosec G304 -- path from t.TempDir
					require.NoError(t, err)
					var manifest map[string]any
					require.NoError(t, json.Unmarshal(data, &manifest))
					assert.Equal(t, "2.0.0", manifest["installedVersion"],
						"update --all should upgrade %s/%s from 1.0.0 to 2.0.0", harness, slug)
				}
			}
		})
	}
}

// TestAgentPluginsUpdateAllNonInteractive verifies that `update --all` without
// --quiet proceeds automatically when CI=true for each harness condition.
func TestAgentPluginsUpdateAllNonInteractive(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "update-all-ci-plugin"
	// Real codex requires the manifest under .codex-plugin/plugin.json (mirroring
	// claude's .claude-plugin/ convention); a flat root plugin.json fails
	// "codex plugin add" with "missing plugin.json". Use the harness-aware fixture
	// since these cases install with harness=codex (via agentPluginHarnessCases()).
	v1Path := createTestHarnessPlugin(t, slug, "1.0.0", allAgentHarnesses)
	require.NoError(t, runAgentPluginsCmd(t, "publish", v1Path, "--repo="+tests.AgentPluginsLocalRepo))
	v2Path := createTestHarnessPlugin(t, slug, "2.0.0", allAgentHarnesses)
	require.NoError(t, runAgentPluginsCmd(t, "publish", v2Path, "--repo="+tests.AgentPluginsLocalRepo))

	for _, tc := range agentPluginHarnessCases() {
		t.Run(tc.name, func(t *testing.T) {
			homeDir := setIsolatedHome(t)
			require.NoError(t, runAgentPluginsCmd(t,
				"install", slug,
				"--repo="+tests.AgentPluginsLocalRepo,
				"--harness="+harnessFlag(tc.harnesses),
				"--global",
				"--version=1.0.0",
			))

			t.Setenv("CI", "true")
			require.NoError(t, runAgentPluginsCmd(t,
				"update",
				"--all",
				"--harness="+harnessFlag(tc.harnesses),
				"--global",
				"--repo="+tests.AgentPluginsLocalRepo,
			), "update --all should proceed without --quiet when CI=true")

			for _, harness := range tc.harnesses {
				manifestPath := filepath.Join(globalPluginInstallDir(homeDir, harness, tests.AgentPluginsLocalRepo, slug), ".jfrog", "plugin-info.json")
				require.FileExists(t, manifestPath)
				data, err := os.ReadFile(manifestPath) // #nosec G304 -- path from t.TempDir
				require.NoError(t, err)
				var manifest map[string]any
				require.NoError(t, json.Unmarshal(data, &manifest))
				assert.Equal(t, "2.0.0", manifest["installedVersion"],
					"update --all with CI=true should upgrade %s to 2.0.0", harness)
			}
		})
	}
}

// TestAgentPluginsUpdateFormatJSON verifies that `update --slug --format=json`
// and `update --all --format=json` succeed for each harness condition.
func TestAgentPluginsUpdateFormatJSON(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "update-format-json-plugin"
	// Real codex requires the manifest under .codex-plugin/plugin.json (mirroring
	// claude's .claude-plugin/ convention); a flat root plugin.json fails
	// "codex plugin add" with "missing plugin.json". Use the harness-aware fixture
	// since these cases install with harness=codex (via agentPluginHarnessCases()).
	v1Path := createTestHarnessPlugin(t, slug, "1.0.0", allAgentHarnesses)
	require.NoError(t, runAgentPluginsCmd(t, "publish", v1Path, "--repo="+tests.AgentPluginsLocalRepo))
	v2Path := createTestHarnessPlugin(t, slug, "2.0.0", allAgentHarnesses)
	require.NoError(t, runAgentPluginsCmd(t, "publish", v2Path, "--repo="+tests.AgentPluginsLocalRepo))

	for _, tc := range agentPluginHarnessCases() {
		t.Run(tc.name, func(t *testing.T) {
			homeDir := setIsolatedHome(t)
			require.NoError(t, runAgentPluginsCmd(t,
				"install", slug,
				"--repo="+tests.AgentPluginsLocalRepo,
				"--harness="+harnessFlag(tc.harnesses),
				"--global",
				"--version=1.0.0",
			))

			out, err := runAgentPluginsCmdWithOutput(t,
				"update",
				"--slug="+slug,
				"--harness="+harnessFlag(tc.harnesses),
				"--global",
				"--repo="+tests.AgentPluginsLocalRepo,
				"--format=json",
			)
			require.NoError(t, err, "update --slug --format=json should succeed")
			assertInstallSummaryJSON(t, out, slug, "2.0.0")
			assertPluginsInstalledGlobally(t, homeDir, tc.harnesses, slug, "2.0.0")
			assertPluginsInstalledNatively(t, tc.harnesses, slug, "2.0.0")

			require.NoError(t, runAgentPluginsCmd(t,
				"install", slug,
				"--repo="+tests.AgentPluginsLocalRepo,
				"--harness="+harnessFlag(tc.harnesses),
				"--global",
				"--version=1.0.0",
				"--force",
			))

			out, err = runAgentPluginsCmdWithOutput(t,
				"update",
				"--all",
				"--harness="+harnessFlag(tc.harnesses),
				"--global",
				"--repo="+tests.AgentPluginsLocalRepo,
				"--quiet",
				"--format=json",
			)
			require.NoError(t, err, "update --all --format=json should succeed")
			assertUpdateAllSummaryJSONContains(t, out, slug)
			assertPluginsInstalledGlobally(t, homeDir, tc.harnesses, slug, "2.0.0")
			assertPluginsInstalledNatively(t, tc.harnesses, slug, "2.0.0")
		})
	}
}

// TestAgentPluginsUpdateProjectScopeRejected verifies built-in agents reject
// project-scoped update the same way install/list do.
func TestAgentPluginsUpdateProjectScopeRejected(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	projectDir := t.TempDir()
	for _, tc := range agentPluginHarnessCases() {
		t.Run(tc.name, func(t *testing.T) {
			err := runAgentPluginsCmd(t,
				"update",
				"--slug=any-plugin",
				"--harness="+harnessFlag(tc.harnesses),
				"--project-dir="+projectDir,
				"--repo="+tests.AgentPluginsLocalRepo,
			)
			require.Error(t, err, "built-in harnesses must reject project-scoped update")
			assertErrorContainsAll(t, err,
				"does not support project-scoped plugin updates",
				"Use --global instead",
			)
		})
	}
}

// TestAgentPluginsUpdateFlags exercises the mutually exclusive and required
// flag combinations for the update subcommand.
func TestAgentPluginsUpdateFlags(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	projectDir := t.TempDir()

	cases := []struct {
		name        string
		args        []string
		expectError bool
		errContains []string
		description string
	}{
		{
			name:        "no-slug-no-all",
			args:        []string{"update", "--repo=" + tests.AgentPluginsLocalRepo, "--path=" + projectDir},
			expectError: true,
			errContains: []string{"usage: jf agent plugins update"},
			description: "update without --slug or --all should fail",
		},
		{
			name:        "all-with-slug",
			args:        []string{"update", "--all", "--slug=some-plugin", "--repo=" + tests.AgentPluginsLocalRepo, "--harness=claude", "--global", "--quiet"},
			expectError: true,
			errContains: []string{"--all cannot be combined with --slug"},
			description: "--all and --slug are mutually exclusive",
		},
		{
			name:        "all-with-version",
			args:        []string{"update", "--all", "--version=1.0.0", "--repo=" + tests.AgentPluginsLocalRepo, "--harness=claude", "--global", "--quiet"},
			expectError: true,
			errContains: []string{"--all cannot be combined with --version"},
			description: "--all and --version are mutually exclusive",
		},
		{
			name:        "invalid-slug-format",
			args:        []string{"update", "--slug=Invalid Slug!", "--repo=" + tests.AgentPluginsLocalRepo, "--path=" + projectDir},
			expectError: true,
			errContains: []string{"invalid slug"},
			description: "--slug with invalid format should be rejected",
		},
		{
			name:        "plugin-not-installed",
			args:        []string{"update", "--slug=notinstalled-xyz-abc", "--repo=" + tests.AgentPluginsLocalRepo, "--path=" + projectDir},
			expectError: true,
			errContains: []string{"plugin 'notinstalled-xyz-abc' not found in repository"},
			description: "update of a plugin that was never installed should fail",
		},
		{
			name:        "all-with-path",
			args:        []string{"update", "--all", "--path=" + projectDir, "--repo=" + tests.AgentPluginsLocalRepo, "--quiet"},
			expectError: true,
			errContains: []string{"--all cannot be combined with --path"},
			description: "--all and --path are mutually exclusive",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := runAgentPluginsCmd(t, tc.args...)
			if tc.expectError {
				assertErrorContainsAll(t, err, tc.errContains...)
			} else {
				assert.NoError(t, err, tc.description)
			}
		})
	}
}

// TestAgentPluginsListCheckUpdates installs a plugin then runs list
// --check-updates for each harness condition and verifies JSON.
func TestAgentPluginsListCheckUpdates(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "check-updates-plugin"
	version := "1.0.0"
	pluginPath := createTestHarnessPlugin(t, slug, version, allAgentHarnesses)
	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	))

	for _, tc := range agentPluginHarnessCases() {
		t.Run(tc.name, func(t *testing.T) {
			homeDir := setIsolatedHome(t)
			require.NoError(t, installViaMarketplaceWithRetry(t, slug, harnessFlag(tc.harnesses)))
			assertPluginsInstalledGlobally(t, homeDir, tc.harnesses, slug, version)

			out, err := runAgentPluginsCmdWithOutput(t,
				"list",
				"--harness="+harnessFlag(tc.harnesses),
				"--global",
				"--check-updates",
				"--format=json",
			)
			require.NoError(t, err, "list --check-updates --harness should run without error")
			assertListContainsInstalledPlugin(t, out, tc.harnesses, slug, version)
		})
	}
}

// TestAgentPluginsListCheckUpdatesStatus installs a plugin at v1 while v2 is
// available, then verifies list --check-updates reports status "behind".
func TestAgentPluginsListCheckUpdatesStatus(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "check-status-plugin"
	// Real codex requires the manifest under .codex-plugin/plugin.json (mirroring
	// claude's .claude-plugin/ convention); a flat root plugin.json fails
	// "codex plugin add" with "missing plugin.json". Use the harness-aware fixture
	// since these cases install with harness=codex (via agentPluginHarnessCases()).
	v1Path := createTestHarnessPlugin(t, slug, "1.0.0", allAgentHarnesses)
	require.NoError(t, runAgentPluginsCmd(t, "publish", v1Path, "--repo="+tests.AgentPluginsLocalRepo))
	v2Path := createTestHarnessPlugin(t, slug, "2.0.0", allAgentHarnesses)
	require.NoError(t, runAgentPluginsCmd(t, "publish", v2Path, "--repo="+tests.AgentPluginsLocalRepo))

	for _, tc := range agentPluginHarnessCases() {
		t.Run(tc.name, func(t *testing.T) {
			homeDir := setIsolatedHome(t)
			require.NoError(t, runAgentPluginsCmd(t,
				"install", slug,
				"--repo="+tests.AgentPluginsLocalRepo,
				"--harness="+harnessFlag(tc.harnesses),
				"--global",
				"--version=1.0.0",
			))
			assertPluginsInstalledGlobally(t, homeDir, tc.harnesses, slug)

			out, err := runAgentPluginsCmdWithOutput(t,
				"list",
				"--harness="+harnessFlag(tc.harnesses),
				"--global",
				"--check-updates",
				"--format=json",
			)
			require.NoError(t, err, "list --check-updates should succeed")
			assertListContainsPluginStatus(t, out, tc.harnesses, slug, "behind", "2.0.0")
		})
	}
}

// TestAgentPluginsListCheckUpdatesCurrent installs a plugin at the latest
// available version then verifies list --check-updates reports status "current".
func TestAgentPluginsListCheckUpdatesCurrent(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "check-current-plugin"
	version := "1.0.0"
	pluginPath := createTestHarnessPlugin(t, slug, version, allAgentHarnesses)
	require.NoError(t, runAgentPluginsCmd(t, "publish", pluginPath, "--repo="+tests.AgentPluginsLocalRepo))

	for _, tc := range agentPluginHarnessCases() {
		t.Run(tc.name, func(t *testing.T) {
			homeDir := setIsolatedHome(t)
			require.NoError(t, installViaMarketplaceWithRetry(t, slug, harnessFlag(tc.harnesses)))
			assertPluginsInstalledGlobally(t, homeDir, tc.harnesses, slug)

			out, err := runAgentPluginsCmdWithOutput(t,
				"list",
				"--harness="+harnessFlag(tc.harnesses),
				"--global",
				"--check-updates",
				"--format=json",
			)
			require.NoError(t, err, "list --check-updates should succeed")
			assertListContainsPluginStatus(t, out, tc.harnesses, slug, "current", "")
		})
	}
}

// assertListContainsPluginStatus validates list --format=json output for single or
// multi-harness responses (array vs map keyed by harness name).
func assertListContainsPluginStatus(t *testing.T, out string, harnesses []string, slug, status, registryLatest string) {
	t.Helper()
	if len(harnesses) > 1 {
		var byHarness map[string][]map[string]any
		require.NoError(t, json.Unmarshal([]byte(extractJSONObjectOrArray(t, out)), &byHarness),
			"multi-harness output must be a JSON object")
		require.NotEmpty(t, byHarness)
		for _, harness := range harnesses {
			rows, found := byHarness[harness]
			require.True(t, found, "list --json should contain harness %q", harness)
			assertListRowsHavePluginStatus(t, rows, slug, status, registryLatest)
		}
		return
	}

	var rows []map[string]any
	require.NoError(t, json.Unmarshal([]byte(extractJSONObjectOrArray(t, out)), &rows), "output must be valid JSON")
	assertListRowsHavePluginStatus(t, rows, slug, status, registryLatest)
}

func assertListRowsHavePluginStatus(t *testing.T, rows []map[string]any, slug, status, registryLatest string) {
	t.Helper()
	require.NotEmpty(t, rows, "at least one row expected")
	for _, row := range rows {
		if row["name"] == slug {
			assert.Equal(t, status, row["status"])
			if registryLatest != "" {
				assert.Equal(t, registryLatest, row["registryLatest"])
			}
			return
		}
	}
	t.Fatalf("plugin %s should appear in list output", slug)
}

// ---------------------------------------------------------------------------
// Delete
// ---------------------------------------------------------------------------

// TestAgentPluginsDelete verifies that deleting a specific version removes
// that version folder from Artifactory (--version is always required).
func TestAgentPluginsDelete(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "delete-plugin"
	version := "1.0.0"
	pluginPath := createTestPlugin(t, slug, version)

	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	))
	assertPluginExists(t, slug, version)

	require.NoError(t, runAgentPluginsCmd(t,
		"delete", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--version="+version,
	))
	assertPluginAbsent(t, slug, version)
}

// TestAgentPluginsDeleteDryRun verifies that --dry-run does not remove the
// artifact from Artifactory when the version exists.
func TestAgentPluginsDeleteDryRun(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "delete-dryrun-plugin"
	version := "1.0.0"
	pluginPath := createTestPlugin(t, slug, version)

	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	))

	require.NoError(t, runAgentPluginsCmd(t,
		"delete", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--version="+version,
		"--dry-run",
	))

	assertPluginExists(t, slug, version)
}

// TestAgentPluginsDeleteDryRunMultipleVersions verifies that --dry-run on a
// multi-version plugin only targets the specified version and leaves all
// versions intact on disk.
func TestAgentPluginsDeleteDryRunMultipleVersions(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "delete-dryrun-multi-plugin"
	for _, version := range []string{"1.0.0", "2.0.0"} {
		pluginPath := createTestPlugin(t, slug, version)
		require.NoError(t, runAgentPluginsCmd(t, "publish", pluginPath, "--repo="+tests.AgentPluginsLocalRepo))
	}

	require.NoError(t, runAgentPluginsCmd(t,
		"delete", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--version=2.0.0",
		"--dry-run",
	))

	assertPluginExists(t, slug, "1.0.0")
	assertPluginExists(t, slug, "2.0.0")
}

// TestAgentPluginsDeleteDryRunNotFound verifies that delete --dry-run on a
// missing plugin returns PackageVersionExists' not-found error (delete.go).
func TestAgentPluginsDeleteDryRunNotFound(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	err := runAgentPluginsCmd(t,
		"delete", "nonexistent-dryrun-plugin",
		"--repo="+tests.AgentPluginsLocalRepo,
		"--version=1.0.0",
		"--dry-run",
	)
	assertErrorContainsAll(t, err,
		"plugin 'nonexistent-dryrun-plugin' v1.0.0 not found in repository '"+tests.AgentPluginsLocalRepo+"'",
	)
}

// TestAgentPluginsDeleteMissing verifies that deleting a nonexistent slug/version
// fails via DeleteVersion (HTTP error from Artifactory).
func TestAgentPluginsDeleteMissing(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	err := runAgentPluginsCmd(t,
		"delete", "nonexistent-slug-xyzzy",
		"--repo="+tests.AgentPluginsLocalRepo,
		"--version=1.0.0",
	)
	assertErrorContainsAll(t, err,
		"failed to delete",
		"nonexistent-slug-xyzzy",
		"1.0.0",
	)
}

// TestAgentPluginsDeleteMissingVersionOfExistingPlugin verifies that deleting a
// version that was never published for an otherwise-known slug fails.
func TestAgentPluginsDeleteMissingVersionOfExistingPlugin(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "delete-missing-ver-plugin"
	pluginPath := createTestPlugin(t, slug, "1.0.0")
	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	))

	err := runAgentPluginsCmd(t,
		"delete", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--version=9.9.9",
	)
	assertErrorContainsAll(t, err, "failed to delete", slug, "9.9.9")
	assertPluginExists(t, slug, "1.0.0")
}

// TestAgentPluginsDeleteOnlySpecifiedVersion verifies that deleting one version
// leaves other versions of the same plugin intact.
func TestAgentPluginsDeleteOnlySpecifiedVersion(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "delete-versioned-plugin"
	keepVersion := "1.0.0"
	deleteVersion := "2.0.0"

	for _, version := range []string{keepVersion, deleteVersion} {
		pluginPath := createTestPlugin(t, slug, version)
		require.NoError(t, runAgentPluginsCmd(t, "publish", pluginPath, "--repo="+tests.AgentPluginsLocalRepo))
	}

	require.NoError(t, runAgentPluginsCmd(t,
		"delete", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--version="+deleteVersion,
	))

	assertPluginAbsent(t, slug, deleteVersion)
	assertPluginExists(t, slug, keepVersion)
}

// TestAgentPluginsDeleteMissingVersionFlag verifies that omitting --version
// produces the exact delete.go validation error.
func TestAgentPluginsDeleteMissingVersionFlag(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "delete-no-version-plugin"
	pluginPath := createTestPlugin(t, slug, "1.0.0")
	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	))

	err := runAgentPluginsCmd(t,
		"delete", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
	)
	assertErrorContainsAll(t, err, "--version is required for delete")
	assertPluginExists(t, slug, "1.0.0")
}

// TestAgentPluginsDeleteMissingSlugArg verifies usage when the required slug
// positional argument is omitted.
func TestAgentPluginsDeleteMissingSlugArg(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	err := runAgentPluginsCmd(t,
		"delete",
		"--repo="+tests.AgentPluginsLocalRepo,
		"--version=1.0.0",
	)
	assertErrorContainsAll(t, err, "usage: jf agent plugins delete")
}

// TestAgentPluginsDeleteRepoFromEnvVar verifies delete resolves the repo from
// JFROG_AGENT_PLUGINS_REPO when --repo is omitted.
func TestAgentPluginsDeleteRepoFromEnvVar(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "delete-env-repo-plugin"
	version := "1.0.0"
	pluginPath := createTestPlugin(t, slug, version)
	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	))

	t.Setenv("JFROG_AGENT_PLUGINS_REPO", tests.AgentPluginsLocalRepo)
	require.NoError(t, runAgentPluginsCmd(t,
		"delete", slug,
		"--version="+version,
	), "delete should resolve repo from JFROG_AGENT_PLUGINS_REPO")
	assertPluginAbsent(t, slug, version)
}

// ---------------------------------------------------------------------------
// List
// ---------------------------------------------------------------------------

// TestAgentPluginsListRemote verifies list --repo --format=json returns the
// published plugin with name, latest version, and "Repo: <repo>" source.
func TestAgentPluginsListRemote(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "list-plugin"
	version := "1.0.0"
	pluginPath := createTestPlugin(t, slug, version)
	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	))

	out, err := runAgentPluginsCmdWithOutput(t,
		"list",
		"--repo="+tests.AgentPluginsLocalRepo,
		"--format=json",
	)
	require.NoError(t, err, "list --repo --format=json should succeed after publish")
	assertListRepoJSONContains(t, out, slug, version, tests.AgentPluginsLocalRepo)
}

// TestAgentPluginsListRemoteLatestVersionOnly publishes two versions and verifies
// list --repo reports only the latest version for that slug.
func TestAgentPluginsListRemoteLatestVersionOnly(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "list-latest-plugin"
	require.NoError(t, runAgentPluginsCmd(t, "publish", createTestPlugin(t, slug, "1.0.0"),
		"--repo="+tests.AgentPluginsLocalRepo))
	require.NoError(t, runAgentPluginsCmd(t, "publish", createTestPlugin(t, slug, "2.0.0"),
		"--repo="+tests.AgentPluginsLocalRepo))

	out, err := runAgentPluginsCmdWithOutput(t,
		"list",
		"--repo="+tests.AgentPluginsLocalRepo,
		"--format=json",
	)
	require.NoError(t, err, "list --repo should succeed with multiple versions published")
	assertListRepoJSONContains(t, out, slug, "2.0.0", tests.AgentPluginsLocalRepo)
}

// TestAgentPluginsListLocal verifies list --harness --format=json returns the
// installed plugin for each harness condition. Omitting --global exercises
// resolveListScope's default-to-global behavior for built-in agents.
func TestAgentPluginsListLocal(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "list-local-plugin"
	version := "1.0.0"
	pluginPath := createTestHarnessPlugin(t, slug, version, allAgentHarnesses)
	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	))

	for _, tc := range agentPluginHarnessCases() {
		t.Run(tc.name, func(t *testing.T) {
			homeDir := setIsolatedHome(t)
			require.NoError(t, installViaMarketplaceWithRetry(t, slug, harnessFlag(tc.harnesses)))
			assertPluginsInstalledGlobally(t, homeDir, tc.harnesses, slug, version)

			// Intentionally omit --global: list should still default to global scope.
			out, err := runAgentPluginsCmdWithOutput(t,
				"list",
				"--harness="+harnessFlag(tc.harnesses),
				"--format=json",
			)
			require.NoError(t, err, "list --harness --format=json should succeed after install")
			assertListContainsInstalledPlugin(t, out, tc.harnesses, slug, version)
		})
	}
}

// TestAgentPluginsListEmptyLocal verifies list --harness --global --format=json
// returns an empty JSON array when nothing is installed.
func TestAgentPluginsListEmptyLocal(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	// Isolate HOME so list --global does not scan the real user profile.
	_ = setIsolatedHome(t)
	out, err := runAgentPluginsCmdWithOutput(t,
		"list",
		"--harness=cursor",
		"--global",
		"--format=json",
	)
	require.NoError(t, err, "list with no installed plugins should succeed")
	rows := parseListLocalJSONArray(t, out)
	assert.Empty(t, rows, "empty install directory should list as []")
}

// TestAgentPluginsListNeitherRepoNorHarness verifies the usage error when neither
// --repo nor --harness is provided.
func TestAgentPluginsListNeitherRepoNorHarness(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	err := runAgentPluginsCmd(t, "list")
	assertErrorContainsAll(t, err,
		"jf agent plugins list requires exactly one of:",
		"Registry: jf agent plugins list --repo",
		"Local:    jf agent plugins list --harness",
	)
}

// TestAgentPluginsListProjectScopeRejected verifies built-in agents reject
// --project-dir (project-scoped list) with RejectUnsupportedProjectScope messages.
func TestAgentPluginsListProjectScopeRejected(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	projectDir := t.TempDir()
	for _, tc := range agentPluginHarnessCases() {
		t.Run(tc.name, func(t *testing.T) {
			err := runAgentPluginsCmd(t,
				"list",
				"--harness="+harnessFlag(tc.harnesses),
				"--project-dir="+projectDir,
			)
			require.Error(t, err, "built-in harnesses must reject project-scoped list")
			assertErrorContainsAll(t, err,
				"does not support project-scoped plugin lists",
				"Use --global instead",
			)
		})
	}
}

// TestAgentPluginsListFlags exercises list flag combinations that must either
// succeed or produce a descriptive error (never panic).
func TestAgentPluginsListFlags(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "list-flags-plugin"
	pluginPath := createTestPlugin(t, slug, "1.0.0")
	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	))

	cases := []struct {
		name        string
		args        []string
		expectError bool
		errContains []string
		description string
	}{
		{
			name:        "format-json",
			args:        []string{"list", "--repo=" + tests.AgentPluginsLocalRepo, "--format=json"},
			expectError: false,
			description: "--format json with --repo should produce JSON output without error",
		},
		{
			name:        "limit-positive",
			args:        []string{"list", "--repo=" + tests.AgentPluginsLocalRepo, "--limit=5"},
			expectError: false,
			description: "--limit with a positive value should succeed",
		},
		{
			name:        "sort-by-updated",
			args:        []string{"list", "--repo=" + tests.AgentPluginsLocalRepo, "--sort-by=updated"},
			expectError: false,
			description: "--sort-by updated is a valid value for --repo mode",
		},
		{
			name:        "sort-by-downloads",
			args:        []string{"list", "--repo=" + tests.AgentPluginsLocalRepo, "--sort-by=downloads"},
			expectError: false,
			description: "--sort-by downloads is a valid value for --repo mode",
		},
		{
			name:        "sort-by-invalid-repo",
			args:        []string{"list", "--repo=" + tests.AgentPluginsLocalRepo, "--sort-by=invalid-field"},
			expectError: true,
			errContains: []string{`--sort-by for --repo accepts 'updated' or 'downloads'`},
			description: "--sort-by with unknown field must produce an error in --repo mode",
		},
		{
			name:        "sort-by-invalid-harness",
			args:        []string{"list", "--harness=cursor", "--global", "--sort-by=updated"},
			expectError: true,
			errContains: []string{`--sort-by for --harness only accepts 'name'`},
			description: "--sort-by updated is invalid in --harness mode",
		},
		{
			name:        "sort-order-desc-harness",
			args:        []string{"list", "--harness=cursor", "--global", "--sort-order=desc", "--format=json"},
			expectError: false,
			description: "--sort-order desc is valid for --harness mode",
		},
		{
			name:        "sort-order-invalid-harness",
			args:        []string{"list", "--harness=cursor", "--global", "--sort-order=sideways"},
			expectError: true,
			errContains: []string{`--sort-order must be 'asc' or 'desc'`},
			description: "--sort-order is validated in --harness mode",
		},
		{
			name:        "sort-order-ignored-in-repo-mode",
			args:        []string{"list", "--repo=" + tests.AgentPluginsLocalRepo, "--sort-order=sideways"},
			expectError: false,
			description: "--sort-order is ignored (not validated) in --repo mode",
		},
		{
			name:        "check-updates-without-harness",
			args:        []string{"list", "--repo=" + tests.AgentPluginsLocalRepo, "--check-updates"},
			expectError: true,
			errContains: []string{"--check-updates is only supported with --harness, not with --repo"},
			description: "--check-updates requires --harness; using it with --repo alone must error",
		},
		{
			name:        "repo-and-harness-mutually-exclusive",
			args:        []string{"list", "--repo=" + tests.AgentPluginsLocalRepo, "--harness=claude", "--global"},
			expectError: true,
			errContains: []string{"--repo and --harness are mutually exclusive; specify only one"},
			description: "--repo and --harness together must error",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := runAgentPluginsCmd(t, tc.args...)
			if tc.expectError {
				assertErrorContainsAll(t, err, tc.errContains...)
			} else {
				assert.NoError(t, err, tc.description)
			}
		})
	}
}

// TestAgentPluginsListGlobalProjectDirMutuallyExclusive verifies that passing
// both --global and --project-dir to list returns a clear error.
func TestAgentPluginsListGlobalProjectDirMutuallyExclusive(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	err := runAgentPluginsCmd(t,
		"list",
		"--harness=claude",
		"--global",
		"--project-dir="+t.TempDir(),
	)
	assertErrorContainsAll(t, err,
		"--global and --project-dir are mutually exclusive",
		"please choose either --global or --project-dir",
	)
}

// TestAgentPluginsListLimitHarnessMode verifies --limit truncates JSON rows when
// listing installed plugins via --harness.
func TestAgentPluginsListLimitHarnessMode(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slugs := []string{"limit-a-plugin", "limit-b-plugin", "limit-c-plugin"}
	for _, slug := range slugs {
		p := createTestHarnessPlugin(t, slug, "1.0.0", allAgentHarnesses)
		require.NoError(t, runAgentPluginsCmd(t, "publish", p, "--repo="+tests.AgentPluginsLocalRepo))
	}

	for _, tc := range agentPluginHarnessCases() {
		t.Run(tc.name, func(t *testing.T) {
			homeDir := setIsolatedHome(t)
			for _, slug := range slugs {
				require.NoError(t, installViaMarketplaceWithRetry(t, slug, harnessFlag(tc.harnesses)))
				assertPluginsInstalledGlobally(t, homeDir, tc.harnesses, slug)
			}

			out, err := runAgentPluginsCmdWithOutput(t,
				"list",
				"--harness="+harnessFlag(tc.harnesses),
				"--global",
				"--limit=2",
				"--format=json",
			)
			require.NoError(t, err, "list --harness --limit should succeed")
			assertListLocalRowCountEqual(t, out, tc.harnesses, 2)
		})
	}
}

// TestAgentPluginsListLimitRepoMode verifies --limit truncates registry list
// results when using --repo --format=json.
func TestAgentPluginsListLimitRepoMode(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	for _, slug := range []string{"limit-repo-a", "limit-repo-b", "limit-repo-c"} {
		require.NoError(t, runAgentPluginsCmd(t,
			"publish", createTestPlugin(t, slug, "1.0.0"),
			"--repo="+tests.AgentPluginsLocalRepo,
		))
	}

	out, err := runAgentPluginsCmdWithOutput(t,
		"list",
		"--repo="+tests.AgentPluginsLocalRepo,
		"--limit=2",
		"--format=json",
	)
	require.NoError(t, err, "list --repo --limit=2 should succeed")
	rows := parseListRepoJSON(t, out)
	assert.Equal(t, 2, len(rows), "list --repo --limit=2 must return exactly 2 rows when more plugins exist")
}

// TestAgentPluginsListLimitZero verifies that --limit=0 (or a negative value)
// returns an error rather than returning an empty result set.
func TestAgentPluginsListLimitZero(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	err := runAgentPluginsCmd(t,
		"list",
		"--repo="+tests.AgentPluginsLocalRepo,
		"--limit=0",
	)
	assertErrorContainsAll(t, err, `--limit must be a positive integer`)
}

// ---------------------------------------------------------------------------
// Search
// ---------------------------------------------------------------------------

// TestAgentPluginsSearch verifies that search finds a published plugin by
// agentplugins.name and returns JSON rows with name, version, and repository.
func TestAgentPluginsSearch(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "search-plugin"
	version := "1.0.0"
	pluginPath := createTestPlugin(t, slug, version)
	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	))

	rows := searchRowsWithRetry(t, slug, version,
		"search", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--format=json",
	)
	assertSearchRowsContain(t, rows, slug, version, tests.AgentPluginsLocalRepo)
}

// TestAgentPluginsSearchSubstringMatch verifies wildcard wrapping: a partial
// query matches plugin names (search.go wraps non-wildcard queries in *...*).
func TestAgentPluginsSearchSubstringMatch(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "search-substring-plugin"
	version := "1.0.0"
	pluginPath := createTestPlugin(t, slug, version)
	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	))

	rows := searchRowsWithRetry(t, slug, version,
		"search", "substring",
		"--repo="+tests.AgentPluginsLocalRepo,
		"--format=json",
	)
	assertSearchRowsContain(t, rows, slug, version, tests.AgentPluginsLocalRepo)
}

// TestAgentPluginsSearchLatestVersionOnly publishes two versions and verifies
// search returns only the highest semver (SearchLatestRowsByProperty).
func TestAgentPluginsSearchLatestVersionOnly(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "search-latest-plugin"
	for _, version := range []string{"1.0.0", "2.0.0"} {
		pluginPath := createTestPlugin(t, slug, version)
		require.NoError(t, runAgentPluginsCmd(t, "publish", pluginPath, "--repo="+tests.AgentPluginsLocalRepo))
	}

	// Wait for the highest version specifically: the lower one can be indexed first,
	// and a search that sees only 1.0.0 would fail the latest-only assertion below.
	rows := searchRowsWithRetry(t, slug, "2.0.0",
		"search", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--format=json",
	)
	require.Len(t, rows, 1, "search should return one row per plugin name (latest only)")
	assert.Equal(t, slug, rows[0].Name)
	assert.Equal(t, "2.0.0", rows[0].Version, "search should keep the highest semver")
	assert.Equal(t, tests.AgentPluginsLocalRepo, rows[0].Repository)
}

// TestAgentPluginsSearchNoMatches verifies that searching with a query that
// matches nothing succeeds with an empty result and logs the not-found message.
func TestAgentPluginsSearchNoMatches(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	query := "nonexistent-plugin-xyzzy-abc123"
	searchArgs := []string{"search", query, "--repo=" + tests.AgentPluginsLocalRepo, "--format=json"}

	out, err := runAgentPluginsCmdWithOutput(t, searchArgs...)
	require.NoError(t, err, "search with no matches should return empty result, not an error")
	assert.Empty(t, strings.TrimSpace(out),
		"no-match search should not print JSON rows, got: %q", out)

	// PrintSearchResults reports the no-match case through log.Info, so it lands on the
	// logs writer rather than stdout. This needs a second run with a plain Exec:
	// RunCliCmdWithOutputs installs its own logger writing to the real stderr, which
	// would discard the buffer redirect.
	_, logBuf, previousLog := coretests.RedirectLogOutputToBuffer()
	t.Cleanup(func() { log.SetLogger(previousLog) })
	require.NoError(t, runAgentPluginsCmd(t, searchArgs...))
	assert.Contains(t, logBuf.String(), fmt.Sprintf("No plugins found matching '%s'.", query))
}

// TestAgentPluginsSearchEmptyQuery verifies that omitting the query argument
// returns the usage error from RunSearch.
func TestAgentPluginsSearchEmptyQuery(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	err := runAgentPluginsCmd(t, "search", "--repo="+tests.AgentPluginsLocalRepo)
	assertErrorContainsAll(t, err, "usage: jf agent plugins search")
}

// TestAgentPluginsSearchBlankQuery verifies that a whitespace-only query is
// rejected after TrimSpace (search.go).
func TestAgentPluginsSearchBlankQuery(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	err := runAgentPluginsCmd(t, "search", "   ", "--repo="+tests.AgentPluginsLocalRepo)
	assertErrorContainsAll(t, err, "search query cannot be empty")
}

// TestAgentPluginsSearchRepoFromEnvVar verifies that search picks up the repo
// from JFROG_AGENT_PLUGINS_REPO when --repo is omitted.
func TestAgentPluginsSearchRepoFromEnvVar(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "search-envvar-plugin"
	version := "1.0.0"
	pluginPath := createTestPlugin(t, slug, version)
	require.NoError(t, runAgentPluginsCmd(t, "publish", pluginPath, "--repo="+tests.AgentPluginsLocalRepo))

	t.Setenv("JFROG_AGENT_PLUGINS_REPO", tests.AgentPluginsLocalRepo)

	rows := searchRowsWithRetry(t, slug, version, "search", slug, "--format=json")
	assertSearchRowsContain(t, rows, slug, version, tests.AgentPluginsLocalRepo)
}

type agentPluginsSearchRow struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Repository  string `json:"repository"`
	Description string `json:"description"`
}

func parseSearchRows(out string) ([]agentPluginsSearchRow, error) {
	// CLI may log the command line before JSON; extract the JSON array. A search that
	// matched nothing reports that through the log instead, leaving no array at all.
	start := strings.Index(out, "[")
	end := strings.LastIndex(out, "]")
	if start < 0 || end <= start {
		return nil, fmt.Errorf("search output does not contain a JSON array, got: %q", out)
	}
	var rows []agentPluginsSearchRow
	if err := json.Unmarshal([]byte(out[start:end+1]), &rows); err != nil {
		return nil, fmt.Errorf("parse search JSON %q: %w", out[start:end+1], err)
	}
	return rows, nil
}

// searchRowsWithRetry runs a search command until it lists slug at version. Artifactory sets the
// agentplugins.name property that search queries asynchronously after the upload response returns,
// so a search issued right after publish can legitimately come back empty. This is the same
// wait-for-async-indexing pattern as assertMarketplaceContainsPlugin.
func searchRowsWithRetry(t *testing.T, slug, version string, searchArgs ...string) []agentPluginsSearchRow {
	t.Helper()
	var found []agentPluginsSearchRow
	description := fmt.Sprintf("wait for property search to list %s %s", slug, version)
	require.NoError(t, retryWithBackoff(t, description, func() error {
		out, err := runAgentPluginsCmdWithOutput(t, searchArgs...)
		if err != nil {
			return err
		}
		rows, err := parseSearchRows(out)
		if err != nil {
			return err
		}
		for _, row := range rows {
			if row.Name == slug && row.Version == version {
				found = rows
				return nil
			}
		}
		return fmt.Errorf("search does not list plugin %q at version %q yet; output: %s", slug, version, out)
	}))
	return found
}

func assertSearchRowsContain(t *testing.T, rows []agentPluginsSearchRow, slug, version, repo string) {
	t.Helper()
	for _, row := range rows {
		if row.Name == slug {
			assert.Equal(t, version, row.Version, "search row version for %q", slug)
			assert.Equal(t, repo, row.Repository, "search row repository for %q", slug)
			return
		}
	}
	t.Fatalf("search rows did not contain plugin %q; rows: %+v", slug, rows)
}

// ---------------------------------------------------------------------------
// Config
// ---------------------------------------------------------------------------

// TestAgentPluginsRepoFromEnvVar verifies that JFROG_AGENT_PLUGINS_REPO is
// respected when --repo is omitted.
func TestAgentPluginsRepoFromEnvVar(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "envvar-repo-plugin"
	pluginPath := createTestPlugin(t, slug, "1.0.0")

	t.Setenv("JFROG_AGENT_PLUGINS_REPO", tests.AgentPluginsLocalRepo)

	// --repo is intentionally omitted; the env var should supply it.
	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
	), "publish should succeed using repo from JFROG_AGENT_PLUGINS_REPO env var")

	assertPluginExists(t, slug, "1.0.0")
}

// TestAgentPluginsInstallRepoFromEnvVar verifies install resolves the repo from
// JFROG_AGENT_PLUGINS_REPO when --repo is omitted.
func TestAgentPluginsInstallRepoFromEnvVar(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "install-env-repo-plugin"
	version := "1.0.0"
	require.NoError(t, runAgentPluginsCmd(t,
		"publish", createTestHarnessPlugin(t, slug, version, []string{"cursor"}),
		"--repo="+tests.AgentPluginsLocalRepo,
	))

	t.Setenv("JFROG_AGENT_PLUGINS_REPO", tests.AgentPluginsLocalRepo)
	homeDir := setIsolatedHome(t)
	// Omit --repo and --version: repo from env, version from cursor-marketplace.json.
	require.NoError(t, retryWithBackoff(t, "install "+slug+" via env repo and marketplace", func() error {
		return runAgentPluginsCmd(t,
			"install", slug,
			"--harness=cursor",
			"--global",
		)
	}), "install should resolve repo from JFROG_AGENT_PLUGINS_REPO")
	assertPluginsInstalledGlobally(t, homeDir, []string{"cursor"}, slug, version)
}

// TestAgentPluginsUpdateRepoFromEnvVar verifies update resolves the repo from
// JFROG_AGENT_PLUGINS_REPO when --repo is omitted.
func TestAgentPluginsUpdateRepoFromEnvVar(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "update-env-repo-plugin"
	require.NoError(t, runAgentPluginsCmd(t, "publish", createTestPlugin(t, slug, "1.0.0"),
		"--repo="+tests.AgentPluginsLocalRepo))
	require.NoError(t, runAgentPluginsCmd(t, "publish", createTestPlugin(t, slug, "2.0.0"),
		"--repo="+tests.AgentPluginsLocalRepo))

	homeDir := setIsolatedHome(t)
	require.NoError(t, runAgentPluginsCmd(t,
		"install", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--harness=cursor",
		"--global",
		"--version=1.0.0",
	))

	t.Setenv("JFROG_AGENT_PLUGINS_REPO", tests.AgentPluginsLocalRepo)
	require.NoError(t, runAgentPluginsCmd(t,
		"update",
		"--slug="+slug,
		"--harness=cursor",
		"--global",
	), "update should resolve repo from JFROG_AGENT_PLUGINS_REPO")
	assertPluginsInstalledGlobally(t, homeDir, []string{"cursor"}, slug, "2.0.0")
	assertPluginsInstalledNatively(t, []string{"cursor"}, slug, "2.0.0")
}

// TestAgentPluginsRepoFlagOverridesEnvVar verifies that --repo takes precedence
// over the JFROG_AGENT_PLUGINS_REPO environment variable.
func TestAgentPluginsRepoFlagOverridesEnvVar(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "flag-override-repo-plugin"
	pluginPath := createTestPlugin(t, slug, "1.0.0")

	// Set env var to a nonexistent repo; --repo flag with the real repo should win.
	t.Setenv("JFROG_AGENT_PLUGINS_REPO", "nonexistent-env-repo-xyz")

	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	), "--repo flag must override JFROG_AGENT_PLUGINS_REPO env var")

	assertPluginExists(t, slug, "1.0.0")
}

// TestAgentPluginsNoRepoConfigured verifies that omitting both --repo and
// JFROG_AGENT_PLUGINS_REPO produces ResolveRepo's discovery error when no
// agentplugins repositories exist (see agent/common/resolve_repo.go).
func TestAgentPluginsNoRepoConfigured(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	t.Setenv("JFROG_AGENT_PLUGINS_REPO", "")

	// Temporarily remove the suite's shared agentplugins repo so auto-discovery finds nothing.
	// Safe only while this package runs sequentially (no t.Parallel): a hard process exit
	// before Cleanup would leave later tests without the suite repo.
	require.True(t, isRepoExist(tests.AgentPluginsLocalRepo))
	execDeleteRepo(tests.AgentPluginsLocalRepo)
	require.False(t, isRepoExist(tests.AgentPluginsLocalRepo))
	t.Cleanup(func() { recreateAgentPluginsLocalRepo(t) })

	pluginPath := createTestPlugin(t, "no-repo-plugin", "1.0.0")
	err := runAgentPluginsCmd(t, "publish", pluginPath)
	assertErrorContainsAll(t, err,
		"no agent plugins repositories found",
	)
}

// TestAgentPluginsServerIDValid verifies that an explicit --server-id pointing
// to the configured test server succeeds for publish and install.
func TestAgentPluginsServerIDValid(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "serverid-valid-plugin"
	version := "1.0.0"
	pluginPath := createTestPlugin(t, slug, version)

	// createJfrogHomeConfig registers the test server as "default".
	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--server-id=default",
	), "publish with a valid --server-id should succeed")
	assertPluginExists(t, slug, version)

	installDir := t.TempDir()
	require.NoError(t, runAgentPluginsCmd(t,
		"install", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--path="+installDir,
		"--server-id=default",
		"--version="+version,
	), "install with a valid --server-id should succeed")
	require.FileExists(t, filepath.Join(installDir, slug, ".jfrog", "plugin-info.json"))
}

// TestAgentPluginsServerIDUnknown verifies that an unknown --server-id produces
// a clear error before any network call is attempted.
func TestAgentPluginsServerIDUnknown(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	pluginPath := createTestPlugin(t, "serverid-bad-plugin", "1.0.0")
	err := runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--server-id=nonexistent-server-id-xyz",
	)
	assertErrorContainsAll(t, err, "Server ID 'nonexistent-server-id-xyz' does not exist.")
}

// ---------------------------------------------------------------------------
// Round-trip
// ---------------------------------------------------------------------------

// TestAgentPluginsRoundTrip publishes a plugin then installs it and verifies
// the installed manifest matches slug and version.
func TestAgentPluginsRoundTrip(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "roundtrip-plugin"
	version := "1.0.0"
	pluginPath := createTestPlugin(t, slug, version)

	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	))

	installDir := t.TempDir()
	require.NoError(t, runAgentPluginsCmd(t,
		"install", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--path="+installDir,
	))

	installedManifest := filepath.Join(installDir, slug, "plugin.json")
	require.FileExists(t, installedManifest, "plugin.json should exist after install")
	data, err := os.ReadFile(installedManifest) // #nosec G304 -- path from t.TempDir
	require.NoError(t, err)
	var manifest struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}
	require.NoError(t, json.Unmarshal(data, &manifest))
	assert.Equal(t, slug, manifest.Name, "installed plugin name should match published slug")
	assert.Equal(t, version, manifest.Version, "installed plugin version should match published version")
}

// TestAgentPluginsRoundTripWithUpdate extends the basic round-trip by also
// running update and verifying the installed version advances to the latest.
func TestAgentPluginsRoundTripWithUpdate(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "rt-update-plugin"
	v1 := "1.0.0"
	v2 := "2.0.0"

	v1Path := createTestPlugin(t, slug, v1)
	require.NoError(t, runAgentPluginsCmd(t, "publish", v1Path, "--repo="+tests.AgentPluginsLocalRepo))

	installDir := t.TempDir()
	require.NoError(t, runAgentPluginsCmd(t,
		"install", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--path="+installDir,
		"--version="+v1,
	))

	v2Path := createTestPlugin(t, slug, v2)
	require.NoError(t, runAgentPluginsCmd(t, "publish", v2Path, "--repo="+tests.AgentPluginsLocalRepo))

	assert.NoError(t, runAgentPluginsCmd(t,
		"update",
		"--slug="+slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--path="+installDir,
	))

	manifestPath := filepath.Join(installDir, slug, ".jfrog", "plugin-info.json")
	require.FileExists(t, manifestPath)
	data, err := os.ReadFile(manifestPath) // #nosec G304 -- path from t.TempDir
	require.NoError(t, err)
	var manifest map[string]any
	require.NoError(t, json.Unmarshal(data, &manifest))
	assert.Equal(t, v2, manifest["installedVersion"],
		"after round-trip update, installed version should be %s", v2)
}

// TestAgentPluginsRoundTripDeleteThenInstall publishes, deletes a specific
// version, then verifies that installing that version fails with not-found.
func TestAgentPluginsRoundTripDeleteThenInstall(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	slug := "rt-delete-plugin"
	deletedVersion := "1.0.0"
	keepVersion := "2.0.0"

	for _, version := range []string{deletedVersion, keepVersion} {
		pluginPath := createTestPlugin(t, slug, version)
		require.NoError(t, runAgentPluginsCmd(t, "publish", pluginPath, "--repo="+tests.AgentPluginsLocalRepo))
	}

	// Delete v1.
	require.NoError(t, runAgentPluginsCmd(t,
		"delete", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--version="+deletedVersion,
	))

	// Attempting to install the deleted version should now fail.
	installDir := t.TempDir()
	err := runAgentPluginsCmd(t,
		"install", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--path="+installDir,
		"--version="+deletedVersion,
	)
	assertErrorContainsAll(t, err, "not found in repository")
}

// ---------------------------------------------------------------------------
// CI/CD
// ---------------------------------------------------------------------------

// TestAgentPluginsArtifactoryUnreachable verifies that pointing --repo at a
// nonexistent Artifactory URL fails with a clear error and does not silently
// succeed.
func TestAgentPluginsArtifactoryUnreachable(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	// Register a server entry that points to an unreachable host.
	const bogusServerID = "unreachable-rt-server"
	bogusServerURL := "https://nonexistent-artifactory-host-xyzzy.example.com/artifactory/"
	configCli := coretests.NewJfrogCli(execMain, "jfrog config", "")
	require.NoError(t, configCli.Exec("add", bogusServerID,
		"--interactive=false",
		"--url="+bogusServerURL,
		"--access-token=dummytoken",
	))
	t.Cleanup(func() {
		// Best-effort removal of the temporary unreachable server entry.
		_ = configCli.Exec("rm", bogusServerID, "--quiet")
	})

	installDir := t.TempDir()
	err := runAgentPluginsCmd(t,
		"install", "any-plugin",
		"--repo=nonexistent-repo-on-unreachable-server",
		"--server-id="+bogusServerID,
		"--path="+installDir,
	)
	require.Error(t, err,
		"install against an unreachable server should fail with a clear error")
	assert.Contains(t, err.Error(), "nonexistent-artifactory-host-xyzzy.example.com",
		"error should identify the unreachable host")
}

// TestAgentPluginsCIPipeline simulates a minimal CI/CD workflow:
//  1. publish with build info flags (--quiet mirrors CI)
//  2. jf rt bp to push build info to Artifactory
//  3. install the same slug to confirm end-to-end availability
func TestAgentPluginsCIPipeline(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	const (
		slug    = "ci-pipeline-plugin"
		version = "1.0.0"
	)
	buildNumber := t.Name()
	// Best-effort teardown of local build-info dir; leftover dirs do not affect assertions.
	t.Cleanup(func() { _ = coreBuild.RemoveBuildDir(tests.AgentPluginsBuildName, buildNumber, "") })

	pluginPath := createTestPlugin(t, slug, version)

	// Step 1 — publish with build info, simulating CI (--quiet suppresses prompts).
	require.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--build-name="+tests.AgentPluginsBuildName,
		"--build-number="+buildNumber,
		"--quiet",
	), "CI publish step must succeed")

	assertPluginExists(t, slug, version)

	// Step 2 — push build info to Artifactory.
	require.NoError(t, artifactoryCli.Exec("bp", tests.AgentPluginsBuildName, buildNumber),
		"jf rt bp must succeed after publish")

	_, found, err := tests.GetBuildInfo(serverDetails, tests.AgentPluginsBuildName, buildNumber)
	require.NoError(t, err)
	require.True(t, found, "build info must be retrievable after bp")

	// Step 3 — install into a scratch directory, simulating a downstream CI job.
	installBase := t.TempDir()
	require.NoError(t, runAgentPluginsCmd(t,
		"install", slug,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--path="+installBase,
		"--quiet",
	), "CI install step must succeed")

	pluginDir := filepath.Join(installBase, slug)
	_, err = os.Stat(pluginDir)
	assert.NoError(t, err, "installed plugin directory must exist after CI pipeline")
}

// ---------------------------------------------------------------------------
// Proxy
// ---------------------------------------------------------------------------

// TestAgentPluginsWithProxy verifies that install and publish work when
// HTTPS_PROXY is configured. Skipped unless PROXY_HTTPS_PORT is set.
func TestAgentPluginsWithProxy(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	proxyPort := os.Getenv(tests.HttpsProxyEnvVar)
	if proxyPort == "" {
		t.Skip("Skipping proxy test: set " + tests.HttpsProxyEnvVar + " env var to enable.")
	}

	slug := "proxy-plugin"
	pluginPath := createTestPlugin(t, slug, "1.0.0")

	assert.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	), "publish through proxy should succeed")

	assertPluginExists(t, slug, "1.0.0")
}

// TestAgentPluginsNoProxy verifies that when NO_PROXY includes the Artifactory
// host, the proxy is bypassed and the command connects directly.
func TestAgentPluginsNoProxy(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	proxyPort := os.Getenv(tests.HttpsProxyEnvVar)
	if proxyPort == "" {
		t.Skip("Skipping NO_PROXY test: set " + tests.HttpsProxyEnvVar + " env var to enable.")
	}

	// Bypass proxy for the Artifactory host.
	clientTestUtils.SetEnvWithCallbackAndAssert(t, "NO_PROXY", "*")

	slug := "no-proxy-plugin"
	pluginPath := createTestPlugin(t, slug, "1.0.0")

	assert.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	), "publish should bypass proxy and connect directly when NO_PROXY=*")

	assertPluginExists(t, slug, "1.0.0")
}

// ---------------------------------------------------------------------------
// TLS
// ---------------------------------------------------------------------------

// TestAgentPluginsInsecureTLS verifies --insecure-tls behaviour.
// Without the flag a self-signed cert connection should fail; with it it should
// succeed. Skipped unless an HTTPS Artifactory with a self-signed cert is
// configured via JFROG_CLI_TESTS_INSECURE_TLS_URL.
func TestAgentPluginsInsecureTLS(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	if os.Getenv("JFROG_CLI_TESTS_INSECURE_TLS_URL") == "" {
		t.Skip("Skipping TLS test: set JFROG_CLI_TESTS_INSECURE_TLS_URL to an Artifactory with a self-signed cert.")
	}

	slug := "insecure-tls-plugin"
	pluginPath := createTestPlugin(t, slug, "1.0.0")

	// Without --insecure-tls the cert error should surface.
	errWithout := runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
	)
	assert.Error(t, errWithout, "publish to self-signed Artifactory without --insecure-tls should fail")

	// With --insecure-tls it should succeed.
	assert.NoError(t, runAgentPluginsCmd(t,
		"publish", pluginPath,
		"--repo="+tests.AgentPluginsLocalRepo,
		"--insecure-tls",
	), "publish to self-signed Artifactory with --insecure-tls should succeed")
}

// ---------------------------------------------------------------------------
// Multi-Harness Marketplace Indexing
// ---------------------------------------------------------------------------

// TestAgentPluginsPublishMultiHarnessMarketplaceIndexing verifies that publishing
// creates the marketplace entries needed to install without an explicit version.
func TestAgentPluginsPublishMultiHarnessMarketplaceIndexing(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	for _, tc := range agentPluginHarnessCases() {
		t.Run(tc.name, func(t *testing.T) {
			version := "1.0.0"
			slug := "marketplace-index-" + strings.ReplaceAll(tc.name, ",", "-") + "-plugin"
			pluginDir := createTestHarnessPlugin(t, slug, version, tc.harnesses)

			require.NoError(t, runAgentPluginsCmd(t,
				"publish", pluginDir,
				"--repo="+tests.AgentPluginsLocalRepo,
				"--version="+version,
			), "publishing %s must succeed", slug)
			assertPluginExists(t, slug, version)

			for _, harness := range tc.harnesses {
				assertMarketplaceContainsPlugin(t, harness, slug, version)
			}

			homeDir := setIsolatedHome(t)
			require.NoError(t, runAgentPluginsCmd(t,
				"install", slug,
				"--repo="+tests.AgentPluginsLocalRepo,
				"--harness="+harnessFlag(tc.harnesses),
				"--global",
			), "install %s without --version must succeed", slug)
			assertPluginsInstalledGlobally(t, homeDir, tc.harnesses, slug, "1.0.0")
			assertPluginsInstalledNatively(t, tc.harnesses, slug, "1.0.0")

			out, err := runAgentPluginsCmdWithOutput(t,
				"list",
				"--harness="+harnessFlag(tc.harnesses),
				"--global",
				"--format=json",
			)
			require.NoError(t, err, "list --json must succeed after installing %s", slug)
			assertListContainsInstalledPlugin(t, out, tc.harnesses, slug, "1.0.0")
		})
	}
}

// assertMarketplaceContainsPlugin waits until <harness>-marketplace.json lists slug at version.
// Both the file's creation and its later rewrites are asynchronous on the Artifactory side, so a
// missing file and a stale file that predates this publish are equally expected early on and are
// retried rather than failed.
func assertMarketplaceContainsPlugin(t *testing.T, harness, slug, version string) {
	t.Helper()
	fileName := plugincommon.MarketplaceFileName(harness)
	description := fmt.Sprintf("wait for %s to list %s %s", fileName, slug, version)
	require.NoError(t, retryWithBackoff(t, description, func() error {
		// Download into a fresh directory each attempt: a previously fetched copy would
		// make `dl --fail-no-op` a no-op and hide the newly indexed content.
		downloadDir := t.TempDir()
		if err := downloadMarketplaceJSON(fileName, downloadDir); err != nil {
			return err
		}
		return marketplaceListsPluginVersion(filepath.Join(downloadDir, fileName), slug, version)
	}))
}

// marketplaceListsPluginVersion reports whether the marketplace index at path lists slug at the
// given version, returning a descriptive error when the index has not caught up yet.
func marketplaceListsPluginVersion(path, slug, version string) error {
	data, err := os.ReadFile(path) // #nosec G304 -- path is under t.TempDir
	if err != nil {
		return err
	}
	var marketplace struct {
		Plugins []struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"plugins"`
	}
	if err := json.Unmarshal(data, &marketplace); err != nil {
		return fmt.Errorf("parse %s: %w", filepath.Base(path), err)
	}
	for _, plugin := range marketplace.Plugins {
		if plugin.Name != slug {
			continue
		}
		if plugin.Version != version {
			return fmt.Errorf("%s lists %s at version %q, want %q",
				filepath.Base(path), slug, plugin.Version, version)
		}
		return nil
	}
	return fmt.Errorf("%s does not list plugin %q yet", filepath.Base(path), slug)
}

func downloadMarketplaceJSON(fileName, downloadDir string) error {
	return artifactoryCli.Exec(
		"dl",
		tests.AgentPluginsLocalRepo+"/"+fileName,
		downloadDir+"/",
		"--flat=true",
		"--fail-no-op=true",
	)
}

// Artifactory builds the per-harness marketplace index asynchronously after a publish, and the
// wait is unpredictable: on a loaded CI runner it can take far longer than the publish response.
// Back off geometrically within a fixed budget so a fast instance costs a couple of seconds while
// a slow one still gets a fair chance.
const (
	marketplaceIndexFirstWait = 2 * time.Second
	marketplaceIndexMaxWait   = 10 * time.Second
	marketplaceIndexBudget    = 90 * time.Second
)

// retryWithBackoff runs operation until it succeeds or the marketplace indexing budget is spent,
// doubling the wait between attempts. It always performs at least one attempt and wraps the last
// error so failures name the operation that timed out.
func retryWithBackoff(t *testing.T, description string, operation func() error) error {
	t.Helper()
	deadline := time.Now().Add(marketplaceIndexBudget)
	wait := marketplaceIndexFirstWait
	for attempt := 1; ; attempt++ {
		err := operation()
		if err == nil {
			return nil
		}
		if time.Now().Add(wait).After(deadline) {
			return fmt.Errorf("%s did not succeed within %s (%d attempts): %w",
				description, marketplaceIndexBudget, attempt, err)
		}
		t.Logf("%s: attempt %d failed (%v); retrying in %s", description, attempt, err, wait)
		time.Sleep(wait)
		if wait *= 2; wait > marketplaceIndexMaxWait {
			wait = marketplaceIndexMaxWait
		}
	}
}

func assertListContainsInstalledPlugin(t *testing.T, out string, harnesses []string, slug, version string) {
	t.Helper()
	if len(harnesses) == 1 {
		rows := parseListLocalJSONArray(t, out)
		assertListRowsContainPluginVersion(t, rows, slug, version)
		return
	}

	byHarness := parseListLocalJSONObject(t, out)
	for _, harness := range harnesses {
		rows, found := byHarness[harness]
		require.True(t, found, "list --json output should contain harness %q", harness)
		assertListRowsContainPluginVersion(t, rows, slug, version)
	}
}

func assertListRowsContainPluginVersion(t *testing.T, rows []map[string]any, slug, version string) {
	t.Helper()
	for _, row := range rows {
		if row["name"] == slug {
			assert.Equal(t, version, row["version"], "list row version for %q", slug)
			return
		}
	}
	t.Fatalf("list --json should contain installed plugin %q version %q", slug, version)
}

type agentPluginsListRepoRow struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Source  string `json:"source"`
}

func extractJSONObjectOrArray(t *testing.T, out string) string {
	t.Helper()
	trimmed := strings.TrimSpace(out)
	arrayStart := strings.Index(trimmed, "[")
	objectStart := strings.Index(trimmed, "{")
	switch {
	case arrayStart >= 0 && (objectStart < 0 || arrayStart < objectStart):
		end := strings.LastIndex(trimmed, "]")
		require.Greater(t, end, arrayStart, "JSON array must have a closing bracket, got: %q", out)
		return trimmed[arrayStart : end+1]
	case objectStart >= 0:
		end := strings.LastIndex(trimmed, "}")
		require.Greater(t, end, objectStart, "JSON object must have a closing brace, got: %q", out)
		return trimmed[objectStart : end+1]
	default:
		t.Fatalf("list/search output must contain JSON, got: %q", out)
		return ""
	}
}

func parseListRepoJSON(t *testing.T, out string) []agentPluginsListRepoRow {
	t.Helper()
	var rows []agentPluginsListRepoRow
	require.NoError(t, json.Unmarshal([]byte(extractJSONObjectOrArray(t, out)), &rows),
		"list --repo output must be a JSON array")
	return rows
}

func assertListRepoJSONContains(t *testing.T, out, slug, version, repo string) {
	t.Helper()
	rows := parseListRepoJSON(t, out)
	expectedSource := "Repo: " + repo
	for _, row := range rows {
		if row.Name == slug {
			assert.Equal(t, version, row.Version, "list --repo version for %q", slug)
			assert.Equal(t, expectedSource, row.Source, "list --repo source for %q", slug)
			return
		}
	}
	t.Fatalf("list --repo JSON did not contain plugin %q; output: %s", slug, out)
}

func parseListLocalJSONArray(t *testing.T, out string) []map[string]any {
	t.Helper()
	var rows []map[string]any
	require.NoError(t, json.Unmarshal([]byte(extractJSONObjectOrArray(t, out)), &rows),
		"single-harness list output must be a JSON array")
	return rows
}

func parseListLocalJSONObject(t *testing.T, out string) map[string][]map[string]any {
	t.Helper()
	var byHarness map[string][]map[string]any
	require.NoError(t, json.Unmarshal([]byte(extractJSONObjectOrArray(t, out)), &byHarness),
		"multi-harness list output must be a JSON object")
	return byHarness
}

func assertListLocalRowCountEqual(t *testing.T, out string, harnesses []string, limit int) {
	t.Helper()
	if len(harnesses) == 1 {
		rows := parseListLocalJSONArray(t, out)
		assert.Equal(t, limit, len(rows),
			"list --harness --limit=%d must return exactly %d rows when more plugins are installed", limit, limit)
		return
	}
	byHarness := parseListLocalJSONObject(t, out)
	for _, harness := range harnesses {
		rows, found := byHarness[harness]
		require.True(t, found, "list --json should contain harness %q", harness)
		assert.Equal(t, limit, len(rows),
			"list --harness --limit=%d must return exactly %d rows for harness %q", limit, limit, harness)
	}
}

type agentPluginsInstallSummary struct {
	Slug    string `json:"slug"`
	Version string `json:"version"`
	Results []struct {
		Agent  string `json:"agent"`
		Status string `json:"status"`
	} `json:"results"`
}

func assertInstallSummaryJSON(t *testing.T, out, slug, version string) {
	t.Helper()
	var summary agentPluginsInstallSummary
	require.NoError(t, json.Unmarshal([]byte(extractJSONObjectOrArray(t, out)), &summary),
		"install/update --format=json must emit a summary object")
	assert.Equal(t, slug, summary.Slug)
	assert.Equal(t, version, summary.Version)
	require.NotEmpty(t, summary.Results, "summary must include at least one result row")
}

func assertUpdateAllSummaryJSONContains(t *testing.T, out, slug string) {
	t.Helper()
	var summary struct {
		Results []struct {
			Name   string `json:"name"`
			Status string `json:"status"`
		} `json:"results"`
	}
	require.NoError(t, json.Unmarshal([]byte(extractJSONObjectOrArray(t, out)), &summary),
		"update --all --format=json must emit a results object")
	for _, row := range summary.Results {
		if row.Name == slug {
			return
		}
	}
	t.Fatalf("update --all JSON did not contain plugin %q; output: %s", slug, out)
}

// ---------------------------------------------------------------------------
// Flag validation
// ---------------------------------------------------------------------------

// TestAgentPluginsInvalidFormatFlag verifies that specifying an unsupported
// --format value falls back to table output without error.
// The list command treats any non-"json" format as table, so this is not an error path.
func TestAgentPluginsInvalidFormatFlag(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	err := runAgentPluginsCmd(t,
		"list",
		"--repo="+tests.AgentPluginsLocalRepo,
		"--format=invalid-format-value",
	)
	assert.NoError(t, err, "list with unrecognised --format falls back to table output")
}

// TestAgentPluginsUnknownFlag verifies that passing an unrecognised flag to
// any subcommand results in a non-zero exit (error), not a panic or silent
// ignore.
func TestAgentPluginsUnknownFlag(t *testing.T) {
	initAgentPluginsTest(t)
	defer cleanAgentPluginsTest()

	subcommands := []string{"publish", "install", "update", "delete", "list", "search"}
	for _, sub := range subcommands {
		t.Run(sub, func(t *testing.T) {
			err := runAgentPluginsCmd(t, sub, "--this-flag-does-not-exist=xyz")
			assert.Error(t, err,
				"subcommand %q must reject unknown flags", sub)
		})
	}
}

// ---------------------------------------------------------------------------
// Test fixture helpers
// ---------------------------------------------------------------------------

// createMinimalPlugin writes a minimal plugin.json with the given slug and version
// (either of which may be intentionally invalid for error-path tests). Raw JSON is
// used so values are stored verbatim without json.Marshal normalisation.
func createMinimalPlugin(t *testing.T, slug, version string) string {
	t.Helper()
	dir := t.TempDir()
	raw := fmt.Sprintf(`{"name":%q,"version":%q,"description":"test"}`, slug, version)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "plugin.json"), []byte(raw), 0644)) // #nosec G306 -- test fixture
	return dir
}
