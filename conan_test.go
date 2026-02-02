package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"

	buildinfo "github.com/jfrog/build-info-go/entities"
	biutils "github.com/jfrog/build-info-go/utils"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	coretests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	"github.com/jfrog/jfrog-cli/inttestutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	clientTestUtils "github.com/jfrog/jfrog-client-go/utils/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConanInstall tests 'jf conan install' command with build info collection.
// This is Scenario 1: Install + Build Publish (dependencies only)
func TestConanInstall(t *testing.T) {
	initConanTest(t)
	buildNumber := "1"

	// Prepare project
	projectPath := createConanProject(t, "conan-install-test")
	wd, err := os.Getwd()
	require.NoError(t, err)
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)
	defer chdirCallback()

	// Configure Conan remote
	configureConanRemote(t)
	defer cleanupConanRemote()

	// Run conan install with build info
	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	args := []string{
		"conan", "install", ".",
		"--build=missing",
		"-r", tests.ConanVirtualRepo,
		"--build-name=" + tests.ConanBuildName,
		"--build-number=" + buildNumber,
	}
	require.NoError(t, jfrogCli.Exec(args...))

	// Publish build info
	require.NoError(t, artifactoryCli.Exec("bp", tests.ConanBuildName, buildNumber))
	defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.ConanBuildName, artHttpDetails)

	// Validate build info
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.ConanBuildName, buildNumber)
	require.NoError(t, err)
	require.True(t, found, "build info was expected to be found")

	buildInfoModules := publishedBuildInfo.BuildInfo.Modules
	require.Len(t, buildInfoModules, 1, "Expected 1 module")
	assert.Equal(t, buildinfo.Conan, buildInfoModules[0].Type, "Module type should be conan")

	// Should have dependencies (at least zlib)
	assert.GreaterOrEqual(t, len(buildInfoModules[0].Dependencies), 1, "Expected at least 1 dependency (zlib)")

	// No artifacts should be present (only install was run, no upload)
	assert.Len(t, buildInfoModules[0].Artifacts, 0, "No artifacts expected for install-only scenario")
}

// TestConanInstallCreate tests 'jf conan install' + 'jf conan create' command flow.
// This is Scenario 2: Install + Create + Build Publish (dependencies only)
func TestConanInstallCreate(t *testing.T) {
	initConanTest(t)
	buildNumber := "1"

	// Prepare project
	projectPath := createConanProject(t, "conan-install-create-test")
	wd, err := os.Getwd()
	require.NoError(t, err)
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)
	defer chdirCallback()

	// Configure Conan remote
	configureConanRemote(t)
	defer cleanupConanRemote()

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")

	// Run conan install
	installArgs := []string{
		"conan", "install", ".",
		"--build=missing",
		"-r", tests.ConanVirtualRepo,
		"--build-name=" + tests.ConanBuildName,
		"--build-number=" + buildNumber,
	}
	assert.NoError(t, jfrogCli.Exec(installArgs...))

	// Run conan create
	createArgs := []string{
		"conan", "create", ".",
		"--build=missing",
		"--build-name=" + tests.ConanBuildName,
		"--build-number=" + buildNumber,
	}
	assert.NoError(t, jfrogCli.Exec(createArgs...))

	// Publish build info
	assert.NoError(t, artifactoryCli.Exec("bp", tests.ConanBuildName, buildNumber))
	defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.ConanBuildName, artHttpDetails)

	// Validate build info
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.ConanBuildName, buildNumber)
	require.NoError(t, err)
	require.True(t, found, "build info was expected to be found")

	buildInfoModules := publishedBuildInfo.BuildInfo.Modules
	assert.Len(t, buildInfoModules, 1, "Expected 1 module")

	// Should have dependencies
	assert.GreaterOrEqual(t, len(buildInfoModules[0].Dependencies), 1, "Expected at least 1 dependency")

	// No artifacts without upload
	assert.Len(t, buildInfoModules[0].Artifacts, 0, "No artifacts expected without upload")
}

// TestConanFullFlow tests the complete flow: install + create + upload + build publish.
// This is Scenario 3: Install + Create + Upload + Build Publish (deps + artifacts)
func TestConanFullFlow(t *testing.T) {
	initConanTest(t)
	buildNumber := "1"

	// Prepare project
	projectPath := createConanProject(t, "conan-full-flow-test")
	wd, err := os.Getwd()
	require.NoError(t, err)
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)
	defer chdirCallback()

	// Configure Conan remote
	configureConanRemote(t)
	defer cleanupConanRemote()

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")

	// Run conan install
	installArgs := []string{
		"conan", "install", ".",
		"--build=missing",
		"-r", tests.ConanVirtualRepo,
		"--build-name=" + tests.ConanBuildName,
		"--build-number=" + buildNumber,
	}
	assert.NoError(t, jfrogCli.Exec(installArgs...))

	// Run conan create
	createArgs := []string{
		"conan", "create", ".",
		"--build=missing",
		"--build-name=" + tests.ConanBuildName,
		"--build-number=" + buildNumber,
	}
	assert.NoError(t, jfrogCli.Exec(createArgs...))

	// Run conan upload
	uploadArgs := []string{
		"conan", "upload", "cli-test-package/*",
		"-r", tests.ConanLocalRepo,
		"--confirm",
		"--build-name=" + tests.ConanBuildName,
		"--build-number=" + buildNumber,
	}
	assert.NoError(t, jfrogCli.Exec(uploadArgs...))

	// Publish build info
	assert.NoError(t, artifactoryCli.Exec("bp", tests.ConanBuildName, buildNumber))
	defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.ConanBuildName, artHttpDetails)

	// Validate build info
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.ConanBuildName, buildNumber)
	require.NoError(t, err)
	require.True(t, found, "build info was expected to be found")

	buildInfoModules := publishedBuildInfo.BuildInfo.Modules
	assert.Len(t, buildInfoModules, 1, "Expected 1 module")

	// Should have both dependencies and artifacts
	assert.GreaterOrEqual(t, len(buildInfoModules[0].Dependencies), 1, "Expected at least 1 dependency")
	assert.GreaterOrEqual(t, len(buildInfoModules[0].Artifacts), 1, "Expected at least 1 artifact after upload")

	// Validate artifacts have checksums
	for _, artifact := range buildInfoModules[0].Artifacts {
		assert.NotEmpty(t, artifact.Sha1, "Artifact %s should have SHA1", artifact.Name)
	}
}

// TestConanCreateUpload tests create + upload flow without install.
// This is Scenario 4: Create + Upload + Build Publish (deps + artifacts)
func TestConanCreateUpload(t *testing.T) {
	initConanTest(t)
	buildNumber := "1"

	// Prepare project
	projectPath := createConanProject(t, "conan-create-upload-test")
	wd, err := os.Getwd()
	require.NoError(t, err)
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)
	defer chdirCallback()

	// Configure Conan remote
	configureConanRemote(t)
	defer cleanupConanRemote()

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")

	// Run conan create (this also installs dependencies)
	createArgs := []string{
		"conan", "create", ".",
		"--build=missing",
		"--build-name=" + tests.ConanBuildName,
		"--build-number=" + buildNumber,
	}
	assert.NoError(t, jfrogCli.Exec(createArgs...))

	// Run conan upload
	uploadArgs := []string{
		"conan", "upload", "cli-test-package/*",
		"-r", tests.ConanLocalRepo,
		"--confirm",
		"--build-name=" + tests.ConanBuildName,
		"--build-number=" + buildNumber,
	}
	assert.NoError(t, jfrogCli.Exec(uploadArgs...))

	// Publish build info
	assert.NoError(t, artifactoryCli.Exec("bp", tests.ConanBuildName, buildNumber))
	defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.ConanBuildName, artHttpDetails)

	// Validate build info
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.ConanBuildName, buildNumber)
	require.NoError(t, err)
	require.True(t, found, "build info was expected to be found")

	buildInfoModules := publishedBuildInfo.BuildInfo.Modules
	assert.Len(t, buildInfoModules, 1, "Expected 1 module")

	// Should have both dependencies and artifacts
	assert.GreaterOrEqual(t, len(buildInfoModules[0].Dependencies), 1, "Expected at least 1 dependency")
	assert.GreaterOrEqual(t, len(buildInfoModules[0].Artifacts), 1, "Expected at least 1 artifact after upload")
}

// TestConanAutoLogin tests that auto-login works for Artifactory remotes.
func TestConanAutoLogin(t *testing.T) {
	initConanTest(t)

	// Prepare project
	projectPath := createConanProject(t, "conan-auto-login-test")
	wd, err := os.Getwd()
	require.NoError(t, err)
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)
	defer chdirCallback()

	// Configure Conan remote (without manual login)
	configureConanRemote(t)
	defer cleanupConanRemote()

	// Run conan install - auto-login should happen
	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	args := []string{
		"conan", "install", ".",
		"--build=missing",
		"-r", tests.ConanVirtualRepo,
	}

	// If auto-login works, this should not fail with authentication error
	err = jfrogCli.Exec(args...)
	assert.NoError(t, err, "Conan install with auto-login should succeed")
}

// TestConanBuildInfoModuleFromProject tests that module ID is derived from the conanfile.py project info.
func TestConanBuildInfoModuleFromProject(t *testing.T) {
	initConanTest(t)
	buildNumber := "1"

	// Prepare project
	projectPath := createConanProject(t, "conan-module-test")
	wd, err := os.Getwd()
	require.NoError(t, err)
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)
	defer chdirCallback()

	// Configure Conan remote
	configureConanRemote(t)
	defer cleanupConanRemote()

	// Run conan install
	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	args := []string{
		"conan", "install", ".",
		"--build=missing",
		"-r", tests.ConanVirtualRepo,
		"--build-name=" + tests.ConanBuildName,
		"--build-number=" + buildNumber,
	}
	assert.NoError(t, jfrogCli.Exec(args...))

	// Publish build info
	assert.NoError(t, artifactoryCli.Exec("bp", tests.ConanBuildName, buildNumber))
	defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.ConanBuildName, artHttpDetails)

	// Validate build info module ID comes from conanfile.py (name:version)
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.ConanBuildName, buildNumber)
	require.NoError(t, err)
	require.True(t, found, "build info was expected to be found")

	buildInfoModules := publishedBuildInfo.BuildInfo.Modules
	assert.Len(t, buildInfoModules, 1, "Expected 1 module")
	// Module ID should be derived from conanfile.py: "cli-test-package:1.0.0"
	assert.Equal(t, "cli-test-package:1.0.0", buildInfoModules[0].Id, "Module ID should be derived from conanfile.py")
}

// TestConanMultipleBuilds tests that multiple Conan builds don't interfere with each other.
func TestConanMultipleBuilds(t *testing.T) {
	initConanTest(t)

	// Prepare project
	projectPath := createConanProject(t, "conan-multi-build-test")
	wd, err := os.Getwd()
	require.NoError(t, err)
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)
	defer chdirCallback()

	// Configure Conan remote
	configureConanRemote(t)
	defer cleanupConanRemote()

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")

	// Run multiple builds
	for i := 1; i <= 3; i++ {
		buildNumber := strconv.Itoa(i)

		args := []string{
			"conan", "install", ".",
			"--build=missing",
			"-r", tests.ConanVirtualRepo,
			"--build-name=" + tests.ConanBuildName,
			"--build-number=" + buildNumber,
		}
		assert.NoError(t, jfrogCli.Exec(args...))

		// Publish each build
		assert.NoError(t, artifactoryCli.Exec("bp", tests.ConanBuildName, buildNumber))
	}

	// Cleanup all builds
	defer func() {
		for i := 1; i <= 3; i++ {
			inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.ConanBuildName, artHttpDetails)
		}
	}()

	// Verify all builds exist
	for i := 1; i <= 3; i++ {
		buildNumber := strconv.Itoa(i)
		publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.ConanBuildName, buildNumber)
		require.NoError(t, err)
		require.True(t, found, "build info %s/%s was expected to be found", tests.ConanBuildName, buildNumber)
		assert.Len(t, publishedBuildInfo.BuildInfo.Modules, 1)
	}
}

// TestConanDependencyChecksums tests that dependencies have proper checksums.
func TestConanDependencyChecksums(t *testing.T) {
	initConanTest(t)
	buildNumber := "1"

	// Prepare project
	projectPath := createConanProject(t, "conan-checksum-test")
	wd, err := os.Getwd()
	require.NoError(t, err)
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)
	defer chdirCallback()

	// Configure Conan remote
	configureConanRemote(t)
	defer cleanupConanRemote()

	// Run conan install
	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	args := []string{
		"conan", "install", ".",
		"--build=missing",
		"-r", tests.ConanVirtualRepo,
		"--build-name=" + tests.ConanBuildName,
		"--build-number=" + buildNumber,
	}
	assert.NoError(t, jfrogCli.Exec(args...))

	// Publish build info
	assert.NoError(t, artifactoryCli.Exec("bp", tests.ConanBuildName, buildNumber))
	defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.ConanBuildName, artHttpDetails)

	// Validate checksums
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.ConanBuildName, buildNumber)
	require.NoError(t, err)
	require.True(t, found, "build info was expected to be found")

	buildInfoModules := publishedBuildInfo.BuildInfo.Modules
	require.Len(t, buildInfoModules, 1)

	checksumCount := 0
	for _, dep := range buildInfoModules[0].Dependencies {
		if dep.Sha1 != "" || dep.Sha256 != "" {
			checksumCount++
			t.Logf("Dependency %s has checksums: SHA1=%s, SHA256=%s", dep.Id, dep.Sha1, dep.Sha256)
		}
	}

	assert.Greater(t, checksumCount, 0, "At least some dependencies should have checksums")
}

// Helper functions

func initConanTest(t *testing.T) {
	if !*tests.TestConan {
		t.Skip("Skipping Conan test. To run Conan test add the '-test.conan=true' option.")
	}
	// Ensure Conan is installed
	_, err := exec.LookPath("conan")
	require.NoError(t, err, "Conan must be installed to run Conan tests")
	// Ensure Conan default profile exists (required for conan install/create)
	_ = exec.Command("conan", "profile", "detect").Run()
	// Initialize CLI if not already done
	if artifactoryCli == nil {
		initArtifactoryCli()
	}
	// Set up home directory configuration
	createJfrogHomeConfig(t, true)
}

func createConanProject(t *testing.T, outputFolder string) string {
	projectSrc := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "conan", "conanproject")
	tmpDir, cleanupCallback := coretests.CreateTempDirWithCallbackAndAssert(t)

	projectPath := filepath.Join(tmpDir, outputFolder)
	require.NoError(t, biutils.CopyDir(projectSrc, projectPath, true, nil))

	// Register cleanup to run at the end of the test
	t.Cleanup(cleanupCallback)

	return projectPath
}

func configureConanRemote(t *testing.T) {
	// Remove existing remote if any
	_ = exec.Command("conan", "remote", "remove", tests.ConanVirtualRepo).Run()
	_ = exec.Command("conan", "remote", "remove", tests.ConanLocalRepo).Run()
	// Add Conan remotes pointing to Artifactory
	virtualUrl := serverDetails.ArtifactoryUrl + "api/conan/" + tests.ConanVirtualRepo
	localUrl := serverDetails.ArtifactoryUrl + "api/conan/" + tests.ConanLocalRepo
	addVirtualCmd := exec.Command("conan", "remote", "add", tests.ConanVirtualRepo, virtualUrl)
	require.NoError(t, addVirtualCmd.Run(), "Failed to add Conan virtual remote")
	addLocalCmd := exec.Command("conan", "remote", "add", tests.ConanLocalRepo, localUrl)
	require.NoError(t, addLocalCmd.Run(), "Failed to add Conan local remote")
}

func cleanupConanRemote() {
	_ = exec.Command("conan", "remote", "remove", tests.ConanVirtualRepo).Run()
	_ = exec.Command("conan", "remote", "remove", tests.ConanLocalRepo).Run()
}

// TestConanBuildPublishWithCIVcsProps tests that CI VCS properties are set on Conan artifacts
// when running build-publish in a CI environment (GitHub Actions).
// Conan packages are published via Conan client; build-publish retrieves artifact paths
// from Build Info and applies properties via batch API.
func TestConanBuildPublishWithCIVcsProps(t *testing.T) {
	initConanTest(t)

	buildName := tests.ConanBuildName + "-civcs"
	buildNumber := "1"

	// Setup GitHub Actions environment (uses real env vars on CI, mock values locally)
	cleanupEnv, actualOrg, actualRepo := tests.SetupGitHubActionsEnv(t)
	defer cleanupEnv()

	// Clean old build
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
	defer inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)

	// Prepare project
	projectPath := createConanProject(t, "conan-civcs-test")
	wd, err := os.Getwd()
	require.NoError(t, err)
	chdirCallback := clientTestUtils.ChangeDirWithCallback(t, wd, projectPath)
	defer chdirCallback()

	// Configure Conan remote
	configureConanRemote(t)
	defer cleanupConanRemote()

	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")

	// Run conan create
	createArgs := []string{
		"conan", "create", ".",
		"--build=missing",
		"--build-name=" + buildName,
		"--build-number=" + buildNumber,
	}
	assert.NoError(t, jfrogCli.Exec(createArgs...))

	// Run conan upload
	uploadArgs := []string{
		"conan", "upload", "cli-test-package/*",
		"-r", tests.ConanLocalRepo,
		"--confirm",
		"--build-name=" + buildName,
		"--build-number=" + buildNumber,
	}
	assert.NoError(t, jfrogCli.Exec(uploadArgs...))

	// Publish build info - should set CI VCS props on artifacts
	assert.NoError(t, artifactoryCli.Exec("bp", buildName, buildNumber))

	// Get the published build info to find artifact paths
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	require.NoError(t, err)
	require.True(t, found, "Build info was not found")

	// Create service manager for getting artifact properties
	serviceManager, err := utils.CreateServiceManager(serverDetails, 3, 1000, false)
	require.NoError(t, err)

	// Verify VCS properties on each artifact from build info
	// Note: Conan artifacts may not have OriginalDeploymentRepo set, so we use Path directly as fallback
	artifactCount := 0
	for _, module := range publishedBuildInfo.BuildInfo.Modules {
		for _, artifact := range module.Artifacts {
			// Use same fallback logic as CI VCS: OriginalDeploymentRepo + Path, or Path directly
			var fullPath string
			if artifact.OriginalDeploymentRepo != "" {
				fullPath = artifact.OriginalDeploymentRepo + "/" + artifact.Path
			} else if artifact.Path != "" {
				fullPath = artifact.Path
			} else {
				continue // Skip artifacts without any path info
			}

			props, err := serviceManager.GetItemProps(fullPath)
			assert.NoError(t, err, "Failed to get properties for artifact: %s", fullPath)
			if props == nil {
				continue
			}

			// Validate VCS properties
			assert.Contains(t, props.Properties, "vcs.provider", "Missing vcs.provider on %s", artifact.Name)
			assert.Contains(t, props.Properties["vcs.provider"], "github", "Wrong vcs.provider on %s", artifact.Name)

			assert.Contains(t, props.Properties, "vcs.org", "Missing vcs.org on %s", artifact.Name)
			assert.Contains(t, props.Properties["vcs.org"], actualOrg, "Wrong vcs.org on %s", artifact.Name)

			assert.Contains(t, props.Properties, "vcs.repo", "Missing vcs.repo on %s", artifact.Name)
			assert.Contains(t, props.Properties["vcs.repo"], actualRepo, "Wrong vcs.repo on %s", artifact.Name)

			artifactCount++
		}
	}

	assert.Greater(t, artifactCount, 0, "No artifacts were validated for CI VCS properties")
}
