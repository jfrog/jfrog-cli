package mcp

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/stretchr/testify/assert"
)

func TestGetMCPServerArgs(t *testing.T) {
	testRuns := []struct {
		name              string
		flags             []string
		envVars           map[string]string
		expectedToolSets  string
		expectedToolAccess string
		expectedVersion   string
	}{
		{
			name:              "FlagsOnly",
			flags:             []string{"toolsets=test", "tools-access=read", "mcp-server-version=1.0.0"},
			envVars:           map[string]string{},
			expectedToolSets:  "test",
			expectedToolAccess: "read",
			expectedVersion:   "1.0.0",
		},
		{
			name:              "EnvVarsOnly",
			flags:             []string{},
			envVars:           map[string]string{mcpToolSetsEnvVar: "test-env", mcpToolAccessEnvVar: "read-env"},
			expectedToolSets:  "test-env",
			expectedToolAccess: "read-env",
			expectedVersion:   defaultServerVersion,
		},
		{
			name:              "FlagsOverrideEnvVars",
			flags:             []string{"toolsets=test-flag", "tools-access=read-flag"},
			envVars:           map[string]string{mcpToolSetsEnvVar: "test-env", mcpToolAccessEnvVar: "read-env"},
			expectedToolSets:  "test-flag",
			expectedToolAccess: "read-flag",
			expectedVersion:   defaultServerVersion,
		},
		{
			name:              "NoFlagsOrEnvVars",
			flags:             []string{},
			envVars:           map[string]string{},
			expectedToolSets:  "",
			expectedToolAccess: "",
			expectedVersion:   defaultServerVersion,
		},
	}

	for _, test := range testRuns {
		t.Run(test.name, func(t *testing.T) {
			// Save current environment and restore it after the test
			originalEnv := make(map[string]string)
			for key := range test.envVars {
				originalEnv[key] = os.Getenv(key)
			}
			defer func() {
				for key, value := range originalEnv {
					os.Setenv(key, value)
				}
			}()

			// Set environment variables for the test
			for key, value := range test.envVars {
				os.Setenv(key, value)
			}

			// Create CLI context
			context, _ := tests.CreateContext(t, test.flags, []string{})
			
			// Test getMCPServerArgs
			cmd := NewMcpCommand()
			cmd.getMCPServerArgs(context)
			
			// Assert results
			assert.Equal(t, test.expectedToolSets, cmd.toolSets)
			assert.Equal(t, test.expectedToolAccess, cmd.toolAccess)
			assert.Equal(t, test.expectedVersion, cmd.serverVersion)
		})
	}
}

func TestGetOsArchBinaryInfo(t *testing.T) {
	osName, arch, binaryName := getOsArchBinaryInfo()
	
	// Verify OS and architecture are not empty
	assert.NotEmpty(t, osName)
	assert.NotEmpty(t, arch)
	
	// Verify binary name has correct format
	if osName == "windows" {
		assert.Equal(t, mcpServerBinaryName+".exe", binaryName)
	} else {
		assert.Equal(t, mcpServerBinaryName, binaryName)
	}
}

func TestExtractVersionFromHeader(t *testing.T) {
	testCases := []struct {
		name        string
		headerValue string
		expected    string
	}{
		{
			name:        "ValidHeader",
			headerValue: `attachment; filename="cli-mcp-server-0.1.0"`,
			expected:    "0.1.0",
		},
		{
			name:        "NoVersion",
			headerValue: `attachment; filename="cli-mcp-server"`,
			expected:    "",
		},
		{
			name:        "NoFilename",
			headerValue: `attachment;`,
			expected:    "",
		},
		{
			name:        "EmptyHeader",
			headerValue: "",
			expected:    "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := extractVersionFromHeader(tc.headerValue)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestCmd(t *testing.T) {
	testCases := []struct {
		name        string
		args        []string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "NoArgs",
			args:        []string{},
			expectError: true,
			errorMsg:    "Unknown subcommand: ",
		},
		{
			name:        "InvalidSubcommand",
			args:        []string{"invalid"},
			expectError: true,
			errorMsg:    "Unknown subcommand: invalid",
		},
		// Update and start subcommands require more complex mocking to test properly
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			context, _ := tests.CreateContext(t, []string{}, tc.args)
			err := Cmd(context)
			
			if tc.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetLocalBinaryPath(t *testing.T) {
	// Skip test since we can't mock coreutils.GetJfrogHomeDir directly
	t.Skip("Skipping test as it requires mocking package-level functions")
}

func TestGetMcpServerVersion(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "mcp-version-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a fake binary that outputs a version when called with --version
	binaryPath := filepath.Join(tempDir, "fake-binary")
	if runtime.GOOS == "windows" {
		binaryPath += ".exe"
	}

	// Create a simple shell script that echoes a version
	var scriptContent string
	if runtime.GOOS == "windows" {
		scriptContent = "@echo off\r\necho 1.2.3\r\n"
	} else {
		scriptContent = "#!/bin/sh\necho \"1.2.3\"\n"
	}

	err = os.WriteFile(binaryPath, []byte(scriptContent), 0755)
	assert.NoError(t, err)

	// Test the getMcpServerVersion function
	version, err := getMcpServerVersion(binaryPath)
	assert.NoError(t, err)
	assert.Equal(t, "1.2.3", version)

	// Test with non-existent binary
	_, err = getMcpServerVersion(filepath.Join(tempDir, "non-existent"))
	assert.Error(t, err)
}
