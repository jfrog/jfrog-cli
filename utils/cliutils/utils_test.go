package cliutils

import (
	"errors"
	"fmt"
	biutils "github.com/jfrog/build-info-go/utils"
	configtests "github.com/jfrog/jfrog-cli-core/v2/utils/config/tests"
	clientTestUtils "github.com/jfrog/jfrog-client-go/utils/tests"
	"path/filepath"
	"strings"
	"testing"
	"time"

	coretests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	"github.com/jfrog/jfrog-cli/utils/tests"

	commandUtils "github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/utils"
	"github.com/jfrog/jfrog-client-go/utils/io/content"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/stretchr/testify/assert"
)

func TestSplitAgentNameAndVersion(t *testing.T) {
	tests := []struct {
		fullAgentName        string
		expectedAgentName    string
		expectedAgentVersion string
	}{
		{"abc/1.2.3", "abc", "1.2.3"},
		{"abc/def/1.2.3", "abc/def", "1.2.3"},
		{"abc\\1.2.3", "abc\\1.2.3", ""},
		{"abc:1.2.3", "abc:1.2.3", ""},
		{"", "", ""},
	}

	for _, test := range tests {
		actualAgentName, actualAgentVersion := splitAgentNameAndVersion(test.fullAgentName)
		assert.Equal(t, test.expectedAgentName, actualAgentName)
		assert.Equal(t, test.expectedAgentVersion, actualAgentVersion)
	}
}

func TestPrintCommandSummary(t *testing.T) {
	outputBuffer, stderrBuffer, previousLog := coretests.RedirectLogOutputToBuffer()
	// Restore previous logger when the function returns
	defer log.SetLogger(previousLog)

	result := &commandUtils.Result{}
	result.SetSuccessCount(1)
	result.SetFailCount(0)
	testdata := filepath.Join(tests.GetTestResourcesPath(), "reader", "printcommandsummary.json")
	tmpDir, createTempDirCallback := coretests.CreateTempDirWithCallbackAndAssert(t)
	defer createTempDirCallback()
	err := biutils.CopyFile(tmpDir, testdata)
	assert.NoError(t, err)

	reader := content.NewContentReader(filepath.Join(tmpDir, "printcommandsummary.json"), content.DefaultKey)
	result.SetReader(reader)
	assert.NoError(t, err)
	tests := []struct {
		isDetailedSummary bool
		isDeploymentView  bool
		expectedString    string
		expectedError     error
	}{
		{true, false, `"status": "success",`, nil},
		{true, false, `"status": "failure",`, errors.New("test")},
		{false, true, "These files were uploaded:", nil},
		{false, true, ``, errors.New("test")},
	}
	for _, test := range tests {
		err = PrintCommandSummary(result, test.isDetailedSummary, test.isDeploymentView, false, test.expectedError)
		if test.expectedError != nil {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
		var output []byte
		if test.isDetailedSummary {
			output = outputBuffer.Bytes()
			outputBuffer.Truncate(0)
		} else {
			output = stderrBuffer.Bytes()
			stderrBuffer.Truncate(0)
		}
		assert.True(t, strings.Contains(string(output), test.expectedString), fmt.Sprintf("cant find '%s' in '%s'", test.expectedString, string(output)))
	}
}

func TestCheckNewCliVersionAvailable(t *testing.T) {
	// Run the following tests on Artifactory tests suite only, to avoid reaching the GitHub API allowed rate limit (60 requests per hour)
	// More info on https://docs.github.com/en/rest/overview/resources-in-the-rest-api?#rate-limiting
	if !*tests.TestArtifactory {
		return
	}

	testCheckNewCliVersionAvailable(t, "0.0.0", true)
	testCheckNewCliVersionAvailable(t, "100.100.100", false)
}

func testCheckNewCliVersionAvailable(t *testing.T, version string, shouldWarn bool) {
	// Create temp JFROG_HOME
	cleanUpTempEnv := configtests.CreateTempEnv(t, false)
	defer cleanUpTempEnv()

	// First run, should warn if needed
	warningMessage, err := CheckNewCliVersionAvailable(version)
	assert.NoError(t, err)
	assert.Equal(t, warningMessage != "", shouldWarn)

	// Second run, shouldn't warn
	warningMessage, err = CheckNewCliVersionAvailable(version)
	assert.NoError(t, err)
	assert.Empty(t, warningMessage)
}

func TestShouldCheckLatestCliVersion(t *testing.T) {
	persistenceFilePath = filepath.Join(t.TempDir(), persistenceFileName)

	// Validate that avoiding the version check using an environment variable is working
	setEnvCallback := clientTestUtils.SetEnvWithCallbackAndAssert(t, JfrogCliAvoidNewVersionWarning, "true")
	shouldCheck, err := shouldCheckLatestCliVersion()
	assert.NoError(t, err)
	assert.False(t, shouldCheck)
	setEnvCallback()

	// First run, should be true
	shouldCheck, err = shouldCheckLatestCliVersion()
	assert.NoError(t, err)
	assert.True(t, shouldCheck)

	// Second run, less than 6 hours between runs, so should return false
	shouldCheck, err = shouldCheckLatestCliVersion()
	assert.NoError(t, err)
	assert.False(t, shouldCheck)

	assert.NoError(t, setCliLatestVersionCheckTime(time.Now().UnixMilli()-LatestCliVersionCheckInterval.Milliseconds()))
	// Third run, more than 6 hours between runs, so should return true
	shouldCheck, err = shouldCheckLatestCliVersion()
	assert.NoError(t, err)
	assert.True(t, shouldCheck)
}

func TestExtractBoolFlagFromArgs(t *testing.T) {
	testCases := []struct {
		name          string
		args          []string
		flagName      string
		expectedValue bool
		expectedErr   bool
		expectedArgs  []string
	}{
		{
			name:          "Flag present as --flagName (implied true)",
			args:          []string{"somecmd", "--flagName", "otherarg"},
			flagName:      "flagName",
			expectedValue: true,
			expectedErr:   false,
			expectedArgs:  []string{"somecmd", "otherarg"},
		},
		{
			name:          "Flag present as --flagName=true",
			args:          []string{"somecmd", "--flagName=true", "otherarg"},
			flagName:      "flagName",
			expectedValue: true,
			expectedErr:   false,
			expectedArgs:  []string{"somecmd", "otherarg"},
		},
		{
			name:          "Flag present as --flagName=false",
			args:          []string{"somecmd", "--flagName=false", "otherarg"},
			flagName:      "flagName",
			expectedValue: false,
			expectedErr:   false,
			expectedArgs:  []string{"somecmd", "otherarg"},
		},
		{
			name:          "Flag not present",
			args:          []string{"somecmd", "otherarg"},
			flagName:      "flagName",
			expectedValue: false,
			expectedErr:   false,
			expectedArgs:  []string{"somecmd", "otherarg"},
		},
		{
			name:          "Flag present with invalid value",
			args:          []string{"somecmd", "--flagName=invalid", "otherarg"},
			flagName:      "flagName",
			expectedValue: false,
			expectedErr:   true,
			expectedArgs:  []string{"somecmd", "--flagName=invalid", "otherarg"},
		},
		{
			name:          "Flag present as -flagName (should not be found)",
			args:          []string{"somecmd", "-flagName", "otherarg"},
			flagName:      "flagName",
			expectedValue: false,
			expectedErr:   false,
			expectedArgs:  []string{"somecmd", "-flagName", "otherarg"},
		},
		{
			name:          "Flag present multiple times",
			args:          []string{"somecmd", "--flagName", "--flagName=false", "otherarg"},
			flagName:      "flagName",
			expectedValue: false,
			expectedErr:   false,
			expectedArgs:  []string{"somecmd", "--flagName=false", "otherarg"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Make a copy to avoid modifying the original
			argsCopy := append([]string(nil), tc.args...)
			value, err := ExtractBoolFlagFromArgs(&argsCopy, tc.flagName)

			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tc.expectedValue, value)
			assert.Equal(t, tc.expectedArgs, argsCopy)
		})
	}
}
