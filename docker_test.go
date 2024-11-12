package main

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/mount"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	biutils "github.com/jfrog/build-info-go/utils"

	"github.com/jfrog/build-info-go/entities"
	"github.com/jfrog/gofrog/version"
	coreContainer "github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/container"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/generic"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils/container"
	"github.com/jfrog/jfrog-cli-core/v2/common/build"
	"github.com/jfrog/jfrog-cli-core/v2/common/spec"
	commonTests "github.com/jfrog/jfrog-cli-core/v2/common/tests"
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
				defer commonTests.DeleteTestImage(t, imageTag, containerManager)
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
				validateContainerBuild(tests.DockerBuildName, buildNumber, imagePath, module, 7, 5, 7, t)

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
	defer commonTests.DeleteTestImage(t, imageTag, containerManager)
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
	validateContainerBuild(tests.DockerBuildName, buildNumber, imagePath, module, 7, 5, 7, t)
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
	ctx := context.Background()
	testContainer, err := tests.NewContainerRequest().
		SetDockerfile(workspace, "Dockerfile.Buildx.Fatmanifest", nil).
		Privileged().
		Networks(rtNetwork).
		Name("buildx_container").
		Mount(mount.Mount{Type: mount.TypeBind, Source: workspace, Target: "/workspace", ReadOnly: false}).
		Cmd("--insecure-registry", tests.RtContainerHostName).
		// Docker daemon take times to load. In order to check if it's available we wait for a log message to indications that the Docker daemon has finished initializing.
		WaitFor(wait.ForLog("API listen on /var/run/docker.sock").WithStartupTimeout(5*time.Minute)).
		Remove().
		Build(ctx, true)
	assert.NoError(t, err, "Couldn't run create buildx image.")
	defer func() { assert.NoError(t, testContainer.Terminate(ctx)) }()

	// Enable the builder util in the container.
	err = testContainer.Exec(ctx, "sh", "script.sh")
	assert.NoError(t, err)

	// login from the builder container toward the Artifactory.
	password := *tests.JfrogPassword
	user := *tests.JfrogUser
	if *tests.JfrogAccessToken != "" {
		user = auth.ExtractUsernameFromAccessToken(*tests.JfrogAccessToken)
		password = *tests.JfrogAccessToken
	}
	assert.NoError(t, testContainer.Exec(
		ctx,
		"docker",
		"login",
		tests.RtContainerHostName,
		"--username="+user,
		"--password="+password))
	buildxOutputFile := "buildmetadata"

	// Run the builder in the container and push the fat-manifest image to artifactory
	assert.NoError(t, testContainer.Exec(
		ctx,
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

	// Validate the published build-info exits
	publishedBuildInfo, found, err := tests.GetBuildInfo(serverDetails, buildName, buildNumber)
	if err != nil {
		assert.NoError(t, err)
		return
	}
	if !found {
		assert.True(t, found, "build info was expected to be found")
		return
	}

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

func TestContainerPushBuildNameNumberFromEnv(t *testing.T) {
	containerManagers := initContainerTest(t)
	for _, containerManager := range containerManagers {
		func() {
			imageTag, err := inttestutils.BuildTestImage(tests.DockerImageName+":1", "", tests.DockerLocalRepo, containerManager)
			assert.NoError(t, err)
			defer commonTests.DeleteTestImage(t, imageTag, containerManager)
			buildNumber := "1"
			setEnvCallBack := clientTestUtils.SetEnvWithCallbackAndAssert(t, coreutils.BuildName, tests.DockerBuildName)
			defer setEnvCallBack()
			setEnvCallBack = clientTestUtils.SetEnvWithCallbackAndAssert(t, coreutils.BuildNumber, buildNumber)
			defer setEnvCallBack()
			// Push container image
			runCmdWithRetries(t, jfrogRtCliTask(containerManager.String()+"-push", imageTag, tests.DockerLocalRepo))
			runRt(t, "build-publish")

			imagePath := path.Join(tests.DockerLocalRepo, tests.DockerImageName, "1") + "/"
			validateContainerBuild(tests.DockerBuildName, buildNumber, imagePath, tests.DockerImageName+":1", 7, 5, 7, t)
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
			defer commonTests.DeleteTestImage(t, imageName, containerManager)
			for _, repo := range []string{tests.DockerVirtualRepo, tests.DockerLocalRepo} {

				// Push container image
				runCmdWithRetries(t, jfrogRtCliTask(containerManager.String()+"-push", imageName, repo))
				buildNumber := "1"

				// Pull container image
				runCmdWithRetries(t, jfrogRtCliTask(containerManager.String()+"-pull", imageName, repo, "--build-name="+tests.DockerBuildName, "--build-number="+buildNumber))
				runRt(t, "build-publish", tests.DockerBuildName, buildNumber)

				imagePath := path.Join(repo, tests.DockerImageName, "1") + "/"
				validateContainerBuild(tests.DockerBuildName, buildNumber, imagePath, tests.DockerImageName+":1", 0, 7, 7, t)

				buildNumber = "2"
				runCmdWithRetries(t, jfrogRtCliTask(containerManager.String()+"-pull", imageName, repo, "--build-name="+tests.DockerBuildName, "--build-number="+buildNumber, "--module="+ModuleNameJFrogTest))
				runRt(t, "build-publish", tests.DockerBuildName, buildNumber)
				validateContainerBuild(tests.DockerBuildName, buildNumber, imagePath, ModuleNameJFrogTest, 0, 7, 7, t)

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
				defer commonTests.DeleteTestImage(t, imageTag, containerManager)
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
	defer commonTests.DeleteTestImage(t, imageName, container.DockerClient)

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

func validateContainerBuild(buildName, buildNumber, imagePath, module string, expectedArtifacts, expectedDependencies, expectedItemsInArtifactory int, t *testing.T) {
	validateContainerImage(t, imagePath, expectedItemsInArtifactory)
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
	pullBuildNumber := "3"
	module := "native-docker-module"
	image, err := inttestutils.BuildTestImage(tests.DockerImageName+":"+pushBuildNumber, "", tests.DockerLocalRepo, container.DockerClient)
	assert.NoError(t, err)
	// Add docker cli flag '-D' to check we ignore them
	runCmdWithRetries(t, jfCliTask("docker", "-D", "push", image, "--build-name="+tests.DockerBuildName, "--build-number="+pushBuildNumber, "--module="+module))

	inttestutils.ValidateGeneratedBuildInfoModule(t, tests.DockerBuildName, pushBuildNumber, "", []string{module}, entities.Docker)
	runRt(t, "build-publish", tests.DockerBuildName, pushBuildNumber)
	imagePath := path.Join(tests.DockerLocalRepo, tests.DockerImageName, pushBuildNumber) + "/"
	validateContainerBuild(tests.DockerBuildName, pushBuildNumber, imagePath, module, 7, 5, 7, t)
	commonTests.DeleteTestImage(t, image, container.DockerClient)

	runCmdWithRetries(t, jfCliTask("docker", "-D", "pull", image, "--build-name="+tests.DockerBuildName, "--build-number="+pullBuildNumber, "--module="+module))
	runRt(t, "build-publish", tests.DockerBuildName, pullBuildNumber)
	imagePath = path.Join(tests.DockerLocalRepo, tests.DockerImageName, pullBuildNumber) + "/"
	validateContainerBuild(tests.DockerBuildName, pullBuildNumber, imagePath, module, 0, 7, 0, t)
	commonTests.DeleteTestImage(t, image, container.DockerClient)

	inttestutils.ContainerTestCleanup(t, serverDetails, artHttpDetails, tests.DockerImageName, tests.DockerBuildName, tests.DockerLocalRepo)
}

func TestNativeDockerFlagParsing(t *testing.T) {
	cleanup := initNativeDockerWithArtTest(t)
	defer cleanup()

	dockerTestCases := []struct {
		name string
		args []string
	}{
		{"docker", []string{"docker"}},
		{"docker version", []string{"docker", "version"}},
		{"docker scan", []string{"docker", "scan"}},
		{"cli flags after args", []string{"docker", "version", "--build-name=d", "--build-number=1", "--module=1"}},
		{"cli flags before args", []string{"docker", "--build-name=d", "--build-number=1", "--module=1", "version"}},
	}
	for _, testCase := range dockerTestCases {
		t.Run(testCase.name, func(t *testing.T) {
			runCmdWithRetries(t, jfCliTask(testCase.args...))
		})
	}
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
