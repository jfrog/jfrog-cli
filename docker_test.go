package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jfrog/jfrog-client-go/utils/log"

	tests2 "github.com/jfrog/jfrog-cli-artifactory/utils/tests"

	"github.com/docker/docker/api/types/mount"

	biutils "github.com/jfrog/build-info-go/utils"

	"github.com/jfrog/build-info-go/entities"
	"github.com/jfrog/gofrog/version"
	coreContainer "github.com/jfrog/jfrog-cli-artifactory/artifactory/commands/container"
	"github.com/jfrog/jfrog-cli-artifactory/artifactory/commands/generic"
	container "github.com/jfrog/jfrog-cli-artifactory/artifactory/commands/ocicontainer"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/v2/common/build"
	"github.com/jfrog/jfrog-cli-core/v2/common/spec"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	coreTests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	"github.com/jfrog/jfrog-cli/inttestutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/jfrog/jfrog-client-go/auth"
	clientUtils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	clientTestUtils "github.com/jfrog/jfrog-client-go/utils/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	kanikoImage = "gcr.io/kaniko-project/executor:latest"
	rtNetwork   = "test-network"
)

func InitContainerTests() {
	initArtifactoryCli()
	cleanUpOldBuilds()
	inttestutils.CleanUpOldImages(serverDetails)
	cleanUpOldRepositories()
	tests.AddTimestampToGlobalVars()
	createRequiredRepos()
	cleanUpOldRepositories()
}

func initContainerTest(t *testing.T) (containerManagers []container.ContainerManagerType) {
	if *tests.TestDocker {
		containerManagers = append(containerManagers, container.DockerClient)
	}
	if *tests.TestPodman {
		containerManagers = append(containerManagers, container.Podman)
	}
	if len(containerManagers) == 0 {
		t.Skip("Skipping docker and podman tests.")
	}
	return containerManagers
}

func initNativeDockerWithArtTest(t *testing.T) func() {
	if !*tests.TestDocker {
		t.Skip("Skipping native docker test. To run docker test add the '-test.docker=true' option.")
	}
	oldHomeDir := os.Getenv(coreutils.HomeDir)
	servicesManager, err := utils.CreateServiceManager(serverDetails, -1, 0, false)
	require.NoError(t, err)
	rtVersion, err := servicesManager.GetVersion()
	require.NoError(t, err)
	if !version.NewVersion(rtVersion).AtLeast(coreContainer.MinRtVersionForRepoFetching) {
		t.Skip("Skipping native docker test. Artifactory version " + coreContainer.MinRtVersionForRepoFetching + " or higher is required (actual is'" + rtVersion + "').")
	}
	// Create server config to use with the command.
	createJfrogHomeConfig(t, true)
	return func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
	}
}

// initDockerBuildTest initializes test environment for docker build tests with JFROG_RUN_NATIVE enabled
func initDockerBuildTest(t *testing.T) func() {
	// Set JFROG_RUN_NATIVE=true for docker build tests
	clientTestUtils.SetEnvAndAssert(t, "JFROG_RUN_NATIVE", "true")

	// Initialize native docker test setup
	cleanupNativeDocker := initNativeDockerWithArtTest(t)

	// Setup buildx builder with insecure registry config for localhost
	builderName := "jfrog-test-builder"
	cleanupBuilder := setupInsecureBuildxBuilder(t, builderName)

	// Return combined cleanup function
	return func() {
		// Cleanup buildx builder
		cleanupBuilder()
		// Restore JFROG_RUN_NATIVE
		clientTestUtils.UnSetEnvAndAssert(t, "JFROG_RUN_NATIVE")
		// Run native docker cleanup
		cleanupNativeDocker()
	}
}

// setupInsecureBuildxBuilder creates a buildx builder with insecure registry config
func setupInsecureBuildxBuilder(t *testing.T, builderName string) func() {
	// Get registry host from ContainerRegistry
	registryHost := *tests.ContainerRegistry
	if parsedURL, err := url.Parse(registryHost); err == nil && parsedURL.Host != "" {
		registryHost = parsedURL.Host
	}

	// Create temporary buildkitd.toml config
	tmpDir, err := os.MkdirTemp("", "buildkit-config")
	require.NoError(t, err)

	configPath := filepath.Join(tmpDir, "buildkitd.toml")
	configContent := fmt.Sprintf(`[registry."%s"]
  http = true
  insecure = true
`, registryHost)
	require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

	// Remove builder if it exists (stop first, then remove)
	_ = exec.Command("docker", "buildx", "stop", builderName).Run()
	_ = exec.Command("docker", "buildx", "rm", "-f", builderName).Run()

	// Also remove any leftover buildkit container
	_ = exec.Command("docker", "rm", "-f", "buildx_buildkit_"+builderName+"0").Run()

	// Create buildx builder with insecure config
	// Pin to moby/buildkit:v0.12.2 as v0.12.3+ has issues with private HTTP insecure registries
	createCmd := exec.Command("docker", "buildx", "create",
		"--name", builderName,
		"--driver", "docker-container",
		"--driver-opt", "network=host",
		"--driver-opt", "image=moby/buildkit:v0.12.2",
		"--config", configPath,
		"--bootstrap",
		"--use")
	output, err := createCmd.CombinedOutput()
	require.NoError(t, err, "Failed to create buildx builder: %s", string(output))

	// Verify builder is using our config
	inspectCmd := exec.Command("docker", "buildx", "inspect", builderName)
	output, err = inspectCmd.CombinedOutput()
	require.NoError(t, err, "Failed to inspect buildx builder: %s", string(output))
	log.Info("Builder inspect output: %s", string(output))

	// Set BUILDX_BUILDER env var to force 'docker build' to use our builder
	oldBuilder := os.Getenv("BUILDX_BUILDER")
	require.NoError(t, os.Setenv("BUILDX_BUILDER", builderName))

	log.Info("Created buildx builder '%s' with insecure registry config for '%s'", builderName, registryHost)

	// Return cleanup function
	return func() {
		// Restore original BUILDX_BUILDER env var
		if oldBuilder == "" {
			_ = os.Unsetenv("BUILDX_BUILDER")
		} else {
			_ = os.Setenv("BUILDX_BUILDER", oldBuilder)
		}
		// Remove the builder
		_ = exec.Command("docker", "buildx", "rm", builderName).Run()
		// Cleanup temp directory
		_ = os.RemoveAll(tmpDir)
	}
}

func TestContainerPush(t *testing.T) {
	containerManagers := initContainerTest(t)
	for _, repo := range []string{tests.DockerLocalRepo, tests.DockerVirtualRepo} {
		for _, containerManager := range containerManagers {
			runPushTest(containerManager, tests.DockerImageName, tests.DockerImageName+":1", false, t, repo)
		}
	}
}

func TestContainerPushWithModuleName(t *testing.T) {
	containerManagers := initContainerTest(t)
	for _, repo := range []string{tests.DockerLocalRepo, tests.DockerVirtualRepo} {
		for _, containerManager := range containerManagers {
			runPushTest(containerManager, tests.DockerImageName, ModuleNameJFrogTest, true, t, repo)
		}
	}
}

func TestContainerPushWithDetailedSummary(t *testing.T) {
	containerManagers := initContainerTest(t)
	for _, containerManager := range containerManagers {
		imageName := tests.DockerImageName
		module := tests.DockerImageName + ":1"
		buildNumber := "1"
		for _, repo := range []string{tests.DockerLocalRepo, tests.DockerVirtualRepo} {
			func() {
				imageTag, err := inttestutils.BuildTestImage(imageName+":1", "", repo, containerManager)
				assert.NoError(t, err)
				defer tests2.DeleteTestImage(t, imageTag, containerManager)
				// Testing detailed summary without build-info
				pushCommand := coreContainer.NewPushCommand(containerManager)
				pushCommand.SetThreads(1).SetDetailedSummary(true).SetCmdParams([]string{"push", imageTag}).SetBuildConfiguration(new(build.BuildConfiguration)).SetRepo(tests.DockerLocalRepo).SetServerDetails(serverDetails).SetImageTag(imageTag)
				assert.NoError(t, pushCommand.Run())
				result := pushCommand.Result()
				reader := result.Reader()
				defer readerCloseAndAssert(t, reader)
				readerGetErrorAndAssert(t, reader)
				for transferDetails := new(clientUtils.FileTransferDetails); reader.NextRecord(transferDetails) == nil; transferDetails = new(clientUtils.FileTransferDetails) {
					assert.Equal(t, 64, len(transferDetails.Sha256), "Summary validation failed - invalid sha256 has returned from artifactory")
				}
				// Testing detailed summary with build-info
				pushCommand.SetBuildConfiguration(build.NewBuildConfiguration(tests.DockerBuildName, buildNumber, "", ""))
				assert.NoError(t, pushCommand.Run())
				anotherResult := pushCommand.Result()
				anotherReader := anotherResult.Reader()
				defer readerCloseAndAssert(t, anotherReader)

				readerGetErrorAndAssert(t, anotherReader)
				for transferDetails := new(clientUtils.FileTransferDetails); anotherReader.NextRecord(transferDetails) == nil; transferDetails = new(clientUtils.FileTransferDetails) {
					assert.Equal(t, 64, len(transferDetails.Sha256), "Summary validation failed - invalid sha256 has returned from artifactory")
				}

				inttestutils.ValidateGeneratedBuildInfoModule(t, tests.DockerBuildName, buildNumber, "", []string{module}, entities.Docker)
				runRt(t, "build-publish", tests.DockerBuildName, buildNumber)

				imagePath := path.Join(repo, imageName, "1") + "/"
				validateContainerBuild(tests.DockerBuildName, buildNumber, imagePath, module, 7, 5, t)

				// Check deployment view
				assertPrintedDeploymentViewFunc, cleanupFunc := initDeploymentViewTest(t)
				defer cleanupFunc()
				runCmdWithRetries(t, jfrogRtCliTask(containerManager.String()+"-push", pushCommand.ImageTag(), tests.DockerLocalRepo))
				assertPrintedDeploymentViewFunc()
				inttestutils.ContainerTestCleanup(t, serverDetails, artHttpDetails, imageName, tests.DockerBuildName, repo)
			}()
		}
	}
}

func TestContainerPushWithMultipleSlash(t *testing.T) {
	containerManagers := initContainerTest(t)
	for _, repo := range []string{tests.DockerLocalRepo, tests.DockerVirtualRepo} {
		for _, containerManager := range containerManagers {
			runPushTest(containerManager, tests.DockerImageName+"/multiple", "multiple:1", false, t, repo)
		}
	}
}

func runPushTest(containerManager container.ContainerManagerType, imageName, module string, withModule bool, t *testing.T, repo string) {
	imageTag, err := inttestutils.BuildTestImage(imageName+":1", "", repo, containerManager)
	assert.NoError(t, err)
	defer tests2.DeleteTestImage(t, imageTag, containerManager)
	buildNumber := "1"

	if withModule {
		runCmdWithRetries(t, jfrogRtCliTask(containerManager.String()+"-push", imageTag, repo, "--build-name="+tests.DockerBuildName, "--build-number="+buildNumber, "--module="+module))
	} else {
		runCmdWithRetries(t, jfrogRtCliTask(containerManager.String()+"-push", imageTag, repo, "--build-name="+tests.DockerBuildName, "--build-number="+buildNumber))
	}
	defer inttestutils.ContainerTestCleanup(t, serverDetails, artHttpDetails, imageName, tests.DockerBuildName, repo)

	inttestutils.ValidateGeneratedBuildInfoModule(t, tests.DockerBuildName, buildNumber, "", []string{module}, entities.Docker)
	runRt(t, "build-publish", tests.DockerBuildName, buildNumber)

	imagePath := path.Join(repo, imageName, "1") + "/"
	validateContainerBuild(tests.DockerBuildName, buildNumber, imagePath, module, 7, 5, t)
}

func loginToArtifactory(t *testing.T, container *tests.TestContainer) {
	password := *tests.JfrogPassword
	user := *tests.JfrogUser
	if *tests.JfrogAccessToken != "" {
		user = auth.ExtractUsernameFromAccessToken(*tests.JfrogAccessToken)
		password = *tests.JfrogAccessToken
	}
	assert.NoError(t, container.Exec(
		context.Background(),
		"docker",
		"login",
		tests.RtContainerHostName,
		"--username="+user,
		"--password="+password,
	))
}

func buildBuilderImage(t *testing.T, workspace, dockerfile, containerName string) *tests.TestContainer {
	ctx := context.Background()
	testContainer, err := tests.NewContainerRequest().
		SetDockerfile(workspace, dockerfile, nil).
		Privileged().
		Networks(rtNetwork).
		Name(containerName).
		Mount(mount.Mount{Type: mount.TypeBind, Source: workspace, Target: "/workspace", ReadOnly: false}).
		Cmd("--insecure-registry", tests.RtContainerHostName).
		WaitFor(wait.ForLog("API listen on /var/run/docker.sock").WithStartupTimeout(5*time.Minute)).
		Remove().
		Build(ctx, true)
	assert.NoError(t, err, "Couldn't create builder image.")
	return testContainer
}

// This test validate the collect build-info flow for fat-manifest images.
// The way we build the fat manifest and push it to Artifactory is not important.
// Therefore, this test runs only with docker.
func TestPushFatManifestImage(t *testing.T) {
	if !*tests.TestDocker {
		t.Skip("Skipping test. To run it, add the '-test.docker=true' option.")
	}
	buildName := "push-fat-manifest" + tests.DockerBuildName

	// Create temp test dir.
	workspace, err := filepath.Abs(tests.Out)
	assert.NoError(t, err)
	assert.NoError(t, fileutils.CreateDirIfNotExist(workspace))
	testDataDir := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "docker")
	files, err := os.ReadDir(testDataDir)
	assert.NoError(t, err)
	for _, file := range files {
		if !file.IsDir() {
			_, err := tests.ReplaceTemplateVariables(filepath.Join(testDataDir, file.Name()), tests.Out)
			assert.NoError(t, err)
		}
	}

	// Build the builder image locally.
	testContainer := buildBuilderImage(t, workspace, "Dockerfile.Buildx.Fatmanifest", "buildx_container")
	defer func() { assert.NoError(t, testContainer.Terminate(context.Background())) }()

	// Enable the builder util in the container.
	err = testContainer.Exec(context.Background(), "sh", "script.sh")
	assert.NoError(t, err)

	// Login to Artifactory
	loginToArtifactory(t, testContainer)

	buildxOutputFile := "buildmetadata"

	// Run the builder in the container and push the fat-manifest image to Artifactory
	assert.NoError(t, testContainer.Exec(
		context.Background(),
		"docker",
		"buildx",
		"build",
		"--platform",
		"linux/amd64,linux/arm64,linux/arm/v7",
		"--tag", path.Join(tests.RtContainerHostName,
			tests.DockerLocalRepo,
			tests.DockerImageName+"-multiarch-image"),
		"-f",
		"Dockerfile.Fatmanifest",
		"--metadata-file",
		"/workspace/"+buildxOutputFile,
		"--push",
		".",
	))

	// Run 'build-docker-create' & publish the results to Artifactory.
	buildxOutput := filepath.Join(workspace, buildxOutputFile)
	buildNumber := "1"
	assert.NoError(t, artifactoryCli.Exec("build-docker-create", tests.DockerLocalRepo, "--image-file="+buildxOutput, "--build-name="+buildName, "--build-number="+buildNumber))
	assert.NoError(t, artifactoryCli.Exec("build-publish", buildName, buildNumber))

	// Validate the published build-info exists
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found, "build info was expected to be found")
	assert.True(t, len(publishedBuildInfo.BuildInfo.Modules) > 1)

	// Validate build-name & build-number properties in all image layers
	searchSpec := spec.NewBuilder().Pattern(tests.DockerLocalRepo + "/*").Build(buildName).Recursive(true).BuildSpec()
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetServerDetails(serverDetails).SetSpec(searchSpec)
	reader, err := searchCmd.Search()
	assert.NoError(t, err)
	totalResults, err := reader.Length()
	assert.NoError(t, err)
	assert.True(t, totalResults > 1)
}

func TestPushMultiTaggedImage(t *testing.T) {
	if !*tests.TestDocker {
		t.Skip("Skipping test. To run it, add the '-test.docker=true' option.")
	}

	buildName := "push-multi-tagged" + tests.DockerBuildName
	buildNumber := "1"

	// Setup workspace
	workspace, err := filepath.Abs(tests.Out)
	assert.NoError(t, err)
	assert.NoError(t, fileutils.CreateDirIfNotExist(workspace))
	testDataDir := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "docker")
	files, err := os.ReadDir(testDataDir)
	assert.NoError(t, err)
	for _, file := range files {
		if !file.IsDir() {
			_, err := tests.ReplaceTemplateVariables(filepath.Join(testDataDir, file.Name()), tests.Out)
			assert.NoError(t, err)
		}
	}

	// Build the builder image locally
	testContainer := buildBuilderImage(t, workspace, "Dockerfile.Buildx.Fatmanifest", "buildx_multi_tag_container")
	defer func() { assert.NoError(t, testContainer.Terminate(context.Background())) }()

	// Enable builder
	assert.NoError(t, testContainer.Exec(context.Background(), "sh", "script.sh"))

	// Login to Artifactory
	loginToArtifactory(t, testContainer)

	// Build & push image with multiple tags
	imageName1 := path.Join(tests.RtContainerHostName, tests.DockerLocalRepo, tests.DockerImageName+"-multi:v1")
	imageName2 := path.Join(tests.RtContainerHostName, tests.DockerLocalRepo, tests.DockerImageName+"-multi:latest")
	buildxOutputFile := "multi-build-metadata"

	assert.NoError(t, testContainer.Exec(
		context.Background(),
		"docker", "buildx", "build",
		"--platform", "linux/amd64,linux/arm64",
		"-t", imageName1,
		"-t", imageName2,
		"-f", "Dockerfile.Fatmanifest",
		"--metadata-file", "/workspace/"+buildxOutputFile,
		"--push", ".",
	))

	// Run build-docker-create & publish
	buildxOutput := filepath.Join(workspace, buildxOutputFile)
	assert.NoError(t, artifactoryCli.Exec("build-docker-create", tests.DockerLocalRepo, "--image-file="+buildxOutput, "--build-name="+buildName, "--build-number="+buildNumber))
	assert.NoError(t, artifactoryCli.Exec("build-publish", buildName, buildNumber))

	// Validate build-info is published
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found, "build info was expected to be found")
	assert.True(t, len(publishedBuildInfo.BuildInfo.Modules) >= 1)

	// Validate build-name & build-number properties in all layers
	searchSpec := spec.NewBuilder().Pattern(tests.DockerLocalRepo + "/*").Build(buildName).Recursive(true).BuildSpec()
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetServerDetails(serverDetails).SetSpec(searchSpec)
	reader, err := searchCmd.Search()
	assert.NoError(t, err)
	totalResults, err := reader.Length()
	assert.NoError(t, err)
	assert.True(t, totalResults > 1)
}

func TestContainerPushBuildNameNumberFromEnv(t *testing.T) {
	containerManagers := initContainerTest(t)
	for _, containerManager := range containerManagers {
		func() {
			imageTag, err := inttestutils.BuildTestImage(tests.DockerImageName+":1", "", tests.DockerLocalRepo, containerManager)
			assert.NoError(t, err)
			defer tests2.DeleteTestImage(t, imageTag, containerManager)
			buildNumber := "1"
			setEnvCallBack := clientTestUtils.SetEnvWithCallbackAndAssert(t, coreutils.BuildName, tests.DockerBuildName)
			defer setEnvCallBack()
			setEnvCallBack = clientTestUtils.SetEnvWithCallbackAndAssert(t, coreutils.BuildNumber, buildNumber)
			defer setEnvCallBack()
			// Push container image
			runCmdWithRetries(t, jfrogRtCliTask(containerManager.String()+"-push", imageTag, tests.DockerLocalRepo))
			runRt(t, "build-publish")

			imagePath := path.Join(tests.DockerLocalRepo, tests.DockerImageName, "1") + "/"
			validateContainerBuild(tests.DockerBuildName, buildNumber, imagePath, tests.DockerImageName+":1", 7, 5, t)
			inttestutils.ContainerTestCleanup(t, serverDetails, artHttpDetails, tests.DockerImageName, tests.DockerBuildName, tests.DockerLocalRepo)
		}()
	}
}

func TestContainerPull(t *testing.T) {
	containerManagers := initContainerTest(t)
	for _, containerManager := range containerManagers {
		func() {
			imageName, err := inttestutils.BuildTestImage(tests.DockerImageName+":1", "", tests.DockerLocalRepo, containerManager)
			assert.NoError(t, err)
			defer tests2.DeleteTestImage(t, imageName, containerManager)
			for _, repo := range []string{tests.DockerVirtualRepo, tests.DockerLocalRepo} {

				// Push container image
				runCmdWithRetries(t, jfrogRtCliTask(containerManager.String()+"-push", imageName, repo))
				buildNumber := "1"

				// Pull container image
				runCmdWithRetries(t, jfrogRtCliTask(containerManager.String()+"-pull", imageName, repo, "--build-name="+tests.DockerBuildName, "--build-number="+buildNumber))
				runRt(t, "build-publish", tests.DockerBuildName, buildNumber)

				imagePath := path.Join(repo, tests.DockerImageName, "1") + "/"
				validateContainerBuild(tests.DockerBuildName, buildNumber, imagePath, tests.DockerImageName+":1", 0, 7, t)

				buildNumber = "2"
				runCmdWithRetries(t, jfrogRtCliTask(containerManager.String()+"-pull", imageName, repo, "--build-name="+tests.DockerBuildName, "--build-number="+buildNumber, "--module="+ModuleNameJFrogTest))
				runRt(t, "build-publish", tests.DockerBuildName, buildNumber)
				validateContainerBuild(tests.DockerBuildName, buildNumber, imagePath, ModuleNameJFrogTest, 0, 7, t)

				inttestutils.ContainerTestCleanup(t, serverDetails, artHttpDetails, tests.DockerImageName, tests.DockerBuildName, repo)
			}
		}()
	}
}

func TestDockerClientApiVersionCmd(t *testing.T) {
	initContainerTest(t)
	// Assert docker min API version
	assert.NoError(t, container.ValidateClientApiVersion())
}

func TestContainerFatManifestPull(t *testing.T) {
	containerManagers := initContainerTest(t)
	for _, containerManager := range containerManagers {
		imageName := "traefik"
		buildNumber := "1"
		for _, dockerRepo := range [...]string{tests.DockerRemoteRepo, tests.DockerVirtualRepo} {
			func() {
				// Pull container image
				imageTag := path.Join(*tests.ContainerRegistry, dockerRepo, imageName+":2.2")
				defer tests2.DeleteTestImage(t, imageTag, containerManager)
				runCmdWithRetries(t, jfrogRtCliTask(containerManager.String()+"-pull", imageTag, dockerRepo, "--build-name="+tests.DockerBuildName, "--build-number="+buildNumber))
				runRt(t, "build-publish", tests.DockerBuildName, buildNumber)

				// Validate
				publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.DockerBuildName, buildNumber)
				if err != nil {
					assert.NoError(t, err)
					return
				}
				if !found {
					assert.True(t, found, "build info was expected to be found")
					return
				}
				buildInfo := publishedBuildInfo.BuildInfo
				validateBuildInfo(buildInfo, t, 6, 0, imageName+":2.2", entities.Docker)

				inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.DockerBuildName, artHttpDetails)
			}()
		}
	}
}

func TestDockerPromote(t *testing.T) {
	initContainerTest(t)

	// Build and push image
	imageName, err := inttestutils.BuildTestImage(tests.DockerImageName+":1", "", tests.DockerLocalRepo, container.DockerClient)
	assert.NoError(t, err)
	defer tests2.DeleteTestImage(t, imageName, container.DockerClient)

	// Push image
	runCmdWithRetries(t, jfrogRtCliTask("docker-push", imageName, tests.DockerLocalRepo))
	assert.NoError(t, err)

	// Promote image
	runRt(t, "docker-promote", tests.DockerImageName, tests.DockerLocalRepo, tests.DockerLocalPromoteRepo, "--source-tag=1", "--target-tag=2", "--target-docker-image="+tests.DockerImageName+"promotion", "--copy")
	defer inttestutils.ContainerTestCleanup(t, serverDetails, artHttpDetails, tests.DockerImageName, tests.DockerBuildName, tests.DockerLocalRepo)

	// Verify image in source
	imagePath := path.Join(tests.DockerLocalRepo, tests.DockerImageName, "1") + "/"
	validateContainerImage(t, imagePath, 7)

	// Verify image promoted
	searchSpec, err := tests.CreateSpec(tests.SearchPromotedDocker)
	assert.NoError(t, err)
	inttestutils.VerifyExistInArtifactory(tests.GetDockerDeployedManifest(), searchSpec, serverDetails, t)
}

func validateContainerBuild(buildName, buildNumber, imagePath, module string, expectedArtifacts, expectedDependencies int, t *testing.T) {
	validateContainerImage(t, imagePath, 7)
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	if err != nil {
		assert.NoError(t, err)
		return
	}
	if !found {
		assert.True(t, found, "build info was expected to be found")
		return
	}
	buildInfo := publishedBuildInfo.BuildInfo
	validateBuildInfo(buildInfo, t, expectedDependencies, expectedArtifacts, module, entities.Docker)
}

func validateContainerImage(t *testing.T, imagePath string, expectedItemsInArtifactory int) {
	specFile := spec.NewBuilder().Pattern(imagePath + "*").BuildSpec()
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetServerDetails(serverDetails).SetSpec(specFile)
	reader, err := searchCmd.Search()
	assert.NoError(t, err)
	length, err := reader.Length()
	assert.NoError(t, err)
	assert.Equal(t, expectedItemsInArtifactory, length, "Container build info was not pushed correctly")
	readerCloseAndAssert(t, reader)
}

func TestKanikoBuildCollect(t *testing.T) {
	if !*tests.TestDocker {
		t.Skip("Skipping test. To run it, add the '-test.docker=true' option.")
	}
	for _, repo := range []string{tests.DockerVirtualRepo, tests.DockerLocalRepo} {
		imageName := "hello-world-or"
		imageTag := imageName + ":latest"
		buildNumber := "1"
		registryDestination := path.Join(tests.RtContainerHostName, repo, imageTag)
		kanikoOutput := runKaniko(t, registryDestination)

		// Run 'build-docker-create' & publish the results to Artifactory.
		runRt(t, "build-docker-create", repo, "--image-file="+kanikoOutput, "--build-name="+tests.DockerBuildName, "--build-number="+buildNumber)
		runRt(t, "build-publish", tests.DockerBuildName, buildNumber)

		// Validate.
		publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, tests.DockerBuildName, buildNumber)
		if err != nil {
			assert.NoError(t, err)
			return
		}
		if !found {
			assert.True(t, found, "build info was expected to be found")
			return
		}
		buildInfo := publishedBuildInfo.BuildInfo
		validateBuildInfo(buildInfo, t, 0, 3, imageTag, entities.Docker)

		// Cleanup.
		inttestutils.ContainerTestCleanup(t, serverDetails, artHttpDetails, imageName, tests.DockerBuildName, repo)
		assert.NoError(t, fileutils.RemoveTempDir(tests.Out))
	}
}

// Kaniko is an image builder. We use it to build an image and push it to Artifactory.
// Returns the built image metadata file
func runKaniko(t *testing.T, imageToPush string) string {
	testDir := tests.GetTestResourcesPath()
	dockerFile := "TestKanikoBuildCollect"
	KanikoOutputFile := "image-file"
	if *tests.JfrogAccessToken != "" {
		origUsername, origPassword := tests.SetBasicAuthFromAccessToken()
		defer func() {
			*tests.JfrogUser = origUsername
			*tests.JfrogPassword = origPassword
		}()
	}

	// Create Artifactory credentials file.
	credentialsFile, err := tests.ReplaceTemplateVariables(filepath.Join(testDir, tests.KanikoConfig), tests.Out)
	assert.NoError(t, err)
	credentialsFile, err = filepath.Abs(credentialsFile)
	assert.NoError(t, err)
	workspace, err := filepath.Abs(tests.Out)
	assert.NoError(t, err)
	assert.NoError(t, biutils.CopyFile(workspace, filepath.Join(testDir, "docker", dockerFile)))

	// Run Kaniko to build the test image and push it to Artifactory.
	_, err = tests.NewContainerRequest().
		Image(kanikoImage).
		Networks(rtNetwork).
		Mount(
			mount.Mount{Type: mount.TypeBind, Source: workspace, Target: "/workspace", ReadOnly: false},
			mount.Mount{Type: mount.TypeBind, Source: credentialsFile, Target: "/kaniko/.docker/config.json", ReadOnly: true}).
		Cmd("--dockerfile=/workspace/"+dockerFile, "--destination="+imageToPush, "--insecure", "--skip-tls-verify", "--image-name-with-digest-file="+KanikoOutputFile).
		WaitFor(wait.ForExit().WithExitTimeout(300000*time.Millisecond)).
		Build(context.Background(), true)
	assert.NoError(t, err)

	// Return a file contains the image metadata which was built by Kaniko.
	return filepath.Join(workspace, KanikoOutputFile)
}

func TestNativeDockerPushPull(t *testing.T) {
	cleanup := initNativeDockerWithArtTest(t)
	defer cleanup()
	pushBuildNumber := "2"
	image, err := inttestutils.BuildTestImage(tests.DockerImageName+":"+pushBuildNumber, "", tests.DockerLocalRepo, container.DockerClient)
	assert.NoError(t, err)
	defer tests2.DeleteTestImage(t, image, container.DockerClient)
	defer inttestutils.ContainerTestCleanup(t, serverDetails, artHttpDetails, tests.DockerImageName, tests.DockerBuildName, tests.DockerLocalRepo)
	runCmdWithRetries(t, jfCliTask("docker", "-D", "push", image, "--detailed-summary=true"))
	imagePath := path.Join(tests.DockerLocalRepo, tests.DockerImageName, pushBuildNumber) + "/"
	validateContainerImage(t, imagePath, 7)
	tests2.DeleteTestImage(t, image, container.DockerClient)
	runCmdWithRetries(t, jfCliTask("docker", "-D", "pull", image, "--detailed-summary=true"))
}

func TestNativeDockerFlagParsing(t *testing.T) {
	cleanup := initNativeDockerWithArtTest(t)
	defer cleanup()

	dockerTestCases := []struct {
		name        string
		args        []string
		expectedErr error
	}{
		{"docker", []string{"docker"}, nil},
		{"docker version", []string{"docker", "version"}, nil},
		{"docker scan", []string{"docker", "scan", "--min-severity=low"}, errors.New("a docker image name must be provided")},
		{"cli flags after args", []string{"docker", "version", "--build-name=d", "--build-number=1", "--module=1"}, nil},
		{"cli flags before args", []string{"docker", "--build-name=d", "--build-number=1", "--module=1", "version"}, nil},
	}
	for _, testCase := range dockerTestCases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.expectedErr != nil {
				err := runJfrogCliWithoutAssertion(testCase.args...)
				assert.EqualError(t, err, testCase.expectedErr.Error())
				return
			}
			runCmdWithRetries(t, jfCliTask(testCase.args...))
		})
	}
}

func TestDockerLoginWithServer(t *testing.T) {
	cleanup := initNativeDockerWithArtTest(t)
	defer cleanup()

	var credentials string
	if *tests.JfrogAccessToken != "" {
		credentials = "--access-token=" + *tests.JfrogAccessToken
	} else {
		credentials = "--user=" + *tests.JfrogUser + " --password=" + *tests.JfrogPassword
	}
	err := coreTests.NewJfrogCli(execMain, "jfrog config", credentials).Exec("add", "artDocker", "--interactive=false", "--url="+"http://localhost:8082", "--enc-password="+strconv.FormatBool(true))
	assert.NoError(t, err)

	imageName := path.Join(*tests.ContainerRegistry, tests.DockerRemoteRepo, "alpine:latest")

	// Ensure we're logged out first
	cmd := exec.Command("docker", "logout", *tests.ContainerRegistry)
	_, err = cmd.CombinedOutput()
	assert.NoError(t, err)

	// since we're logged out, pulling should fail
	cmd = exec.Command("docker", "pull", imageName)
	output, err := cmd.CombinedOutput()
	assert.Error(t, err)
	assert.Contains(t, string(output), "Authentication is required")

	// Login (perform jf docker login)
	err = runJfrogCliWithoutAssertion("docker", "login", "--server-id=artDocker")
	assert.NoError(t, err)

	// pull should work now
	cmd = exec.Command("docker", "pull", imageName)
	output, err = cmd.CombinedOutput()
	assert.NoError(t, err)
	assert.Contains(t, string(output), "Downloaded newer image")
}

func TestDockerLoginWithRegistry(t *testing.T) {
	cleanup := initNativeDockerWithArtTest(t)
	defer cleanup()

	imageName := path.Join(*tests.ContainerRegistry, tests.DockerRemoteRepo, "busybox:latest")

	// Ensure we're logged out first
	cmd := exec.Command("docker", "logout", *tests.ContainerRegistry)
	_, err := cmd.CombinedOutput()
	assert.NoError(t, err)

	// since we're logged out, pulling should fail
	cmd = exec.Command("docker", "pull", imageName)
	output, err := cmd.CombinedOutput()
	assert.Error(t, err)
	assert.Contains(t, string(output), "Authentication is required")

	// Login (perform jf docker login)
	err = runJfrogCliWithoutAssertion("docker", "login", *tests.ContainerRegistry)
	assert.NoError(t, err)

	// pull should work now
	cmd = exec.Command("docker", "pull", imageName)
	output, err = cmd.CombinedOutput()
	assert.NoError(t, err)
	assert.Contains(t, string(output), "Downloaded newer image")
}

func TestDockerLoginWithRegistryUserAndPass(t *testing.T) {
	cleanup := initNativeDockerWithArtTest(t)
	defer cleanup()

	imageName := path.Join(*tests.ContainerRegistry, tests.DockerRemoteRepo, "hello-world:linux")

	// Ensure we're logged out first
	cmd := exec.Command("docker", "logout", *tests.ContainerRegistry)
	_, err := cmd.CombinedOutput()
	assert.NoError(t, err)

	// since we're logged out, pulling should fail
	cmd = exec.Command("docker", "pull", imageName)
	output, err := cmd.CombinedOutput()
	assert.Error(t, err)
	assert.Contains(t, string(output), "Authentication is required")

	// Login (perform jf docker login)
	password := *tests.JfrogPassword
	if *tests.JfrogAccessToken != "" {
		password = *tests.JfrogAccessToken
	}
	err = runJfrogCliWithoutAssertion("docker", "login", *tests.ContainerRegistry, "--username="+*tests.JfrogUser, "--password="+password)
	assert.NoError(t, err)

	// pull should work now
	cmd = exec.Command("docker", "pull", imageName)
	output, err = cmd.CombinedOutput()
	assert.NoError(t, err)
	assert.Contains(t, string(output), "Downloaded newer image")
}

func jfrogRtCliTask(args ...string) func() error {
	return func() error {
		return artifactoryCli.Exec(args...)
	}
}

func jfCliTask(args ...string) func() error {
	return func() error {
		return coreTests.NewJfrogCli(execMain, "jf", "").WithoutCredentials().Exec(args...)
	}
}

func runCmdWithRetries(t *testing.T, task func() error) {
	maxRetries := 10
	executor := &clientUtils.RetryExecutor{
		MaxRetries:               maxRetries,
		RetriesIntervalMilliSecs: 10000,
		LogMsgPrefix:             "[retries on]",
		ErrorMessage:             fmt.Sprintf("failed to run the command with %d retries.\n", maxRetries),
		ExecutionHandler: func() (bool, error) {
			err := task()
			return err != nil, err
		},
	}
	assert.NoError(t, executor.Execute())
}

func validateDockerBuildInfo(t *testing.T, buildName, buildNumber string, expectedArtifacts bool) {
	// Get and validate build-info
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	assert.NoError(t, err)
	assert.True(t, found, "build info was expected to be found")

	buildInfo := publishedBuildInfo.BuildInfo
	assert.NotNil(t, buildInfo)
	assert.NotEmpty(t, buildInfo.Modules)

	// Check module
	module := buildInfo.Modules[0]
	assert.Equal(t, entities.Docker, module.Type)

	// Check dependencies count
	assert.NotEmpty(t, module.Dependencies)
	assert.True(t, len(module.Dependencies) > 0, "expected dependencies but found none")

	// Check artifacts count
	if expectedArtifacts {
		assert.NotEmpty(t, module.Artifacts)
		assert.True(t, len(module.Artifacts) > 0, "expected artifacts but found none")
	} else {
		assert.Empty(t, module.Artifacts, "expected no artifacts but found some")
	}

	// Check properties
	assert.NotNil(t, module.Properties)
	assert.Contains(t, module.Properties, "docker.image.tag")
	assert.Contains(t, module.Properties, "docker.build.command")

	// Check config digest - should be present when image was pushed
	if expectedArtifacts {
		assert.Contains(t, module.Properties, "docker.image.id")
		props, ok := module.Properties.(map[string]string)
		if ok {
			configDigest := props["docker.image.id"]
			assert.NotEmpty(t, configDigest)
			assert.True(t, strings.HasPrefix(configDigest, "sha256:"))
		}
	}
}

// TestDockerBuildWithBuildInfo tests basic docker build command with build-info collection
func TestDockerBuildWithBuildInfo(t *testing.T) {
	cleanup := initDockerBuildTest(t)
	defer cleanup()

	buildName := tests.DockerBuildName
	buildNumber := "1"
	// Extract hostname from ContainerRegistry (remove protocol if present)
	registryHost := *tests.ContainerRegistry
	if parsedURL, err := url.Parse(registryHost); err == nil && parsedURL.Host != "" {
		registryHost = parsedURL.Host
	}
	// Construct image name for Docker (hostname/repo/image:tag format, no protocol)
	imageName := path.Join(registryHost, tests.OciLocalRepo, "test-docker-build")
	imageTag := imageName + ":v1"

	log.Info("Building image with oci", imageTag)

	// Create test workspace
	workspace, err := filepath.Abs(tests.Out)
	assert.NoError(t, err)
	assert.NoError(t, fileutils.CreateDirIfNotExist(workspace))

	// Create simple Dockerfile
	baseImage := path.Join(registryHost, tests.OciRemoteRepo, "nginx:1.28.0")
	dockerfileContent := fmt.Sprintf(`FROM %s
RUN echo "Hello from test"
CMD ["sh"]`, baseImage)

	dockerfilePath := filepath.Join(workspace, "Dockerfile")
	assert.NoError(t, os.WriteFile(dockerfilePath, []byte(dockerfileContent), 0644))

	// clean build before test
	runJfrogCli(t, "rt", "bc", buildName, buildNumber)

	// Run docker build with build-info
	runJfrogCli(t, "docker", "build", "-t", imageTag, "-f", dockerfilePath, "--build-name="+buildName, "--build-number="+buildNumber, workspace)

	// Publish build info
	runJfrogCli(t, "rt", "build-publish", buildName, buildNumber)

	// Validate build info
	validateDockerBuildInfo(t, buildName, buildNumber, false) // Should have dependencies from base image, no artifacts

	// Cleanup
	tests2.DeleteTestImage(t, imageTag, container.DockerClient)
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

// TestDockerBuildAndPushWithBuildInfo tests docker build followed by push with full build-info tracking
func TestDockerBuildAndPushWithBuildInfo(t *testing.T) {
	cleanup := initDockerBuildTest(t)
	defer cleanup()

	buildName := tests.DockerBuildName
	buildNumber := "2"
	// Extract hostname from ContainerRegistry (remove protocol if present)
	registryHost := *tests.ContainerRegistry
	if parsedURL, err := url.Parse(registryHost); err == nil && parsedURL.Host != "" {
		registryHost = parsedURL.Host
	}
	// Construct image name for Docker (hostname/repo/image:tag format, no protocol)
	imageName := path.Join(registryHost, tests.OciLocalRepo, "test-docker-build")
	imageTag := imageName + ":v1"

	log.Info("Building image with oci", imageTag)

	// Create test workspace
	workspace, err := filepath.Abs(tests.Out)
	assert.NoError(t, err)
	assert.NoError(t, fileutils.CreateDirIfNotExist(workspace))

	// Create simple Dockerfile
	baseImage := path.Join(registryHost, tests.OciRemoteRepo, "nginx:1.28.0")
	dockerfileContent := fmt.Sprintf(`FROM %s
RUN echo "Hello from test"
CMD ["sh"]`, baseImage)

	dockerfilePath := filepath.Join(workspace, "Dockerfile")
	assert.NoError(t, os.WriteFile(dockerfilePath, []byte(dockerfileContent), 0644))

	// Create test file
	testFilePath := filepath.Join(workspace, "test.txt")
	assert.NoError(t, os.WriteFile(testFilePath, []byte("Hello from Docker build test"), 0644))

	// clean build before test
	runJfrogCli(t, "rt", "bc", buildName, buildNumber)

	// Run docker build with build-info
	runCmdWithRetries(t, jfCliTask("docker", "build", "-t", imageTag, "--push", "-f", dockerfilePath, "--build-name="+buildName, "--build-number="+buildNumber, workspace))

	// Publish build info
	runRt(t, "build-publish", buildName, buildNumber)

	// Validate build info - should have both dependencies and artifacts
	validateDockerBuildInfo(t, buildName, buildNumber, true)

	// Cleanup
	tests2.DeleteTestImage(t, imageTag, container.DockerClient)
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

// TestDockerBuildMultiStageDockerfile tests multi-stage Dockerfile parsing and dependency collection
func TestDockerBuildMultiStageDockerfile(t *testing.T) {
	cleanup := initDockerBuildTest(t)
	defer cleanup()

	buildName := tests.DockerBuildName
	buildNumber := "1"
	// Extract hostname from ContainerRegistry (remove protocol if present)
	registryHost := *tests.ContainerRegistry
	if parsedURL, err := url.Parse(registryHost); err == nil && parsedURL.Host != "" {
		registryHost = parsedURL.Host
	}
	// Construct image name for Docker (hostname/repo/image:tag format, no protocol)
	imageName := path.Join(registryHost, tests.OciLocalRepo, "test-docker-build")
	imageTag := imageName + ":v1"

	log.Info("Building image with oci", imageTag)

	// Create test workspace
	workspace, err := filepath.Abs(tests.Out)
	assert.NoError(t, err)
	assert.NoError(t, fileutils.CreateDirIfNotExist(workspace))

	// Construct base images with hostname (just like imageTag construction)
	golangImage := path.Join(registryHost, tests.OciRemoteRepo, "alpine:latest")
	alpineImage := path.Join(registryHost, tests.OciRemoteRepo, "nginx:latest")

	// Create multi-stage Dockerfile
	dockerfileContent := fmt.Sprintf(`# First stage - builder
FROM %s AS builder
CMD ["hello"]

# second stage - final
FROM %s
CMD ["hello"]`, golangImage, alpineImage)

	dockerfilePath := filepath.Join(workspace, "Dockerfile")
	assert.NoError(t, os.WriteFile(dockerfilePath, []byte(dockerfileContent), 0644))

	// clean build before test
	runJfrogCli(t, "rt", "bc", buildName, buildNumber)

	// Run docker build with build-info
	runCmdWithRetries(t, jfCliTask("docker", "build", "-t", imageTag, "-f", dockerfilePath, "--build-name="+buildName, "--build-number="+buildNumber, workspace))

	// Publish build info
	runRt(t, "build-publish", buildName, buildNumber)

	// Validate build info - should have dependencies from golang:1.19-alpine and alpine:3.18
	validateDockerBuildInfo(t, buildName, buildNumber, false)

	// Cleanup
	tests2.DeleteTestImage(t, imageTag, container.DockerClient)
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}

// TestDockerBuildxWithBuildInfo tests buildx build command with build-info collection
func TestDockerBuildxWithBuildInfo(t *testing.T) {
	cleanup := initDockerBuildTest(t)
	defer cleanup()

	buildName := tests.DockerBuildName
	buildNumber := "1"
	// Extract hostname from ContainerRegistry (remove protocol if present)
	registryHost := *tests.ContainerRegistry
	if parsedURL, err := url.Parse(registryHost); err == nil && parsedURL.Host != "" {
		registryHost = parsedURL.Host
	}
	// Construct image name for Docker (hostname/repo/image:tag format, no protocol)
	imageName := path.Join(registryHost, tests.OciLocalRepo, "test-docker-build")
	imageTag := imageName + ":v1"

	log.Info("Building image with oci", imageTag)
	fullImageName := imageTag

	// Create test workspace
	workspace, err := filepath.Abs(tests.Out)
	assert.NoError(t, err)
	assert.NoError(t, fileutils.CreateDirIfNotExist(workspace))

	// Construct base image with hostname (just like imageTag construction)
	baseImage := path.Join(registryHost, tests.OciRemoteRepo, "alpine:latest")

	// Create Dockerfile for buildx
	dockerfileContent := fmt.Sprintf(`FROM %s
RUN echo "Built with buildx"
CMD ["echo", "Hello from buildx"]`, baseImage)

	dockerfilePath := filepath.Join(workspace, "Dockerfile")
	assert.NoError(t, os.WriteFile(dockerfilePath, []byte(dockerfileContent), 0644))

	// Check if buildx is available
	cmd := exec.Command("docker", "buildx", "version")
	if err := cmd.Run(); err != nil {
		t.Error("Docker buildx not available, skipping test")
	}

	// clean build before test
	runJfrogCli(t, "rt", "bc", buildName, buildNumber)

	// Run docker buildx build with build-info and push
	runJfrogCli(t, "docker", "buildx", "build", "--platform", "linux/amd64",
		"-t", fullImageName, "-f", dockerfilePath, "--push", "--build-name="+buildName, "--build-number="+buildNumber, workspace)

	// Publish build info
	runJfrogCli(t, "rt", "build-publish", buildName, buildNumber)

	// Validate build info - buildx with --push should have both dependencies and artifacts
	validateDockerBuildInfo(t, buildName, buildNumber, true)

	// Cleanup
	// Extract just the image name (last part) for cleanup
	imageNameOnly := "test-buildx"
	inttestutils.ContainerTestCleanup(t, serverDetails, artHttpDetails, imageNameOnly, buildName, tests.OciLocalRepo)
}

// TestDockerBuildWithVirtualRepo tests docker build with virtual repository
func TestDockerBuildWithVirtualRepo(t *testing.T) {
	cleanup := initDockerBuildTest(t)
	defer cleanup()

	buildName := tests.DockerBuildName
	buildNumber := "1"
	// Extract hostname from ContainerRegistry (remove protocol if present)
	registryHost := *tests.ContainerRegistry
	if parsedURL, err := url.Parse(registryHost); err == nil && parsedURL.Host != "" {
		registryHost = parsedURL.Host
	}
	// Construct image name for Docker (hostname/repo/image:tag format, no protocol)
	imageName := path.Join(registryHost, tests.OciLocalRepo, "test-docker-build")
	imageTag := imageName + ":v1"

	log.Info("Building image with oci", imageTag)

	// Create test workspace
	workspace, err := filepath.Abs(tests.Out)
	assert.NoError(t, err)
	assert.NoError(t, fileutils.CreateDirIfNotExist(workspace))

	// Construct base image with hostname (just like imageTag construction)
	baseImage := path.Join(registryHost, tests.OciRemoteRepo, "alpine:latest")

	// Create Dockerfile that uses image from virtual repo
	dockerfileContent := fmt.Sprintf(`FROM %s
RUN echo "Testing virtual repo"
CMD ["sh"]`, baseImage)

	dockerfilePath := filepath.Join(workspace, "Dockerfile")
	assert.NoError(t, os.WriteFile(dockerfilePath, []byte(dockerfileContent), 0644))

	// clean build before test
	runJfrogCli(t, "rt", "bc", buildName, buildNumber)

	// Run docker build
	runJfrogCli(t, "docker", "build", "-t", imageTag, "-f", dockerfilePath, "--push",
		"--build-name="+buildName, "--build-number="+buildNumber, workspace)

	// Publish build info
	runJfrogCli(t, "rt", "build-publish", buildName, buildNumber)

	// Validate build info - virtual repo with push should have both dependencies and artifacts
	validateDockerBuildInfo(t, buildName, buildNumber, true)

	// Cleanup
	tests2.DeleteTestImage(t, imageTag, container.DockerClient)
	inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)
}
