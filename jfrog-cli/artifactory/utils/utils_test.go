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
		regex        RegExpStruct
		expectedLine string
		matched      bool
	}{
		{"http", RegExpStruct{RegExp: regExpProtocol, Separator: "//", Replacer: "//***.***@", line: "This is an example line http://user:password@127.0.0.1:8081/artifactory/path/to/repo"}, "This is an example line http://***.***@127.0.0.1:8081/artifactory/path/to/repo", true},
		{"https", RegExpStruct{RegExp: regExpProtocol, Separator: "//", Replacer: "//***.***@", line: "This is an example line https://user:password@127.0.0.1:8081/artifactory/path/to/repo"}, "This is an example line https://***.***@127.0.0.1:8081/artifactory/path/to/repo", true},
		{"No credentials", RegExpStruct{RegExp: regExpProtocol, Separator: "//", Replacer: "//***.***@", line: "This is an example line https://127.0.0.1:8081/artifactory/path/to/repo"}, "This is an example line https://127.0.0.1:8081/artifactory/path/to/repo", false},
		{"No http", RegExpStruct{RegExp: regExpProtocol, Separator: "//", Replacer: "//***.***@", line: "This is an example line"}, "This is an example line", false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.regex.matcher = test.regex.RegExp.FindString(test.regex.line)
			if test.matched && test.regex.matcher == "" {
				t.Error("Expected to find a match.")
			}
			if test.matched && test.regex.matcher != "" {
				actual, _ := test.regex.MaskCredentials()
				if !strings.EqualFold(actual, test.expectedLine) {
					t.Errorf("Expected: %s, Got: %s", test.expectedLine, actual)
				}
			}
			if !test.matched && test.regex.matcher != "" {
				t.Error("Expected to find zero match, found:", test.regex.matcher)
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
		regex RegExpStruct
		error bool
	}{
		{"Without Error", RegExpStruct{RegExp: regExpProtocol, line: "This is an example line http://user:password@127.0.0.1:8081/artifactory/path/to/repo"}, false},
		{"With Error", RegExpStruct{RegExp: regExpProtocol, line: "This is an example line http://user:password@127.0.0.1:8081/artifactory/path/to/repo: 404 Not Found"}, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.regex.matcher = test.regex.RegExp.FindString(test.regex.line)
			if test.regex.matcher == "" && test.error {
				t.Error("Expected to find 404 not found")
			}
			if test.regex.matcher != "" && !test.error {
				t.Error("Expected regex to return empty result")
			}
		})
	}
}
