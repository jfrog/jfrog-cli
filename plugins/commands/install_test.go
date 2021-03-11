package commands

import (
	"github.com/jfrog/jfrog-cli/plugins/commands/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetNameAndVersion(t *testing.T) {
	tests := []struct {
		name            string
		providedPlugin  string
		isValid         bool
		expectedName    string
		expectedVersion string
	}{
		{"latest", "hello-frog", true, "hello-frog", utils.LatestVersionName},
		{"latestWithSeparator", "hello-frog@", true, "hello-frog", utils.LatestVersionName},
		{"version", "hello-frog@1.0.0", true, "hello-frog", "1.0.0"},
		{"tooManySeparators", "hello-frog@1.0.0@", false, "", ""},
		{"tooManySeparators2", "hello-frog@@1.0.0", false, "", ""},
		{"tooManySeparatorsLatest", "hello-frog@@", false, "", ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pluginName, version, err := getNameAndVersion(test.providedPlugin)
			if !test.isValid {
				assert.Error(t, err)
				return
			}
			assert.Equal(t, test.expectedName, pluginName)
			assert.Equal(t, test.expectedVersion, version)
		})
	}
}
