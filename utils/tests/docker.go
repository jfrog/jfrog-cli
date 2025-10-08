package tests

import (
	"io"
	"os/exec"
	"path"
	"testing"

	gofrogcmd "github.com/jfrog/gofrog/io"
	container "github.com/jfrog/jfrog-cli-artifactory/artifactory/commands/ocicontainer"
	"github.com/stretchr/testify/assert"
)

// Image get parent image id command
type BuildDockerImage struct {
	// The build command builds images from a Dockerfile and a context.
	// A build's context is the set of files located in the specified PATH.
	// The build process can refer to any of the files in the context.
	// For example, The build can use a COPY instruction to reference a file in the context.
	buildContext     string
	dockerFileName   string
	imageName        string
	containerManager container.ContainerManagerType
}

func NewBuildDockerImage(imageTag, dockerFilePath string, containerManager container.ContainerManagerType) *BuildDockerImage {
	return &BuildDockerImage{imageName: imageTag, buildContext: dockerFilePath, containerManager: containerManager}
}

func (image *BuildDockerImage) SetDockerFileName(name string) *BuildDockerImage {
	image.dockerFileName = name
	return image
}

func (image *BuildDockerImage) GetCmd() *exec.Cmd {
	var cmd []string
	cmd = append(cmd, "build")
	cmd = append(cmd, "--tag", image.imageName)
	if image.dockerFileName != "" {
		cmd = append(cmd, "--file", path.Join(image.buildContext, image.dockerFileName))

	}
	cmd = append(cmd, image.buildContext)
	return exec.Command(image.containerManager.String(), cmd...)
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

// The ExecDockerImage command runs a new command in a running container.
type ExecDockerImage struct {
	Args             []string
	errCloser        io.WriteCloser
	stdWriter        io.WriteCloser
	containerManager container.ContainerManagerType
}

func NewExecDockerImage(containerManager container.ContainerManagerType, args ...string) *ExecDockerImage {
	return &ExecDockerImage{Args: args, containerManager: containerManager}
}

func (e *ExecDockerImage) GetCmd() *exec.Cmd {
	var cmd []string
	cmd = append(cmd, "exec")
	cmd = append(cmd, e.Args...)
	return exec.Command(e.containerManager.String(), cmd...)
}

func (e *ExecDockerImage) GetEnv() map[string]string {
	return map[string]string{}
}

func (e *ExecDockerImage) GetStdWriter() io.WriteCloser {
	return e.stdWriter
}

func (e *ExecDockerImage) SetStdWriter(writer io.WriteCloser) {
	e.stdWriter = writer
}

func (e *ExecDockerImage) GetErrWriter() io.WriteCloser {
	return e.errCloser
}

func (e *ExecDockerImage) SetErrWriter(writer io.WriteCloser) {
	e.errCloser = writer
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
	return exec.Command(run.containerManager.String(), cmd...)
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
	return exec.Command(image.containerManager.String(), cmd...)
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

type DeleteContainer struct {
	containerName    string
	containerManager container.ContainerManagerType
}

func NewDeleteContainer(containerName string, containerManager container.ContainerManagerType) *DeleteContainer {
	return &DeleteContainer{containerName: containerName, containerManager: containerManager}
}

func (image *DeleteContainer) GetCmd() *exec.Cmd {
	var cmd []string
	cmd = append(cmd, "rm")
	cmd = append(cmd, "--force")
	cmd = append(cmd, image.containerName)
	return exec.Command(image.containerManager.String(), cmd...)
}

func (image *DeleteContainer) GetEnv() map[string]string {
	return map[string]string{}
}

func (image *DeleteContainer) GetStdWriter() io.WriteCloser {
	return nil
}

func (image *DeleteContainer) GetErrWriter() io.WriteCloser {
	return nil
}

func DeleteTestImage(t *testing.T, imageTag string, containerManagerType container.ContainerManagerType) {
	imageBuilder := NewDeleteDockerImage(imageTag, containerManagerType)
	assert.NoError(t, gofrogcmd.RunCmd(imageBuilder))
}

func DeleteTestContainer(t *testing.T, containerName string, containerManagerType container.ContainerManagerType) {
	containerDelete := NewDeleteContainer(containerName, containerManagerType)
	assert.NoError(t, gofrogcmd.RunCmd(containerDelete))
}
