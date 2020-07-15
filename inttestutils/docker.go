package inttestutils

import (
	"io"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	gofrogcmd "github.com/jfrog/gofrog/io"
	"github.com/jfrog/jfrog-cli/artifactory/commands/generic"
	"github.com/jfrog/jfrog-cli/artifactory/spec"
	"github.com/jfrog/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/jfrog/jfrog-client-go/utils/io/httputils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

// Image get parent image id command
type BuildDockerImage struct {
	DockerFilePath string
	DockerTag      string
}

func (image *BuildDockerImage) GetCmd() *exec.Cmd {
	var cmd []string
	cmd = append(cmd, "docker")
	cmd = append(cmd, "build")
	cmd = append(cmd, image.DockerFilePath)
	cmd = append(cmd, "--tag", image.DockerTag)
	return exec.Command(cmd[0], cmd[1:]...)
}

func (image *BuildDockerImage) GetEnv() map[string]string {
	return map[string]string{}
}

func (image *BuildDockerImage) GetStdWriter() io.WriteCloser {
	return nil
}

func (image *BuildDockerImage) GetErrWriter() io.WriteCloser {
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

func BuildTestDockerImage(imageName string) string {
	imageTag := path.Join(*tests.DockerRepoDomain, imageName+":1")
	dockerFilePath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "docker")
	imageBuilder := &BuildDockerImage{DockerTag: imageTag, DockerFilePath: dockerFilePath}
	gofrogcmd.RunCmd(imageBuilder)
	return imageTag
}

func DeleteTestDockerImage(imageTag string) {
	imageBuilder := &deleteDockerImage{dockerImageTag: imageTag}
	gofrogcmd.RunCmd(imageBuilder)
}

func DockerTestCleanup(artifactoryDetails *config.ArtifactoryDetails, artHttpDetails httputils.HttpClientDetails, imageName, buildName string) {
	// Remove build from Artifactory
	DeleteBuild(artifactoryDetails.Url, buildName, artHttpDetails)

	// Remove image from Artifactory
	deleteSpec := spec.NewBuilder().Pattern(path.Join(*tests.DockerTargetRepo, imageName)).BuildSpec()
	tests.DeleteFiles(deleteSpec, artifactoryDetails)
}

func getAllImagesNames(artifactoryDetails *config.ArtifactoryDetails) ([]string, error) {
	var imageNames []string
	prefix := *tests.DockerTargetRepo + "/"
	specFile := spec.NewBuilder().Pattern(prefix + tests.DockerImageName + "*").IncludeDirs(true).BuildSpec()
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetRtDetails(artifactoryDetails).SetSpec(specFile)
	err := searchCmd.Search()
	if err != nil {
		return nil, err
	}
	for _, v := range searchCmd.SearchResult() {
		imageNames = append(imageNames, strings.TrimPrefix(v.Path, prefix))
	}

	return imageNames, err
}

func CleanUpOldImages(artifactoryDetails *config.ArtifactoryDetails, artHttpDetails httputils.HttpClientDetails) {
	getActualItems := func() ([]string, error) { return getAllImagesNames(artifactoryDetails) }
	deleteItem := func(imageName string) {
		deleteSpec := spec.NewBuilder().Pattern(path.Join(*tests.DockerTargetRepo, imageName)).BuildSpec()
		tests.DeleteFiles(deleteSpec, artifactoryDetails)
		log.Info("Image", imageName, "deleted.")
	}
	tests.CleanUpOldItems([]string{tests.DockerImageName}, getActualItems, deleteItem)
}
