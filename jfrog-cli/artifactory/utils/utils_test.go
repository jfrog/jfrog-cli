package utils

import (
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/config"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetHomeDir(t *testing.T) {
	path, err := GetJfrogSecurityDir()
	if err != nil {
		t.Error(err.Error())
	}
	homeDir, err := config.GetJfrogHomeDir()
	if err != nil {
		t.Error(err.Error())
	}
	expectedPath := filepath.Join(homeDir, "security")
	if path != expectedPath {
		t.Error("Expecting", expectedPath, "got:", path)
	}
}

func TestRemoveCredentialsFromURL(t *testing.T) {

	tests := []struct {
		name     string
		line     string
		expected string
	}{
		{"http", "This is an example line http://user:password@127.0.0.1:8081/artifactory/path/to/repo", "This is an example line http://***:***@127.0.0.1:8081/artifactory/path/to/repo"},
		{"https", "This is an example line https://user:password@127.0.0.1:8081/artifactory/path/to/repo", "This is an example line https://***:***@127.0.0.1:8081/artifactory/path/to/repo"},
		{"No credentials", "This is an example line https://127.0.0.1:8081/artifactory/path/to/repo", "This is an example line https://127.0.0.1:8081/artifactory/path/to/repo"},
		{"No http", "This is an example line", "This is an example line"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := removeCredentialsFromLine(test.line)
			if !strings.EqualFold(actual, test.expected) {
				t.Errorf("Expected: %s, Got: %s", test.expected, actual)
			}
		})
	}

}
