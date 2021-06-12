package artifactory

import (
	"bytes"
	"flag"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codegangsta/cli"
	"github.com/jfrog/jfrog-cli-core/artifactory/spec"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/stretchr/testify/assert"
)

func TestValidateGoNativeCommand(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected bool
	}{
		{"withInvalidCommand", []string{"build", "-test", "-another", "--publish-deps"}, true},
		{"withSimilarCommand", []string{"build", "-test", "-another", "publish-deps"}, false},
		{"withoutAnyFlags", []string{"build", "-test", "-another"}, false},
		{"withFlagAndValue", []string{"build", "-test", "-another", "--url=http://another.com"}, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := validateCommand(test.args, cliutils.GetLegacyGoFlags())
			if result != nil && !test.expected {
				t.Errorf("Expected error nil, got the following error %s", result)
			}

			if result == nil && test.expected {
				t.Errorf("Expected error, howerver, got nil")
			}
		})
	}
}

func TestPrepareSearchDownloadDeleteCommands(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		flags           []string
		expectedPattern string
		expectedBuild   string
		expectedBundle  string
		expectError     bool
	}{
		{"withoutArgs", []string{}, []string{}, "TestPattern", "", "", true},
		{"withPattern", []string{"TestPattern"}, []string{}, "TestPattern", "", "", false},
		{"withBuild", []string{}, []string{"build=buildName/buildNumber"}, "", "buildName/buildNumber", "", false},
		{"withBundle", []string{}, []string{"bundle=bundleName/bundleVersion"}, "", "", "bundleName/bundleVersion", false},
		{"withSpec", []string{}, []string{"spec=" + getSpecPath(t, tests.SearchAllRepo1)}, "${REPO1}/*", "", "", false},
		{"withSpecAndPattern", []string{"TestPattern"}, []string{"spec=" + getSpecPath(t, tests.SearchAllRepo1)}, "", "", "", true},
		{"withBuildAndPattern", []string{"TestPattern"}, []string{"build=buildName/buildNumber"}, "TestPattern", "buildName/buildNumber", "", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			context, buffer := createContext(test.flags, test.args)
			funcArray := []func(c *cli.Context) (*spec.SpecFiles, error){
				prepareSearchCommand, prepareDownloadCommand, prepareDeleteCommand,
			}
			for _, prepareCommandFunc := range funcArray {
				specFiles, err := prepareCommandFunc(context)
				assertGenericCommand(t, err, buffer, test.expectError, test.expectedPattern, test.expectedBuild, test.expectedBundle, specFiles)
			}
		})
	}
}

func TestPrepareCopyMoveCommand(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		flags           []string
		expectedPattern string
		expectedTarget  string
		expectedBuild   string
		expectedBundle  string
		expectError     bool
	}{
		{"withoutArguments", []string{}, []string{}, "", "", "", "", true},
		{"withPatternAndTarget", []string{"TestPattern", "TestTarget"}, []string{}, "TestPattern", "TestTarget", "", "", false},
		{"withSpec", []string{}, []string{"spec=" + getSpecPath(t, tests.CopyItemsSpec)}, "${REPO1}/*/", "${REPO2}/", "", "", false},
		{"withSpecAndPattern", []string{"TestPattern"}, []string{"spec=" + getSpecPath(t, tests.CopyItemsSpec)}, "", "", "", "", true},
		{"withPatternTargetAndBuild", []string{"TestPattern", "TestTarget"}, []string{"build=buildName/buildNumber"}, "TestPattern", "", "buildName/buildNumber", "", false},
		{"withPatternTargetAndBundle", []string{"TestPattern", "TestTarget"}, []string{"bundle=bundleName/bundleVersion"}, "TestPattern", "", "", "bundleName/bundleVersion", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			context, buffer := createContext(test.flags, test.args)
			specFiles, err := prepareCopyMoveCommand(context)
			assertGenericCommand(t, err, buffer, test.expectError, test.expectedPattern, test.expectedBuild, test.expectedBundle, specFiles)
		})
	}
}

func TestPreparePropsCmd(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		flags           []string
		expectedProps   string
		expectedPattern string
		expectedBuild   string
		expectedBundle  string
		expectError     bool
	}{
		{"withoutPattern", []string{"key1=val1"}, []string{}, "key1=val1", "", "", "", true},
		{"withPattern", []string{"TestPattern", "key1=val1"}, []string{}, "key1=val1", "TestPattern", "", "", false},
		{"withBuild", []string{"key1=val1"}, []string{"build=buildName/buildNumber"}, "key1=val1", "*", "buildName/buildNumber", "", false},
		{"withBundle", []string{"key1=val1"}, []string{"bundle=bundleName/bundleVersion"}, "key1=val1", "*", "", "bundleName/bundleVersion", false},
		{"withSpec", []string{"key1=val1"}, []string{"spec=" + getSpecPath(t, tests.SetDeletePropsSpec)}, "key1=val1", "${REPO1}/", "", "", false},
		{"withSpecAndPattern", []string{"TestPattern", "key1=val1"}, []string{"spec=" + getSpecPath(t, tests.SetDeletePropsSpec)}, "key1=val1", "", "", "", true},
		{"withPatternAndBuild", []string{"TestPattern", "key1=val1"}, []string{"build=buildName/buildNumber"}, "key1=val1", "TestPattern", "buildName/buildNumber", "", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			context, buffer := createContext(test.flags, test.args)
			propsCommand, err := preparePropsCmd(context)
			var actualSpec *spec.SpecFiles
			if !test.expectError {
				actualSpec = propsCommand.Spec()
				assert.Equal(t, test.expectedProps, propsCommand.Props())
			}
			assertGenericCommand(t, err, buffer, test.expectError, test.expectedPattern, test.expectedBuild, test.expectedBundle, actualSpec)
		})
	}
}

func assertGenericCommand(t *testing.T, err error, buffer *bytes.Buffer, expectError bool, expectedPattern, expectedBuild, expectedBundle string, actualSpec *spec.SpecFiles) {
	if expectError {
		assert.Error(t, err, buffer)
	} else {
		assert.NoError(t, err, buffer)
		assert.Equal(t, expectedPattern, actualSpec.Get(0).Pattern)
		assert.Equal(t, expectedBuild, actualSpec.Get(0).Build)
		assert.Equal(t, expectedBundle, actualSpec.Get(0).Bundle)
	}
}

func createContext(testFlags, testArgs []string) (*cli.Context, *bytes.Buffer) {
	flagSet := flag.NewFlagSet("TestFlagSet", flag.ContinueOnError)
	flags := append(setStringFlags(flagSet, testFlags...), testArgs...)
	flagSet.Parse(flags)

	buffer := &bytes.Buffer{}
	app := cli.NewApp()
	app.Writer = &bytes.Buffer{}
	return cli.NewContext(app, flagSet, nil), buffer
}

func getSpecPath(t *testing.T, spec string) string {
	dir, err := os.Getwd()
	assert.NoError(t, err)
	return filepath.Join(path.Dir(dir), "testdata", "filespecs", spec)
}

// Set string flags. Return a slice of their values.
func setStringFlags(flagSet *flag.FlagSet, flags ...string) []string {
	cmdFlags := []string{}
	for _, flag := range flags {
		flagSet.String(strings.Split(flag, "=")[0], "", "")
		cmdFlags = append(cmdFlags, "--"+flag)
	}
	flagSet.Parse(cmdFlags)

	return cmdFlags
}
