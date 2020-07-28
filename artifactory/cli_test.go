package artifactory

import (
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"testing"
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
