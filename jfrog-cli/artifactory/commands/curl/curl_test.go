package curl

import (
	"reflect"
	"testing"
)

func TestGetFlagValueAndValueIndex(t *testing.T) {
	tests := getFlagTestCases()
	command := &CurlCommand{}
	for _, test := range tests {
		command.Arguments = test.arguments
		t.Run(test.name, func(t *testing.T) {
			actualValue, actualIndex, err := command.getFlagValueAndValueIndex(test.flagName, test.flagIndex)

			// Validate errors.
			if err != nil && !test.expectErr {
				t.Error(err)
			}
			if err == nil && test.expectErr {
				t.Errorf("Expecting: error, Got: nil")
			}

			// Validate results.
			if actualValue != test.flagValue {
				t.Errorf("Expected value of: %s, got: %s.", test.flagValue, actualValue)
			}
			if actualIndex != test.flagValueIndex {
				t.Errorf("Expected value of: %d, got: %d.", test.flagValueIndex, actualIndex)
			}
		})
	}
}

func TestFindFlag(t *testing.T) {
	tests := getFlagTestCases()
	command := &CurlCommand{}
	for _, test := range tests {
		command.Arguments = test.arguments
		t.Run(test.name, func(t *testing.T) {
			actualIndex, actualValueIndex, actualValue, err := command.findFlag(test.flagName)

			// Check errors.
			if err != nil && !test.expectErr {
				t.Error(err)
			}
			if err == nil && test.expectErr {
				t.Errorf("Expecting: error, Got: nil")
			}

			if err == nil {
				// Validate results.
				if actualValue != test.flagValue {
					t.Errorf("Expected flag value of: %s, got: %s.", test.flagValue, actualValue)
				}
				if actualValueIndex != test.flagValueIndex {
					t.Errorf("Expected flag value index of: %d, got: %d.", test.flagValueIndex, actualValueIndex)
				}
				if actualIndex != test.flagIndex {
					t.Errorf("Expected flag index of: %d, got: %d.", test.flagIndex, actualIndex)
				}
			}
		})
	}
}

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
		command.Arguments = test
		actualArgIndex, actualArg := command.findNextArg()

		if actualArgIndex != expected[index].int {
			t.Errorf("Expected arg index of: %d, got: %d.", expected[index].int, actualArgIndex)
		}
		if actualArg != expected[index].string {
			t.Errorf("Expected arg index of: %s, got: %s.", expected[index].string, actualArg)
		}
	}
}

func TestGetAndRemoveServerIdFromCommand(t *testing.T) {
	command := &CurlCommand{}
	args := [][]string{
		{"-X", "GET", "/api/build/test1", "--server-id", "test1", "--foo", "bar"},
		{"-X", "GET", "/api/build/test2", "--server-idea", "foo", "--server-id=test2"},
		{"-X", "GET", "api/build/test3", "--server-id", "test3", "--foo", "bar"},
	}

	expected := []struct {
		value   string
		command []string
	}{
		{"test1", []string{"-X", "GET", "/api/build/test1", "--foo", "bar"}},
		{"test2", []string{"-X", "GET", "/api/build/test2", "--server-idea", "foo"}},
		{"test3", []string{"-X", "GET", "api/build/test3", "--foo", "bar"}},
	}

	for index, test := range args {
		command.Arguments = test
		serverIdValue, err := command.getAndRemoveServerIdFromCommand()
		if err != nil {
			t.Error(err)
		}
		if serverIdValue != expected[index].value {
			t.Errorf("Expected --server-id value: %s, got: %s.", expected[index].value, serverIdValue)
		}
		if !reflect.DeepEqual(command.Arguments, expected[index].command) {
			t.Errorf("Expected command arguments: %v, got: %v.", expected[index].command, command.Arguments)
		}
	}
}

func TestBuildCommandUrl(t *testing.T) {
	tests := []struct {
		name      string
		arguments []string
		urlIndex  int
		urlValue  string
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
		command.Arguments = test.arguments
		t.Run(test.name, func(t *testing.T) {
			urlIndex, urlValue, err := command.buildCommandUrl(urlPrefix)

			// Check errors.
			if err != nil && !test.expectErr {
				t.Error(err)
			}
			if err == nil && test.expectErr {
				t.Errorf("Expecting: error, Got: nil")
			}

			if err == nil {
				// Validate results.
				if urlValue != test.urlValue {
					t.Errorf("Expected url value of: %s, got: %s.", test.urlValue, urlValue)
				}
				if urlIndex != test.urlIndex {
					t.Errorf("Expected url index of: %d, got: %d.", test.urlIndex, urlIndex)
				}
			}
		})
	}
}

func getFlagTestCases() []testCase {
	return []testCase{
		{"test1", []string{"-X", "GET", "/api/build/test1", "--server-id", "test1", "--foo", "bar"}, "--server-id", 3, "test1", 4, false},
		{"test2", []string{"-X", "GET", "/api/build/test2", "--server-idea", "foo", "--server-id=test2"}, "--server-id", 5, "test2", 5, false},
		{"test3", []string{"-XGET", "/api/build/test3", "--server-id="}, "--server-id", 2, "", -1, true},
		{"test4", []string{"-XGET", "/api/build/test4", "--server-id", "--foo", "bar"}, "--server-id", 2, "", -1, true},
		{"test5", []string{"-X", "GET", "api/build/test5", "--server-id", "test5", "--foo", "bar"}, "--server-id", 3, "test5", 4, false},
	}
}

type testCase struct {
	name           string
	arguments      []string
	flagName       string
	flagIndex      int
	flagValue      string
	flagValueIndex int
	expectErr      bool
}
