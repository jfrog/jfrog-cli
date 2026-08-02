package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	biutils "github.com/jfrog/build-info-go/utils"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	coretests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	clientTestUtils "github.com/jfrog/jfrog-client-go/utils/tests"

	"github.com/jfrog/jfrog-cli/inttestutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Cargo integration tests. Scenario numbers (#N) refer to CARGO_TEST_PLAN.md.
//
// These tests drive the native `jf cargo` FlexPack command against a live Artifactory.
// Build-info collection is enabled by JFROG_RUN_NATIVE=true (flexpack.IsFlexPackEnabled).
// The command buckets that COLLECT are:
//
//	deps    -> install, update, add, fetch, generate-lockfile, run, test, check
//	publish -> publish (deploys via native cargo publish + records the .crate via AQL)
//
// `build` and `package` intentionally collect NOTHING, so there are no tests asserting
// build-info for them (that would contradict the implementation — see the reconciled plan).
//
// Resolution goes through the auto-created remote repo (proxies crates.io); publishing
// targets the auto-created local repo. Both are pointed at via a .cargo/config.toml that
// each test writes into its copied fixture (the repo names carry a per-run timestamp).
const (
	cargoModuleType = "cargo"
	// Fixed cargo registry names written into each fixture's .cargo/config.toml.
	cargoResolveRegistry = "jfrog"       // -> remote repo (resolve)
	cargoDeployRegistry  = "jfrog-local" // -> local repo (publish)
)

// ==================== Initialization ====================

func initCargoTest(t *testing.T) {
	if !*tests.TestCargo {
		t.Skip("Skipping Cargo test. To run Cargo test add the '-test.cargo=true' option.")
	}
	require.True(t, isRepoExist(tests.CargoLocalRepo), "Cargo test local repository doesn't exist.")
	require.True(t, isRepoExist(tests.CargoRemoteRepo), "Cargo test remote repository doesn't exist.")
}

// ==================== Project / config helpers ====================

// cargoSparseIndex builds the Artifactory sparse index URL for a cargo repo.
func cargoSparseIndex(repo string) string {
	base := strings.TrimSuffix(serverDetails.ArtifactoryUrl, "/")
	return fmt.Sprintf("sparse+%s/api/cargo/%s/index/", base, repo)
}

// writeCargoConfig writes .cargo/config.toml into a copied fixture, redirecting crates.io
// to the test Artifactory: [registries.jfrog] resolves via the remote repo, [registries.jfrog-local]
// publishes to the local repo. `jf cargo` injects the Bearer token for both (host-matched).
func writeCargoConfig(t *testing.T, projectPath string) {
	dir := filepath.Join(projectPath, ".cargo")
	assert.NoError(t, os.MkdirAll(dir, 0o755))
	content := fmt.Sprintf(`[registries.%s]
index = "%s"

[registries.%s]
index = "%s"

[source.crates-io]
replace-with = "%s"
`,
		cargoResolveRegistry, cargoSparseIndex(tests.CargoRemoteRepo),
		cargoDeployRegistry, cargoSparseIndex(tests.CargoLocalRepo),
		cargoResolveRegistry)
	assert.NoError(t, os.WriteFile(filepath.Join(dir, "config.toml"), []byte(content), 0o644))
}

// createCargoProject copies a fixture from testdata/cargo/<projectName> into a fresh temp dir.
func createCargoProject(t *testing.T, outputFolder, projectName string) (string, func()) {
	projectSrc := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "cargo", projectName)
	tmpDir, cleanupCallback := coretests.CreateTempDirWithCallbackAndAssert(t)

	projectPath := filepath.Join(tmpDir, outputFolder)
	assert.NoError(t, biutils.CopyDir(projectSrc, projectPath, true, nil))
	return projectPath, cleanupCallback
}

// runInCargoProject prepares an isolated home dir (default server = test Artifactory), copies the
// fixture, writes .cargo/config.toml, and chdirs into it. It returns the JfrogCli client and a
// teardown func that restores home/cwd.
//
// Note: cargo is FlexPack-only (there is no legacy cargo build-info path), so build-info collection
// does NOT depend on JFROG_RUN_NATIVE. The env is only relevant for package managers that have both
// a legacy and a FlexPack path; it is intentionally not set here.
func runInCargoProject(t *testing.T, outputFolder, projectName string) (*coretests.JfrogCli, func()) {
	oldHomeDir, newHomeDir := prepareHomeDir(t)

	projectPath, cleanupProject := createCargoProject(t, outputFolder, projectName)
	writeCargoConfig(t, projectPath)

	wd, err := os.Getwd()
	assert.NoError(t, err)
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)

	teardown := func() {
		chdirCallback()
		cleanupProject()
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	}
	return coretests.NewJfrogCli(execMain, "jfrog", ""), teardown
}

// cargoInstallRoot returns a fresh temp dir to use as `cargo install --root` (keeps the installed
// binary out of the real ~/.cargo/bin), plus a cleanup callback.
func cargoInstallRoot(t *testing.T) (string, func()) {
	dir, cleanup := coretests.CreateTempDirWithCallbackAndAssert(t)
	return dir, cleanup
}

// ==================== Dependency collection (deps bucket) ====================

func TestCargoInstallCollectsDeps(t *testing.T) {
	// Scenarios: #14 install resolves deps from Artifactory; #22 deps captured in a cargo module;
	//            #36 dep module complete (scopes); #42 all deps carry sha256; dev-dep excluded.
	initCargoTest(t)

	jfrogCli, teardown := runInCargoProject(t, "cargo-install-deps", "simple")
	defer teardown()

	root, rootCleanup := cargoInstallRoot(t)
	defer rootCleanup()

	buildName := tests.CargoBuildName
	buildNumber := "1"

	require.NoError(t, jfrogCli.Exec("cargo", "install", "--path", ".", "--root", root, "--force",
		"--build-name="+buildName, "--build-number="+buildNumber))

	require.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))
	defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	require.NoError(t, err)
	require.True(t, found, "build-info should be found after cargo install")
	require.NotEmpty(t, publishedBuildInfo.BuildInfo.Modules)

	module := publishedBuildInfo.BuildInfo.Modules[0]
	assert.Equal(t, cargoModuleType, string(module.Type), "module type should be cargo")
	require.NotEmpty(t, module.Dependencies, "install should capture dependencies")

	// The dev-dependency `rand` and its EXCLUSIVE transitive subtree must all be excluded — a crate
	// reachable only through a dev-dependency does not belong in a non-test build's build-info.
	devOnly := []string{"rand-", "rand_core", "rand_chacha", "getrandom", "ppv-lite86"}
	scopes := map[string][]string{}
	sha256Count := 0
	for _, dep := range module.Dependencies {
		scopes[dep.Id] = dep.Scopes
		if dep.Sha256 != "" {
			sha256Count++
		}
		for _, d := range devOnly {
			assert.NotContains(t, dep.Id, d, "dev-dependency (or its exclusive transitive) must be excluded: %s", dep.Id)
		}
	}
	// Scopes are cargo's own dep-kind names verbatim ("normal"/"build"/"dev") — no synthesized
	// "prod" or "transitive" markers, matching the Q3 bughunt correction: cargo's normal
	// dependency (kind: "") is surfaced as "normal", not "prod", and indirect deps keep the same
	// kind as their reaching edge rather than being relabelled as "transitive".
	assertScope(t, scopes, "serde_json-", "normal")
	// cc is the build-dependency.
	assertScope(t, scopes, "cc-", "build")
	// At least one transitive normal dep (e.g. serde_core/itoa/memchr/zmij), reached only via
	// serde_json — carries "normal" scope, not "transitive".
	assert.True(t, hasScope(scopes, "normal"), "expected at least one normal-scoped transitive dependency")
	// #42 — every dependency should carry a sha256 (local cache + AQL enrichment).
	assert.Equal(t, len(module.Dependencies), sha256Count,
		"all dependencies should have sha256 (got %d/%d)", sha256Count, len(module.Dependencies))
}

// ==================== Publish (publish bucket) ====================

func TestCargoPublish(t *testing.T) {
	// Scenarios: #6 publish uploads the .crate; #7 crate lands at crates/<name>/<name>-<ver>.crate;
	//            #21 artifact captured in module; #24 build props stamped; #40 sha256 stored.
	initCargoTest(t)

	jfrogCli, teardown := runInCargoProject(t, "cargo-publish", "simple")
	defer teardown()

	buildName := tests.CargoBuildName
	buildNumber := "2"

	require.NoError(t, jfrogCli.Exec("cargo", "publish", "--registry", cargoDeployRegistry, "--allow-dirty",
		"--build-name="+buildName, "--build-number="+buildNumber))

	require.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))
	defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	require.NoError(t, err)
	require.True(t, found)
	require.NotEmpty(t, publishedBuildInfo.BuildInfo.Modules)

	module := publishedBuildInfo.BuildInfo.Modules[0]
	assert.Equal(t, cargoModuleType, string(module.Type))
	require.NotEmpty(t, module.Artifacts, "publish should record the .crate artifact")
	art := module.Artifacts[0]
	assert.Equal(t, "crate", art.Type, "artifact type should be crate")
	assert.Contains(t, art.Path, "crates/cli-cargo-lib/cli-cargo-lib-1.0.0.crate",
		"crate must follow crates/<name>/<name>-<version>.crate layout")
	assert.NotEmpty(t, art.Sha256, "crate artifact must have sha256")
}

// ==================== Workspace / multi-module ====================

func TestCargoWorkspaceModules(t *testing.T) {
	// Scenarios: #38/#75 — a workspace produces one build-info module per member, each with its
	// own dependencies. Only install/publish collect build-info; member-a has a binary, so
	// `cargo install --path member-a` drives collection while `cargo metadata` still enumerates the
	// whole workspace (both members become modules).
	initCargoTest(t)

	jfrogCli, teardown := runInCargoProject(t, "cargo-workspace", "workspace")
	defer teardown()

	root, rootCleanup := cargoInstallRoot(t)
	defer rootCleanup()

	buildName := tests.CargoBuildName
	buildNumber := "3"

	require.NoError(t, jfrogCli.Exec("cargo", "install", "--path", "member-a", "--root", root, "--force",
		"--build-name="+buildName, "--build-number="+buildNumber))

	require.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))
	defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	require.NoError(t, err)
	require.True(t, found)

	moduleIds := map[string]bool{}
	for _, m := range publishedBuildInfo.BuildInfo.Modules {
		moduleIds[m.Id] = true
		assert.Equal(t, cargoModuleType, string(m.Type))
	}
	assert.True(t, moduleIds["cli-cargo-member-a:1.0.0"], "expected member-a module, got %v", moduleIds)
	assert.True(t, moduleIds["cli-cargo-member-b:1.0.0"], "expected member-b module, got %v", moduleIds)
}

func TestCargoWorkspacePublishMember(t *testing.T) {
	// Scenarios: #39/#76 — selective member publish (`-p <member>`) records the crate as an
	// artifact under THAT member's module (not module 0).
	initCargoTest(t)

	jfrogCli, teardown := runInCargoProject(t, "cargo-ws-publish", "workspace")
	defer teardown()

	buildName := tests.CargoBuildName
	buildNumber := "4"

	// member-b has no path deps, so it publishes cleanly.
	require.NoError(t, jfrogCli.Exec("cargo", "publish", "-p", "cli-cargo-member-b",
		"--registry", cargoDeployRegistry, "--allow-dirty",
		"--build-name="+buildName, "--build-number="+buildNumber))

	require.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))
	defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	require.NoError(t, err)
	require.True(t, found)

	// The artifact must be attached to member-b's module, and member-a's module must have none.
	var memberBArtifacts, memberAArtifacts int
	for _, m := range publishedBuildInfo.BuildInfo.Modules {
		switch m.Id {
		case "cli-cargo-member-b:1.0.0":
			memberBArtifacts = len(m.Artifacts)
		case "cli-cargo-member-a:1.0.0":
			memberAArtifacts = len(m.Artifacts)
		}
	}
	assert.Equal(t, 1, memberBArtifacts, "member-b module should carry the published .crate")
	assert.Equal(t, 0, memberAArtifacts, "member-a module should carry no artifact")
}

// ==================== Build-info flag combinations ====================

func TestCargoBuildInfoFlags(t *testing.T) {
	// Scenarios: #26 both set -> build-info captured; #27 name-only and #28 number-only ->
	// the CLI rejects them (build-name/build-number cannot be provided separately), so no
	// build-info is possible; #29 neither -> cargo runs as a pass-through with no build-info.
	initCargoTest(t)

	cases := []struct {
		name        string
		buildName   string
		buildNumber string
		expectErr   bool // CLI rejects name/number given alone
		expectBI    bool // build-info captured
	}{
		{"both-set", tests.CargoBuildName, "10", false, true},
		{"name-only", tests.CargoBuildName, "", true, false},
		{"number-only", "", "11", true, false},
		{"neither", "", "", false, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			jfrogCli, teardown := runInCargoProject(t, "cargo-flags", "simple")
			defer teardown()
			root, rootCleanup := cargoInstallRoot(t)
			defer rootCleanup()

			args := []string{"cargo", "install", "--path", ".", "--root", root, "--force"}
			if tc.buildName != "" {
				args = append(args, "--build-name="+tc.buildName)
			}
			if tc.buildNumber != "" {
				args = append(args, "--build-number="+tc.buildNumber)
			}
			err := jfrogCli.Exec(args...)
			if tc.expectErr {
				assert.Error(t, err, "CLI should reject build-name/build-number given separately")
				return
			}
			require.NoError(t, err)

			if tc.expectBI {
				require.NoError(t, artifactoryCli.Exec("bp", tc.buildName, tc.buildNumber))
				defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tc.buildName, artHttpDetails)
				_, found, err := tests.GetBuildInfo(serverDetails, tc.buildName, tc.buildNumber)
				require.NoError(t, err)
				assert.True(t, found, "build-info should be captured when both flags are set")
			}
		})
	}
}

func TestCargoModuleOverride(t *testing.T) {
	// Scenario: #30 — --module overrides the build-info module id (mirrors nix/go). Without it the
	// module id is the crate's "<name>:<version>"; with it, the id is the supplied module name.
	initCargoTest(t)

	jfrogCli, teardown := runInCargoProject(t, "cargo-module", "simple")
	defer teardown()

	root, rootCleanup := cargoInstallRoot(t)
	defer rootCleanup()

	buildName := tests.CargoBuildName
	buildNumber := "12"
	const customModule = "my-custom-cargo-module"

	require.NoError(t, jfrogCli.Exec("cargo", "install", "--path", ".", "--root", root, "--force",
		"--module="+customModule, "--build-name="+buildName, "--build-number="+buildNumber))

	require.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))
	defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	require.NoError(t, err)
	require.True(t, found)
	require.NotEmpty(t, publishedBuildInfo.BuildInfo.Modules)

	module := publishedBuildInfo.BuildInfo.Modules[0]
	assert.Equal(t, customModule, module.Id, "--module should override the module id")
	assert.Equal(t, cargoModuleType, string(module.Type))
}

// ==================== Flag / server validation ====================

func TestCargoUnknownServerId(t *testing.T) {
	// Scenario: #49 — an unknown --server-id produces a clear error.
	initCargoTest(t)

	jfrogCli, teardown := runInCargoProject(t, "cargo-bad-server", "simple")
	defer teardown()

	err := jfrogCli.Exec("cargo", "install", "--path", ".", "--server-id=nonexistent-server-xyz")
	assert.Error(t, err, "unknown --server-id should fail")
}

// ==================== helpers ====================

func assertScope(t *testing.T, scopes map[string][]string, idPrefix, wantScope string) {
	for id, sc := range scopes {
		if strings.HasPrefix(id, idPrefix) {
			assert.Contains(t, sc, wantScope, "dependency %s should have scope %s", id, wantScope)
			return
		}
	}
	assert.Failf(t, "dependency not found", "no dependency with prefix %q (scopes: %v)", idPrefix, scopes)
}

func hasScope(scopes map[string][]string, wantScope string) bool {
	for _, sc := range scopes {
		for _, s := range sc {
			if s == wantScope {
				return true
			}
		}
	}
	return false
}
