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
	corecommands "github.com/jfrog/jfrog-cli-core/v2/common/commands"
	configtests "github.com/jfrog/jfrog-cli-core/v2/utils/config/tests"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
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
	// Explicitly unset the env var in case it was pre-set in the environment (e.g. in CI),
	// so it does not interfere with the remaining assertions in this test.
	assert.NoError(t, os.Unsetenv(JfrogCliAvoidNewVersionWarning))

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
		_, err := w.Write([]byte(`{"url":"https://api.github.com/repos/jfrog/jfrog-cli/releases/1","tag_name":"v1.0.0"}`))
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
		// #nosec G704 -- redirectURL is a controlled test value, not user input
		redirectReq, err := http.NewRequest(req.Method, t.redirectURL, req.Body) //nolint:gosec // G704 - URL is a test-controlled constant
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

// agentDetectorEnvVars lists every env var jfrog-cli-core's agent detector consults
// (see ExecutionContext in jfrog-cli-core/common/commands). Tests clear these so
// ShouldHideSurveyLink's agent check is deterministic regardless of the shell
// running `go test` (e.g. running inside Claude Code, Cursor, etc.).
var agentDetectorEnvVars = []string{
	"CLAUDECODE", "CLAUDE_CODE", "CLAUDE_CODE_ENTRYPOINT",
	"GEMINI_CLI",
	"GOOSE_TERMINAL",
	"CURSOR_AGENT", "CURSOR_TRACE_ID", "CURSOR_EXTENSION_HOST_ROLE",
	"COPILOT_CLI", "COPILOT_AGENT_SESSION_ID", "COPILOT_MODEL", "COPILOT_ALLOW_ALL",
	"KILOCODE_FEATURE", "KILO_PID",
	"ROO_ACTIVE", "ROO_CLI_RUNTIME",
	"CODEX_CI", "CODEX_THREAD_ID", "CODEX_SANDBOX",
	"WINDSURF_CASCADE_TERMINAL",
	"CLINE_ACTIVE", "OPENCODE", "OPENCODE_CLIENT",
	"AMP_CURRENT_THREAD_ID", "AUGMENT_AGENT", "QWEN_CODE",
	"ANTIGRAVITY_AGENT", "CRUSH", "IFLOW_CLI", "TRAE_AI_SHELL_ID",
	"AI_AGENT", "AGENT",
	// Host editor and model axes — cleared so the wire format is deterministic
	// regardless of the shell running `go test`.
	"TERM_PROGRAM", "JFROG_CLI_AI_MODEL",
}

func clearAgentEnvVarsForTest(t *testing.T) {
	t.Helper()
	for _, e := range agentDetectorEnvVars {
		t.Setenv(e, "")
	}
	corecommands.ResetExecutionContextForTest()
	t.Cleanup(corecommands.ResetExecutionContextForTest)
}

// TestGetHasDisplayedSurveyLink tests the survey link environment variable check with parametrized test cases
func TestGetHasDisplayedSurveyLink(t *testing.T) {
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
	t.Setenv(coreutils.CI, "")
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			clearAgentEnvVarsForTest(t)
			t.Setenv(JfrogCliHideSurvey, tc.envValue)

			shouldHide := ShouldHideSurveyLink()

			if tc.shouldHide {
				assert.True(t, shouldHide, "Expected survey to be hidden for test case: %s", tc.name)
			} else {
				assert.False(t, shouldHide, "Expected survey to not be hidden for test case: %s", tc.name)
			}
		})
	}
}

func TestSettingCIFlagRemovesSurvey(t *testing.T) {
	t.Setenv(coreutils.CI, "true")
	shouldHide := ShouldHideSurveyLink()
	assert.True(t, shouldHide, "Expected survey to be hidden when CI flag is set")
}

func TestSurveyHiddenForAgent(t *testing.T) {
	t.Setenv(coreutils.CI, "")
	t.Setenv(JfrogCliHideSurvey, "")
	clearAgentEnvVarsForTest(t)
	t.Setenv("CLAUDECODE", "true")
	corecommands.ResetExecutionContextForTest()

	assert.True(t, ShouldHideSurveyLink(), "Expected survey to be hidden when invoked by an agent")
}

func TestLoginCommandFlagsIncludeServerId(t *testing.T) {
	flags := GetCommandFlags(Login)
	assert.NotEmpty(t, flags, "Expected login command to have flags")

	var flagNames []string
	for _, f := range flags {
		flagNames = append(flagNames, f.GetName())
	}
	assert.Contains(t, flagNames, "server-id", "Expected login command flags to include 'server-id'")
}

func TestTransferFilesTimestampFilterFlags(t *testing.T) {
	flags := GetCommandFlags(TransferFiles)
	assert.NotEmpty(t, flags)

	var flagNames []string
	usageByName := map[string]string{}
	for _, f := range flags {
		flagNames = append(flagNames, f.GetName())
		usageByName[f.GetName()] = f.String()
	}

	assert.Contains(t, flagNames, CreatedAfter)
	assert.Contains(t, flagNames, DownloadedAfter)
	assert.Contains(t, usageByName[CreatedAfter], "YYYY-MM-DDTHH:mm:ss.sssZ")
	assert.Contains(t, usageByName[DownloadedAfter], "YYYY-MM-DDTHH:mm:ss.sssZ")
}

// --- AGW-86: User-Agent enrichment with the detected AI agent ---

// withCliUserAgent pins the CLI user-agent name/version for one test. The real values are
// set once in init() from JFROG_CLI_USER_AGENT, so tests drive the setters directly rather
// than trying to re-run init.
func withCliUserAgent(t *testing.T, name, version string) {
	t.Helper()
	prevName, prevVersion := coreutils.GetCliUserAgentName(), coreutils.GetCliUserAgentVersion()
	coreutils.SetCliUserAgentName(name)
	coreutils.SetCliUserAgentVersion(version)
	t.Cleanup(func() {
		coreutils.SetCliUserAgentName(prevName)
		coreutils.SetCliUserAgentVersion(prevVersion)
	})
}

func TestGetCliUserAgentWithAgentNoAgentDetected(t *testing.T) {
	clearAgentEnvVarsForTest(t)
	withCliUserAgent(t, "jfrog-cli-go", "2.117.0")
	corecommands.ResetExecutionContextForTest()

	assert.Equal(t, "jfrog-cli-go/2.117.0", GetCliUserAgentWithAgent(),
		"a human invocation must stay byte-identical to today's behaviour")
}

func TestGetCliUserAgentWithAgentPerDetector(t *testing.T) {
	// One case per row of jfrog-cli-core's agentEnvDetectors table, plus the generic
	// AGENT fallback that is deliberately collapsed to "unknown".
	testCases := []struct {
		name      string
		envVar    string
		envValue  string // empty → "1"
		wantAgent string
	}{
		{"claude code", "CLAUDECODE", "", "claude"},
		{"claude code entrypoint", "CLAUDE_CODE_ENTRYPOINT", "", "claude"},
		{"gemini", "GEMINI_CLI", "", "gemini"},
		{"goose", "GOOSE_TERMINAL", "", "goose"},
		{"cursor agent", "CURSOR_AGENT", "", "cursor"},
		{"cursor extension host", "CURSOR_EXTENSION_HOST_ROLE", "agent-exec", "cursor"},
		{"copilot", "COPILOT_CLI", "", "copilot"},
		{"kilocode", "KILO_PID", "", "kilocode"},
		{"roo code", "ROO_ACTIVE", "", "roo_code"},
		{"codex", "CODEX_CI", "", "codex"},
		{"generic agent collapses to unknown", "AGENT", "", "unknown"},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			clearAgentEnvVarsForTest(t)
			withCliUserAgent(t, "jfrog-cli-go", "2.117.0")
			val := testCase.envValue
			if val == "" {
				val = "1"
			}
			t.Setenv(testCase.envVar, val)
			corecommands.ResetExecutionContextForTest()

			assert.Equal(t, "jfrog-cli-go/2.117.0 ai-agent/"+testCase.wantAgent, GetCliUserAgentWithAgent())
		})
	}
}

func TestGetCliUserAgentWithAgentPreservesCustomUserAgent(t *testing.T) {
	// JFROG_CLI_USER_AGENT lets an operator replace the product token entirely. The agent
	// marker must be appended to whatever that resolves to, never replace it.
	clearAgentEnvVarsForTest(t)
	withCliUserAgent(t, "my-wrapper", "9.9.9")
	t.Setenv("CLAUDECODE", "true")
	corecommands.ResetExecutionContextForTest()

	assert.Equal(t, "my-wrapper/9.9.9 ai-agent/claude", GetCliUserAgentWithAgent())
}

func TestGetCliUserAgentWithAgentNoVersion(t *testing.T) {
	// GetCliUserAgent omits the slash when no version is set; the marker still appends.
	clearAgentEnvVarsForTest(t)
	withCliUserAgent(t, "jfrog-cli-go", "")
	t.Setenv("CLAUDECODE", "true")
	corecommands.ResetExecutionContextForTest()

	assert.Equal(t, "jfrog-cli-go ai-agent/claude", GetCliUserAgentWithAgent())
}

func TestGetCliUserAgentWithAgentAppendsHostAndModel(t *testing.T) {
	clearAgentEnvVarsForTest(t)
	withCliUserAgent(t, "jfrog-cli-go", "2.117.0")
	t.Setenv("CURSOR_AGENT", "1")
	t.Setenv("TERM_PROGRAM", "vscode")
	t.Setenv("JFROG_CLI_AI_MODEL", "opus-4.7")
	corecommands.ResetExecutionContextForTest()

	assert.Equal(t, "jfrog-cli-go/2.117.0 ai-agent/cursor ai-client/vscode ai-model/opus-4.7",
		GetCliUserAgentWithAgent())
}

func TestGetCliUserAgentWithAgentOmitsAbsentAxes(t *testing.T) {
	// Host and model are optional: with neither advertised, the suffix is just
	// the agent token — byte-identical to the pre-host/model behaviour.
	clearAgentEnvVarsForTest(t)
	withCliUserAgent(t, "jfrog-cli-go", "2.117.0")
	t.Setenv("CLAUDECODE", "1")
	corecommands.ResetExecutionContextForTest()

	assert.Equal(t, "jfrog-cli-go/2.117.0 ai-agent/claude", GetCliUserAgentWithAgent())
}

func TestGetCliUserAgentWithAgentMarkerIsWellFormed(t *testing.T) {
	clearAgentEnvVarsForTest(t)
	withCliUserAgent(t, "jfrog-cli-go", "2.117.0")
	t.Setenv("CURSOR_AGENT", "1")
	corecommands.ResetExecutionContextForTest()

	userAgent := GetCliUserAgentWithAgent()
	// The product token stays first, so parsers that read only it are unaffected.
	assert.True(t, strings.HasPrefix(userAgent, "jfrog-cli-go/2.117.0"), "got %q", userAgent)
	assert.True(t, strings.HasSuffix(userAgent, "ai-agent/cursor"), "got %q", userAgent)
	// The detector only ever returns fixed table names, so no raw env value — and
	// therefore no header-splitting sequence — can reach the wire.
	assert.NotContains(t, userAgent, "\n")
	assert.NotContains(t, userAgent, "\r")
}
