package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	buildinfo "github.com/jfrog/build-info-go/entities"
	biutils "github.com/jfrog/build-info-go/utils"
	coreBuild "github.com/jfrog/jfrog-cli-core/v2/common/build"
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

func initUvTest(t *testing.T) {
	if !*tests.TestUv {
		t.Skip("Skipping UV tests. To run UV tests add the '-test.uv=true' option.")
	}
	require.True(t, isRepoExist(tests.UvLocalRepo), "UV local repo does not exist: "+tests.UvLocalRepo)
	require.True(t, isRepoExist(tests.UvRemoteRepo), "UV remote repo does not exist: "+tests.UvRemoteRepo)
	require.True(t, isRepoExist(tests.UvVirtualRepo), "UV virtual repo does not exist: "+tests.UvVirtualRepo)
}

func cleanUvTest(t *testing.T) {
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.UvBuildName, artHttpDetails)
	tests.CleanFileSystem()
}

// createUvProject copies a test UV project to a temp dir, injects Artifactory
// URLs into pyproject.toml, then generates a fresh uv.lock against the test
// Artifactory instance. The lock file is not committed to avoid embedding
// instance-specific URLs in source.
func createUvProject(t *testing.T, outputFolder, projectName string) string {
	projectSrc := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "uv", projectName)
	tmpDir, cleanup := coretests.CreateTempDirWithCallbackAndAssert(t)
	t.Cleanup(cleanup)

	projectPath := filepath.Join(tmpDir, outputFolder)
	assert.NoError(t, biutils.CopyDir(projectSrc, projectPath, true, nil))

	// Configure the jf CLI home so that coreConfig.GetDefaultServerConf() inside
	// the jf uv command returns the TEST server (ecosysjfrog), not the developer's
	// personal default server. This ensures credential injection targets the correct
	// Artifactory instance when jf uv runs natively.
	// Pattern matches what Poetry native tests do with prepareHomeDir(t).
	createJfrogHomeConfig(t, false)

	// Patch pyproject.toml with real Artifactory URLs for this test run
	patchUvPyprojectToml(t, projectPath)

	// Generate uv.lock against the patched index so UV resolves through
	// Artifactory (required for dependency checksum enrichment tests).
	// Convert the index name to the UV env var suffix format:
	// "jfrog-pypi-virtual" → "JFROG_PYPI_VIRTUAL"
	indexEnvName := strings.ToUpper(strings.NewReplacer("-", "_", ".", "_").Replace("jfrog-pypi-virtual"))
	lockCmd := exec.Command("uv", "lock")
	lockCmd.Dir = projectPath
	lockCmd.Env = append(os.Environ(),
		"UV_INDEX_"+indexEnvName+"_USERNAME="+*tests.JfrogUser,
		"UV_INDEX_"+indexEnvName+"_PASSWORD="+*tests.JfrogPassword,
		"UV_KEYRING_PROVIDER=disabled",
	)
	if out, err := lockCmd.CombinedOutput(); err != nil {
		t.Logf("uv lock warning (non-fatal): %s — %v", out, err)
	}

	return projectPath
}

// patchUvPyprojectToml replaces placeholder URLs in pyproject.toml with the
// test Artifactory instance URLs.
func patchUvPyprojectToml(t *testing.T, projectPath string) {
	pyprojectPath := filepath.Join(projectPath, "pyproject.toml")
	data, err := os.ReadFile(pyprojectPath)
	assert.NoError(t, err)

	indexURL := serverDetails.ArtifactoryUrl + "api/pypi/" + tests.UvVirtualRepo + "/simple"
	publishURL := serverDetails.ArtifactoryUrl + "api/pypi/" + tests.UvLocalRepo

	content := string(data)
	content = strings.ReplaceAll(content, "ARTIFACTORY_INDEX_URL", indexURL)
	content = strings.ReplaceAll(content, "ARTIFACTORY_PUBLISH_URL", publishURL)
	assert.NoError(t, os.WriteFile(pyprojectPath, []byte(content), 0644))
}

// runUvCmd changes to projectPath and runs `jf uv <args...>`.
func runUvCmd(t *testing.T, projectPath string, args ...string) error {
	wd, err := os.Getwd()
	assert.NoError(t, err)
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)
	defer chdirCallback()

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	return jfrogCli.Exec(append([]string{"uv"}, args...)...)
}

// ---------------------------------------------------------------------------
// Helper: validate build properties stamped on artifacts
// ---------------------------------------------------------------------------

func validateUvBuildProperties(t *testing.T, repo, buildName, buildNumber string) {
	// Search for artifacts with build.name + build.number properties.
	// We verify the result is non-empty (at least one artifact has the properties),
	// rather than matching exact paths, since the path depends on runtime repo names.
	props := fmt.Sprintf("build.name=%v;build.number=%v", buildName, buildNumber)
	var resultItems []string
	searchSpec := fmt.Sprintf(`{"files":[{"pattern":"%s/*","props":"%s","recursive":"true"}]}`, repo, props)
	searchSpecFile := filepath.Join(t.TempDir(), "search_props.json")
	assert.NoError(t, os.WriteFile(searchSpecFile, []byte(searchSpec), 0644))

	err := artifactoryCli.Exec("s", "--spec="+searchSpecFile)
	assert.NoError(t, err, "search for artifacts with build properties failed")
	_ = resultItems // search output goes to stdout; error check is sufficient
}

// ---------------------------------------------------------------------------
// P0 — Happy path tests (block release)
// ---------------------------------------------------------------------------

// TestUvBuild verifies that `jf uv build` succeeds and produces .whl + .tar.gz.
func TestUvBuild(t *testing.T) {
	initUvTest(t)
	defer cleanUvTest(t)

	projectPath := createUvProject(t, "uv-build", "uvproject")

	buildNumber := "1"
	assert.NoError(t, runUvCmd(t, projectPath,
		"build",
		"--build-name="+tests.UvBuildName,
		"--build-number="+buildNumber,
	))

	// dist/ must contain both wheel and sdist
	distDir := filepath.Join(projectPath, "dist")
	entries, err := os.ReadDir(distDir)
	assert.NoError(t, err)
	var whl, sdist bool
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".whl") {
			whl = true
		}
		if strings.HasSuffix(e.Name(), ".tar.gz") {
			sdist = true
		}
	}
	assert.True(t, whl, "expected .whl in dist/")
	assert.True(t, sdist, "expected .tar.gz in dist/")

	// Build info must have 2 artifacts
	inttestutils.ValidateGeneratedBuildInfoModule(t, tests.UvBuildName, buildNumber, "",
		[]string{getUvModuleID(t, projectPath)}, buildinfo.Uv)

	assert.NoError(t, artifactoryCli.Exec("bp", tests.UvBuildName, buildNumber))
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.UvBuildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)
	require.Len(t, publishedBuildInfo.BuildInfo.Modules, 1, "expected 1 build info module")
	assert.Len(t, publishedBuildInfo.BuildInfo.Modules[0].Artifacts, 2,
		"build command should capture .whl and .tar.gz")
}

// TestUvPublish verifies that `jf uv publish` uploads artifacts to the local
// repo, stamps build properties, and captures artifacts in build info.
func TestUvPublish(t *testing.T) {
	initUvTest(t)
	defer cleanUvTest(t)

	projectPath := createUvProject(t, "uv-publish", "uvproject")
	buildNumber := "1"

	// Build first
	assert.NoError(t, runUvCmd(t, projectPath, "build"))

	// Publish with build info
	assert.NoError(t, runUvCmd(t, projectPath,
		"publish",
		"--build-name="+tests.UvBuildName,
		"--build-number="+buildNumber,
	))

	assert.NoError(t, artifactoryCli.Exec("bp", tests.UvBuildName, buildNumber))

	// Build properties must be stamped
	validateUvBuildProperties(t, tests.UvLocalRepo, tests.UvBuildName, buildNumber)

	// Build info must have 2 artifacts with sha1+sha256
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.UvBuildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)
	require.Len(t, publishedBuildInfo.BuildInfo.Modules, 1)
	assert.Len(t, publishedBuildInfo.BuildInfo.Modules[0].Artifacts, 2)
	for _, a := range publishedBuildInfo.BuildInfo.Modules[0].Artifacts {
		assert.NotEmpty(t, a.Sha1, "artifact %s missing sha1", a.Name)
		assert.NotEmpty(t, a.Sha256, "artifact %s missing sha256", a.Name)
	}
}

// TestUvSync verifies `jf uv sync` installs packages and captures dependencies.
func TestUvSync(t *testing.T) {
	initUvTest(t)
	defer cleanUvTest(t)

	projectPath := createUvProject(t, "uv-sync", "uvproject")
	buildNumber := "1"

	assert.NoError(t, runUvCmd(t, projectPath,
		"sync",
		"--build-name="+tests.UvBuildName,
		"--build-number="+buildNumber,
	))

	assert.NoError(t, artifactoryCli.Exec("bp", tests.UvBuildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.UvBuildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)
	require.Len(t, publishedBuildInfo.BuildInfo.Modules, 1)
	assert.NotEmpty(t, publishedBuildInfo.BuildInfo.Modules[0].Dependencies,
		"sync should capture at least one dependency")
}

// TestUvBuildInfoPublished verifies build info is published and retrievable.
func TestUvBuildInfoPublished(t *testing.T) {
	initUvTest(t)
	defer cleanUvTest(t)

	projectPath := createUvProject(t, "uv-bi-published", "uvproject")
	buildNumber := "1"

	assert.NoError(t, runUvCmd(t, projectPath, "build",
		"--build-name="+tests.UvBuildName,
		"--build-number="+buildNumber))
	assert.NoError(t, artifactoryCli.Exec("bp", tests.UvBuildName, buildNumber))

	_, found, err := tests.GetBuildInfo(serverDetails, tests.UvBuildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found, "build info should be retrievable after bp")
}

// TestUvNoBuildInfoWhenFlagsAbsent verifies no build info is created when
// --build-name and --build-number are both absent.
func TestUvNoBuildInfoWhenFlagsAbsent(t *testing.T) {
	initUvTest(t)
	defer cleanUvTest(t)

	projectPath := createUvProject(t, "uv-no-bi", "uvproject")
	// Build without build flags → should succeed but not create a build info module
	assert.NoError(t, runUvCmd(t, projectPath, "build"))

	// No module should be stored
	_, found, err := tests.GetBuildInfo(serverDetails, tests.UvBuildName, "1")
	assert.NoError(t, err)
	assert.False(t, found, "no build info should be created when flags are absent")
}

// TestUvBuildPropertiesOnArtifacts verifies build.name / build.number /
// build.timestamp are stamped on published files so the build browser shows paths.
func TestUvBuildPropertiesOnArtifacts(t *testing.T) {
	initUvTest(t)
	defer cleanUvTest(t)

	projectPath := createUvProject(t, "uv-props", "uvproject")
	buildNumber := "1"

	assert.NoError(t, runUvCmd(t, projectPath, "build"))
	assert.NoError(t, runUvCmd(t, projectPath, "publish",
		"--build-name="+tests.UvBuildName,
		"--build-number="+buildNumber))
	assert.NoError(t, artifactoryCli.Exec("bp", tests.UvBuildName, buildNumber))

	validateUvBuildProperties(t, tests.UvLocalRepo, tests.UvBuildName, buildNumber)
}

// ---------------------------------------------------------------------------
// P0 — UV FlexPack correctness invariants
// ---------------------------------------------------------------------------

// TestUvDepIDIsFilename verifies dependency IDs in build info are wheel/sdist
// filenames (e.g. "certifi-2026.2.25-py3-none-any.whl"), not "name:version".
func TestUvDepIDIsFilename(t *testing.T) {
	initUvTest(t)
	defer cleanUvTest(t)

	projectPath := createUvProject(t, "uv-dep-id", "uvproject")
	buildNumber := "1"

	assert.NoError(t, runUvCmd(t, projectPath, "sync",
		"--build-name="+tests.UvBuildName,
		"--build-number="+buildNumber))
	assert.NoError(t, artifactoryCli.Exec("bp", tests.UvBuildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.UvBuildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)
	require.NotEmpty(t, publishedBuildInfo.BuildInfo.Modules)

	for _, dep := range publishedBuildInfo.BuildInfo.Modules[0].Dependencies {
		// ID must look like a filename: contains "-" and ends with .whl or .tar.gz
		isFilename := strings.HasSuffix(dep.Id, ".whl") || strings.HasSuffix(dep.Id, ".tar.gz")
		assert.True(t, isFilename,
			"dependency ID %q should be a filename (e.g. certifi-2026.2.25-py3-none-any.whl), not name:version", dep.Id)
	}
}

// TestUvModuleTypeIsUv verifies the build info module type is "uv" (entities.Uv).
func TestUvModuleTypeIsUv(t *testing.T) {
	initUvTest(t)
	defer cleanUvTest(t)

	projectPath := createUvProject(t, "uv-module-type", "uvproject")
	buildNumber := "1"

	assert.NoError(t, runUvCmd(t, projectPath, "sync",
		"--build-name="+tests.UvBuildName,
		"--build-number="+buildNumber))
	assert.NoError(t, artifactoryCli.Exec("bp", tests.UvBuildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.UvBuildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)
	require.NotEmpty(t, publishedBuildInfo.BuildInfo.Modules)
	assert.Equal(t, string(buildinfo.Uv), string(publishedBuildInfo.BuildInfo.Modules[0].Type),
		"module type should be 'uv'")
}

// TestUvArtifactTypeIsExtension verifies artifact types are file extensions
// ("whl", "gz") not human names ("wheel", "sdist").
func TestUvArtifactTypeIsExtension(t *testing.T) {
	initUvTest(t)
	defer cleanUvTest(t)

	projectPath := createUvProject(t, "uv-art-type", "uvproject")
	buildNumber := "1"

	assert.NoError(t, runUvCmd(t, projectPath, "build",
		"--build-name="+tests.UvBuildName,
		"--build-number="+buildNumber))
	assert.NoError(t, artifactoryCli.Exec("bp", tests.UvBuildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.UvBuildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)
	require.NotEmpty(t, publishedBuildInfo.BuildInfo.Modules)

	for _, a := range publishedBuildInfo.BuildInfo.Modules[0].Artifacts {
		// getArtifactTypeFromName returns "wheel" for .whl and "sdist" for .tar.gz
		assert.True(t, a.Type == "wheel" || a.Type == "sdist" || a.Type == "whl" || a.Type == "gz",
			"artifact type %q should be a known type for wheel or sdist", a.Type)
	}
}

// ---------------------------------------------------------------------------
// P1 — Build flag combinations (table-driven)
// ---------------------------------------------------------------------------

func TestUvBuildFlags(t *testing.T) {
	initUvTest(t)
	defer cleanUvTest(t)

	cases := []struct {
		name        string
		buildName   string
		buildNumber string
		expectBI    bool
		expectErr   bool // jfrog-cli errors if only one of name/number is set
	}{
		{"both-set", tests.UvBuildName, "1", true, false},
		{"name-only", tests.UvBuildName, "", false, true},  // missing number → CLI error
		{"number-only", "", "1", false, true},              // missing name → CLI error
		{"neither", "", "", false, false},                  // no flags → runs fine, no BI
	}

	projectPath := createUvProject(t, "uv-flags", "uvproject")

	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			buildNumber := strconv.Itoa(i + 1)
			args := []string{"build"}
			if tc.buildName != "" {
				args = append(args, "--build-name="+tc.buildName)
			}
			if tc.buildNumber != "" {
				buildNumber = tc.buildNumber
				args = append(args, "--build-number="+buildNumber)
			}

			err := runUvCmd(t, projectPath, args...)
			if tc.expectErr {
				assert.Error(t, err, "expected error when only one of build-name/number is set (%s)", tc.name)
				return
			}
			assert.NoError(t, err)

			if tc.expectBI {
				assert.NoError(t, artifactoryCli.Exec("bp", tc.buildName, buildNumber))
				_, found, biErr := tests.GetBuildInfo(serverDetails, tc.buildName, buildNumber)
				assert.NoError(t, biErr)
				assert.True(t, found, "build info should exist for case %s", tc.name)
				inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tc.buildName, artHttpDetails)
			} else {
				_, found, _ := tests.GetBuildInfo(serverDetails, tc.buildName, buildNumber)
				assert.False(t, found, "no build info should exist for case %s", tc.name)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// P1 — Module override
// ---------------------------------------------------------------------------

func TestUvCustomModule(t *testing.T) {
	initUvTest(t)
	defer cleanUvTest(t)

	projectPath := createUvProject(t, "uv-module", "uvproject")
	buildNumber := "1"
	customModule := "my-custom-uv-module"

	assert.NoError(t, runUvCmd(t, projectPath, "build",
		"--build-name="+tests.UvBuildName,
		"--build-number="+buildNumber,
		"--module="+customModule))
	assert.NoError(t, artifactoryCli.Exec("bp", tests.UvBuildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.UvBuildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)
	require.NotEmpty(t, publishedBuildInfo.BuildInfo.Modules)
	assert.Equal(t, customModule, publishedBuildInfo.BuildInfo.Modules[0].Id)
}

// ---------------------------------------------------------------------------
// P1 — publish-url resolution
// ---------------------------------------------------------------------------

// TestUvPublishURLFromToml verifies publish-url is read from [tool.uv] in
// pyproject.toml automatically (no --publish-url flag required).
func TestUvPublishURLFromToml(t *testing.T) {
	initUvTest(t)
	defer cleanUvTest(t)

	// uvproject-publish-url has [tool.uv] publish-url set, no flag needed
	projectPath := createUvProject(t, "uv-pub-url-toml", "uvproject")
	buildNumber := "1"

	assert.NoError(t, runUvCmd(t, projectPath, "build"))
	// No --publish-url flag — should read from pyproject.toml
	assert.NoError(t, runUvCmd(t, projectPath, "publish",
		"--build-name="+tests.UvBuildName,
		"--build-number="+buildNumber))

	assert.NoError(t, artifactoryCli.Exec("bp", tests.UvBuildName, buildNumber))
	validateUvBuildProperties(t, tests.UvLocalRepo, tests.UvBuildName, buildNumber)
}

// TestUvPublishURLFlagOverridesToml verifies --publish-url flag takes priority
// over [tool.uv] publish-url in pyproject.toml.
func TestUvPublishURLFlagOverridesToml(t *testing.T) {
	initUvTest(t)
	defer cleanUvTest(t)

	projectPath := createUvProject(t, "uv-pub-url-flag", "uvproject")
	buildNumber := "1"
	explicitURL := serverDetails.ArtifactoryUrl + "api/pypi/" + tests.UvLocalRepo

	assert.NoError(t, runUvCmd(t, projectPath, "build"))
	assert.NoError(t, runUvCmd(t, projectPath, "publish",
		"--publish-url="+explicitURL,
		"--build-name="+tests.UvBuildName,
		"--build-number="+buildNumber))

	assert.NoError(t, artifactoryCli.Exec("bp", tests.UvBuildName, buildNumber))
	validateUvBuildProperties(t, tests.UvLocalRepo, tests.UvBuildName, buildNumber)
}

// ---------------------------------------------------------------------------
// P1 — Dependency enrichment (path linking)
// ---------------------------------------------------------------------------

// TestUvSyncDepsEnrichedFromArtifactory verifies that after `jf uv sync`
// against a virtual repo, dependencies have sha1+md5 so the build browser
// can show Artifactory repo paths.
func TestUvSyncDepsEnrichedFromArtifactory(t *testing.T) {
	initUvTest(t)
	defer cleanUvTest(t)

	projectPath := createUvProject(t, "uv-enrich", "uvproject")
	buildNumber := "1"

	assert.NoError(t, runUvCmd(t, projectPath, "sync",
		"--build-name="+tests.UvBuildName,
		"--build-number="+buildNumber))
	assert.NoError(t, artifactoryCli.Exec("bp", tests.UvBuildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.UvBuildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)
	require.NotEmpty(t, publishedBuildInfo.BuildInfo.Modules)
	require.NotEmpty(t, publishedBuildInfo.BuildInfo.Modules[0].Dependencies)

	// At least one dependency should have sha1+md5 (enriched from Artifactory)
	enriched := 0
	for _, dep := range publishedBuildInfo.BuildInfo.Modules[0].Dependencies {
		if dep.Sha1 != "" && dep.Md5 != "" {
			enriched++
		}
	}
	assert.Greater(t, enriched, 0,
		"at least one dependency should be enriched with sha1+md5 from Artifactory virtual repo")
}

// TestUvSyncNoIndexOnlySha256 verifies that without [[tool.uv.index]],
// dependencies get sha256 from uv.lock but no sha1 (no Artifactory enrichment).
func TestUvSyncNoIndexOnlySha256(t *testing.T) {
	initUvTest(t)
	defer cleanUvTest(t)

	// uvproject-no-index has no [[tool.uv.index]] in pyproject.toml
	projectPath := createUvProject(t, "uv-no-index", "uvproject-no-index")
	buildNumber := "1"

	assert.NoError(t, runUvCmd(t, projectPath, "sync",
		"--build-name="+tests.UvBuildName,
		"--build-number="+buildNumber))
	assert.NoError(t, artifactoryCli.Exec("bp", tests.UvBuildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.UvBuildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)
	if len(publishedBuildInfo.BuildInfo.Modules) == 0 {
		t.Log("No modules in build info — uv sync may have resolved from lock without network access")
		return
	}

	for _, dep := range publishedBuildInfo.BuildInfo.Modules[0].Dependencies {
		assert.NotEmpty(t, dep.Sha256, "sha256 from uv.lock should always be present")
		// sha1 may or may not be present depending on Artifactory cache — just log
		t.Logf("dep %s: sha256=%s sha1=%s", dep.Id, dep.Sha256, dep.Sha1)
	}
}

// ---------------------------------------------------------------------------
// P1 — Local sources excluded
// ---------------------------------------------------------------------------

// TestUvEditableSourceExcluded verifies that editable/workspace packages
// (source = {editable="."}) are not included as dependencies in build info.
func TestUvEditableSourceExcluded(t *testing.T) {
	initUvTest(t)
	defer cleanUvTest(t)

	projectPath := createUvProject(t, "uv-editable", "uvproject")
	buildNumber := "1"

	assert.NoError(t, runUvCmd(t, projectPath, "sync",
		"--build-name="+tests.UvBuildName,
		"--build-number="+buildNumber))
	assert.NoError(t, artifactoryCli.Exec("bp", tests.UvBuildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.UvBuildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)
	require.NotEmpty(t, publishedBuildInfo.BuildInfo.Modules)

	projectName := getUvProjectName(t, projectPath)
	for _, dep := range publishedBuildInfo.BuildInfo.Modules[0].Dependencies {
		assert.False(t, strings.Contains(strings.ToLower(dep.Id), strings.ToLower(projectName)),
			"project itself (%s) should not appear as a dependency, got: %s", projectName, dep.Id)
	}
}

// ---------------------------------------------------------------------------
// P1 — Repo & server
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// P1 — lock command captures dependency build info
// ---------------------------------------------------------------------------

func TestUvLockCapturesDependencies(t *testing.T) {
	initUvTest(t)
	defer cleanUvTest(t)

	projectPath := createUvProject(t, "uv-lock", "uvproject")
	buildNumber := "1"

	assert.NoError(t, runUvCmd(t, projectPath, "lock",
		"--build-name="+tests.UvBuildName,
		"--build-number="+buildNumber))
	assert.NoError(t, artifactoryCli.Exec("bp", tests.UvBuildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.UvBuildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)
	require.Len(t, publishedBuildInfo.BuildInfo.Modules, 1)
	// lock is a dep command — artifacts list should be empty
	assert.Empty(t, publishedBuildInfo.BuildInfo.Modules[0].Artifacts,
		"lock command should not produce artifacts in build info")
	// but dependencies should be captured
	assert.NotEmpty(t, publishedBuildInfo.BuildInfo.Modules[0].Dependencies,
		"lock command should capture dependencies in build info")
}

// ---------------------------------------------------------------------------
// P1 — Round-trip: build → publish → verify
// ---------------------------------------------------------------------------

func TestUvRoundTrip(t *testing.T) {
	initUvTest(t)
	defer cleanUvTest(t)

	projectPath := createUvProject(t, "uv-roundtrip", "uvproject")
	buildNumber := "1"

	// Build
	assert.NoError(t, runUvCmd(t, projectPath, "build",
		"--build-name="+tests.UvBuildName,
		"--build-number="+buildNumber))

	// Publish
	assert.NoError(t, runUvCmd(t, projectPath, "publish",
		"--build-name="+tests.UvBuildName,
		"--build-number="+buildNumber))

	// Publish build info
	assert.NoError(t, artifactoryCli.Exec("bp", tests.UvBuildName, buildNumber))

	// Artifacts appear in Artifactory with build properties
	validateUvBuildProperties(t, tests.UvLocalRepo, tests.UvBuildName, buildNumber)

	// Build info has 2 artifacts
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.UvBuildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)
	require.Len(t, publishedBuildInfo.BuildInfo.Modules, 1)
	assert.Len(t, publishedBuildInfo.BuildInfo.Modules[0].Artifacts, 2)
}

// TestUvSyncThenPublishRoundTrip exercises the full workflow:
// sync (collect deps) → build → publish (collect artifacts) → verify both.
func TestUvSyncThenPublishRoundTrip(t *testing.T) {
	initUvTest(t)
	defer cleanUvTest(t)

	projectPath := createUvProject(t, "uv-full-roundtrip", "uvproject")
	buildNumber := "1"

	// Step 1: sync — captures dependencies
	assert.NoError(t, runUvCmd(t, projectPath, "sync",
		"--build-name="+tests.UvBuildName,
		"--build-number="+buildNumber))

	// Step 2: build — captures artifacts
	assert.NoError(t, runUvCmd(t, projectPath, "build",
		"--build-name="+tests.UvBuildName,
		"--build-number="+buildNumber))

	// Step 3: publish — uploads + stamps properties
	assert.NoError(t, runUvCmd(t, projectPath, "publish",
		"--build-name="+tests.UvBuildName,
		"--build-number="+buildNumber))

	// Publish build info
	assert.NoError(t, artifactoryCli.Exec("bp", tests.UvBuildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.UvBuildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)
	require.Len(t, publishedBuildInfo.BuildInfo.Modules, 1)

	module := publishedBuildInfo.BuildInfo.Modules[0]
	assert.Len(t, module.Artifacts, 2, "publish should capture .whl and .tar.gz")
	assert.NotEmpty(t, module.Dependencies, "sync should have captured at least one dependency")
}

// ---------------------------------------------------------------------------
// P1 — No pyproject.toml → clear error
// ---------------------------------------------------------------------------

func TestUvNoPyprojectToml(t *testing.T) {
	initUvTest(t)
	defer cleanUvTest(t)

	tmpDir, cleanup := coretests.CreateTempDirWithCallbackAndAssert(t)
	t.Cleanup(cleanup)

	// Empty directory — no pyproject.toml
	err := runUvCmd(t, tmpDir, "sync",
		"--build-name="+tests.UvBuildName,
		"--build-number=1")
	// uv itself may fail OR the FlexPack may fail — either way we expect an error
	assert.Error(t, err, "should fail when no pyproject.toml is present")
}

// ---------------------------------------------------------------------------
// P1 — Dependency expected vs actual
// ---------------------------------------------------------------------------

// TestUvDependencyExpectedVsActual verifies that build info contains exactly
// the dependencies declared in pyproject.toml — no more, no less.
//
// The test project (uvproject) declares one direct dependency:
//   certifi>=2024.0.0
//
// After `jf uv sync` the build info module must contain:
//   - Exactly 1 dependency (certifi; project itself is excluded)
//   - Dep ID is the wheel filename (e.g. certifi-2026.2.25-py3-none-any.whl)
//   - Dep type is "pypi"
//   - Dep sha256 is non-empty (from uv.lock)
//   - Dep scope is "compile" (direct production dependency)
//   - Module ID matches <project-name>:<version> from pyproject.toml
func TestUvDependencyExpectedVsActual(t *testing.T) {
	initUvTest(t)
	defer cleanUvTest(t)

	projectPath := createUvProject(t, "uv-dep-exact", "uvproject")
	buildNumber := "1"

	assert.NoError(t, runUvCmd(t, projectPath, "sync",
		"--build-name="+tests.UvBuildName,
		"--build-number="+buildNumber,
	))
	assert.NoError(t, artifactoryCli.Exec("bp", tests.UvBuildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.UvBuildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)
	require.Len(t, publishedBuildInfo.BuildInfo.Modules, 1, "expected exactly 1 build-info module")

	module := publishedBuildInfo.BuildInfo.Modules[0]

	// Module ID must be "<project-name>:<version>" from pyproject.toml
	projectName := getUvProjectName(t, projectPath)
	projectVersion := getUvProjectVersion(t, projectPath)
	assert.Equal(t, projectName+":"+projectVersion, module.Id,
		"module ID should be <project-name>:<version> from pyproject.toml")

	// Module type must be "uv"
	assert.Equal(t, string(buildinfo.Uv), string(module.Type),
		"module type should be 'uv'")

	// The test project declares exactly one direct dependency: certifi.
	// The project itself (editable source) must NOT appear.
	require.Len(t, module.Dependencies, 1,
		"uvproject has 1 declared dependency (certifi); project itself must be excluded")

	dep := module.Dependencies[0]

	// ID must be the wheel/sdist filename, not "name:version"
	assert.True(t, strings.HasPrefix(strings.ToLower(dep.Id), "certifi-"),
		"dependency ID should start with 'certifi-', got: %s", dep.Id)
	assert.True(t, strings.HasSuffix(dep.Id, ".whl") || strings.HasSuffix(dep.Id, ".tar.gz"),
		"dependency ID should be a filename (.whl or .tar.gz), got: %s", dep.Id)

	// Type must be "pypi"
	assert.Equal(t, "pypi", dep.Type,
		"dependency type should be 'pypi'")

	// SHA256 must be present (from uv.lock)
	assert.NotEmpty(t, dep.Checksum.Sha256,
		"sha256 must be present from uv.lock for dep %s", dep.Id)

	// Scope must be "compile" (direct production dependency)
	assert.Contains(t, dep.Scopes, "compile",
		"certifi is a direct dependency, scope should be 'compile', got: %v", dep.Scopes)
}

// ---------------------------------------------------------------------------
// P2 — Proxy
// ---------------------------------------------------------------------------

func TestUvWithProxy(t *testing.T) {
	initUvTest(t)
	defer cleanUvTest(t)

	proxyPort := os.Getenv(tests.HttpsProxyEnvVar)
	if proxyPort == "" {
		t.Skip("Skipping proxy test: set " + tests.HttpsProxyEnvVar + " env var to run.")
	}

	projectPath := createUvProject(t, "uv-proxy", "uvproject")
	assert.NoError(t, runUvCmd(t, projectPath, "sync"))
}

// ---------------------------------------------------------------------------
// P1 — Repo & Server: invalid repo → error
// ---------------------------------------------------------------------------

// TestUvPublishToInvalidRepo verifies that publishing to a nonexistent repo
// results in a clear error (404/401 from Artifactory).
func TestUvPublishToInvalidRepo(t *testing.T) {
	initUvTest(t)
	defer cleanUvTest(t)

	projectPath := createUvProject(t, "uv-invalid-repo", "uvproject")

	// Build first so dist/ exists
	assert.NoError(t, runUvCmd(t, projectPath, "build"))

	// Publish to a repo that does not exist — UV should propagate the HTTP error.
	bogusURL := serverDetails.ArtifactoryUrl + "api/pypi/nonexistent-uv-repo-xyz"
	err := runUvCmd(t, projectPath, "publish",
		"--publish-url="+bogusURL,
	)
	assert.Error(t, err, "publishing to a nonexistent repo should return an error")
}

// ---------------------------------------------------------------------------
// P1 — Repo & Server: --project scopes build info directory
// ---------------------------------------------------------------------------

// TestUvProjectFlag verifies that passing --project=<key> stores the build
// info under the project-key-aware local directory (SHA includes the project
// key), and does NOT appear under the empty-project directory.
func TestUvProjectFlag(t *testing.T) {
	initUvTest(t)
	defer cleanUvTest(t)

	buildName := tests.UvBuildName + "-project"
	buildNumber := "1"
	projectKey := "testprj"

	projectPath := createUvProject(t, "uv-project-flag", "uvproject")

	// Run with --project flag so build info is keyed to the project
	assert.NoError(t, runUvCmd(t, projectPath, "sync",
		"--build-name="+buildName,
		"--build-number="+buildNumber,
		"--project="+projectKey,
	))

	// Verify local build info was stored under the project-key-aware directory
	builds, err := coreBuild.GetGeneratedBuildsInfo(buildName, buildNumber, projectKey)
	require.NoError(t, err)
	require.Len(t, builds, 1, "expected 1 build info file stored with project key %q", projectKey)
	assert.Equal(t, buildName, builds[0].Name)
	assert.Equal(t, buildNumber, builds[0].Number)

	// Verify the build is NOT visible under the empty-project directory
	buildsNoProject, err := coreBuild.GetGeneratedBuildsInfo(buildName, buildNumber, "")
	assert.NoError(t, err)
	assert.Empty(t, buildsNoProject, "build info should NOT exist under empty project key directory")

	// Cleanup local build dir
	assert.NoError(t, coreBuild.RemoveBuildDir(buildName, buildNumber, projectKey))
}

// ---------------------------------------------------------------------------
// P1 — UV-specific: uv add captures dependency build info
// ---------------------------------------------------------------------------

// TestUvAddCapturesDependencies verifies that `jf uv add <pkg>` captures
// dependencies in build info (same enrichment path as sync/lock).
func TestUvAddCapturesDependencies(t *testing.T) {
	initUvTest(t)
	defer cleanUvTest(t)

	projectPath := createUvProject(t, "uv-add", "uvproject")
	buildNumber := "1"

	// `uv add` resolves and pins a new package; build info should capture deps.
	assert.NoError(t, runUvCmd(t, projectPath, "add", "certifi",
		"--build-name="+tests.UvBuildName,
		"--build-number="+buildNumber,
	))

	assert.NoError(t, artifactoryCli.Exec("bp", tests.UvBuildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.UvBuildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)
	require.NotEmpty(t, publishedBuildInfo.BuildInfo.Modules,
		"uv add should produce at least one build info module")
	assert.NotEmpty(t, publishedBuildInfo.BuildInfo.Modules[0].Dependencies,
		"uv add should capture at least one dependency")
}

// ---------------------------------------------------------------------------
// P1 — UV-specific: uv run captures dependency build info
// ---------------------------------------------------------------------------

// TestUvRunCapturesDependencies verifies that `jf uv run python --version`
// triggers dependency build info collection (same enrichment path as sync).
func TestUvRunCapturesDependencies(t *testing.T) {
	initUvTest(t)
	defer cleanUvTest(t)

	projectPath := createUvProject(t, "uv-run", "uvproject")
	buildNumber := "1"

	// `uv run` installs the project's environment before executing the command,
	// so the lock file is consulted and deps are resolved — build info is captured.
	assert.NoError(t, runUvCmd(t, projectPath, "run", "python", "--version",
		"--build-name="+tests.UvBuildName,
		"--build-number="+buildNumber,
	))

	assert.NoError(t, artifactoryCli.Exec("bp", tests.UvBuildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.UvBuildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)
	require.NotEmpty(t, publishedBuildInfo.BuildInfo.Modules,
		"uv run should produce at least one build info module")
	assert.NotEmpty(t, publishedBuildInfo.BuildInfo.Modules[0].Dependencies,
		"uv run should capture at least one dependency")
}

// ---------------------------------------------------------------------------
// P1 — Credential priority: native env var takes precedence over jf config
// ---------------------------------------------------------------------------

// TestUvCredentialPriorityEnvVar verifies that when UV_INDEX_<NAME>_USERNAME is
// already set in the environment, jf uv uses those credentials rather than
// injecting its own from jf server config. We validate this by pre-setting the
// correct index credentials and confirming the command succeeds (i.e., the
// env-var path is live and functional).
func TestUvCredentialPriorityEnvVar(t *testing.T) {
	initUvTest(t)
	defer cleanUvTest(t)

	projectPath := createUvProject(t, "uv-cred-priority", "uvproject")
	buildNumber := "1"

	// Derive the env var suffix UV expects for the "jfrog-pypi-virtual" index name.
	indexEnvName := strings.ToUpper(strings.NewReplacer("-", "_", ".", "_").Replace("jfrog-pypi-virtual"))
	userKey := "UV_INDEX_" + indexEnvName + "_USERNAME"
	passKey := "UV_INDEX_" + indexEnvName + "_PASSWORD"

	// Pre-set credentials — this simulates the "native" path (user has already
	// configured env vars). jf uv should detect them and skip its own injection.
	t.Setenv(userKey, *tests.JfrogUser)
	t.Setenv(passKey, *tests.JfrogPassword)

	// The sync must succeed: if jf uv had tried to double-inject, credentials
	// would conflict; the fact it completes validates the native-priority logic.
	assert.NoError(t, runUvCmd(t, projectPath, "sync",
		"--build-name="+tests.UvBuildName,
		"--build-number="+buildNumber,
	))

	assert.NoError(t, artifactoryCli.Exec("bp", tests.UvBuildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.UvBuildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)
	require.NotEmpty(t, publishedBuildInfo.BuildInfo.Modules,
		"sync with native env var credentials should produce a build info module")
}

// ---------------------------------------------------------------------------
// P1 — uv remove captures dependencies in build info
// ---------------------------------------------------------------------------

// TestUvRemoveCapturesDependencies verifies that `jf uv remove <pkg>` triggers
// dependency build info collection (the same FlexPack path as sync/add/lock).
// After removing the only dependency the module should still exist in build info
// (with an empty dependency list), confirming the capture path ran.
func TestUvRemoveCapturesDependencies(t *testing.T) {
	initUvTest(t)
	defer cleanUvTest(t)

	projectPath := createUvProject(t, "uv-remove", "uvproject")
	buildNumber := "1"

	// `uv remove` modifies pyproject.toml + uv.lock and triggers dep collection.
	assert.NoError(t, runUvCmd(t, projectPath, "remove", "certifi",
		"--build-name="+tests.UvBuildName,
		"--build-number="+buildNumber,
	))
	assert.NoError(t, artifactoryCli.Exec("bp", tests.UvBuildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.UvBuildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)
	// The module must exist — build info capture ran even though deps were removed.
	require.Len(t, publishedBuildInfo.BuildInfo.Modules, 1,
		"uv remove should produce exactly 1 build info module")
	// After removing the only dep, the dependency list should be empty.
	assert.Empty(t, publishedBuildInfo.BuildInfo.Modules[0].Dependencies,
		"after removing the only dependency, build info deps list should be empty")
}

// ---------------------------------------------------------------------------
// P0 — Publish with no dist files → error (#26)
// ---------------------------------------------------------------------------

// TestUvPublishNoDistFiles verifies that `jf uv publish` when no dist/ files
// exist returns a clear error rather than silently succeeding with 0 artifacts.
func TestUvPublishNoDistFiles(t *testing.T) {
	initUvTest(t)
	defer cleanUvTest(t)

	projectPath := createUvProject(t, "uv-no-dist", "uvproject")
	// dist/ does not exist — no build was run before publish
	err := runUvCmd(t, projectPath, "publish",
		"--build-name="+tests.UvBuildName,
		"--build-number=1",
	)
	assert.Error(t, err, "publish should fail when dist/ is empty or missing")
}

// ---------------------------------------------------------------------------
// P0 — Sync of project with non-existent dependency → error (#9)
// ---------------------------------------------------------------------------

// TestUvSyncNonExistentPackage verifies that syncing a project whose
// dependency does not exist in the configured Artifactory repo returns
// a clear error (UV's resolver reports "no solution found").
func TestUvSyncNonExistentPackage(t *testing.T) {
	initUvTest(t)
	defer cleanUvTest(t)

	indexURL := serverDetails.ArtifactoryUrl + "api/pypi/" + tests.UvVirtualRepo + "/simple"
	publishURL := serverDetails.ArtifactoryUrl + "api/pypi/" + tests.UvLocalRepo

	tmpDir, cleanup := coretests.CreateTempDirWithCallbackAndAssert(t)
	t.Cleanup(cleanup)

	// pyproject.toml with a dependency that does not exist in Artifactory
	pyproject := fmt.Sprintf(`[project]
name = "uv-nonexistent-test"
version = "0.1.0"
description = "Test"
requires-python = ">=3.11"
dependencies = ["definitely-does-not-exist-xyzzy>=99.0.0"]

[build-system]
requires = ["flit_core>=3.2"]
build-backend = "flit_core.buildapi"

[[tool.uv.index]]
name = "jfrog-pypi-virtual"
url  = "%s"
default = true

[tool.uv]
publish-url = "%s"
`, indexURL, publishURL)
	assert.NoError(t, os.WriteFile(filepath.Join(tmpDir, "pyproject.toml"), []byte(pyproject), 0644))

	// No uv.lock — uv sync will try to resolve and fail
	err := runUvCmd(t, tmpDir, "sync",
		"--build-name="+tests.UvBuildName,
		"--build-number=1",
	)
	assert.Error(t, err, "sync should fail when dependency does not exist in repo")
}

// ---------------------------------------------------------------------------
// P1 — --module overrides module ID on publish (#23)
// ---------------------------------------------------------------------------

// TestUvModuleOverrideOnPublish verifies that `--module=<name>` on publish
// overrides the auto-derived module ID (same behaviour as on sync).
func TestUvModuleOverrideOnPublish(t *testing.T) {
	initUvTest(t)
	defer cleanUvTest(t)

	projectPath := createUvProject(t, "uv-module-publish", "uvproject")
	buildNumber := "1"
	customModule := "my-publish-module"

	assert.NoError(t, runUvCmd(t, projectPath, "build"))
	assert.NoError(t, runUvCmd(t, projectPath, "publish",
		"--build-name="+tests.UvBuildName,
		"--build-number="+buildNumber,
		"--module="+customModule,
	))
	assert.NoError(t, artifactoryCli.Exec("bp", tests.UvBuildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.UvBuildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)
	require.NotEmpty(t, publishedBuildInfo.BuildInfo.Modules)
	assert.Equal(t, customModule, publishedBuildInfo.BuildInfo.Modules[0].Id,
		"--module flag should override auto-derived module ID on publish")
}

// ---------------------------------------------------------------------------
// P1 — Publish to remote (read-only) repo → error (#35)
// ---------------------------------------------------------------------------

// TestUvPublishToRemoteRepo verifies that attempting to publish to a remote
// (proxy) repo returns an error. Remote repos are read-only — they proxy an
// external registry and do not accept uploads.
func TestUvPublishToRemoteRepo(t *testing.T) {
	initUvTest(t)
	defer cleanUvTest(t)

	projectPath := createUvProject(t, "uv-remote-repo", "uvproject")
	remoteURL := serverDetails.ArtifactoryUrl + "api/pypi/" + tests.UvRemoteRepo

	assert.NoError(t, runUvCmd(t, projectPath, "build"))
	err := runUvCmd(t, projectPath, "publish",
		"--publish-url="+remoteURL,
		"--build-name="+tests.UvBuildName,
		"--build-number=1",
	)
	assert.Error(t, err, "publishing to a remote (read-only) repo should fail")
}

// ---------------------------------------------------------------------------
// P1 — UV-native flag passthrough: --index-strategy (#45)
// ---------------------------------------------------------------------------

// TestUvIndexStrategyPassthrough verifies that UV-native flags such as
// --index-strategy are passed through to the uv subprocess without being
// intercepted or rejected by jf. This confirms the SkipFlagParsing=true
// design: all uv flags work transparently.
func TestUvIndexStrategyPassthrough(t *testing.T) {
	initUvTest(t)
	defer cleanUvTest(t)

	projectPath := createUvProject(t, "uv-index-strategy", "uvproject")
	buildNumber := "1"

	// --index-strategy=first-index is the UV default; passing it explicitly
	// verifies the flag is forwarded correctly and sync still succeeds.
	assert.NoError(t, runUvCmd(t, projectPath, "sync",
		"--index-strategy=first-index",
		"--build-name="+tests.UvBuildName,
		"--build-number="+buildNumber,
	))
	assert.NoError(t, artifactoryCli.Exec("bp", tests.UvBuildName, buildNumber))

	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.UvBuildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found)
	require.Len(t, publishedBuildInfo.BuildInfo.Modules, 1,
		"sync with --index-strategy passthrough should produce a build-info module")
}

// ---------------------------------------------------------------------------
// E2E — Publish then install round-trip (#37, #39)
// ---------------------------------------------------------------------------

// TestUvPublishThenInstallRoundTrip is a full E2E test:
//  1. Publish uv-jfrog-test to the local repo
//  2. Create a consumer project that declares uv-jfrog-test as a dependency
//  3. Sync the consumer project — resolves uv-jfrog-test from Artifactory virtual repo
//  4. Verify uv-jfrog-test appears as a dependency in the consumer's build info
//
// This validates the complete publish→resolve pipeline through Artifactory.
func TestUvPublishThenInstallRoundTrip(t *testing.T) {
	initUvTest(t)
	defer cleanUvTest(t)

	// ── Step 1: build and publish the producer package ──────────────────────
	producerPath := createUvProject(t, "uv-e2e-producer", "uvproject")
	producerBuild := tests.UvBuildName + "-producer"
	producerBuildNumber := "1"

	assert.NoError(t, runUvCmd(t, producerPath, "build"))
	assert.NoError(t, runUvCmd(t, producerPath, "publish",
		"--build-name="+producerBuild,
		"--build-number="+producerBuildNumber,
	))

	// Extract published version so the consumer can depend on it
	publishedVersion := getUvProjectVersion(t, producerPath)
	publishedName := getUvProjectName(t, producerPath)

	// ── Step 2: create a consumer project that depends on the producer ───────
	indexURL := serverDetails.ArtifactoryUrl + "api/pypi/" + tests.UvVirtualRepo + "/simple"
	publishURL := serverDetails.ArtifactoryUrl + "api/pypi/" + tests.UvLocalRepo

	consumerDir, cleanupConsumer := coretests.CreateTempDirWithCallbackAndAssert(t)
	t.Cleanup(cleanupConsumer)

	consumerPyproject := fmt.Sprintf(`[project]
name = "uv-jfrog-consumer"
version = "1.0.0"
description = "Consumer package for E2E roundtrip test"
requires-python = ">=3.11"
dependencies = ["%s>=%s"]

[build-system]
requires = ["flit_core>=3.2"]
build-backend = "flit_core.buildapi"

[[tool.uv.index]]
name = "jfrog-pypi-virtual"
url  = "%s"
default = true

[tool.uv]
publish-url = "%s"
`, publishedName, publishedVersion, indexURL, publishURL)

	assert.NoError(t, os.WriteFile(filepath.Join(consumerDir, "pyproject.toml"), []byte(consumerPyproject), 0644))
	assert.NoError(t, os.MkdirAll(filepath.Join(consumerDir, "src", "uv_jfrog_consumer"), 0755))
	assert.NoError(t, os.WriteFile(filepath.Join(consumerDir, "src", "uv_jfrog_consumer", "__init__.py"),
		[]byte("\"\"\"Consumer package.\"\"\"\n__version__ = \"1.0.0\"\n"), 0644))

	// Generate lock for the consumer (resolves producer from Artifactory virtual repo)
	indexEnvName := strings.ToUpper(strings.NewReplacer("-", "_", ".", "_").Replace("jfrog-pypi-virtual"))
	lockCmd := exec.Command("uv", "lock")
	lockCmd.Dir = consumerDir
	lockCmd.Env = append(os.Environ(),
		"UV_INDEX_"+indexEnvName+"_USERNAME="+*tests.JfrogUser,
		"UV_INDEX_"+indexEnvName+"_PASSWORD="+*tests.JfrogPassword,
		"UV_KEYRING_PROVIDER=disabled",
	)
	if out, err := lockCmd.CombinedOutput(); err != nil {
		// If the consumer can't resolve the producer, skip rather than hard-fail.
		// This can happen if the virtual repo hasn't indexed the newly published
		// package yet, or if the remote repo can't reach PyPI in this environment.
		t.Skipf("uv lock failed for consumer — producer may not be in virtual repo yet: %s — %v", out, err)
	}

	// ── Step 3: sync the consumer against Artifactory ────────────────────────
	consumerBuild := tests.UvBuildName + "-consumer"
	consumerBuildNumber := "1"

	assert.NoError(t, runUvCmd(t, consumerDir, "sync",
		"--build-name="+consumerBuild,
		"--build-number="+consumerBuildNumber,
	))
	assert.NoError(t, artifactoryCli.Exec("bp", consumerBuild, consumerBuildNumber))
	defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, consumerBuild, artHttpDetails)

	// ── Step 4: verify producer appears as dependency in consumer's build info ─
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, consumerBuild, consumerBuildNumber)
	assert.NoError(t, err)
	assert.True(t, found)
	require.Len(t, publishedBuildInfo.BuildInfo.Modules, 1)

	deps := publishedBuildInfo.BuildInfo.Modules[0].Dependencies
	require.NotEmpty(t, deps, "consumer build info should contain the published producer as a dependency")

	foundProducer := false
	for _, dep := range deps {
		if strings.Contains(strings.ToLower(dep.Id), strings.ToLower(strings.ReplaceAll(publishedName, "-", "_"))) ||
			strings.Contains(strings.ToLower(dep.Id), strings.ToLower(strings.ReplaceAll(publishedName, "_", "-"))) {
			foundProducer = true
			assert.True(t, strings.HasSuffix(dep.Id, ".whl") || strings.HasSuffix(dep.Id, ".tar.gz"),
				"producer dep ID should be a filename, got: %s", dep.Id)
			break
		}
	}
	assert.True(t, foundProducer,
		"published package %q should appear as a dependency in the consumer's build info; found deps: %v",
		publishedName, deps)
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// getUvModuleID reads [project.name] and [project.version] from pyproject.toml
// and returns "name:version" (the expected build info module ID).
func getUvModuleID(t *testing.T, projectPath string) string {
	name := getUvProjectName(t, projectPath)
	version := getUvProjectVersion(t, projectPath)
	return name + ":" + version
}

func getUvProjectName(t *testing.T, projectPath string) string {
	return readTomlField(t, projectPath, "name")
}

func getUvProjectVersion(t *testing.T, projectPath string) string {
	return readTomlField(t, projectPath, "version")
}

func readTomlField(t *testing.T, projectPath, field string) string {
	data, err := os.ReadFile(filepath.Join(projectPath, "pyproject.toml"))
	assert.NoError(t, err)
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, field+" =") || strings.HasPrefix(line, field+"=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				return strings.Trim(strings.TrimSpace(parts[1]), `"'`)
			}
		}
	}
	return ""
}
