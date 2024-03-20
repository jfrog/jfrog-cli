package artifactory

import (
	"bytes"
	"path/filepath"
	"testing"

	commonCliUtils "github.com/jfrog/jfrog-cli-core/v2/common/cliutils"
	"github.com/jfrog/jfrog-cli-core/v2/common/spec"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestPrepareSearchDownloadDeleteCommands(t *testing.T) {
	testRuns := []struct {
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
		{"withSpec", []string{}, []string{"spec=" + getSpecPath(tests.SearchAllRepo1)}, "${REPO1}/*", "", "", false},
		{"withSpecAndPattern", []string{"TestPattern"}, []string{"spec=" + getSpecPath(tests.SearchAllRepo1)}, "", "", "", true},
		{"withBuildAndPattern", []string{"TestPattern"}, []string{"build=buildName/buildNumber"}, "TestPattern", "buildName/buildNumber", "", false},
	}

	for _, test := range testRuns {
		t.Run(test.name, func(t *testing.T) {
			context, buffer := tests.CreateContext(t, test.flags, test.args)
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
	testRuns := []struct {
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
		{"withSpec", []string{}, []string{"spec=" + getSpecPath(tests.CopyItemsSpec)}, "${REPO1}/*/", "${REPO2}/", "", "", false},
		{"withSpecAndPattern", []string{"TestPattern"}, []string{"spec=" + getSpecPath(tests.CopyItemsSpec)}, "", "", "", "", true},
		{"withPatternTargetAndBuild", []string{"TestPattern", "TestTarget"}, []string{"build=buildName/buildNumber"}, "TestPattern", "", "buildName/buildNumber", "", false},
		{"withPatternTargetAndBundle", []string{"TestPattern", "TestTarget"}, []string{"bundle=bundleName/bundleVersion"}, "TestPattern", "", "", "bundleName/bundleVersion", false},
	}

	for _, test := range testRuns {
		t.Run(test.name, func(t *testing.T) {
			context, buffer := tests.CreateContext(t, test.flags, test.args)
			specFiles, err := prepareCopyMoveCommand(context)
			assertGenericCommand(t, err, buffer, test.expectError, test.expectedPattern, test.expectedBuild, test.expectedBundle, specFiles)
		})
	}
}

func TestPreparePropsCmd(t *testing.T) {
	testRuns := []struct {
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
		{"withSpec", []string{"key1=val1"}, []string{"spec=" + getSpecPath(tests.SetDeletePropsSpec)}, "key1=val1", "${REPO1}/", "", "", false},
		{"withSpecAndPattern", []string{"TestPattern", "key1=val1"}, []string{"spec=" + getSpecPath(tests.SetDeletePropsSpec)}, "key1=val1", "", "", "", true},
		{"withPatternAndBuild", []string{"TestPattern", "key1=val1"}, []string{"build=buildName/buildNumber"}, "key1=val1", "TestPattern", "buildName/buildNumber", "", false},
	}

	for _, test := range testRuns {
		t.Run(test.name, func(t *testing.T) {
			context, buffer := tests.CreateContext(t, test.flags, test.args)
			propsCommand, err := preparePropsCmd(context)
			var actualSpec *spec.SpecFiles
			if propsCommand != nil {
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

func getSpecPath(spec string) string {
	return filepath.Join("..", "testdata", "filespecs", spec)
}

var createUploadConfigurationCases = []struct {
	name               string
	flags              []string
	expectedMinSplit   int64
	expectedSplitCount int
	expectedThreads    int
	expectedDeb        string
}{
	{"empty", []string{}, cliutils.UploadMinSplitMb, cliutils.UploadSplitCount, commonCliUtils.Threads, ""},
	{"min-split", []string{"min-split=101"}, 101, cliutils.UploadSplitCount, commonCliUtils.Threads, ""},
	{"split-count", []string{"split-count=6"}, cliutils.UploadMinSplitMb, 6, commonCliUtils.Threads, ""},
	{"threads", []string{"threads=6"}, cliutils.UploadMinSplitMb, cliutils.UploadSplitCount, 6, ""},
	{"deb", []string{"deb=jammy/main/i386"}, cliutils.UploadMinSplitMb, cliutils.UploadSplitCount, commonCliUtils.Threads, "jammy/main/i386"},
}

func TestCreateUploadConfiguration(t *testing.T) {
	for _, testCase := range createUploadConfigurationCases {
		t.Run(testCase.name, func(t *testing.T) {
			context, _ := tests.CreateContext(t, testCase.flags, []string{})
			uploadConfiguration, err := cliutils.CreateUploadConfiguration(context)
			assert.NoError(t, err)
			assert.Equal(t, testCase.expectedMinSplit, uploadConfiguration.MinSplitSizeMB)
			assert.Equal(t, testCase.expectedSplitCount, uploadConfiguration.SplitCount)
			assert.Equal(t, testCase.expectedThreads, uploadConfiguration.Threads)
			assert.Equal(t, testCase.expectedDeb, uploadConfiguration.Deb)
		})
	}
}
