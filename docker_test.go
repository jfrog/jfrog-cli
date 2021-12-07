package main

import (
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	buildinfo "github.com/jfrog/build-info-go/entities"

	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/v2/common/spec"
	clientutils "github.com/jfrog/jfrog-client-go/utils"

	gofrogcmd "github.com/jfrog/gofrog/io"
	corecontainer "github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/container"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/generic"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils/container"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-cli/inttestutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/stretchr/testify/assert"
)

const (
	kanikoImage = "gcr.io/kaniko-project/executor:latest"
)

func InitDockerTests() {
	initArtifactoryCli()
	cleanUpOldBuilds()
	inttestutils.CleanUpOldImages(serverDetails, artHttpDetails)
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
			imageTag := inttestutils.BuildTestContainerImage(t, imageName, containerManager)
			buildNumber := "1"
			dockerPushCommand := corecontainer.NewPushCommand(containerManager)

			// Testing detailed summary without buildinfo
			dockerPushCommand.SetThreads(1).SetDetailedSummary(true).SetBuildConfiguration(new(utils.BuildConfiguration)).SetRepo(*tests.DockerLocalRepo).SetServerDetails(serverDetails).SetImageTag(imageTag)
			assert.NoError(t, dockerPushCommand.Run())
			result := dockerPushCommand.Result()
			result.Reader()
			reader := result.Reader()
			defer func() {
				assert.NoError(t, reader.Close())
			}()
			assert.NoError(t, reader.GetError())
			for transferDetails := new(clientutils.FileTransferDetails); reader.NextRecord(transferDetails) == nil; transferDetails = new(clientutils.FileTransferDetails) {
				assert.Equal(t, 64, len(transferDetails.Sha256), "Summary validation failed - invalid sha256 has returned from artifactory")
			}
			// Testing detailed summary with buildinfo
			assert.NoError(t, reader.Close())
			dockerPushCommand.SetBuildConfiguration(utils.NewBuildConfiguration(tests.DockerBuildName, buildNumber, "", ""))
			assert.NoError(t, dockerPushCommand.Run())
			result = dockerPushCommand.Result()
			reader = result.Reader()
			assert.NoError(t, reader.GetError())
			for transferDetails := new(clientutils.FileTransferDetails); reader.NextRecord(transferDetails) == nil; transferDetails = new(clientutils.FileTransferDetails) {
				assert.Equal(t, 64, len(transferDetails.Sha256), "Summary validation failed - invalid sha256 has returned from artifactory")
			}

			inttestutils.ValidateGeneratedBuildInfoModule(t, tests.DockerBuildName, buildNumber, "", []string{module}, buildinfo.Docker)
			RunRt(t, "build-publish", tests.DockerBuildName, buildNumber)

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
	imageTag := inttestutils.BuildTestContainerImage(t, imageName, containerManager)
	buildNumber := "1"

	// Push image
	if withModule {
		RunRt(t, containerManager.String()+"-push", imageTag, repo, "--build-name="+tests.DockerBuildName, "--build-number="+buildNumber, "--module="+module)
	} else {
		RunRt(t, containerManager.String()+"-push", imageTag, repo, "--build-name="+tests.DockerBuildName, "--build-number="+buildNumber)
	}

	inttestutils.ValidateGeneratedBuildInfoModule(t, tests.DockerBuildName, buildNumber, "", []string{module}, buildinfo.Docker)
	RunRt(t, "build-publish", tests.DockerBuildName, buildNumber)

	imagePath := path.Join(repo, imageName, "1") + "/"
	validateContainerBuild(tests.DockerBuildName, buildNumber, imagePath, module, 7, 5, 7, t)
	inttestutils.ContainerTestCleanup(t, serverDetails, artHttpDetails, imageName, tests.DockerBuildName, repo)
}

func TestContainerPushBuildNameNumberFromEnv(t *testing.T) {
	containerManagers := initContainerTest(t)
	for _, containerManager := range containerManagers {
		imageTag := inttestutils.BuildTestContainerImage(t, tests.DockerImageName, containerManager)
		buildNumber := "1"
		assert.NoError(t, os.Setenv(coreutils.BuildName, tests.DockerBuildName))
		assert.NoError(t, os.Setenv(coreutils.BuildNumber, buildNumber))
		defer func() {
			assert.NoError(t, os.Unsetenv(coreutils.BuildName))
			assert.NoError(t, os.Unsetenv(coreutils.BuildNumber))
		}()
		// Push container image
		RunRt(t, containerManager.String()+"-push", imageTag, *tests.DockerLocalRepo)
		RunRt(t, "build-publish")

		imagePath := path.Join(*tests.DockerLocalRepo, tests.DockerImageName, "1") + "/"
		validateContainerBuild(tests.DockerBuildName, buildNumber, imagePath, tests.DockerImageName+":1", 7, 5, 7, t)

		inttestutils.ContainerTestCleanup(t, serverDetails, artHttpDetails, tests.DockerImageName, tests.DockerBuildName, *tests.DockerLocalRepo)
	}

}

func TestContainerPull(t *testing.T) {
	containerManagers := initContainerTest(t)
	for _, repo := range []string{*tests.DockerVirtualRepo, *tests.DockerLocalRepo} {
		for _, containerManager := range containerManagers {
			imageTag := inttestutils.BuildTestContainerImage(t, tests.DockerImageName, containerManager)

			// Push container image
			RunRt(t, containerManager.String()+"-push", imageTag, repo)

			buildNumber := "1"

			// Pull container image
			RunRt(t, containerManager.String()+"-pull", imageTag, repo, "--build-name="+tests.DockerBuildName, "--build-number="+buildNumber)
			RunRt(t, "build-publish", tests.DockerBuildName, buildNumber)

			imagePath := path.Join(repo, tests.DockerImageName, "1") + "/"
			validateContainerBuild(tests.DockerBuildName, buildNumber, imagePath, tests.DockerImageName+":1", 0, 7, 7, t)

			buildNumber = "2"
			RunRt(t, containerManager.String()+"-pull", imageTag, repo, "--build-name="+tests.DockerBuildName, "--build-number="+buildNumber, "--module="+ModuleNameJFrogTest)
			RunRt(t, "build-publish", tests.DockerBuildName, buildNumber)
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
			RunRt(t, containerManager.String()+"-pull", imageTag, dockerRepo, "--build-name="+tests.DockerBuildName, "--build-number="+buildNumber)
			RunRt(t, "build-publish", tests.DockerBuildName, buildNumber)

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
			inttestutils.DeleteTestContainerImage(t, imageTag, containerManager)
		}
	}
}

func TestDockerPromote(t *testing.T) {
	initContainerTest(t)

	// Build and push image
	imageTag := inttestutils.BuildTestContainerImage(t, tests.DockerImageName, container.DockerClient)
	RunRt(t, "docker-push", imageTag, *tests.DockerLocalRepo)

	// Promote image
	RunRt(t, "docker-promote", tests.DockerImageName, *tests.DockerLocalRepo, tests.DockerRepo, "--source-tag=1", "--target-tag=2", "--target-docker-image=docker-target-image", "--copy")

	// Verify image in source
	imagePath := path.Join(*tests.DockerLocalRepo, tests.DockerImageName, "1") + "/"
	validateContainerImage(t, imagePath, 7)

	// Verify image promoted
	searchSpec, err := tests.CreateSpec(tests.SearchAllDocker)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetDockerDeployedManifest(), searchSpec, t)

	inttestutils.ContainerTestCleanup(t, serverDetails, artHttpDetails, tests.DockerImageName, tests.DockerBuildName, *tests.DockerLocalRepo)
	inttestutils.DeleteTestContainerImage(t, imageTag, container.DockerClient)
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
	assert.NoError(t, reader.Close())
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
		RunRt(t, "build-docker-create", repo, "--image-file="+kanikoOutput, "--build-name="+tests.DockerBuildName, "--build-number="+buildNumber)
		RunRt(t, "build-publish", tests.DockerBuildName, buildNumber)

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
		inttestutils.DeleteTestContainerImage(t, kanikoImage, container.DockerClient)
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
