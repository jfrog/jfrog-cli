package artifactory

import (
	"testing"
)

func TestValidateGoNativeCommand(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{"withInvalidCommand", []string{"build", "-test", "-another", "--publish-deps"}, "--publish-deps"},
		{"withSimilarCommand", []string{"build", "-test", "-another", "publish-deps"}, ""},
		{"withoutAnyFlags", []string{"build", "-test", "-another"}, ""},
		{"withFlagAndValue", []string{"build", "-test", "-another", "--url=http://another.com"}, "--url"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := validateGoNativeCommand(test.args)
			if result != test.expected {
				t.Errorf("Expected %s, got %s", test.expected, result)
			}
		})
	}
}
