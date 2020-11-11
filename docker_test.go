package main

import (
	"os"
	"path"
	"strings"
	"testing"

	gofrogcmd "github.com/jfrog/gofrog/io"
	"github.com/jfrog/jfrog-cli-core/artifactory/commands/generic"
	"github.com/jfrog/jfrog-cli-core/artifactory/spec"
	"github.com/jfrog/jfrog-cli-core/artifactory/utils/docker"
	"github.com/jfrog/jfrog-cli-core/utils/coreutils"
	"github.com/jfrog/jfrog-cli/inttestutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
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

func initDockerTest(t *testing.T) {
	if !*tests.TestDocker {
		t.Skip("Skipping docker test. To run docker test add the '-test.docker=true' option.")
	}
}

func TestDockerPush(t *testing.T) {
	initDockerTest(t)
	runDockerPushTest(tests.DockerImageName, tests.DockerImageName+":1", false, t)
}

func TestDockerPushWithModuleName(t *testing.T) {
	initDockerTest(t)
	runDockerPushTest(tests.DockerImageName, ModuleNameJFrogTest, true, t)
}

func TestDockerPushWithMultipleSlash(t *testing.T) {
	initDockerTest(t)
	runDockerPushTest(tests.DockerImageName+"/multiple", "multiple:1", false, t)
}

// Run docker push to Artifactory
func runDockerPushTest(imageName, module string, withModule bool, t *testing.T) {
	imageTag := inttestutils.BuildTestDockerImage(imageName)
	buildNumber := "1"

	// Push docker image using docker client
	if withModule {
		artifactoryCli.Exec("docker-push", imageTag, *tests.DockerTargetRepo, "--build-name="+tests.DockerBuildName, "--build-number="+buildNumber, "--module="+module)
	} else {
		artifactoryCli.Exec("docker-push", imageTag, *tests.DockerTargetRepo, "--build-name="+tests.DockerBuildName, "--build-number="+buildNumber)
	}
	artifactoryCli.Exec("build-publish", tests.DockerBuildName, buildNumber)

	imagePath := path.Join(*tests.DockerTargetRepo, imageName, "1") + "/"
	validateDockerBuild(tests.DockerBuildName, buildNumber, imagePath, module, 7, 5, 7, t)
	inttestutils.DockerTestCleanup(artifactoryDetails, artHttpDetails, imageName, tests.DockerBuildName)

}
func TestDockerPushBuildNameNumberFromEnv(t *testing.T) {
	initDockerTest(t)
	imageTag := inttestutils.BuildTestDockerImage(tests.DockerImageName)
	buildNumber := "1"
	os.Setenv(coreutils.BuildName, tests.DockerBuildName)
	os.Setenv(coreutils.BuildNumber, buildNumber)
	defer os.Unsetenv(coreutils.BuildName)
	defer os.Unsetenv(coreutils.BuildNumber)

	// Push docker image using docker client
	artifactoryCli.Exec("docker-push", imageTag, *tests.DockerTargetRepo)
	artifactoryCli.Exec("build-publish")

	imagePath := path.Join(*tests.DockerTargetRepo, tests.DockerImageName, "1") + "/"
	validateDockerBuild(tests.DockerBuildName, buildNumber, imagePath, tests.DockerImageName+":1", 7, 5, 7, t)

	inttestutils.DockerTestCleanup(artifactoryDetails, artHttpDetails, tests.DockerImageName, tests.DockerBuildName)
}

func TestDockerPull(t *testing.T) {
	initDockerTest(t)

	imageTag := inttestutils.BuildTestDockerImage(tests.DockerImageName)

	// Push docker image using docker client
	artifactoryCli.Exec("docker-push", imageTag, *tests.DockerTargetRepo)

	buildNumber := "1"

	// Pull docker image using docker client
	artifactoryCli.Exec("docker-pull", imageTag, *tests.DockerTargetRepo, "--build-name="+tests.DockerBuildName, "--build-number="+buildNumber)
	artifactoryCli.Exec("build-publish", tests.DockerBuildName, buildNumber)

	imagePath := path.Join(*tests.DockerTargetRepo, tests.DockerImageName, "1") + "/"
	validateDockerBuild(tests.DockerBuildName, buildNumber, imagePath, tests.DockerImageName+":1", 0, 7, 7, t)

	buildNumber = "2"
	artifactoryCli.Exec("docker-pull", imageTag, *tests.DockerTargetRepo, "--build-name="+tests.DockerBuildName, "--build-number="+buildNumber, "--module="+ModuleNameJFrogTest)
	artifactoryCli.Exec("build-publish", tests.DockerBuildName, buildNumber)
	validateDockerBuild(tests.DockerBuildName, buildNumber, imagePath, ModuleNameJFrogTest, 0, 7, 7, t)

	inttestutils.DockerTestCleanup(artifactoryDetails, artHttpDetails, tests.DockerImageName, tests.DockerBuildName)
}

func dockerTestCleanup(imageName, buildName string) {
	// Remove build from Artifactory
	inttestutils.DeleteBuild(artifactoryDetails.Url, buildName, artHttpDetails)
	inttestutils.DockerTestCleanup(artifactoryDetails, artHttpDetails, tests.DockerImageName, tests.DockerBuildName)
}

func TestDockerClientApiVersionCmd(t *testing.T) {
	initDockerTest(t)

	// Run docker version command and expect no errors
	cmd := &docker.VersionCmd{}
	content, err := gofrogcmd.RunCmdOutput(cmd)
	assert.NoError(t, err)

	// Expect VersionRegex to match the output API version
	content = strings.TrimSpace(content)
	assert.True(t, docker.ApiVersionRegex.Match([]byte(content)))

	// Assert docker min API version
	assert.True(t, docker.IsCompatibleApiVersion(content))
}

func TestDockerFatManifestPull(t *testing.T) {
	initDockerTest(t)
	for _, dockerRepo := range [...]string{*tests.DockerRemoteRepo, *tests.DockerVirtualRepo} {
		imageName := "traefik"
		imageTag := path.Join(*tests.DockerRepoDomain, imageName+":2.2")
		buildNumber := "1"

		// Pull docker image using docker client
		assert.NoError(t, artifactoryCli.Exec("docker-pull", imageTag, dockerRepo, "--build-name="+tests.DockerBuildName, "--build-number="+buildNumber))
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

		inttestutils.DockerTestCleanup(artifactoryDetails, artHttpDetails, imageName, tests.DockerBuildName)
		inttestutils.DeleteTestDockerImage(imageTag)
	}
}

func TestDockerPromote(t *testing.T) {
	initDockerTest(t)

	// Build and push image
	imageTag := inttestutils.BuildTestDockerImage(tests.DockerImageName)
	err := artifactoryCli.Exec("docker-push", imageTag, *tests.DockerTargetRepo)
	assert.NoError(t, err)

	// Promote image
	err = artifactoryCli.Exec("docker-promote", tests.DockerImageName, *tests.DockerTargetRepo, tests.DockerRepo, "--source-tag=1", "--target-tag=2", "--target-docker-image=docker-target-image", "--copy")
	assert.NoError(t, err)

	// Verify image in source
	imagePath := path.Join(*tests.DockerTargetRepo, tests.DockerImageName, "1") + "/"
	validateDockerImage(t, imagePath, 7)

	// Verify image promoted
	searchSpec, err := tests.CreateSpec(tests.SearchAllDocker)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetDockerDeployedManifest(), searchSpec, t)

	inttestutils.DockerTestCleanup(artifactoryDetails, artHttpDetails, tests.DockerImageName, tests.DockerBuildName)
	inttestutils.DeleteTestDockerImage(imageTag)
}

func validateDockerBuild(buildName, buildNumber, imagePath, module string, expectedArtifacts, expectedDependencies, expectedItemsInArtifactory int, t *testing.T) {
	validateDockerImage(t, imagePath, expectedItemsInArtifactory)
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

func validateDockerImage(t *testing.T, imagePath string, expectedItemsInArtifactory int) {
	specFile := spec.NewBuilder().Pattern(imagePath + "*").BuildSpec()
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetRtDetails(artifactoryDetails).SetSpec(specFile)
	reader, err := searchCmd.Search()
	assert.NoError(t, err)
	length, err := reader.Length()
	assert.NoError(t, err)
	assert.Equal(t, expectedItemsInArtifactory, length, "Docker build info was not pushed correctly")
	assert.NoError(t, reader.Close())
}
