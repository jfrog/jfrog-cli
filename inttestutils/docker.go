package inttestutils

import (
	"io"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"testing"

	gofrogcmd "github.com/jfrog/gofrog/io"
	"github.com/jfrog/jfrog-cli-core/artifactory/commands/generic"
	"github.com/jfrog/jfrog-cli-core/artifactory/spec"
	"github.com/jfrog/jfrog-cli-core/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/artifactory/utils/container"
	"github.com/jfrog/jfrog-cli-core/utils/config"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/jfrog/jfrog-client-go/utils/io/httputils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/stretchr/testify/assert"
)

// Image get parent image id command
type BuildDockerImage struct {
	DockerFilePath   string
	DockerTag        string
	containerManager container.ContainerManagerType
}

func NewBuildDockerImage(imageTag, dockerFilePath string, containerManager container.ContainerManagerType) *BuildDockerImage {
	return &BuildDockerImage{DockerTag: imageTag, DockerFilePath: dockerFilePath, containerManager: containerManager}
}

func (image *BuildDockerImage) GetCmd() *exec.Cmd {
	var cmd []string
	cmd = append(cmd, "build")
	cmd = append(cmd, image.DockerFilePath)
	cmd = append(cmd, "--tag", image.DockerTag)
	return exec.Command(image.containerManager.String(), cmd[:]...)
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

type RunDockerImage struct {
	Args             []string
	containerManager container.ContainerManagerType
}

func NewRunDockerImage(containerManager container.ContainerManagerType, args ...string) *RunDockerImage {
	return &RunDockerImage{Args: args, containerManager: containerManager}
}

func (run *RunDockerImage) GetCmd() *exec.Cmd {
	var cmd []string
	cmd = append(cmd, "run")
	cmd = append(cmd, run.Args...)
	return exec.Command(run.containerManager.String(), cmd[:]...)
}

func (run *RunDockerImage) GetEnv() map[string]string {
	return map[string]string{}
}

func (run *RunDockerImage) GetStdWriter() io.WriteCloser {
	return nil
}

func (run *RunDockerImage) GetErrWriter() io.WriteCloser {
	return nil
}

type DeleteDockerImage struct {
	imageTag         string
	containerManager container.ContainerManagerType
}

func NewDeleteDockerImage(imageTag string, containerManager container.ContainerManagerType) *DeleteDockerImage {
	return &DeleteDockerImage{imageTag: imageTag, containerManager: containerManager}
}

func (image *DeleteDockerImage) GetCmd() *exec.Cmd {
	var cmd []string
	cmd = append(cmd, "image")
	cmd = append(cmd, "rm")
	cmd = append(cmd, image.imageTag)
	return exec.Command(image.containerManager.String(), cmd[:]...)
}

func (image *DeleteDockerImage) GetEnv() map[string]string {
	return map[string]string{}
}

func (image *DeleteDockerImage) GetStdWriter() io.WriteCloser {
	return nil
}

func (image *DeleteDockerImage) GetErrWriter() io.WriteCloser {
	return nil
}

func BuildTestContainerImage(t *testing.T, imageName string, containerManagerType container.ContainerManagerType) string {
	log.Info("Building image", imageName, "with", containerManagerType.String())
	imageTag := path.Join(*tests.DockerRepoDomain, imageName+":1")
	dockerFilePath := filepath.Join(filepath.FromSlash(tests.GetTestResourcesPath()), "docker")
	imageBuilder := NewBuildDockerImage(imageTag, dockerFilePath, containerManagerType)
	assert.NoError(t, gofrogcmd.RunCmd(imageBuilder))
	return imageTag
}

func DeleteTestContainerImage(t *testing.T, imageTag string, containerManagerType container.ContainerManagerType) {
	imageBuilder := NewDeleteDockerImage(imageTag, containerManagerType)
	assert.NoError(t, gofrogcmd.RunCmd(imageBuilder))
}

func ContainerTestCleanup(t *testing.T, serverDetails *config.ServerDetails, artHttpDetails httputils.HttpClientDetails, imageName, buildName, repo string) {
	// Remove build from Artifactory
	DeleteBuild(serverDetails.ArtifactoryUrl, buildName, artHttpDetails)

	// Remove image from Artifactory
	deleteSpec := spec.NewBuilder().Pattern(path.Join(repo, imageName)).BuildSpec()
	successCount, failCount, err := tests.DeleteFiles(deleteSpec, serverDetails)
	assert.Greater(t, successCount, 0)
	assert.Equal(t, failCount, 0)
	assert.NoError(t, err)
}

func getAllImagesNames(serverDetails *config.ServerDetails) ([]string, error) {
	var imageNames []string
	prefix := *tests.DockerLocalRepo + "/"
	specFile := spec.NewBuilder().Pattern(prefix + tests.DockerImageName + "*").IncludeDirs(true).BuildSpec()
	searchCmd := generic.NewSearchCommand()
	searchCmd.SetServerDetails(serverDetails).SetSpec(specFile)
	reader, err := searchCmd.Search()
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	for searchResult := new(utils.SearchResult); reader.NextRecord(searchResult) == nil; searchResult = new(utils.SearchResult) {
		imageNames = append(imageNames, strings.TrimPrefix(searchResult.Path, prefix))
	}
	return imageNames, err
}

func CleanUpOldImages(serverDetails *config.ServerDetails, artHttpDetails httputils.HttpClientDetails) {
	getActualItems := func() ([]string, error) { return getAllImagesNames(serverDetails) }
	deleteItem := func(imageName string) {
		deleteSpec := spec.NewBuilder().Pattern(path.Join(*tests.DockerLocalRepo, imageName)).BuildSpec()
		tests.DeleteFiles(deleteSpec, serverDetails)
		log.Info("Image", imageName, "deleted.")
	}
	tests.CleanUpOldItems([]string{tests.DockerImageName}, getActualItems, deleteItem)
}
