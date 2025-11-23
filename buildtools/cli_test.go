package buildtools

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractDockerBuildOptionsFromArgs(t *testing.T) {
	tests := []struct {
		name               string
		args               []string
		expectedPush       bool
		expectedDockerfile string
		expectedTag        string
		expectedError      error
	}{
		// Push option tests
		{
			name:               "Push option with --push flag",
			args:               []string{"build", "-t", "myimage:latest", "--push"},
			expectedPush:       true,
			expectedDockerfile: "Dockerfile",
			expectedTag:        "myimage:latest",
			expectedError:      nil,
		},
		{
			name:               "Push option with --output type=registry",
			args:               []string{"build", "-t", "myimage:latest", "--output", "type=registry"},
			expectedPush:       true,
			expectedDockerfile: "Dockerfile",
			expectedTag:        "myimage:latest",
			expectedError:      nil,
		},
		{
			name:               "Push option with --output push=true",
			args:               []string{"build", "-t", "myimage:latest", "--output", "push=true"},
			expectedPush:       true,
			expectedDockerfile: "Dockerfile",
			expectedTag:        "myimage:latest",
			expectedError:      nil,
		},
		{
			name:               "Push option with --output containing type=registry",
			args:               []string{"build", "-t", "myimage:latest", "--output", "type=registry,dest=/tmp"},
			expectedPush:       true,
			expectedDockerfile: "Dockerfile",
			expectedTag:        "myimage:latest",
			expectedError:      nil,
		},
		{
			name:               "Push option with --output containing push=true",
			args:               []string{"build", "-t", "myimage:latest", "--output", "push=true,dest=/tmp"},
			expectedPush:       true,
			expectedDockerfile: "Dockerfile",
			expectedTag:        "myimage:latest",
			expectedError:      nil,
		},
		{
			name:               "No push option",
			args:               []string{"build", "-t", "myimage:latest"},
			expectedPush:       false,
			expectedDockerfile: "Dockerfile",
			expectedTag:        "myimage:latest",
			expectedError:      nil,
		},
		{
			name:               "Push option false with --output not matching",
			args:               []string{"build", "-t", "myimage:latest", "--output", "type=local"},
			expectedPush:       false,
			expectedDockerfile: "Dockerfile",
			expectedTag:        "myimage:latest",
			expectedError:      nil,
		},
		{
			name:               "Push option already true, --output should not override",
			args:               []string{"build", "-t", "myimage:latest", "--push", "--output", "type=local"},
			expectedPush:       true,
			expectedDockerfile: "Dockerfile",
			expectedTag:        "myimage:latest",
			expectedError:      nil,
		},
		{
			name:               "Push option with --push and --output type=registry",
			args:               []string{"build", "-t", "myimage:latest", "--push", "--output", "type=registry"},
			expectedPush:       true,
			expectedDockerfile: "Dockerfile",
			expectedTag:        "myimage:latest",
			expectedError:      nil,
		},
		{
			name:               "Push option with --output empty string",
			args:               []string{"build", "-t", "myimage:latest", "--output", ""},
			expectedPush:       false,
			expectedDockerfile: "Dockerfile",
			expectedTag:        "myimage:latest",
			expectedError:      nil,
		},

		// Dockerfile path tests
		{
			name:               "Dockerfile with -f flag",
			args:               []string{"build", "-f", "custom.Dockerfile", "-t", "myimage:latest"},
			expectedPush:       false,
			expectedDockerfile: "custom.Dockerfile",
			expectedTag:        "myimage:latest",
			expectedError:      nil,
		},
		{
			name:               "Dockerfile with --file flag",
			args:               []string{"build", "--file", "custom.Dockerfile", "-t", "myimage:latest"},
			expectedPush:       false,
			expectedDockerfile: "custom.Dockerfile",
			expectedTag:        "myimage:latest",
			expectedError:      nil,
		},
		{
			name:               "Dockerfile default when not specified",
			args:               []string{"build", "-t", "myimage:latest"},
			expectedPush:       false,
			expectedDockerfile: "Dockerfile",
			expectedTag:        "myimage:latest",
			expectedError:      nil,
		},
		{
			name:               "Dockerfile with -f takes precedence over --file",
			args:               []string{"build", "-f", "first.Dockerfile", "--file", "second.Dockerfile", "-t", "myimage:latest"},
			expectedPush:       false,
			expectedDockerfile: "first.Dockerfile",
			expectedTag:        "myimage:latest",
			expectedError:      nil,
		},
		{
			name:               "Dockerfile with --file when -f is empty",
			args:               []string{"build", "-f", "", "--file", "custom.Dockerfile", "-t", "myimage:latest"},
			expectedPush:       false,
			expectedDockerfile: "custom.Dockerfile",
			expectedTag:        "myimage:latest",
			expectedError:      nil,
		},

		// Image tag tests
		{
			name:               "Image tag with -t flag",
			args:               []string{"build", "-t", "myimage:latest"},
			expectedPush:       false,
			expectedDockerfile: "Dockerfile",
			expectedTag:        "myimage:latest",
			expectedError:      nil,
		},
		{
			name:               "Image tag with --tag flag",
			args:               []string{"build", "--tag", "myimage:latest"},
			expectedPush:       false,
			expectedDockerfile: "Dockerfile",
			expectedTag:        "myimage:latest",
			expectedError:      nil,
		},
		{
			name:               "Image tag with -t takes precedence over --tag",
			args:               []string{"build", "-t", "first:tag", "--tag", "second:tag"},
			expectedPush:       false,
			expectedDockerfile: "Dockerfile",
			expectedTag:        "first:tag",
			expectedError:      nil,
		},
		{
			name:               "Image tag with --tag when -t is empty",
			args:               []string{"build", "-t", "", "--tag", "myimage:latest"},
			expectedPush:       false,
			expectedDockerfile: "Dockerfile",
			expectedTag:        "myimage:latest",
			expectedError:      nil,
		},
		{
			name:               "Missing image tag should return error",
			args:               []string{"build"},
			expectedPush:       false,
			expectedDockerfile: "Dockerfile",
			expectedTag:        "",
			expectedError:      errors.New("could not find image tag in the command arguments. Please provide an image tag using the '-t' or '--tag' flag"),
		},
		{
			name:               "Image tag with tag=value format",
			args:               []string{"build", "-t=myimage:latest"},
			expectedPush:       false,
			expectedDockerfile: "Dockerfile",
			expectedTag:        "myimage:latest",
			expectedError:      nil,
		},

		// Combined tests
		{
			name:               "All options combined: push, dockerfile, tag",
			args:               []string{"build", "--push", "-f", "prod.Dockerfile", "-t", "myapp:v1.0"},
			expectedPush:       true,
			expectedDockerfile: "prod.Dockerfile",
			expectedTag:        "myapp:v1.0",
			expectedError:      nil,
		},
		{
			name:               "Push with output type=registry and custom dockerfile",
			args:               []string{"build", "--output", "type=registry", "--file", "dev.Dockerfile", "--tag", "myapp:dev"},
			expectedPush:       true,
			expectedDockerfile: "dev.Dockerfile",
			expectedTag:        "myapp:dev",
			expectedError:      nil,
		},
		{
			name:               "Complex buildx command with all options",
			args:               []string{"buildx", "build", "--platform", "linux/amd64", "--push", "-f", "Dockerfile.multi", "-t", "registry.io/app:latest", "--output", "type=registry"},
			expectedPush:       true,
			expectedDockerfile: "Dockerfile.multi",
			expectedTag:        "registry.io/app:latest",
			expectedError:      nil,
		},
		{
			name:               "Output with multiple values including type=registry",
			args:               []string{"build", "-t", "myimage:latest", "--output", "type=registry,dest=/tmp,compression=gzip"},
			expectedPush:       true,
			expectedDockerfile: "Dockerfile",
			expectedTag:        "myimage:latest",
			expectedError:      nil,
		},
		{
			name:               "Output with multiple values including push=true",
			args:               []string{"build", "-t", "myimage:latest", "--output", "push=true,dest=/tmp,compression=gzip"},
			expectedPush:       true,
			expectedDockerfile: "Dockerfile",
			expectedTag:        "myimage:latest",
			expectedError:      nil,
		},
		{
			name:               "Output with type=registry and push=true both present",
			args:               []string{"build", "-t", "myimage:latest", "--output", "type=registry,push=true"},
			expectedPush:       true,
			expectedDockerfile: "Dockerfile",
			expectedTag:        "myimage:latest",
			expectedError:      nil,
		},
		{
			name:               "Output with partial match should not trigger push",
			args:               []string{"build", "-t", "myimage:latest", "--output", "type=local-registry"},
			expectedPush:       false,
			expectedDockerfile: "Dockerfile",
			expectedTag:        "myimage:latest",
			expectedError:      nil,
		},
		{
			name:               "Output with push=false should not trigger push",
			args:               []string{"build", "-t", "myimage:latest", "--output", "push=false"},
			expectedPush:       false,
			expectedDockerfile: "Dockerfile",
			expectedTag:        "myimage:latest",
			expectedError:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pushOption, dockerfilePath, imageTag, err := extractDockerBuildOptionsFromArgs(tt.args)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedPush, pushOption, "Push option mismatch")
			assert.Equal(t, tt.expectedDockerfile, dockerfilePath, "Dockerfile path mismatch")
			assert.Equal(t, tt.expectedTag, imageTag, "Image tag mismatch")
		})
	}
}
