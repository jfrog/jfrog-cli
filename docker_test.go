package main

import (
	"os"
	"path"
	"strings"
	"testing"

	gofrogcmd "github.com/jfrog/gofrog/io"
	"github.com/jfrog/jfrog-cli-core/artifactory/commands/generic"
	"github.com/jfrog/jfrog-cli-core/artifactory/spec"
	"github.com/jfrog/jfrog-cli-core/artifactory/utils/container"
	"github.com/jfrog/jfrog-cli-core/utils/coreutils"
	"github.com/jfrog/jfrog-cli/inttestutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/stretchr/testify/assert"
)

func InitDockerTests() {
	initArtifactoryCli()
	cleanUpOldBuilds()
	inttestutils.CleanUpOldImages(artifactoryDetails, artHttpDetails)
	cleanUpOldRepositories()
	tests.AddTimestampToGlobalVars()
	createRequiredRepos()
}

func initContainerTest(t *testing.T) {
	if !*tests.TestDocker {
		t.Skip("Skipping docker/podman test. To run docker test add the '-test.docker=true' option.")
	}
}

func TestContainerPush(t *testing.T) {
	initContainerTest(t)
	containerManagers := []container.ContainerManagerType{container.Docker}
	if coreutils.IsLinux() {
		containerManagers = append(containerManagers, container.Podman)
	}
	for _, containerManager := range containerManagers {
		runPushTest(containerManager, tests.DockerImageName, tests.DockerImageName+":1", false, t)
	}
}

func TestContainerPushWithModuleName(t *testing.T) {
	initContainerTest(t)
	containerManagers := []container.ContainerManagerType{container.Docker}
	if coreutils.IsLinux() {
		containerManagers = append(containerManagers, container.Podman)
	}
	for _, containerManager := range containerManagers {
		runPushTest(containerManager, tests.DockerImageName, ModuleNameJFrogTest, true, t)
	}
}

func TestContainerPushWithMultipleSlash(t *testing.T) {
	initContainerTest(t)
	containerManagers := []container.ContainerManagerType{container.Docker}
	if coreutils.IsLinux() {
		containerManagers = append(containerManagers, container.Podman)
	}
	for _, containerManager := range containerManagers {

		runPushTest(containerManager, tests.DockerImageName+"/multiple", "multiple:1", false, t)
	}
}

// Run container push to Artifactory
func runPushTest(containerManager container.ContainerManagerType, imageName, module string, withModule bool, t *testing.T) {
	imageTag := inttestutils.BuildTestContainerImage(imageName, containerManager)
	buildNumber := "1"

	// Push image
	if withModule {
		artifactoryCli.Exec(containerManager.String()+"-push", imageTag, *tests.DockerTargetRepo, "--build-name="+tests.DockerBuildName, "--build-number="+buildNumber, "--module="+module)
	} else {
		artifactoryCli.Exec(containerManager.String()+"-push", imageTag, *tests.DockerTargetRepo, "--build-name="+tests.DockerBuildName, "--build-number="+buildNumber)
	}

	inttestutils.ValidateGeneratedBuildInfoModule(t, tests.DockerBuildName, buildNumber, []string{module}, buildinfo.Docker)
	artifactoryCli.Exec("build-publish", tests.DockerBuildName, buildNumber)

	imagePath := path.Join(*tests.DockerTargetRepo, imageName, "1") + "/"
	validateContainerBuild(tests.DockerBuildName, buildNumber, imagePath, module, 7, 5, 7, t)
	inttestutils.ContainerTestCleanup(artifactoryDetails, artHttpDetails, imageName, tests.DockerBuildName)

}
func TestContainerPushBuildNameNumberFromEnv(t *testing.T) {
	initContainerTest(t)
	containerManagers := []container.ContainerManagerType{container.Docker}
	if coreutils.IsLinux() {
		containerManagers = append(containerManagers, container.Podman)
	}
	for _, containerManager := range containerManagers {
		imageTag := inttestutils.BuildTestContainerImage(tests.DockerImageName, containerManager)
		buildNumber := "1"
		os.Setenv(coreutils.BuildName, tests.DockerBuildName)
		os.Setenv(coreutils.BuildNumber, buildNumber)
		defer os.Unsetenv(coreutils.BuildName)
		defer os.Unsetenv(coreutils.BuildNumber)

		// Push container image
		artifactoryCli.Exec(containerManager.String()+"-push", imageTag, *tests.DockerTargetRepo)
		artifactoryCli.Exec("build-publish")

		imagePath := path.Join(*tests.DockerTargetRepo, tests.DockerImageName, "1") + "/"
		validateContainerBuild(tests.DockerBuildName, buildNumber, imagePath, tests.DockerImageName+":1", 7, 5, 7, t)

		inttestutils.ContainerTestCleanup(artifactoryDetails, artHttpDetails, tests.DockerImageName, tests.DockerBuildName)
	}
}

func TestContainerPull(t *testing.T) {
	initContainerTest(t)
	containerManagers := []container.ContainerManagerType{container.Docker}
	if coreutils.IsLinux() {
		containerManagers = append(containerManagers, container.Podman)
	}
	for _, containerManager := range containerManagers {
		imageTag := inttestutils.BuildTestContainerImage(tests.DockerImageName, containerManager)

		// Push container image
		artifactoryCli.Exec(containerManager.String()+"-push", imageTag, *tests.DockerTargetRepo)

		buildNumber := "1"

		// Pull container image
		artifactoryCli.Exec(containerManager.String()+"-pull", imageTag, *tests.DockerTargetRepo, "--build-name="+tests.DockerBuildName, "--build-number="+buildNumber)
		artifactoryCli.Exec("build-publish", tests.DockerBuildName, buildNumber)

		imagePath := path.Join(*tests.DockerTargetRepo, tests.DockerImageName, "1") + "/"
		validateContainerBuild(tests.DockerBuildName, buildNumber, imagePath, tests.DockerImageName+":1", 0, 7, 7, t)

		buildNumber = "2"
		artifactoryCli.Exec(containerManager.String()+"-pull", imageTag, *tests.DockerTargetRepo, "--build-name="+tests.DockerBuildName, "--build-number="+buildNumber, "--module="+ModuleNameJFrogTest)
		artifactoryCli.Exec("build-publish", tests.DockerBuildName, buildNumber)
		validateContainerBuild(tests.DockerBuildName, buildNumber, imagePath, ModuleNameJFrogTest, 0, 7, 7, t)

		inttestutils.ContainerTestCleanup(artifactoryDetails, artHttpDetails, tests.DockerImageName, tests.DockerBuildName)
	}
}

func containerTestCleanup(imageName, buildName string) {
	// Remove build from Artifactory
	inttestutils.DeleteBuild(artifactoryDetails.Url, buildName, artHttpDetails)
	inttestutils.ContainerTestCleanup(artifactoryDetails, artHttpDetails, tests.DockerImageName, tests.DockerBuildName)
}

func TestDockerClientApiVersionCmd(t *testing.T) {
	initContainerTest(t)

	// Run docker version command and expect no errors
	cmd := &container.VersionCmd{}
	content, err := gofrogcmd.RunCmdOutput(cmd)
	assert.NoError(t, err)

	// Expect VersionRegex to match the output API version
	content = strings.TrimSpace(content)
	assert.True(t, container.ApiVersionRegex.Match([]byte(content)))

	// Assert docker min API version
	assert.True(t, container.IsCompatibleApiVersion(content))
}

func TestContainerFatManifestPull(t *testing.T) {
	initContainerTest(t)
	containerManagers := []container.ContainerManagerType{container.Docker}
	if coreutils.IsLinux() {
		containerManagers = append(containerManagers, container.Podman)
	}
	for _, containerManager := range containerManagers {
		for _, dockerRepo := range [...]string{*tests.DockerRemoteRepo, *tests.DockerVirtualRepo} {
			imageName := "traefik"
			imageTag := path.Join(*tests.DockerRepoDomain, imageName+":2.2")
			buildNumber := "1"

			// Pull container image
			assert.NoError(t, artifactoryCli.Exec(containerManager.String()+"-pull", imageTag, dockerRepo, "--build-name="+tests.DockerBuildName, "--build-number="+buildNumber))
			assert.NoError(t, artifactoryCli.Exec("build-publish", tests.DockerBuildName, buildNumber))

			// Validate
			publishedBuildInfo, found, err := tests.GetBuildInfo(artifactoryDetails, tests.DockerBuildName, buildNumber)
			if err != nil {
				assert.NoError(t, err)
				return
			}
			if !found {
				assert.True(t, found, "build info was expected to be found")
				return
			}
			buildInfo := publishedBuildInfo.BuildInfo
			validateBuildInfo(buildInfo, t, 6, 0, imageName+":2.2")

			inttestutils.ContainerTestCleanup(artifactoryDetails, artHttpDetails, imageName, tests.DockerBuildName)
			inttestutils.DeleteTestContainerImage(imageTag, containerManager)
		}
	}
}

func TestDockerPromote(t *testing.T) {
	initContainerTest(t)

	// Build and push image
	imageTag := inttestutils.BuildTestContainerImage(tests.DockerImageName, container.Docker)
	err := artifactoryCli.Exec("docker-push", imageTag, *tests.DockerTargetRepo)
	assert.NoError(t, err)

	// Promote image
	err = artifactoryCli.Exec("docker-promote", tests.DockerImageName, *tests.DockerTargetRepo, tests.DockerRepo, "--source-tag=1", "--target-tag=2", "--target-docker-image=docker-target-image", "--copy")
	assert.NoError(t, err)

	// Verify image in source
	imagePath := path.Join(*tests.DockerTargetRepo, tests.DockerImageName, "1") + "/"
	validateContainerImage(t, imagePath, 7)

	// Verify image promoted
	searchSpec, err := tests.CreateSpec(tests.SearchAllDocker)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetDockerDeployedManifest(), searchSpec, t)

	inttestutils.ContainerTestCleanup(artifactoryDetails, artHttpDetails, tests.DockerImageName, tests.DockerBuildName)
	inttestutils.DeleteTestContainerImage(imageTag, container.Docker)
}

func validateContainerBuild(buildName, buildNumber, imagePath, module string, expectedArtifacts, expectedDependencies, expectedItemsInArtifactory int, t *testing.T) {
	validateContainerImage(t, imagePath, expectedItemsInArtifactory)
	publishedBuildInfo, found, err := tests.GetBuildInfo(artifactoryDetails, buildName, buildNumber)
	if err != nil {
		assert.NoError(t, err)
		return
	}
	if !found {
		assert.True(t, found, "build info was expected to be found")
		return
	}
	buildInfo := publishedBuildInfo.BuildInfo
	validateBuildInfo(buildInfo, t, expectedDependencies, expectedArtifacts, module)
}

func validateContainerImage(t *testing.T, imagePath string, expectedItemsInArtifactory int) {
	specFile := spec.NewBuilder().Pattern(imagePath + "*").BuildSpec()
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetRtDetails(artifactoryDetails).SetSpec(specFile)
	reader, err := searchCmd.Search()
	assert.NoError(t, err)
	length, err := reader.Length()
	assert.NoError(t, err)
	assert.Equal(t, expectedItemsInArtifactory, length, "Container build info was not pushed correctly")
	assert.NoError(t, reader.Close())
}
