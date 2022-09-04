package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jfrog/build-info-go/entities"
	gofrogcmd "github.com/jfrog/gofrog/io"
	"github.com/jfrog/gofrog/version"
	coreContainer "github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/container"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/generic"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils/container"
	"github.com/jfrog/jfrog-cli-core/v2/common/spec"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-cli-core/v2/xray/commands"
	"github.com/jfrog/jfrog-cli-core/v2/xray/commands/scan"
	"github.com/jfrog/jfrog-cli/inttestutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/jfrog/jfrog-client-go/auth"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	clientTestUtils "github.com/jfrog/jfrog-client-go/utils/tests"
	clientUtils "github.com/jfrog/jfrog-client-go/xray/services/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	kanikoImage = "gcr.io/kaniko-project/executor:latest"
)

func InitDockerTests() {
	initArtifactoryCli()
	cleanUpOldBuilds()
	inttestutils.CleanUpOldImages(serverDetails)
	cleanUpOldRepositories()
	tests.AddTimestampToGlobalVars()
	createRequiredRepos()
}

func initContainerTest(t *testing.T) []container.ContainerManagerType {
	if !*tests.TestDocker {
		t.Skip("Skipping docker/podman test. To run docker test add the '-test.docker=true' option.")
	}
	containerManagers := []container.ContainerManagerType{container.DockerClient}
	if !coreutils.IsWindows() {
		containerManagers = append(containerManagers, container.Podman)
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

func initNativeDockerWithXrayTest(t *testing.T) func() {
	if !*tests.TestDocker {
		t.Skip("Skipping native docker test. To run docker test add the '-test.docker=true' option.")
	}
	oldHomeDir := os.Getenv(coreutils.HomeDir)
	initXrayCli()
	validateXrayVersion(t, scan.DockerScanMinXrayVersion)
	// Create server config to use with the command.
	createJfrogHomeConfig(t, true)
	return func() {
		clientTestUtils.SetEnvAndAssert(t, coreutils.HomeDir, oldHomeDir)
	}
}

func TestContainerPush(t *testing.T) {
	containerManagers := initContainerTest(t)
	for _, repo := range []string{*tests.DockerLocalRepo, *tests.DockerVirtualRepo} {
		for _, containerManager := range containerManagers {
			runPushTest(containerManager, tests.DockerImageName, tests.DockerImageName+":1", false, t, repo)
		}
	}
}

func TestContainerPushWithModuleName(t *testing.T) {
	containerManagers := initContainerTest(t)
	for _, repo := range []string{*tests.DockerLocalRepo, *tests.DockerVirtualRepo} {
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
		imageTag, err := inttestutils.BuildTestImage(imageName+":1", "", containerManager)
		assert.NoError(t, err)
		defer inttestutils.DeleteTestImage(t, imageTag, containerManager)
		buildNumber := "1"
		for _, repo := range []string{*tests.DockerLocalRepo, *tests.DockerVirtualRepo} {
			// Testing detailed summary without buildinfo
			pushCommand := coreContainer.NewPushCommand(containerManager)
			pushCommand.SetThreads(1).SetDetailedSummary(true).SetCmdParams([]string{"push", imageTag}).SetBuildConfiguration(new(utils.BuildConfiguration)).SetRepo(*tests.DockerLocalRepo).SetServerDetails(serverDetails).SetImageTag(imageTag)
			assert.NoError(t, pushCommand.Run())
			result := pushCommand.Result()
			reader := result.Reader()
			defer readerCloseAndAssert(t, reader)
			readerGetErrorAndAssert(t, reader)
			for transferDetails := new(clientutils.FileTransferDetails); reader.NextRecord(transferDetails) == nil; transferDetails = new(clientutils.FileTransferDetails) {
				assert.Equal(t, 64, len(transferDetails.Sha256), "Summary validation failed - invalid sha256 has returned from artifactory")
			}
			// Testing detailed summary with buildinfo
			pushCommand.SetBuildConfiguration(utils.NewBuildConfiguration(tests.DockerBuildName, buildNumber, "", ""))
			assert.NoError(t, pushCommand.Run())
			anotherResult := pushCommand.Result()
			anotherReader := anotherResult.Reader()
			defer readerCloseAndAssert(t, anotherReader)

			readerGetErrorAndAssert(t, anotherReader)
			for transferDetails := new(clientutils.FileTransferDetails); anotherReader.NextRecord(transferDetails) == nil; transferDetails = new(clientutils.FileTransferDetails) {
				assert.Equal(t, 64, len(transferDetails.Sha256), "Summary validation failed - invalid sha256 has returned from artifactory")
			}

			inttestutils.ValidateGeneratedBuildInfoModule(t, tests.DockerBuildName, buildNumber, "", []string{module}, entities.Docker)
			runRt(t, "build-publish", tests.DockerBuildName, buildNumber)

			imagePath := path.Join(repo, imageName, "1") + "/"
			validateContainerBuild(tests.DockerBuildName, buildNumber, imagePath, module, 7, 5, 7, t)

			// Check deployment view
			assertPrintedDeploymentViewFunc, cleanupFunc := initDeploymentViewTest(t)
			defer cleanupFunc()
			runRt(t, containerManager.String()+"-push", pushCommand.ImageTag(), *tests.DockerLocalRepo)
			assertPrintedDeploymentViewFunc()
			inttestutils.ContainerTestCleanup(t, serverDetails, artHttpDetails, imageName, tests.DockerBuildName, repo)
		}
	}
}

func TestContainerPushWithMultipleSlash(t *testing.T) {
	containerManagers := initContainerTest(t)
	for _, repo := range []string{*tests.DockerLocalRepo, *tests.DockerVirtualRepo} {
		for _, containerManager := range containerManagers {
			runPushTest(containerManager, tests.DockerImageName+"/multiple", "multiple:1", false, t, repo)
		}
	}
}

// Run container push to Artifactory
func runPushTest(containerManager container.ContainerManagerType, imageName, module string, withModule bool, t *testing.T, repo string) {
	imageTag, err := inttestutils.BuildTestImage(imageName+":1", "", containerManager)
	assert.NoError(t, err)
	defer inttestutils.DeleteTestImage(t, imageTag, containerManager)
	buildNumber := "1"

	// Push image
	if withModule {
		runRt(t, containerManager.String()+"-push", imageTag, repo, "--build-name="+tests.DockerBuildName, "--build-number="+buildNumber, "--module="+module)
	} else {
		runRt(t, containerManager.String()+"-push", imageTag, repo, "--build-name="+tests.DockerBuildName, "--build-number="+buildNumber)
	}
	defer inttestutils.ContainerTestCleanup(t, serverDetails, artHttpDetails, imageName, tests.DockerBuildName, repo)

	inttestutils.ValidateGeneratedBuildInfoModule(t, tests.DockerBuildName, buildNumber, "", []string{module}, entities.Docker)
	runRt(t, "build-publish", tests.DockerBuildName, buildNumber)

	imagePath := path.Join(repo, imageName, "1") + "/"
	validateContainerBuild(tests.DockerBuildName, buildNumber, imagePath, module, 7, 5, 7, t)
}

func TestRunPushFatManifestImage(t *testing.T) {
	initContainerTest(t)
	// Create new container that includes docker daemon + buildx CLI tool
	// The builder image name.
	builderImageName := "buildx_builder:1"
	// The container name which runs the builder image.
	builderContainerName := "buildx_container"
	// The image name to build for multi platforms.
	multiArchImageName := tests.DockerImageName + "-multiarch-image"
	// The multi platforms image tag .
	multiArchImageTag := ":latest"
	buildName := "push-fat-manifest" + tests.DockerBuildName
	// A name for buildx output file
	buildxOutputFile := "buildmetadata"

	// Setup test env.
	workspace, err := filepath.Abs(tests.Out)
	assert.NoError(t, err)
	assert.NoError(t, fileutils.CreateDirIfNotExist(workspace))

	// Build the builder image locally.
	builderImageName, err = inttestutils.BuildTestImage(builderImageName, "Dockerfile.Buildx.Fatmanifest", container.DockerClient)
	assert.NoError(t, err)
	defer inttestutils.DeleteTestImage(t, builderImageName, container.DockerClient)

	// Run the builder container.
	runCmd := inttestutils.NewRunDockerImage(container.DockerClient, "-d", "--name", builderContainerName, "--privileged", "-v", workspace+":/workspace", builderImageName)
	assert.NoError(t, gofrogcmd.RunCmd(runCmd))
	defer inttestutils.DeleteTestContainer(t, builderContainerName, container.DockerClient)

	// Docker daemon may be lost for the first few seconds, perform 3 retries before failure.
	require.True(t, isDaemonRunning(builderContainerName), "docker daemon is not responding in remote container")

	// Configure buildx in remote container
	execCmd := inttestutils.NewExecDockerImage(container.DockerClient, builderContainerName, "sh", "script.sh")
	require.NoError(t, gofrogcmd.RunCmd(execCmd))

	// login to the Artifactory within the container
	password := *tests.JfrogPassword
	user := *tests.JfrogUser
	if *tests.JfrogAccessToken != "" {
		user, err = auth.ExtractUsernameFromAccessToken(*tests.JfrogAccessToken)
		require.NoError(t, err)
		password = *tests.JfrogAccessToken
	}
	execCmd = inttestutils.NewExecDockerImage(container.DockerClient, builderContainerName, "docker", "login", *tests.DockerRepoDomain, "--username", user, "--password", password)
	err = gofrogcmd.RunCmd(execCmd)
	require.NoError(t, err, "fail to login to container registry")

	// Build & push the multi platform image to Artifactory
	execCmd = inttestutils.NewExecDockerImage(container.DockerClient, builderContainerName, "/buildx", "build", "--platform", "linux/amd64,linux/arm64,linux/arm/v7", "--tag", path.Join(*tests.DockerRepoDomain, multiArchImageName+multiArchImageTag), "-f", "Dockerfile.Fatmanifest", "--metadata-file", "/workspace/"+buildxOutputFile, "--push", ".")
	require.NoError(t, gofrogcmd.RunCmd(execCmd))

	// Run 'build-docker-create' & publish the results to Artifactory.
	buildxOutput := filepath.Join(workspace, buildxOutputFile)
	buildNumber := "1"
	assert.NoError(t, artifactoryCli.Exec("build-docker-create", *tests.DockerLocalRepo, "--image-file="+buildxOutput, "--build-name="+buildName, "--build-number="+buildNumber))
	assert.NoError(t, artifactoryCli.Exec("build-publish", buildName, buildNumber))
	defer inttestutils.ContainerTestCleanup(t, serverDetails, artHttpDetails, multiArchImageName, buildName, *tests.DockerLocalRepo)

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
	match, err := entities.IsEqualModuleSlices(publishedBuildInfo.BuildInfo.Modules, getExpectedFatManifestBuildInfo(t, tests.ExpectedFatManifestBuildInfo).Modules)
	if err != nil {
		assert.NoError(t, err)
		return
	}
	assert.True(t, match, "the actual buildinfo.json is different compared to the expected")

	// Validate build-name & build-number properties in all image layers
	spec := spec.NewBuilder().Pattern(*tests.DockerLocalRepo + "/*").Build(buildName).Recursive(true).BuildSpec()
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetServerDetails(serverDetails).SetSpec(spec)
	reader, err := searchCmd.Search()
	assert.NoError(t, err)
	totalResults, err := reader.Length()
	assert.NoError(t, err)
	assert.Equal(t, 10, totalResults)
}

// Check if Docker daemon is running on a given container
func isDaemonRunning(containerName string) bool {
	execCmd := inttestutils.NewExecDockerImage(container.DockerClient, containerName, "docker", "ps")
	for i := 0; i < 3; i++ {
		if execCmd.GetCmd().Run() != nil {
			time.Sleep(8 * time.Second)
		} else {
			return true
		}
	}
	return false
}

func TestContainerPushBuildNameNumberFromEnv(t *testing.T) {
	containerManagers := initContainerTest(t)
	for _, containerManager := range containerManagers {
		imageTag, err := inttestutils.BuildTestImage(tests.DockerImageName+":1", "", containerManager)
		assert.NoError(t, err)
		defer inttestutils.DeleteTestImage(t, imageTag, containerManager)
		buildNumber := "1"
		setEnvCallBack := clientTestUtils.SetEnvWithCallbackAndAssert(t, coreutils.BuildName, tests.DockerBuildName)
		defer setEnvCallBack()
		setEnvCallBack = clientTestUtils.SetEnvWithCallbackAndAssert(t, coreutils.BuildNumber, buildNumber)
		defer setEnvCallBack()
		// Push container image
		runRt(t, containerManager.String()+"-push", imageTag, *tests.DockerLocalRepo)
		runRt(t, "build-publish")

		imagePath := path.Join(*tests.DockerLocalRepo, tests.DockerImageName, "1") + "/"
		validateContainerBuild(tests.DockerBuildName, buildNumber, imagePath, tests.DockerImageName+":1", 7, 5, 7, t)
		inttestutils.ContainerTestCleanup(t, serverDetails, artHttpDetails, tests.DockerImageName, tests.DockerBuildName, *tests.DockerLocalRepo)
	}

}

func TestContainerPull(t *testing.T) {
	containerManagers := initContainerTest(t)
	for _, containerManager := range containerManagers {
		imageName, err := inttestutils.BuildTestImage(tests.DockerImageName+":1", "", containerManager)
		assert.NoError(t, err)
		defer inttestutils.DeleteTestImage(t, imageName, containerManager)
		for _, repo := range []string{*tests.DockerVirtualRepo, *tests.DockerLocalRepo} {

			// Push container image
			runRt(t, containerManager.String()+"-push", imageName, repo)

			buildNumber := "1"

			// Pull container image
			runRt(t, containerManager.String()+"-pull", imageName, repo, "--build-name="+tests.DockerBuildName, "--build-number="+buildNumber)
			runRt(t, "build-publish", tests.DockerBuildName, buildNumber)

			imagePath := path.Join(repo, tests.DockerImageName, "1") + "/"
			validateContainerBuild(tests.DockerBuildName, buildNumber, imagePath, tests.DockerImageName+":1", 0, 7, 7, t)

			buildNumber = "2"
			runRt(t, containerManager.String()+"-pull", imageName, repo, "--build-name="+tests.DockerBuildName, "--build-number="+buildNumber, "--module="+ModuleNameJFrogTest)
			runRt(t, "build-publish", tests.DockerBuildName, buildNumber)
			validateContainerBuild(tests.DockerBuildName, buildNumber, imagePath, ModuleNameJFrogTest, 0, 7, 7, t)

			inttestutils.ContainerTestCleanup(t, serverDetails, artHttpDetails, tests.DockerImageName, tests.DockerBuildName, repo)
		}
	}
}

func TestDockerClientApiVersionCmd(t *testing.T) {
	initContainerTest(t)

	// Run docker version command and expect no errors
	cmd := &container.VersionCmd{}
	content, err := cmd.RunCmd()
	assert.NoError(t, err)

	// Expect VersionRegex to match the output API version
	content = strings.TrimSpace(content)
	assert.True(t, container.ApiVersionRegex.Match([]byte(content)))

	// Assert docker min API version
	assert.True(t, container.IsCompatibleApiVersion(content))
}

func TestContainerFatManifestPull(t *testing.T) {
	containerManagers := initContainerTest(t)
	for _, containerManager := range containerManagers {
		imageName := "traefik"
		imageTag := path.Join(*tests.DockerRepoDomain, imageName+":2.2")
		buildNumber := "1"
		defer inttestutils.DeleteTestImage(t, imageTag, containerManager)
		for _, dockerRepo := range [...]string{*tests.DockerRemoteRepo, *tests.DockerVirtualRepo} {
			// Pull container image
			runRt(t, containerManager.String()+"-pull", imageTag, dockerRepo, "--build-name="+tests.DockerBuildName, "--build-number="+buildNumber)
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
		}
	}
}

func TestDockerPromote(t *testing.T) {
	initContainerTest(t)

	// Build and push image
	imageName, err := inttestutils.BuildTestImage(tests.DockerImageName+":1", "", container.DockerClient)
	assert.NoError(t, err)
	defer inttestutils.DeleteTestImage(t, imageName, container.DockerClient)

	// Push image
	runRt(t, "docker-push", imageName, *tests.DockerLocalRepo)
	assert.NoError(t, err)

	// Promote image
	runRt(t, "docker-promote", tests.DockerImageName, *tests.DockerLocalRepo, *tests.DockerPromoteLocalRepo, "--source-tag=1", "--target-tag=2", "--target-docker-image="+tests.DockerImageName+"promotion", "--copy")
	defer inttestutils.ContainerTestCleanup(t, serverDetails, artHttpDetails, tests.DockerImageName, tests.DockerBuildName, *tests.DockerLocalRepo)

	// Verify image in source
	imagePath := path.Join(*tests.DockerLocalRepo, tests.DockerImageName, "1") + "/"
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
	initContainerTest(t)
	for _, repo := range []string{*tests.DockerVirtualRepo, *tests.DockerLocalRepo} {
		imageName := "hello-world-or"
		imageTag := imageName + ":latest"
		buildNumber := "1"
		registryDestination := path.Join(*tests.DockerRepoDomain, imageTag)
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
		inttestutils.DeleteTestImage(t, kanikoImage, container.DockerClient)
		assert.NoError(t, fileutils.RemoveTempDir(tests.Out))
	}
}

// t - Test struct.
// kanikoWorkspace - Local path to kaniko's workspace.
// imageToPush - The image to be pushed by kaniko.
// return path to the kaniko's output file.
func runKaniko(t *testing.T, imageToPush string) string {
	testDir := tests.GetTestResourcesPath()
	dockerFile := "TestKanikoBuildCollect"
	imageNameWithDigestFile := "image-file"
	if *tests.JfrogAccessToken != "" {
		origUsername, origPassword := tests.SetBasicAuthFromAccessToken(t)
		defer func() {
			*tests.JfrogUser = origUsername
			*tests.JfrogPassword = origPassword
		}()
	}
	credentialsFile, err := tests.ReplaceTemplateVariables(filepath.Join(testDir, tests.KanikoConfig), tests.Out)
	assert.NoError(t, err)
	credentialsFile, err = filepath.Abs(credentialsFile)
	assert.NoError(t, err)
	workspace, err := filepath.Abs(tests.Out)
	assert.NoError(t, err)
	assert.NoError(t, fileutils.CopyFile(workspace, filepath.Join(testDir, "docker", dockerFile)))
	imageRunnder := inttestutils.NewRunDockerImage(container.DockerClient, "--rm", "-v", workspace+":/workspace", "-v", credentialsFile+":/kaniko/.docker/config.json:ro", kanikoImage, "--dockerfile="+dockerFile, "--destination="+imageToPush, "--image-name-with-digest-file="+imageNameWithDigestFile)
	assert.NoError(t, gofrogcmd.RunCmd(imageRunnder))
	return filepath.Join(workspace, imageNameWithDigestFile)
}

func TestDockerScan(t *testing.T) {
	cleanup := initNativeDockerWithXrayTest(t)
	defer cleanup()

	watchName, deleteWatch := createTestWatch(t)
	defer deleteWatch()

	imagesToScan := []string{
		// Simple image with vulnerabilities
		"bitnami/minio:2022",

		// Image with RPM with vulnerabilities
		"redhat/ubi8-micro:8.5",
	}
	for _, imageName := range imagesToScan {
		runDockerScan(t, imageName, watchName, 3, 3, 3)
	}

	// On Xray 3.40.3 there is a bug whereby xray fails to scan docker image with 0 vulnerabilities,
	// So we skip it for now till the next version will be released
	validateXrayVersion(t, "3.41.0")

	// Image with 0 vulnerabilities
	runDockerScan(t, "busybox:1.35", "", 0, 0, 0)
}

func TestDockerScanWithProgressBar(t *testing.T) {
	callback := tests.MockProgressInitialization()
	defer callback()
	TestDockerScan(t)
}

func runDockerScan(t *testing.T, imageName, watchName string, minViolations, minVulnerabilities, minLicenses int) {
	// Pull image from docker repo
	imageTag := path.Join(*tests.DockerRepoDomain, imageName)
	dockerPullCommand := coreContainer.NewPullCommand(container.DockerClient)
	dockerPullCommand.SetCmdParams([]string{"pull", imageTag}).SetImageTag(imageTag).SetRepo(*tests.DockerVirtualRepo).SetServerDetails(serverDetails).SetBuildConfiguration(new(utils.BuildConfiguration))
	if assert.NoError(t, dockerPullCommand.Run()) {
		defer inttestutils.DeleteTestImage(t, imageTag, container.DockerClient)

		args := []string{"docker", "scan", imageTag, "--server-id=default", "--licenses", "--format=json", "--fail=false"}

		// Run docker scan on image
		output := xrayCli.WithoutCredentials().RunCliCmdWithOutput(t, args...)
		if assert.NotEmpty(t, output) {
			verifyJsonScanResults(t, output, 0, minVulnerabilities, minLicenses)
		}

		// Run docker scan on image with watch
		if watchName != "" {
			args = append(args, "--watches="+watchName)
			output = xrayCli.WithoutCredentials().RunCliCmdWithOutput(t, args...)
			if assert.NotEmpty(t, output) {
				verifyJsonScanResults(t, output, minViolations, 0, 0)
			}
		}
	}
}

func createTestWatch(t *testing.T) (string, func()) {
	trueValue := true
	xrayManager, err := commands.CreateXrayServiceManager(xrayDetails)
	assert.NoError(t, err)
	// Create new default policy.
	policyParams := clientUtils.PolicyParams{
		Name: fmt.Sprintf("%s-%s", "docker-policy", strconv.FormatInt(time.Now().Unix(), 10)),
		Type: clientUtils.Security,
		Rules: []clientUtils.PolicyRule{{
			Name:     "sec_rule",
			Criteria: *clientUtils.CreateSeverityPolicyCriteria(clientUtils.Low),
			Priority: 1,
			Actions: &clientUtils.PolicyAction{
				FailBuild: &trueValue,
			},
		}},
	}
	if !assert.NoError(t, xrayManager.CreatePolicy(policyParams)) {
		return "", func() {}
	}
	// Create new default watch.
	watchParams := clientUtils.NewWatchParams()
	watchParams.Name = fmt.Sprintf("%s-%s", "docker-watch", strconv.FormatInt(time.Now().Unix(), 10))
	watchParams.Active = true
	watchParams.Builds.Type = clientUtils.WatchBuildAll
	watchParams.Policies = []clientUtils.AssignedPolicy{
		{
			Name: policyParams.Name,
			Type: "security",
		},
	}
	assert.NoError(t, xrayManager.CreateWatch(watchParams))
	return watchParams.Name, func() {
		assert.NoError(t, xrayManager.DeleteWatch(watchParams.Name))
		assert.NoError(t, xrayManager.DeletePolicy(policyParams.Name))
	}
}

func getExpectedFatManifestBuildInfo(t *testing.T, fileName string) entities.BuildInfo {
	testDir := tests.GetTestResourcesPath()
	buildinfoFile, err := tests.ReplaceTemplateVariables(filepath.Join(testDir, fileName), tests.Out)
	assert.NoError(t, err)
	buildinfoFile, err = filepath.Abs(buildinfoFile)
	assert.NoError(t, err)
	data, err := ioutil.ReadFile(buildinfoFile)
	assert.NoError(t, err)
	var buildinfo entities.BuildInfo
	assert.NoError(t, json.Unmarshal(data, &buildinfo))
	return buildinfo
}

func TestNativeDockerPushPull(t *testing.T) {
	cleanup := initNativeDockerWithArtTest(t)
	defer cleanup()
	pushBuildNumber := "2"
	pullBuildNumber := "3"
	module := "native-docker-module"
	image, err := inttestutils.BuildTestImage(tests.DockerImageName+":"+pushBuildNumber, "", container.DockerClient)
	assert.NoError(t, err)
	// Add docker cli flag '-D' to check we ignore them
	runNativeDocker(t, "docker", "-D", "push", image, "--build-name="+tests.DockerBuildName, "--build-number="+pushBuildNumber, "--module="+module)

	inttestutils.ValidateGeneratedBuildInfoModule(t, tests.DockerBuildName, pushBuildNumber, "", []string{module}, entities.Docker)
	runRt(t, "build-publish", tests.DockerBuildName, pushBuildNumber)
	imagePath := path.Join(*tests.DockerLocalRepo, tests.DockerImageName, pushBuildNumber) + "/"
	validateContainerBuild(tests.DockerBuildName, pushBuildNumber, imagePath, module, 7, 5, 7, t)
	inttestutils.DeleteTestImage(t, image, container.DockerClient)

	runNativeDocker(t, "docker", "-D", "pull", image, "--build-name="+tests.DockerBuildName, "--build-number="+pullBuildNumber, "--module="+module)
	runRt(t, "build-publish", tests.DockerBuildName, pullBuildNumber)
	imagePath = path.Join(*tests.DockerLocalRepo, tests.DockerImageName, pullBuildNumber) + "/"
	validateContainerBuild(tests.DockerBuildName, pullBuildNumber, imagePath, module, 0, 7, 0, t)
	inttestutils.DeleteTestImage(t, image, container.DockerClient)

	inttestutils.ContainerTestCleanup(t, serverDetails, artHttpDetails, tests.DockerImageName, tests.DockerBuildName, *tests.DockerLocalRepo)
}

func TestNativeDocker(t *testing.T) {
	cleanup := initNativeDockerWithArtTest(t)
	defer cleanup()
	runNativeDocker(t, "docker", "version")
	// Check we don't fail with JFrog flags.
	runNativeDocker(t, "docker", "version", "--build-name=d", "--build-number=1", "--module=1")
}

func runNativeDocker(t *testing.T, args ...string) {
	jfCli := tests.NewJfrogCli(execMain, "jf", "")
	assert.NoError(t, jfCli.WithoutCredentials().Exec(args...))
}
