package services

import (
	"errors"
	"github.com/jfrog/jfrog-cli-core/v2/general"
	"github.com/stretchr/testify/assert"
	"testing"
)

type CommandRunner interface {
	Run() error
}

type mockArtifactoryCommand struct {
	runError     error
	ServerId     string
	AccessToken  string
	FormatOutput string
	DisplayLimit int
}

func (m *mockArtifactoryCommand) Run() error {
	return m.runError
}

var newArtifactoryStatsCommandFunc = func(ss *general.Stats) CommandRunner {
	return ss.NewArtifactoryStatsCommand()
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
			name:    "Artifactory Alias Success",
			product: "artifactory",
		},
		{
			name:           "Artifactory Command Fails",
			product:        "rt",
			mockRunError:   errors.New("artifactory command failed"),
			expectError:    true,
			expectedErrMsg: "artifactory command failed",
		},
		{
			name:        "Unknown Product",
			product:     "xr",
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			statsCmd := general.NewStatsCommand().SetProduct(tc.product)

			mockCmd := &mockArtifactoryCommand{runError: tc.mockRunError}

			originalFactory := newArtifactoryStatsCommandFunc
			newArtifactoryStatsCommandFunc = func(ss *general.Stats) CommandRunner {
				return mockCmd
			}
			defer func() { newArtifactoryStatsCommandFunc = originalFactory }() // Restore original

			err := statsCmd.Run()

			if tc.expectError {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedErrMsg, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
