package lifecycle

import (
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateCreateReleaseBundleContext(t *testing.T) {
	testRuns := []struct {
		name        string
		args        []string
		flags       []string
		expectError bool
	}{
		{"withoutArgs", []string{}, []string{}, true},
		{"oneArg", []string{"one"}, []string{}, true},
		{"twoArgs", []string{"one", "two"}, []string{}, true},
		{"extraArgs", []string{"one", "two", "three", "four"}, []string{}, true},
		{"bothSources", []string{"one", "two", "three"}, []string{cliutils.Builds + "=/path/to/file", cliutils.ReleaseBundles + "=/path/to/file"}, true},
		{"noSources", []string{"one", "two", "three"}, []string{}, true},
		{"builds without signing key", []string{"name", "version"}, []string{cliutils.Builds + "=/path/to/file"}, true},
		{"builds correct", []string{"name", "version"}, []string{
			cliutils.Builds + "=/path/to/file", cliutils.SigningKey + "=key"}, false},
		{"releaseBundles without signing key", []string{"name", "version", "env"}, []string{cliutils.ReleaseBundles + "=/path/to/file"}, true},
		{"releaseBundles correct", []string{"name", "version"}, []string{
			cliutils.ReleaseBundles + "=/path/to/file", cliutils.SigningKey + "=key"}, false},
		{"spec without signing key", []string{"name", "version", "env"}, []string{"spec=/path/to/file"}, true},
		{"spec correct", []string{"name", "version"}, []string{
			"spec=/path/to/file", cliutils.SigningKey + "=key"}, false},
	}

	for _, test := range testRuns {
		t.Run(test.name, func(t *testing.T) {
			context, buffer := tests.CreateContext(t, test.flags, test.args)
			err := validateCreateReleaseBundleContext(context)
			if test.expectError {
				assert.Error(t, err, buffer)
			} else {
				assert.NoError(t, err, buffer)
			}
		})
	}
}

// Validates that the project option does not override the project field in the spec file.
func TestCreateReleaseBundleSpecWithProject(t *testing.T) {
	projectKey := "myproj"
	specFile := filepath.Join("testdata", "specfile.json")
	context, _ := tests.CreateContext(t, []string{"spec=" + specFile, "project=" + projectKey}, []string{})
	creationSpec, err := getReleaseBundleCreationSpec(context)
	assert.NoError(t, err)
	assert.Equal(t, creationSpec.Get(0).Pattern, "path/to/file")
	creationSpec.Get(0).Project = ""
	assert.Equal(t, projectKey, cliutils.GetProject(context))
}

func TestGetProjectPriorities(t *testing.T) {
	testCases := []struct {
		name     string
		flag     string
		envBuild string
		envCli   string
		expected string
	}{
		{"Flag provided", "flagProject", "", "", "flagProject"},
		{"JFROG_CLI_BUILD_PROJECT provided", "", "buildProject", "", "buildProject"},
		{"JFROG_CLI_PROJECT provided", "", "", "cliProject", "cliProject"},
		{"All provided", "flagProject", "buildProject", "cliProject", "flagProject"},
		{"None provided", "", "", "", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set environment variables
			err := os.Setenv("JFROG_CLI_BUILD_PROJECT", tc.envBuild)
			assert.NoError(t, err)
			err = os.Setenv("JFROG_CLI_PROJECT", tc.envCli)
			assert.NoError(t, err)

			// Create context with flag
			context, _ := tests.CreateContext(t, []string{"project=" + tc.flag}, []string{})

			// Get project key
			projectKey := cliutils.GetProject(context)

			// Assert
			assert.Equal(t, tc.expected, projectKey)

			// Unset environment variables
			err = os.Unsetenv("JFROG_CLI_BUILD_PROJECT")
			assert.NoError(t, err)
			err = os.Unsetenv("JFROG_CLI_PROJECT")
			assert.NoError(t, err)
		})
	}
}
