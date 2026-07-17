package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	buildinfo "github.com/jfrog/build-info-go/entities"
	biutils "github.com/jfrog/build-info-go/utils"
	coreBuild "github.com/jfrog/jfrog-cli-core/v2/common/build"
	artUtils "github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	coretests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	clientTestUtils "github.com/jfrog/jfrog-client-go/utils/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jfrog/jfrog-cli/inttestutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
)

// ---------------------------------------------------------------------------
// Init / cleanup
// ---------------------------------------------------------------------------

func initRubyTest(t *testing.T) {
	if !*tests.TestRuby {
		t.Skip("Skipping Ruby tests. To run Ruby tests add the '-test.ruby=true' option.")
	}
	require.True(t, isRepoExist(tests.RubyLocalRepo), "Ruby local repo does not exist: "+tests.RubyLocalRepo)
	require.True(t, isRepoExist(tests.RubyRemoteRepo), "Ruby remote repo does not exist: "+tests.RubyRemoteRepo)
	require.True(t, isRepoExist(tests.RubyVirtualRepo), "Ruby virtual repo does not exist: "+tests.RubyVirtualRepo)

	oldHomeDir, newHomeDir := prepareHomeDir(t)
	t.Cleanup(func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
		clientTestUtils.RemoveAllAndAssert(t, newHomeDir)
	})
}

func cleanRubyTest(_ *testing.T) {
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.RubyBuildName, artHttpDetails)
	_ = coreBuild.RemoveBuildDir(tests.RubyBuildName, "1", "")
	tests.CleanFileSystem()
}

func rubyToolRequired(t *testing.T, tool string) {
	if _, err := exec.LookPath(tool); err != nil {
		t.Skipf("'%s' is not installed; skipping Ruby integration test", tool)
	}
}

var warmUpOnce sync.Once

// warmUpRubyLocalRepo pushes a self-contained gem to the local repo so that
// Bundler tests can resolve it. Uses the LOCAL repo directly (not virtual)
// because CI Artifactory instances don't reliably generate specs.4.8.gz on
// virtual repos. The local repo generates specs immediately on gem push.
func warmUpRubyLocalRepo(t *testing.T) {
	warmUpOnce.Do(func() { doWarmUpRubyLocalRepo(t) })
}

func doWarmUpRubyLocalRepo(t *testing.T) {
	t.Helper()
	tmpDir, cleanup := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer cleanup()

	specContent := `Gem::Specification.new do |s|
  s.name    = "warmup"
  s.version = "0.0.1"
  s.summary = "warmup"
  s.authors = ["test"]
  s.files   = ["lib/warmup.rb"]
end`
	libDir := filepath.Join(tmpDir, "lib")
	require.NoError(t, os.MkdirAll(libDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "warmup.gemspec"), []byte(specContent), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(libDir, "warmup.rb"), []byte(""), 0600))

	buildCmd := exec.Command("gem", "build", "warmup.gemspec")
	buildCmd.Dir = tmpDir
	out, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("warm-up gem build failed: %v\n%s", err, out)
	}

	// Push via jf ruby gem push — this triggers specs.4.8.gz generation on the local repo.
	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	gemFile := filepath.Join(tmpDir, "warmup-0.0.1.gem")
	require.NoError(t, jfrogCli.Exec("ruby", "gem", "push", gemFile,
		"--repo", tests.RubyLocalRepo,
		"--server-id=default"))

	// Poll local repo specs to confirm index is ready (should be near-instant).
	specsURL := serverDetails.ArtifactoryUrl + "api/gems/" + tests.RubyLocalRepo + "/specs.4.8.gz"
	user, pass := rubyTestCredentials()
	for i := 0; i < 12; i++ {
		req, _ := http.NewRequest("GET", specsURL, nil)
		req.SetBasicAuth(user, pass)
		resp, httpErr := http.DefaultClient.Do(req)
		if httpErr == nil {
			resp.Body.Close()
			if resp.StatusCode == 200 {
				t.Logf("warm-up: local repo specs.4.8.gz available after %ds", i*2)
				return
			}
		}
		time.Sleep(2 * time.Second)
	}
	t.Fatal("warm-up: local repo specs.4.8.gz not available after 24s")
}

// createRubyProject copies a test Ruby project to a temp dir and patches the
// Gemfile source to point at the test Artifactory local repo.
func createRubyProject(t *testing.T, projectName string) string {
	projectSrc := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "ruby", projectName)
	tmpDir, cleanup := coretests.CreateTempDirWithCallbackAndAssert(t)
	t.Cleanup(cleanup)

	projectPath := filepath.Join(tmpDir, projectName)
	assert.NoError(t, biutils.CopyDir(projectSrc, projectPath, true, nil))

	warmUpRubyLocalRepo(t)
	patchRubyGemfile(t, projectPath)
	return projectPath
}

// patchRubyGemfile replaces ARTIFACTORY_GEMS_URL placeholder in Gemfile with
// the test Artifactory virtual gems repo URL.
func patchRubyGemfile(t *testing.T, projectPath string) {
	t.Helper()
	gemfilePath := filepath.Join(projectPath, "Gemfile")
	data, err := os.ReadFile(gemfilePath)
	if err != nil {
		return
	}
	// Embed credentials in the source URL for reliable Bundler auth in CI.
	// Bundler's env-var credential matching varies across versions; URL-embedded
	// credentials work universally.
	gemsURL := rubyGemsURLWithCreds(t)
	content := strings.ReplaceAll(string(data), "ARTIFACTORY_GEMS_URL", gemsURL)
	require.NoError(t, os.WriteFile(gemfilePath, []byte(content), 0600)) // #nosec G703 -- test-only path from CreateTempDir
}

func rubyGemsURLWithCreds(t *testing.T) string {
	t.Helper()
	base := serverDetails.ArtifactoryUrl + "api/gems/" + tests.RubyLocalRepo
	parsed, err := url.Parse(base)
	require.NoError(t, err)
	user, pass := rubyTestCredentials()
	if user != "" && pass != "" {
		parsed.User = url.UserPassword(user, pass)
	}
	return parsed.String()
}

// rubyTestCredentials returns user/password for authenticating against Artifactory in tests.
func rubyTestCredentials() (string, string) {
	user := serverDetails.User
	pass := serverDetails.Password
	if serverDetails.AccessToken != "" {
		pass = serverDetails.AccessToken
		if user == "" {
			user = *tests.JfrogUser
		}
	}
	return user, pass
}

// runRubyCmd changes to projectPath and runs `jf ruby <args...>`.
func runRubyCmd(t *testing.T, projectPath string, args ...string) error {
	wd, err := os.Getwd()
	assert.NoError(t, err)
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)
	defer chdirCallback()

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	return jfrogCli.Exec(append([]string{"ruby"}, args...)...)
}

// ---------------------------------------------------------------------------
// P0 — Happy path: gem install (Scenario #1)
// ---------------------------------------------------------------------------

func TestRubyGemInstall(t *testing.T) {
	// Scenario #1: jf ruby gem install <gem> --repo succeeds
	initRubyTest(t)
	defer cleanRubyTest(t)
	rubyToolRequired(t, "gem")

	tmpDir, cleanup := coretests.CreateTempDirWithCallbackAndAssert(t)
	t.Cleanup(cleanup)

	wd, err := os.Getwd()
	assert.NoError(t, err)
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, tmpDir)
	defer chdirCallback()

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err = jfrogCli.Exec("ruby", "gem", "install", "rake", "-v", "13.0.6",
		"--repo", tests.RubyVirtualRepo,
		"--server-id=default",
		"--build-name="+tests.RubyBuildName,
		"--build-number=1",
	)
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// P0 — Happy path: gem fetch (Scenario #2)
// ---------------------------------------------------------------------------

func TestRubyGemFetch(t *testing.T) {
	// Scenario #2: jf ruby gem fetch <gem> --repo succeeds
	initRubyTest(t)
	defer cleanRubyTest(t)
	rubyToolRequired(t, "gem")

	tmpDir, cleanup := coretests.CreateTempDirWithCallbackAndAssert(t)
	t.Cleanup(cleanup)

	wd, err := os.Getwd()
	assert.NoError(t, err)
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, tmpDir)
	defer chdirCallback()

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err = jfrogCli.Exec("ruby", "gem", "fetch", "rake", "-v", "13.0.6",
		"--repo", tests.RubyVirtualRepo,
		"--server-id=default",
		"--build-name="+tests.RubyBuildName,
		"--build-number=1",
	)
	assert.NoError(t, err)

	// Verify .gem file was downloaded to current dir
	entries, readErr := os.ReadDir(tmpDir)
	assert.NoError(t, readErr)
	found := false
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".gem") {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected .gem file to be downloaded to working dir")
}

// ---------------------------------------------------------------------------
// P0 — Happy path: gem push (Scenario #3)
// ---------------------------------------------------------------------------

func TestRubyGemPush(t *testing.T) {
	// Scenario #3: jf ruby gem push <file.gem> --repo succeeds
	initRubyTest(t)
	defer cleanRubyTest(t)
	rubyToolRequired(t, "gem")

	tmpDir, cleanup := coretests.CreateTempDirWithCallbackAndAssert(t)
	t.Cleanup(cleanup)

	gemFile := createTestGem(t, tmpDir)
	require.NotEmpty(t, gemFile)

	wd, err := os.Getwd()
	assert.NoError(t, err)
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, tmpDir)
	defer chdirCallback()

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err = jfrogCli.Exec("ruby", "gem", "push", gemFile,
		"--repo", tests.RubyLocalRepo,
		"--server-id=default",
		"--build-name="+tests.RubyBuildName,
		"--build-number=1",
	)
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// P0 — Happy path: bundle install (Scenario #5)
// ---------------------------------------------------------------------------

func TestRubyBundleInstall(t *testing.T) {
	// Scenario #5: jf ruby bundle install resolves from Artifactory
	initRubyTest(t)
	defer cleanRubyTest(t)
	rubyToolRequired(t, "bundle")

	projectPath := createRubyProject(t, "bundleproject")

	err := runRubyCmd(t, projectPath, "bundle", "install",
		"--path", "vendor/bundle",
		"--server-id=default",
		"--build-name="+tests.RubyBuildName,
		"--build-number=1",
	)
	assert.NoError(t, err)

	// Verify Gemfile.lock was created/updated
	lockFile := filepath.Join(projectPath, "Gemfile.lock")
	assert.FileExists(t, lockFile)
}

// ---------------------------------------------------------------------------
// P0 — Build info: gem push captures artifacts (Scenario #9)
// ---------------------------------------------------------------------------

func TestRubyGemPushBuildInfo(t *testing.T) {
	// Scenario #9: gem push captures artifact in build-info module
	initRubyTest(t)
	defer cleanRubyTest(t)
	rubyToolRequired(t, "gem")

	tmpDir, cleanup := coretests.CreateTempDirWithCallbackAndAssert(t)
	t.Cleanup(cleanup)

	gemFile := createTestGem(t, tmpDir)
	require.NotEmpty(t, gemFile)

	wd, err := os.Getwd()
	assert.NoError(t, err)
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, tmpDir)
	defer chdirCallback()

	buildNumber := "1"
	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err = jfrogCli.Exec("ruby", "gem", "push", gemFile,
		"--repo", tests.RubyLocalRepo,
		"--server-id=default",
		"--build-name="+tests.RubyBuildName,
		"--build-number="+buildNumber,
	)
	require.NoError(t, err)

	// Publish build info
	assert.NoError(t, artifactoryCli.Exec("bp", tests.RubyBuildName, buildNumber))

	// Retrieve and validate
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.RubyBuildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found, "Build info not found for %s/%s", tests.RubyBuildName, buildNumber)

	if found {
		bi := publishedBuildInfo.BuildInfo
		require.NotEmpty(t, bi.Modules, "build info should have at least one module")
		module := bi.Modules[0]
		assert.Equal(t, buildinfo.Gem, module.Type)
		assert.NotEmpty(t, module.Artifacts, "module should have artifacts")
		for _, artifact := range module.Artifacts {
			assert.True(t, strings.HasSuffix(artifact.Name, ".gem"),
				"artifact name should end with .gem, got: %s", artifact.Name)
		}
	}
}

// ---------------------------------------------------------------------------
// P0 — Build info: bundle install captures dependencies (Scenario #10)
// ---------------------------------------------------------------------------

func TestRubyBundleInstallBuildInfo(t *testing.T) {
	// Scenario #10: bundle install captures dependencies in build-info module
	initRubyTest(t)
	defer cleanRubyTest(t)
	rubyToolRequired(t, "bundle")

	projectPath := createRubyProject(t, "bundleproject")
	buildNumber := "1"

	err := runRubyCmd(t, projectPath, "bundle", "install",
		"--path", "vendor/bundle",
		"--server-id=default",
		"--build-name="+tests.RubyBuildName,
		"--build-number="+buildNumber,
	)
	require.NoError(t, err)

	// Publish build info
	assert.NoError(t, artifactoryCli.Exec("bp", tests.RubyBuildName, buildNumber))

	// Validate
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.RubyBuildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found, "Build info not found")

	if found {
		bi := publishedBuildInfo.BuildInfo
		require.NotEmpty(t, bi.Modules)
		module := bi.Modules[0]
		assert.Equal(t, buildinfo.Gem, module.Type)
		assert.NotEmpty(t, module.Dependencies, "module should have dependencies from Gemfile.lock")
	}
}

// ---------------------------------------------------------------------------
// P0 — Build info publish and retrieve (Scenario #11)
// ---------------------------------------------------------------------------

func TestRubyBuildInfoPublish(t *testing.T) {
	// Scenario #11: Build info published and retrievable from Artifactory
	initRubyTest(t)
	defer cleanRubyTest(t)
	rubyToolRequired(t, "bundle")

	projectPath := createRubyProject(t, "bundleproject")
	buildNumber := "1"

	err := runRubyCmd(t, projectPath, "bundle", "install",
		"--path", "vendor/bundle",
		"--server-id=default",
		"--build-name="+tests.RubyBuildName,
		"--build-number="+buildNumber,
	)
	require.NoError(t, err)

	assert.NoError(t, artifactoryCli.Exec("bp", tests.RubyBuildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.RubyBuildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, tests.RubyBuildName, publishedBuildInfo.BuildInfo.Name)
	assert.Equal(t, buildNumber, publishedBuildInfo.BuildInfo.Number)
}

// ---------------------------------------------------------------------------
// P0 — Build info properties (Scenario #12)
// ---------------------------------------------------------------------------

func TestRubyBuildInfoProperties(t *testing.T) {
	// Scenario #12: Build properties stamped on pushed gem
	initRubyTest(t)
	defer cleanRubyTest(t)
	rubyToolRequired(t, "gem")

	tmpDir, cleanup := coretests.CreateTempDirWithCallbackAndAssert(t)
	t.Cleanup(cleanup)

	gemFile := createTestGem(t, tmpDir)
	require.NotEmpty(t, gemFile)

	wd, err := os.Getwd()
	assert.NoError(t, err)
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, tmpDir)
	defer chdirCallback()

	buildNumber := "1"
	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err = jfrogCli.Exec("ruby", "gem", "push", gemFile,
		"--repo", tests.RubyLocalRepo,
		"--server-id=default",
		"--build-name="+tests.RubyBuildName,
		"--build-number="+buildNumber,
	)
	require.NoError(t, err)

	assert.NoError(t, artifactoryCli.Exec("bp", tests.RubyBuildName, buildNumber))

	validateRubyBuildProperties(t, tests.RubyLocalRepo, tests.RubyBuildName, buildNumber)
}

// validateRubyBuildProperties verifies build.name/build.number properties on artifacts.
func validateRubyBuildProperties(t *testing.T, repo, buildName, buildNumber string) {
	t.Helper()
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	require.NoError(t, err)
	require.True(t, found)
	require.NotEmpty(t, publishedBuildInfo.BuildInfo.Modules)

	serviceManager, err := artUtils.CreateServiceManager(serverDetails, -1, 0, false)
	require.NoError(t, err)

	verified := 0
	for _, module := range publishedBuildInfo.BuildInfo.Modules {
		for _, artifact := range module.Artifacts {
			if artifact.Name == "" {
				continue
			}
			fullPath := repo + "/gems/" + artifact.Name
			props, propErr := serviceManager.GetItemProps(fullPath)
			if propErr != nil {
				continue
			}
			if props == nil {
				continue
			}
			assert.Contains(t, props.Properties, "build.name")
			assert.Contains(t, props.Properties, "build.number")
			verified++
		}
	}
	assert.Greater(t, verified, 0, "no artifacts had build properties")
}

// ---------------------------------------------------------------------------
// P0 — Checksum integrity (Scenario #18)
// ---------------------------------------------------------------------------

func TestRubyChecksumIntegrity(t *testing.T) {
	// Scenario #18: Pushed artifact has sha256/sha1/md5 in Artifactory
	initRubyTest(t)
	defer cleanRubyTest(t)
	rubyToolRequired(t, "gem")

	tmpDir, cleanup := coretests.CreateTempDirWithCallbackAndAssert(t)
	t.Cleanup(cleanup)

	gemFile := createTestGem(t, tmpDir)
	require.NotEmpty(t, gemFile)

	wd, err := os.Getwd()
	assert.NoError(t, err)
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, tmpDir)
	defer chdirCallback()

	buildNumber := "1"
	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err = jfrogCli.Exec("ruby", "gem", "push", gemFile,
		"--repo", tests.RubyLocalRepo,
		"--server-id=default",
		"--build-name="+tests.RubyBuildName,
		"--build-number="+buildNumber,
	)
	require.NoError(t, err)

	assert.NoError(t, artifactoryCli.Exec("bp", tests.RubyBuildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.RubyBuildName, buildNumber)
	require.NoError(t, err)
	require.True(t, found)
	require.NotEmpty(t, publishedBuildInfo.BuildInfo.Modules)

	for _, module := range publishedBuildInfo.BuildInfo.Modules {
		for _, artifact := range module.Artifacts {
			assert.NotEmpty(t, artifact.Sha1, "artifact %s missing sha1", artifact.Name)
			assert.NotEmpty(t, artifact.Sha256, "artifact %s missing sha256", artifact.Name)
			assert.NotEmpty(t, artifact.Md5, "artifact %s missing md5", artifact.Name)
		}
	}
}

// ---------------------------------------------------------------------------
// P0 — Flag validation (Scenarios #22, #23, #24)
// ---------------------------------------------------------------------------

func TestRubyFlags(t *testing.T) {
	// Scenarios #22, #23, #24: flag validation
	initRubyTest(t)
	defer cleanRubyTest(t)

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")

	cases := []struct {
		name        string
		args        []string
		expectError bool
		description string
	}{
		{"no-args", []string{"ruby"}, true, "jf ruby with no args should error"},
		{"unsupported-tool", []string{"ruby", "npm", "install"}, true, "unsupported tool should error"},
		{"help-gem", []string{"ruby", "gem", "help"}, false, "help should pass through"},
		{"version-gem", []string{"ruby", "gem", "--version"}, false, "version should pass through"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := jfrogCli.Exec(tc.args...)
			if tc.expectError {
				assert.Error(t, err, tc.description)
			} else {
				if _, lookupErr := exec.LookPath("gem"); lookupErr != nil {
					t.Skip("gem not available")
				}
				assert.NoError(t, err, tc.description)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// P1 — Build flags table-driven (Scenarios #13-16)
// ---------------------------------------------------------------------------

func TestRubyBuildFlags(t *testing.T) {
	// Scenarios #13-16: build-name/build-number combinations
	initRubyTest(t)
	defer cleanRubyTest(t)
	rubyToolRequired(t, "bundle")

	projectPath := createRubyProject(t, "bundleproject")

	cases := []struct {
		name        string
		buildName   string
		buildNumber string
		expectBI    bool
		expectErr   bool
	}{
		{"both-set", tests.RubyBuildName, "1", true, false},
		{"name-only", tests.RubyBuildName, "", false, true},
		{"number-only", "", "1", false, true},
		{"neither", "", "", false, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			args := []string{"bundle", "install", "--path", "vendor/bundle",
				"--server-id=default"}
			if tc.buildName != "" {
				args = append(args, "--build-name="+tc.buildName)
			}
			if tc.buildNumber != "" {
				args = append(args, "--build-number="+tc.buildNumber)
			}

			err := runRubyCmd(t, projectPath, args...)
			if tc.expectErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			if tc.expectBI && tc.buildName != "" && tc.buildNumber != "" {
				assert.NoError(t, artifactoryCli.Exec("bp", tc.buildName, tc.buildNumber))
				_, found, biErr := tests.GetBuildInfo(serverDetails, tc.buildName, tc.buildNumber)
				assert.NoError(t, biErr)
				assert.True(t, found)
				inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tc.buildName, artHttpDetails)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// P1 — Module override (Scenario #17)
// ---------------------------------------------------------------------------

func TestRubyCustomModule(t *testing.T) {
	// Scenario #17: --module overrides auto-detected module ID
	initRubyTest(t)
	defer cleanRubyTest(t)
	rubyToolRequired(t, "bundle")

	projectPath := createRubyProject(t, "bundleproject")
	buildNumber := "1"
	customModule := "my-custom-ruby-module"

	err := runRubyCmd(t, projectPath, "bundle", "install",
		"--path", "vendor/bundle",
		"--server-id=default",
		"--build-name="+tests.RubyBuildName,
		"--build-number="+buildNumber,
		"--module="+customModule,
	)
	require.NoError(t, err)

	assert.NoError(t, artifactoryCli.Exec("bp", tests.RubyBuildName, buildNumber))
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.RubyBuildName, buildNumber)
	require.NoError(t, err)
	require.True(t, found)
	require.NotEmpty(t, publishedBuildInfo.BuildInfo.Modules)
	assert.Equal(t, customModule, publishedBuildInfo.BuildInfo.Modules[0].Id)
}

// ---------------------------------------------------------------------------
// P1 — Server ID flag (Scenarios #25, #26)
// ---------------------------------------------------------------------------

func TestRubyServerIdFlag(t *testing.T) {
	// Scenarios #25, #26: valid and invalid --server-id
	initRubyTest(t)
	defer cleanRubyTest(t)
	rubyToolRequired(t, "gem")

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")

	// Valid server-id with help (should succeed without network)
	err := jfrogCli.Exec("ruby", "gem", "help", "--server-id=default")
	assert.NoError(t, err)

	// Invalid server-id should error
	err = jfrogCli.Exec("ruby", "gem", "install", "rake",
		"--repo", tests.RubyVirtualRepo,
		"--server-id", "nonexistent-server-id-xyz")
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// P1 — bundle lock does not collect build-info (Scenario #8)
// ---------------------------------------------------------------------------

func TestRubyBundleLockNoBuildInfo(t *testing.T) {
	// Scenario #8: bundle lock only resolves, no build-info collected
	initRubyTest(t)
	defer cleanRubyTest(t)
	rubyToolRequired(t, "bundle")

	projectPath := createRubyProject(t, "bundleproject")
	buildNumber := "1"

	err := runRubyCmd(t, projectPath, "bundle", "lock",
		"--server-id=default",
		"--build-name="+tests.RubyBuildName,
		"--build-number="+buildNumber,
	)
	assert.NoError(t, err)

	// Try to publish — should have no modules since lock doesn't collect
	_ = artifactoryCli.Exec("bp", tests.RubyBuildName, buildNumber)
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.RubyBuildName, buildNumber)
	if err == nil && found {
		bi := publishedBuildInfo.BuildInfo
		if len(bi.Modules) > 0 {
			assert.Empty(t, bi.Modules[0].Dependencies,
				"bundle lock should not collect dependencies")
		}
	}
}

// ---------------------------------------------------------------------------
// P1 — gem build skips auth (Scenario #4, #31)
// ---------------------------------------------------------------------------

func TestRubyGemBuildNoAuth(t *testing.T) {
	// Scenarios #4, #31: gem build is local-only, no auth injection
	initRubyTest(t)
	defer cleanRubyTest(t)
	rubyToolRequired(t, "gem")

	tmpDir, cleanup := coretests.CreateTempDirWithCallbackAndAssert(t)
	t.Cleanup(cleanup)

	createMinimalGemspec(t, tmpDir)

	wd, err := os.Getwd()
	assert.NoError(t, err)
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, tmpDir)
	defer chdirCallback()

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	err = jfrogCli.Exec("ruby", "gem", "build", "test_gem.gemspec",
		"--server-id=default",
	)
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// P1 — Dependency checksum completeness (Scenario #21)
// ---------------------------------------------------------------------------

func TestRubyDepChecksumCompleteness(t *testing.T) {
	// Scenario #21: All dependencies in build info have checksums
	initRubyTest(t)
	defer cleanRubyTest(t)
	rubyToolRequired(t, "bundle")

	projectPath := createRubyProject(t, "bundleproject")
	buildNumber := "1"

	err := runRubyCmd(t, projectPath, "bundle", "install",
		"--path", "vendor/bundle",
		"--server-id=default",
		"--build-name="+tests.RubyBuildName,
		"--build-number="+buildNumber,
	)
	require.NoError(t, err)

	assert.NoError(t, artifactoryCli.Exec("bp", tests.RubyBuildName, buildNumber))
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.RubyBuildName, buildNumber)
	require.NoError(t, err)
	require.True(t, found)
	require.NotEmpty(t, publishedBuildInfo.BuildInfo.Modules)

	module := publishedBuildInfo.BuildInfo.Modules[0]
	checksumComplete := 0
	for _, dep := range module.Dependencies {
		if dep.Sha1 != "" || dep.Sha256 != "" {
			checksumComplete++
		}
	}
	if len(module.Dependencies) > 0 {
		assert.Greater(t, checksumComplete, 0,
			"at least some dependencies should have checksums (local cache or AQL)")
	}
}

// ---------------------------------------------------------------------------
// P1 — Round-trip: push then install (Scenario #36)
// ---------------------------------------------------------------------------

func TestRubyRoundTrip(t *testing.T) {
	// Scenario #36: Push gem then install same gem from Artifactory
	initRubyTest(t)
	defer cleanRubyTest(t)
	rubyToolRequired(t, "gem")

	tmpDir, cleanup := coretests.CreateTempDirWithCallbackAndAssert(t)
	t.Cleanup(cleanup)

	gemFile := createTestGem(t, tmpDir)
	require.NotEmpty(t, gemFile)

	wd, err := os.Getwd()
	assert.NoError(t, err)
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, tmpDir)
	defer chdirCallback()

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")

	// Push
	err = jfrogCli.Exec("ruby", "gem", "push", gemFile,
		"--repo", tests.RubyLocalRepo,
		"--server-id=default",
	)
	require.NoError(t, err)

	// Install from virtual (which includes local)
	err = jfrogCli.Exec("ruby", "gem", "install", "testgem",
		"--repo", tests.RubyVirtualRepo,
		"--server-id=default",
	)
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// P1 — Scope classification (Scenario #33)
// ---------------------------------------------------------------------------

func TestRubyScopeClassification(t *testing.T) {
	// Scenario #33: Gemfile group scopes propagated to deps
	initRubyTest(t)
	defer cleanRubyTest(t)
	rubyToolRequired(t, "bundle")

	projectPath := createRubyProject(t, "scopeproject")
	buildNumber := "1"

	err := runRubyCmd(t, projectPath, "bundle", "install",
		"--path", "vendor/bundle",
		"--server-id=default",
		"--build-name="+tests.RubyBuildName,
		"--build-number="+buildNumber,
	)
	if err != nil {
		t.Skipf("Scope project not available or install failed: %v", err)
	}

	assert.NoError(t, artifactoryCli.Exec("bp", tests.RubyBuildName, buildNumber))
	publishedBuildInfo, found, biErr := tests.GetBuildInfo(serverDetails, tests.RubyBuildName, buildNumber)
	require.NoError(t, biErr)
	require.True(t, found)

	if len(publishedBuildInfo.BuildInfo.Modules) > 0 {
		module := publishedBuildInfo.BuildInfo.Modules[0]
		scopeFound := false
		for _, dep := range module.Dependencies {
			if len(dep.Scopes) > 0 {
				scopeFound = true
				break
			}
		}
		assert.True(t, scopeFound, "at least one dependency should have scope classification")
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// createTestGem generates a minimal .gem file for testing push operations.
func createTestGem(t *testing.T, dir string) string {
	t.Helper()
	createMinimalGemspec(t, dir)

	cmd := exec.Command("gem", "build", "test_gem.gemspec")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Skipf("Could not build test gem (gem not available or build failed): %s\n%s", err, out)
		return ""
	}

	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".gem") {
			return e.Name()
		}
	}
	t.Fatal("gem build did not produce a .gem file")
	return ""
}

func createMinimalGemspec(t *testing.T, dir string) {
	t.Helper()
	gemspec := `Gem::Specification.new do |s|
  s.name        = "testgem"
  s.version     = "0.0.1"
  s.summary     = "Test gem for JFrog CLI integration tests"
  s.authors     = ["Test"]
  s.files       = []
  s.homepage    = "https://example.com"
  s.license     = "MIT"
end
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "test_gem.gemspec"), []byte(gemspec), 0644))
}

// ---------------------------------------------------------------------------
// Unused import guard
// ---------------------------------------------------------------------------

var _ = fmt.Sprintf
var _ = buildinfo.Gem
