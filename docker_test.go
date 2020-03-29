package main

import (
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	gofrogcmd "github.com/jfrog/gofrog/io"
	"github.com/jfrog/jfrog-cli/artifactory/commands/generic"
	"github.com/jfrog/jfrog-cli/artifactory/spec"
	"github.com/jfrog/jfrog-cli/artifactory/utils/docker"
	"github.com/jfrog/jfrog-cli/inttestutils"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
)

const DockerTestImage string = "jfrog_cli_test_image"

// Image get parent image id command
type buildDockerImage struct {
	dockerFilePath string
	dockerTag      string
}

func (image *buildDockerImage) GetCmd() *exec.Cmd {
	var cmd []string
	cmd = append(cmd, "docker")
	cmd = append(cmd, "build")
	cmd = append(cmd, image.dockerFilePath)
	cmd = append(cmd, "--tag", image.dockerTag)
	return exec.Command(cmd[0], cmd[1:]...)
}

func (image *buildDockerImage) GetEnv() map[string]string {
	return map[string]string{}
}

func (image *buildDockerImage) GetStdWriter() io.WriteCloser {
	return nil
}

func (image *buildDockerImage) GetErrWriter() io.WriteCloser {
	return nil
}

type deleteDockerImage struct {
	dockerImageTag string
}

func (image *deleteDockerImage) GetCmd() *exec.Cmd {
	var cmd []string
	cmd = append(cmd, "docker")
	cmd = append(cmd, "image")
	cmd = append(cmd, "rm")
	cmd = append(cmd, image.dockerImageTag)
	return exec.Command(cmd[0], cmd[1:]...)
}

func (image *deleteDockerImage) GetEnv() map[string]string {
	return map[string]string{}
}

func (image *deleteDockerImage) GetStdWriter() io.WriteCloser {
	return nil
}

func (image *deleteDockerImage) GetErrWriter() io.WriteCloser {
	return nil
}

func TestDockerPush(t *testing.T) {
	if !*tests.TestDocker {
		t.Skip("Skipping docker test. To run docker test add the '-test.docker=true' option.")
	}
	runDockerPushTest(DockerTestImage, DockerTestImage+":1", false, t)
}

func TestDockerPushWithModuleName(t *testing.T) {
	if !*tests.TestDocker {
		t.Skip("Skipping docker test. To run docker test add the '-test.docker=true' option.")
	}
	runDockerPushTest(DockerTestImage, ModuleNameJFrogTest, true, t)
}

func TestDockerPushWithMultipleSlash(t *testing.T) {
	if !*tests.TestDocker {
		t.Skip("Skipping docker test. To run docker test add the '-test.docker=true' option.")
	}
	runDockerPushTest(DockerTestImage+"/multiple", "multiple:1", false, t)
}

// Run docker push to Artifactory
func runDockerPushTest(imageName, module string, withModule bool, t *testing.T) {
	imageTag := buildTestDockerImage(imageName)
	buildName := "docker-build"
	buildNumber := "1"

	// Push docker image using docker client
	if withModule {
		artifactoryCli.Exec("docker-push", imageTag, *tests.DockerTargetRepo, "--build-name="+buildName, "--build-number="+buildNumber, "--module="+module)
	} else {
		artifactoryCli.Exec("docker-push", imageTag, *tests.DockerTargetRepo, "--build-name="+buildName, "--build-number="+buildNumber)
	}
	artifactoryCli.Exec("build-publish", buildName, buildNumber)

	imagePath := path.Join(*tests.DockerTargetRepo, imageName, "1") + "/"
	validateDockerBuild(buildName, buildNumber, imagePath, module, 7, 5, 7, t)

	dockerTestCleanup(imageName, buildName)
}

func TestDockerPushBuildNameNumberFromEnv(t *testing.T) {
	if !*tests.TestDocker {
		t.Skip("Skipping docker test. To run docker test add the '-test.docker=true' option.")
	}

	imageName := DockerTestImage
	imageTag := buildTestDockerImage(imageName)
	buildName := "docker-build"
	buildNumber := "1"
	os.Setenv(cliutils.BuildName, buildName)
	os.Setenv(cliutils.BuildNumber, buildNumber)
	defer os.Unsetenv(cliutils.BuildName)
	defer os.Unsetenv(cliutils.BuildNumber)

	// Push docker image using docker client
	artifactoryCli.Exec("docker-push", imageTag, *tests.DockerTargetRepo)
	artifactoryCli.Exec("build-publish")

	imagePath := path.Join(*tests.DockerTargetRepo, imageName, "1") + "/"
	validateDockerBuild(buildName, buildNumber, imagePath, DockerTestImage+":1", 7, 5, 7, t)

	dockerTestCleanup(imageName, buildName)
}

func TestDockerPull(t *testing.T) {
	if !*tests.TestDocker {
		t.Skip("Skipping docker test. To run docker test add the '-test.docker=true' option.")
	}

	imageName := DockerTestImage
	imageTag := buildTestDockerImage(imageName)

	// Push docker image using docker client
	artifactoryCli.Exec("docker-push", imageTag, *tests.DockerTargetRepo)

	buildName := "docker-pull"
	buildNumber := "1"

	// Pull docker image using docker client
	artifactoryCli.Exec("docker-pull", imageTag, *tests.DockerTargetRepo, "--build-name="+buildName, "--build-number="+buildNumber)
	artifactoryCli.Exec("build-publish", buildName, buildNumber)

	imagePath := path.Join(*tests.DockerTargetRepo, imageName, "1") + "/"
	validateDockerBuild(buildName, buildNumber, imagePath, imageName+":1", 0, 7, 7, t)

	buildNumber = "2"
	artifactoryCli.Exec("docker-pull", imageTag, *tests.DockerTargetRepo, "--build-name="+buildName, "--build-number="+buildNumber, "--module="+ModuleNameJFrogTest)
	artifactoryCli.Exec("build-publish", buildName, buildNumber)
	validateDockerBuild(buildName, buildNumber, imagePath, ModuleNameJFrogTest, 0, 7, 7, t)

	dockerTestCleanup(imageName, buildName)
}

func buildTestDockerImage(imageName string) string {
	imageTag := path.Join(*tests.DockerRepoDomain, imageName+":1")
	dockerFilePath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "docker")
	imageBuilder := &buildDockerImage{dockerTag: imageTag, dockerFilePath: dockerFilePath}
	gofrogcmd.RunCmd(imageBuilder)
	return imageTag
}

func deleteTestDockerImage(imageTag string) {
	imageBuilder := &deleteDockerImage{dockerImageTag: imageTag}
	gofrogcmd.RunCmd(imageBuilder)
}

func validateDockerBuild(buildName, buildNumber, imagePath, module string, expectedArtifacts, expectedDependencies, expectedItemsInArtifactory int, t *testing.T) {
	specFile := spec.NewBuilder().Pattern(imagePath + "*").BuildSpec()
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetRtDetails(artifactoryDetails).SetSpec(specFile)
	assert.NoError(t, searchCmd.Search())
	assert.Len(t, searchCmd.SearchResult(), expectedItemsInArtifactory, "Docker build info was not pushed correctly")

	buildInfo := inttestutils.GetBuildInfo(artifactoryDetails.Url, buildName, buildNumber, t, artHttpDetails)
	validateBuildInfo(buildInfo, t, expectedDependencies, expectedArtifacts, module)
}

func dockerTestCleanup(imageName, buildName string) {
	// Remove build from Artifactory
	inttestutils.DeleteBuild(artifactoryDetails.Url, buildName, artHttpDetails)

	// Remove image from Artifactory
	deleteSpec := spec.NewBuilder().Pattern(path.Join(*tests.DockerTargetRepo, imageName)).BuildSpec()
	tests.DeleteFiles(deleteSpec, artifactoryDetails)
}

func TestDockerClientApiVersionCmd(t *testing.T) {
	if !*tests.TestDocker {
		t.Skip("Skipping docker test. To run docker test add the '-test.docker=true' option.")
	}

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
	if !*tests.TestDocker {
		t.Skip("Skipping docker test. To run docker test add the '-test.docker=true' option.")
	}

	imageName := "traefik"
	imageTag := path.Join(*tests.DockerRepoDomain, imageName+":2.2")
	buildName := "docker-pull"
	buildNumber := "1"

	// Pull docker image using docker client
	artifactoryCli.Exec("docker-pull", imageTag, *tests.DockerTargetRepo, "--build-name="+buildName, "--build-number="+buildNumber)
	artifactoryCli.Exec("build-publish", buildName, buildNumber)

	//Validate
	buildInfo := inttestutils.GetBuildInfo(artifactoryDetails.Url, buildName, buildNumber, t, artHttpDetails)
	validateBuildInfo(buildInfo, t, 6, 0, imageName+":2.2")

	dockerTestCleanup(imageName, buildName)
	deleteTestDockerImage(imageTag)
}
