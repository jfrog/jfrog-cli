package cliutils

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	biutils "github.com/jfrog/build-info-go/utils"
	configtests "github.com/jfrog/jfrog-cli-core/v2/utils/config/tests"
	clientTestUtils "github.com/jfrog/jfrog-client-go/utils/tests"
	"github.com/urfave/cli"

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
			expectedValue: true,
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

func TestGetFlagOrEnvValue(t *testing.T) {
	// Define test cases
	var envVarName = "test-env-var"
	testCases := []struct {
		name      string
		flagValue string
		envValue  string
		expected  string
		flagName  string
	}{
		{
			name:      "Flag value is set",
			flagValue: "flagValue",
			envValue:  "envValue",
			expected:  "flagValue",
			flagName:  "test-flag",
		},
		{
			name:      "Flag value is not set, env value is set",
			flagValue: "",
			envValue:  "envValue",
			expected:  "envValue",
			flagName:  "test-flag",
		},
		{
			name:      "Neither flag value nor env value is set",
			flagValue: "",
			envValue:  "",
			expected:  "",
			flagName:  "test-flag",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set environment variable
			cleanup := clientTestUtils.SetEnvWithCallbackAndAssert(t, envVarName, tc.envValue)
			defer cleanup()

			// Create a new CLI context with the flag
			set := flag.NewFlagSet("test", 0)
			set.String(tc.flagName, tc.flagValue, "")
			c := cli.NewContext(nil, set, nil)

			// Get the value using the function
			value := GetFlagOrEnvValue(c, tc.flagName, envVarName)

			// Assert the expected value
			assert.Equal(t, tc.expected, value)
		})
	}
}

// TestAuthorizationHeaderInCliVersionCheck tests that the HTTP request for checking new CLI versions
// includes an authorization header when a GitHub token is provided.
func TestAuthorizationHeaderInCliVersionCheck(t *testing.T) {
	// Create a test server that will capture the request headers
	var capturedAuthHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Capture the Authorization header
		capturedAuthHeader = r.Header.Get("Authorization")
		// Return a valid JSON response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"tag_name": "v1.0.0"}`))
		if err != nil {
			return
		}
	}))
	defer server.Close()

	// Create a custom transport that redirects GitHub API requests to our test server
	originalTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = originalTransport }()

	http.DefaultTransport = &redirectingTransport{
		targetURL:     "https://api.github.com/repos/jfrog/jfrog-cli/releases/latest",
		redirectURL:   server.URL,
		baseTransport: originalTransport,
	}

	// Test cases
	testCases := []struct {
		name             string
		githubToken      string
		expectAuthHeader bool
	}{
		{
			name:             "With GitHub token",
			githubToken:      "test-token",
			expectAuthHeader: true,
		},
		{
			name:             "Without GitHub token",
			githubToken:      "",
			expectAuthHeader: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset captured auth header for each test case
			capturedAuthHeader = ""
			err1 := os.Setenv(JfrogCliGithubToken, tc.githubToken)
			if err1 != nil {
				return
			}
			// Call getLatestCliVersionFromGithubAPI directly
			_, err := getLatestCliVersionFromGithubAPI()
			assert.NoError(t, err)

			// Check if the Authorization header was captured correctly by the server
			if tc.expectAuthHeader {
				assert.Equal(t, "Bearer "+tc.githubToken, capturedAuthHeader)
			} else {
				assert.Empty(t, capturedAuthHeader)
			}
		})
	}
}

// redirectingTransport is a custom http.RoundTripper that redirects requests
// from a specific URL to another URL.
type redirectingTransport struct {
	targetURL     string
	redirectURL   string
	baseTransport http.RoundTripper
}

func (t *redirectingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.String() == t.targetURL {
		// Create a new request to the redirect URL
		redirectReq, err := http.NewRequest(req.Method, t.redirectURL, req.Body)
		if err != nil {
			return nil, err
		}

		// Copy all headers from the original request
		redirectReq.Header = req.Header

		// Send the redirected request
		return t.baseTransport.RoundTrip(redirectReq)
	}

	// For all other requests, use the base transport
	return t.baseTransport.RoundTrip(req)
}

// TestGetHasDisplayedSurveyLink tests the survey link environment variable check with parametrized test cases
func TestGetHasDisplayedSurveyLink(t *testing.T) {
	// Save original environment variable value
	originalValue := os.Getenv(JfrogCliHideSurvey)
	defer func() {
		// Restore original value
		if originalValue == "" {
			os.Unsetenv(JfrogCliHideSurvey)
		} else {
			os.Setenv(JfrogCliHideSurvey, originalValue)
		}
	}()

	testCases := []struct {
		name       string
		envValue   string
		shouldHide bool
	}{
		{
			name:       "env_var_not_set",
			envValue:   "", // This will be handled by unsetting the env var
			shouldHide: false,
		},
		{
			name:       "env_var_true",
			envValue:   "true",
			shouldHide: true,
		},
		{
			name:       "env_var_bad_input",
			envValue:   "garbage",
			shouldHide: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up environment variable
			if tc.envValue == "" {
				os.Unsetenv(JfrogCliHideSurvey)
			} else {
				os.Setenv(JfrogCliHideSurvey, tc.envValue)
			}

			// Test the function
			shouldHide := ShouldHideSurveyLink()

			// Assert the result
			if tc.shouldHide {
				assert.True(t, shouldHide, "Expected survey to be hidden for test case: %s", tc.name)
			} else {
				assert.False(t, shouldHide, "Expected survey to not be hidden for test case: %s", tc.name)
			}
		})
	}
}
