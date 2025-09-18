package services

import (
	"errors"
	"testing"
	// Import your artifactory stats package
	"github.com/stretchr/testify/assert"
)

type CommandRunner interface {
	Run() error
}

type mockArtifactoryCommand struct {
	runError error
}

func (m *mockArtifactoryCommand) Run() error {
	return m.runError
}

func TestStatsRun(t *testing.T) {
	testCases := []struct {
		name           string
		product        string
		mockRunError   error
		expectError    bool
		expectedErrMsg string
	}{
		{
			name:    "Artifactory Success",
			product: "rt",
		},
		{
			name:           "Artifactory Command Fails",
			product:        "artifactory",
			mockRunError:   errors.New("artifactory command failed"),
			expectError:    true,
			expectedErrMsg: "artifactory command failed",
		},
		{
			name:    "Unknown Product",
			product: "xr",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockCmd := &mockArtifactoryCommand{runError: tc.mockRunError}

			var err error
			switch tc.product {
			case "rt", "artifactory", "artifactories":
				err = mockCmd.Run()
			default:
				err = nil
			}
			if tc.expectError {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedErrMsg, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
