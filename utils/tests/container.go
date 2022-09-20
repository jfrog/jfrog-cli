package tests

import (
	"context"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

//  TestContainer is a friendly API to run container.
//  It is designed to create runtime environment to use during automatic tests.
type TestContainer struct {
	container testcontainers.Container
}

// Run a command in a running container
func (tc *TestContainer) Exec(ctx context.Context, cmd ...string) error {
	exitCode, reader, err := tc.container.Exec(ctx, cmd)
	if err != nil {
		return err
	}
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	fmt.Printf(string(data))
	if exitCode != 0 {
		return fmt.Errorf("Container Exec command returned exit code %d.", exitCode)
	}
	return nil
}
func (tc *TestContainer) Terminate(ctx context.Context) error {
	return tc.container.Terminate(ctx)
}

// ContainerRequest represents the parameters used to create and run a container.
type ContainerRequest struct {
	request testcontainers.ContainerRequest
}

func NewContainerRequest() *ContainerRequest {
	return &ContainerRequest{
		request: testcontainers.ContainerRequest{}}
}

// FromDockerfile represents the parameters needed to build an image from a Dockerfile
// rather than using a pre-built image.
// This setter cannot be used with 'SetImage' to run a container..
//
// context - The path to the context of of the docker build
// file - The path from the context to the Dockerfile for the image, defaults to "Dockerfile"
// BuildArgs - Args to docker daemon
func (c *ContainerRequest) SetDockerfile(context, file string, BuildArgs map[string]*string) *ContainerRequest {
	c.request.FromDockerfile = testcontainers.FromDockerfile{
		Context:       context,
		Dockerfile:    file,
		BuildArgs:     BuildArgs,
		PrintBuildLog: true,
	}
	return c
}

// Use a pre-built image to run the container.
// rather than using a docker file.
// This setter cannot be used with 'SetDockerfile' to run a container.
func (c *ContainerRequest) Image(image string) *ContainerRequest {
	c.request.Image = image
	return c
}

// Set  tag in the 'name:tag' format
func (c *ContainerRequest) Name(name string) *ContainerRequest {
	c.request.Name = name
	return c
}

// Give extended privileges to this container
func (c *ContainerRequest) Privileged() *ContainerRequest {
	c.request.Privileged = true
	return c
}

// Connect a container to one or more networks
func (c *ContainerRequest) Networks(networks ...string) *ContainerRequest {
	c.request.Networks = networks
	return c
}

// Removed the container from the host when stopped.
func (c *ContainerRequest) Remove() *ContainerRequest {
	c.request.AutoRemove = true
	c.request.SkipReaper = true
	return c
}

// Mounts the 'hostPath' working directory from localhost into the container.
// source - it is the path where the mount should be mounted within the localhost
// Target - It's the path where the mount should be mounted within the container
func (c *ContainerRequest) Mount(hostPath, target string, readOnly bool) *ContainerRequest {
	c.request.Mounts = append(c.request.Mounts, testcontainers.ContainerMount{Source: testcontainers.GenericBindMountSource{HostPath: hostPath}, Target: testcontainers.ContainerMountTarget(target), ReadOnly: readOnly})
	return c
}

//  When the container starts, set command instructions (shell for example).
func (c *ContainerRequest) Cmd(cmd ...string) *ContainerRequest {
	c.request.Cmd = cmd
	return c
}

// Set wait strategy to detect when the container is read.
// For example, the wait.ForHTTP("/home") strategy waits for a 200 response from the container's '/home' path.
func (c *ContainerRequest) WaitFor(waitingFor wait.Strategy) *ContainerRequest {
	c.request.WaitingFor = waitingFor
	return c
}

// Creates a container based on container request parameters.
func (c *ContainerRequest) Build(ctx context.Context, t *testing.T, autoStart bool) (*TestContainer, error) {
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: c.request,
		Started:          autoStart,
		Reuse:            false,
	})
	if err != nil {
		return nil, err
	}
	reader, err := container.Logs(ctx)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	fmt.Printf(string(data))
	return &TestContainer{container: container}, nil
}
