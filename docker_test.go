package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jfrog/jfrog-cli-core/v2/xray/commands/scan"

	"github.com/jfrog/build-info-go/entities"
	buildinfo "github.com/jfrog/build-info-go/entities"

	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/v2/common/spec"

	gofrogcmd "github.com/jfrog/gofrog/io"
	corecontainer "github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/container"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/generic"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils/container"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-cli/inttestutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	clientTestUtils "github.com/jfrog/jfrog-client-go/utils/tests"
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
	if coreutils.IsLinux() {
		containerManagers = append(containerManagers, container.Podman)
	}
	return containerManagers
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
	for _, repo := range []string{*tests.DockerLocalRepo, *tests.DockerVirtualRepo} {
		for _, containerManager := range containerManagers {
			imageName := tests.DockerImageName
			module := tests.DockerImageName + ":1"
			imageTag, err := inttestutils.BuildTestImage(imageName+":1", "", containerManager)
			assert.NoError(t, err)
			buildNumber := "1"
			dockerPushCommand := corecontainer.NewPushCommand(containerManager)

			// Testing detailed summary without buildinfo
			dockerPushCommand.SetThreads(1).SetDetailedSummary(true).SetBuildConfiguration(new(utils.BuildConfiguration)).SetRepo(*tests.DockerLocalRepo).SetServerDetails(serverDetails).SetImageTag(imageTag)
			assert.NoError(t, dockerPushCommand.Run())
			result := dockerPushCommand.Result()
			result.Reader()
			reader := result.Reader()
			defer readerCloseAndAssert(t, reader)
			readerGetErrorAndAssert(t, reader)
			for transferDetails := new(clientutils.FileTransferDetails); reader.NextRecord(transferDetails) == nil; transferDetails = new(clientutils.FileTransferDetails) {
				assert.Equal(t, 64, len(transferDetails.Sha256), "Summary validation failed - invalid sha256 has returned from artifactory")
			}
			// Testing detailed summary with buildinfo
			readerCloseAndAssert(t, reader)
			dockerPushCommand.SetBuildConfiguration(utils.NewBuildConfiguration(tests.DockerBuildName, buildNumber, "", ""))
			assert.NoError(t, dockerPushCommand.Run())
			result = dockerPushCommand.Result()
			reader = result.Reader()
			readerGetErrorAndAssert(t, reader)
			for transferDetails := new(clientutils.FileTransferDetails); reader.NextRecord(transferDetails) == nil; transferDetails = new(clientutils.FileTransferDetails) {
				assert.Equal(t, 64, len(transferDetails.Sha256), "Summary validation failed - invalid sha256 has returned from artifactory")
			}

			inttestutils.ValidateGeneratedBuildInfoModule(t, tests.DockerBuildName, buildNumber, "", []string{module}, buildinfo.Docker)
			runRt(t, "build-publish", tests.DockerBuildName, buildNumber)

			imagePath := path.Join(repo, imageName, "1") + "/"
			validateContainerBuild(tests.DockerBuildName, buildNumber, imagePath, module, 7, 5, 7, t)
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
	buildNumber := "1"

	// Push image
	if withModule {
		runRt(t, containerManager.String()+"-push", imageTag, repo, "--build-name="+tests.DockerBuildName, "--build-number="+buildNumber, "--module="+module)
	} else {
		runRt(t, containerManager.String()+"-push", imageTag, repo, "--build-name="+tests.DockerBuildName, "--build-number="+buildNumber)
	}

	inttestutils.ValidateGeneratedBuildInfoModule(t, tests.DockerBuildName, buildNumber, "", []string{module}, buildinfo.Docker)
	runRt(t, "build-publish", tests.DockerBuildName, buildNumber)

	imagePath := path.Join(repo, imageName, "1") + "/"
	validateContainerBuild(tests.DockerBuildName, buildNumber, imagePath, module, 7, 5, 7, t)
	inttestutils.ContainerTestCleanup(t, serverDetails, artHttpDetails, imageName, tests.DockerBuildName, repo)
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
	defer inttestutils.DeleteTestcontainer(t, builderContainerName, container.DockerClient)

	// Docker daemon may be lost for the first few seconds, perform 3 retries before failure.
	require.True(t, isDaemonRunning(builderContainerName), "docker daemon is not responding in remote container")

	// Configure buildx in remote container
	execCmd := inttestutils.NewExecDockerImage(container.DockerClient, builderContainerName, "sh", "script.sh")
	require.NoError(t, gofrogcmd.RunCmd(execCmd))

	// login to the Artifactory within the container
	password := *tests.JfrogPassword
	if *tests.JfrogAccessToken != "" {
		password = *tests.JfrogAccessToken
	}
	execCmd = inttestutils.NewExecDockerImage(container.DockerClient, builderContainerName, "docker", "login", *tests.DockerRepoDomain, "--username", *tests.JfrogUser, "--password", password)
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
	assert.True(t, entities.IsEqualModuleSlices(publishedBuildInfo.BuildInfo.Modules, getExpectedFatManifestBuildInfo(t).Modules))

	// Validate build-name & build-number properties in all image layers
	spec := spec.NewBuilder().Pattern(*tests.DockerLocalRepo + "/*").Build(buildName).Recursive(true).BuildSpec()
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetServerDetails(serverDetails).SetSpec(spec)
	reader, err := searchCmd.Search()
	totalResults, err := reader.Length()
	assert.NoError(t, err)
	assert.Equal(t, 19, totalResults)

	inttestutils.ContainerTestCleanup(t, serverDetails, artHttpDetails, multiArchImageName, buildName, *tests.DockerLocalRepo)
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
	for _, repo := range []string{*tests.DockerVirtualRepo, *tests.DockerLocalRepo} {
		for _, containerManager := range containerManagers {
			imageName, err := inttestutils.BuildTestImage(tests.DockerImageName+":1", "", containerManager)
			assert.NoError(t, err)

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
		for _, dockerRepo := range [...]string{*tests.DockerRemoteRepo, *tests.DockerVirtualRepo} {
			imageName := "traefik"
			imageTag := path.Join(*tests.DockerRepoDomain, imageName+":2.2")
			buildNumber := "1"

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
			validateBuildInfo(buildInfo, t, 6, 0, imageName+":2.2", buildinfo.Docker)

			inttestutils.DeleteBuild(serverDetails.ArtifactoryUrl, tests.DockerBuildName, artHttpDetails)
			inttestutils.DeleteTestImage(t, imageTag, containerManager)
		}
	}
}

func TestDockerPromote(t *testing.T) {
	initContainerTest(t)

	// Build and push image
	imageName, err := inttestutils.BuildTestImage(tests.DockerImageName+":1", "", container.DockerClient)
	assert.NoError(t, err)
	runRt(t, "docker-push", imageName, *tests.DockerLocalRepo)
	assert.NoError(t, err)

	// Promote image
	runRt(t, "docker-promote", tests.DockerImageName, *tests.DockerLocalRepo, *tests.DockerPromoteLocalRepo, "--source-tag=1", "--target-tag=2", "--target-docker-image=docker-target-image", "--copy")

	// Verify image in source
	imagePath := path.Join(*tests.DockerLocalRepo, tests.DockerImageName, "1") + "/"
	validateContainerImage(t, imagePath, 7)

	// Verify image promoted
	searchSpec, err := tests.CreateSpec(tests.SearchAllDocker)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetDockerDeployedManifest(), searchSpec, t)

	inttestutils.ContainerTestCleanup(t, serverDetails, artHttpDetails, tests.DockerImageName, tests.DockerBuildName, *tests.DockerLocalRepo)
	inttestutils.DeleteTestImage(t, imageName, container.DockerClient)
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
	validateBuildInfo(buildInfo, t, expectedDependencies, expectedArtifacts, module, buildinfo.Docker)
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
		kanikoOutput := runKaniko(t, registryDestination, kanikoImage)

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
		validateBuildInfo(buildInfo, t, 0, 3, imageTag, buildinfo.Docker)

		// Cleanup.
		inttestutils.ContainerTestCleanup(t, serverDetails, artHttpDetails, imageName, tests.DockerBuildName, repo)
		inttestutils.DeleteTestImage(t, kanikoImage, container.DockerClient)
		assert.NoError(t, os.RemoveAll(tests.Out))
	}
}

// t - Test struct.
// kanikoWorkspace - Local path to kaniko's workspace.
// imageToPush - The image to be pushed by kaniko.
// return path to the kaniko's output file.
func runKaniko(t *testing.T, imageToPush, kanikoImage string) string {
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

func TestXrayDockerScan(t *testing.T) {
	initContainerTest(t)
	initXrayCli()
	validateXrayVersion(t, scan.DockerScanMinXrayVersion)

	// Pull alpine image from docker repo
	imageTag := path.Join(*tests.DockerRepoDomain, tests.DockerScanTestImage)
	dockerPullCommand := corecontainer.NewPullCommand(container.DockerClient)
	dockerPullCommand.SetImageTag(imageTag).SetRepo(*tests.DockerVirtualRepo).SetServerDetails(serverDetails).SetBuildConfiguration(new(utils.BuildConfiguration))
	assert.NoError(t, dockerPullCommand.Run())

	// Run docker scan on alpine image
	output := xrayCli.RunCliCmdWithOutput(t, container.DockerClient.String(), "scan", tests.DockerScanTestImage)
	verifyScanResults(t, output, 0, 1, 1)

	// Delete alpine image
	inttestutils.DeleteTestImage(t, imageTag, container.DockerClient)
}

func getExpectedFatManifestBuildInfo(t *testing.T) entities.BuildInfo {
	testDir := tests.GetTestResourcesPath()
	buildinfoFile, err := tests.ReplaceTemplateVariables(filepath.Join(testDir, tests.ExpectedFatManifestBuildInfo), tests.Out)
	assert.NoError(t, err)
	buildinfoFile, err = filepath.Abs(buildinfoFile)
	data, err := ioutil.ReadFile(buildinfoFile)
	assert.NoError(t, err)
	var buildinfo entities.BuildInfo
	assert.NoError(t, json.Unmarshal(data, &buildinfo))
	return buildinfo
}
