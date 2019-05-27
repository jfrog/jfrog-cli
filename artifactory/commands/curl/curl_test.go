package curl

import (
	"testing"
)

func TestFindNextArg(t *testing.T) {
	command := &CurlCommand{}
	args := [][]string{
		{"-X", "GET", "arg1", "--foo", "bar"},
		{"-X", "GET", "--server-idea", "foo", "/api/arg2"},
		{"-XGET", "--foo", "bar", "--foo-bar", "meow", "arg3"},
	}

	expected := []struct {
		int
		string
	}{
		{2, "arg1"},
		{4, "/api/arg2"},
		{5, "arg3"},
	}

	for index, test := range args {
		command.arguments = test
		actualArgIndex, actualArg := command.findUriValueAndIndex()

		if actualArgIndex != expected[index].int {
			t.Errorf("Expected arg index of: %d, got: %d.", expected[index].int, actualArgIndex)
		}
		if actualArg != expected[index].string {
			t.Errorf("Expected arg index of: %s, got: %s.", expected[index].string, actualArg)
		}
	}
}

func TestIsCredsFlagExists(t *testing.T) {
	command := &CurlCommand{}
	args := [][]string{
		{"-X", "GET", "arg1", "--foo", "bar", "-uaaa:ppp"},
		{"-X", "GET", "--server-idea", "foo", "-u", "aaa:ppp", "/api/arg2"},
		{"-XGET", "--foo", "bar", "--foo-bar", "--user", "meow", "-Ttest"},
		{"-XGET", "--foo", "bar", "--foo-bar", "-Ttest"},
	}

	expected := []bool{
		true,
		true,
		true,
		false,
	}

	for index, test := range args {
		command.arguments = test
		flagExists := command.isCredentialsFlagExists()

		if flagExists != expected[index] {
			t.Errorf("Expected flag existstence to be: %t, got: %t.", expected[index], flagExists)
		}
	}
}

func TestBuildCommandUrl(t *testing.T) {
	tests := []struct {
		name      string
		arguments []string
		uriIndex  int
		uriValue  string
		expectErr bool
	}{
		{"test1", []string{"-X", "GET", "/api/build/test1", "--server-id", "test1", "--foo", "bar"}, 2, "http://artifactory:8081/artifactory/api/build/test1", false},
		{"test2", []string{"-X", "GET", "/api/build/test2", "--server-idea", "foo", "--server-id=test2"}, 2, "http://artifactory:8081/artifactory/api/build/test2", false},
		{"test3", []string{"-XGET", "--/api/build/test3", "--server-id="}, 1, "http://artifactory:8081/artifactory/api/build/test3", true},
		{"test4", []string{"-XGET", "-Test4", "--server-id", "bar"}, 1, "http://artifactory:8081/artifactory/api/build/test4", true},
		{"test5", []string{"-X", "GET", "api/build/test5", "--server-id", "test5", "--foo", "bar"}, 2, "http://artifactory:8081/artifactory/api/build/test5", false},
	}

	command := &CurlCommand{}
	urlPrefix := "http://artifactory:8081/artifactory/"
	for _, test := range tests {
		command.arguments = test.arguments
		t.Run(test.name, func(t *testing.T) {
			uriIndex, uriValue, err := command.buildCommandUrl(urlPrefix)

			// Check errors.
			if err != nil && !test.expectErr {
				t.Error(err)
			}
			if err == nil && test.expectErr {
				t.Errorf("Expecting: error, Got: nil")
			}

			if err == nil {
				// Validate results.
				if uriValue != test.uriValue {
					t.Errorf("Expected uri value of: %s, got: %s.", test.uriValue, uriValue)
				}
				if uriIndex != test.uriIndex {
					t.Errorf("Expected uri index of: %d, got: %d.", test.uriIndex, uriIndex)
				}
			}
		})
	}
}
