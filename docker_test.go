package main

import (
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"testing"

	gofrogcmd "github.com/jfrog/gofrog/io"
	"github.com/jfrog/jfrog-cli-go/artifactory/commands/generic"
	"github.com/jfrog/jfrog-cli-go/artifactory/spec"
	"github.com/jfrog/jfrog-cli-go/inttestutils"
	"github.com/jfrog/jfrog-cli-go/utils/cliutils"
	"github.com/jfrog/jfrog-cli-go/utils/tests"
	"github.com/jfrog/jfrog-client-go/utils/log"
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

func InitDockerTests() {
	if !*tests.TestDocker {
		return
	}
	os.Setenv(cliutils.ReportUsage, "false")
	os.Setenv(cliutils.OfferConfig, "false")
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

func validateDockerBuild(buildName, buildNumber, imagePath, module string, expectedArtifacts, expectedDependencies, expectedItemsInArtifactory int, t *testing.T) {
	specFile := spec.NewBuilder().Pattern(imagePath + "*").BuildSpec()
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetRtDetails(artifactoryDetails).SetSpec(specFile)
	err := searchCmd.Search()
	if err != nil {
		log.Error(err)
		t.Error(err)
	}
	if expectedItemsInArtifactory != len(searchCmd.SearchResult()) {
		t.Error("Docker build info was not pushed correctly, expected:", expectedArtifacts, " Found:", len(searchCmd.SearchResult()))
	}

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
