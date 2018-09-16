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
	regExpProtocol, err := GetRegExp(`((http|https):\/\/\w.*?:\w.*?@)`)
	if err != nil {
		t.Error(err)
	}

	tests := []struct {
		name         string
		regex        CmdOutputPattern
		expectedLine string
		matched      bool
	}{
		{"http", CmdOutputPattern{RegExp: regExpProtocol, line: "This is an example line http://user:password@127.0.0.1:8081/artifactory/path/to/repo"}, "This is an example line http://***.***@127.0.0.1:8081/artifactory/path/to/repo", true},
		{"https", CmdOutputPattern{RegExp: regExpProtocol, line: "This is an example line https://user:password@127.0.0.1:8081/artifactory/path/to/repo"}, "This is an example line https://***.***@127.0.0.1:8081/artifactory/path/to/repo", true},
		{"No credentials", CmdOutputPattern{RegExp: regExpProtocol, line: "This is an example line https://127.0.0.1:8081/artifactory/path/to/repo"}, "This is an example line https://127.0.0.1:8081/artifactory/path/to/repo", false},
		{"No http", CmdOutputPattern{RegExp: regExpProtocol, line: "This is an example line"}, "This is an example line", false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.regex.matchedResult = test.regex.RegExp.FindString(test.regex.line)
			if test.matched && test.regex.matchedResult == "" {
				t.Error("Expected to find a match.")
			}
			if test.matched && test.regex.matchedResult != "" {
				actual, _ := test.regex.MaskCredentials()
				if !strings.EqualFold(actual, test.expectedLine) {
					t.Errorf("Expected: %s, The Regex found %s and the masked line: %s", test.expectedLine, test.regex.matchedResult, actual)
				}
			}
			if !test.matched && test.regex.matchedResult != "" {
				t.Error("Expected to find zero match, found:", test.regex.matchedResult)
			}
		})
	}
}

func TestReturnErrorOnNotFound(t *testing.T) {
	regExpProtocol, err := GetRegExp(`(404 Not Found)`)
	if err != nil {
		t.Error(err)
	}

	tests := []struct {
		name  string
		regex CmdOutputPattern
		error bool
	}{
		{"Without Error", CmdOutputPattern{RegExp: regExpProtocol, line: "This is an example line http://user:password@127.0.0.1:8081/artifactory/path/to/repo"}, false},
		{"With Error", CmdOutputPattern{RegExp: regExpProtocol, line: "This is an example line http://user:password@127.0.0.1:8081/artifactory/path/to/repo: 404 Not Found"}, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.regex.matchedResult = test.regex.RegExp.FindString(test.regex.line)
			if test.regex.matchedResult == "" && test.error {
				t.Error("Expected to find 404 not found, found nothing.")
			}
			if test.regex.matchedResult != "" && !test.error {
				t.Error("Expected regex to return empty result. Got:", test.regex.matchedResult)
			}
		})
	}
}
