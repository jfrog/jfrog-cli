package utils

import (
	"os"
	"reflect"
	"testing"

	"github.com/jfrog/jfrog-cli-go/utils/cliutils"
	"github.com/stretchr/testify/assert"
)

func TestFindAndRemoveFlagFromCommand(t *testing.T) {
	args := [][]string{
		{"-X", "GET", "/api/build/test1", "--server-id", "test1", "--foo", "bar"},
		{"-X", "GET", "/api/build/test2", "--server-idea", "foo", "--server-id=test2"},
		{"-X", "GET", "api/build/test3", "--server-id", "test3", "--foo", "bar"},
		{"-X", "GET", "api/build/test3", "--build-name", "name", "--foo", "bar"},
		{"-X", "GET", "api/build/test3", "--build-number", "3", "--foo", "bar"},
	}

	expected := []struct {
		key     string
		value   string
		command []string
	}{
		{"--server-id", "test1", []string{"-X", "GET", "/api/build/test1", "--foo", "bar"}},
		{"--server-id", "test2", []string{"-X", "GET", "/api/build/test2", "--server-idea", "foo"}},
		{"--server-id", "test3", []string{"-X", "GET", "api/build/test3", "--foo", "bar"}},
		{"--build-name", "name", []string{"-X", "GET", "api/build/test3", "--foo", "bar"}},
		{"--build-number", "3", []string{"-X", "GET", "api/build/test3", "--foo", "bar"}},
	}

	for index, test := range args {
		flagIndex, valueIndex, keyValue, err := FindFlag(expected[index].key, test)
		if err != nil {
			t.Error(err)
		}
		if keyValue != expected[index].value {
			t.Errorf("Expected %s value: %s, got: %s.", expected[index].key, expected[index].value, keyValue)
		}
		RemoveFlagFromCommand(&test, flagIndex, valueIndex)
		if !reflect.DeepEqual(test, expected[index].command) {
			t.Errorf("Expected command arguments: %v, got: %v.", expected[index].command, test)
		}
	}
}

func TestFindFlag(t *testing.T) {
	tests := getFlagTestCases()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actualIndex, actualValueIndex, actualValue, err := FindFlag(test.flagName, test.arguments)

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

func TestGetFlagValueAndValueIndex(t *testing.T) {
	tests := getFlagTestCases()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actualValue, actualIndex, err := getFlagValueAndValueIndex(test.flagName, test.arguments, test.flagIndex)

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

func TestFindBooleanFlag(t *testing.T) {
	tests := []struct {
		flagName      string
		command       []string
		expectedIndex int
		expectedValue bool
		shouldFail    bool
	}{
		{"--foo", []string{"-X", "--GET", "--foo/api/build/test1", "--foo", "bar"}, 3, true, false},
		{"--server-id", []string{"-X", "GET", "/api/build/test2", "--server-idea", "foo"}, -1, false, false},
		{"--bar", []string{"-X", "GET", "api/build/test3", "--foo", "bar"}, -1, false, false},
		{"-X", []string{"-X=true", "GET", "api/build/test3", "--foo", "bar"}, 0, true, false},
		{"--json", []string{"-X=true", "GET", "api/build/test3", "--foo", "--json=false"}, 4, false, false},
		{"--dry-run", []string{"-X=falsee", "GET", "api/build/test3", "--dry-run=falsee", "--json"}, 3, false, true},
	}

	for testIndex, test := range tests {
		actualIndex, actualValue, err := FindBooleanFlag(test.flagName, test.command)
		if test.shouldFail && err == nil {
			t.Errorf("Test #%d: Should fail to parse the boolean value, but ended with nil error.", testIndex)
		}
		if actualIndex != test.expectedIndex {
			t.Errorf("Test #%d: Expected index value: %d, got: %d.", testIndex, test.expectedIndex, actualIndex)
		}
		if actualValue != test.expectedValue {
			t.Errorf("Test #%d: Expected value: %t, got: %t.", testIndex, test.expectedValue, actualValue)
		}
	}
}

func TestExtractBuildDetailsFromArgs(t *testing.T) {
	tests := []struct {
		command             []string
		expectedArgs        []string
		expectedBuildConfig *BuildConfiguration
	}{
		{[]string{"-test", "--build-name", "test1", "--foo", "--build-number", "1", "--module", "module1"}, []string{"-test", "--foo"}, &BuildConfiguration{"test1", "1", "module1"}},
		{[]string{"--module=module2", "--build-name", "test2", "--foo", "bar", "--build-number=2"}, []string{"--foo", "bar"}, &BuildConfiguration{"test2", "2", "module2"}},
		{[]string{"foo", "-X", "123", "--build-name", "test3", "--bar", "--build-number=3", "--foox"}, []string{"foo", "-X", "123", "--bar", "--foox"}, &BuildConfiguration{"test3", "3", ""}},
	}

	for _, test := range tests {
		actualArgs, actualBuildConfig, err := ExtractBuildDetailsFromArgs(test.command)
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(actualArgs, test.expectedArgs) {
			t.Errorf("Expected value: %v, got: %v.", test.expectedArgs, actualArgs)
		}
		if !reflect.DeepEqual(actualBuildConfig, test.expectedBuildConfig) {
			t.Errorf("Expected value: %v, got: %v.", test.expectedBuildConfig, actualBuildConfig)
		}
	}
}

func TestExtractBuildDetailsFromEnv(t *testing.T) {
	const buildNameEnv = "envBuildName"
	const buildNumberEnv = "777"
	tests := []struct {
		command             []string
		expectedArgs        []string
		expectedBuildConfig *BuildConfiguration
	}{
		{[]string{"-test", "--build-name", "test1", "--foo", "--build-number", "1", "--module", "module1"}, []string{"-test", "--foo"}, &BuildConfiguration{"test1", "1", "module1"}},
		{[]string{"foo", "-X", "123", "--bar", "--build-name=test3", "--build-number=3", "--foox"}, []string{"foo", "-X", "123", "--bar", "--foox"}, &BuildConfiguration{"test3", "3", ""}},
		{[]string{"foo", "-X", "123", "--bar", "--build-name=test1", "--build-number=1", "--foox"}, []string{"foo", "-X", "123", "--bar", "--foox"}, &BuildConfiguration{"test1", "1", ""}},
		{[]string{"foo", "-X", "123", "--bar", "--foox"}, []string{"foo", "-X", "123", "--bar", "--foox"}, &BuildConfiguration{buildNameEnv, buildNumberEnv, ""}},
	}

	os.Setenv(cliutils.BuildName, buildNameEnv)
	os.Setenv(cliutils.BuildNumber, buildNumberEnv)
	defer os.Unsetenv(cliutils.BuildName)
	defer os.Unsetenv(cliutils.BuildNumber)
	for _, test := range tests {
		actualArgs, actualBuildConfig, err := ExtractBuildDetailsFromArgs(test.command)
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(actualArgs, test.expectedArgs) {
			t.Errorf("Expected value: %v, got: %v.", test.expectedArgs, actualArgs)
		}
		if !reflect.DeepEqual(actualBuildConfig, test.expectedBuildConfig) {
			t.Errorf("Expected value: %v, got: %v.", test.expectedBuildConfig, actualBuildConfig)
		}
	}
}

func TestFindFlagFirstMatch(t *testing.T) {
	tests := []struct {
		command            []string
		flags              []string
		expectedFlagIndex  int
		expectedValueIndex int
		expectedValue      string
	}{
		{[]string{"-test", "--build-name", "test1", "--foo", "--build-number", "1", "--module", "module1"}, []string{"--build", "--build-name"}, 1, 2, "test1"},
		{[]string{"--module=module2", "--build-name", "test2", "--foo", "bar", "--build-number=2"}, []string{"--build-name", "--module"}, 1, 2, "test2"},
		{[]string{"foo", "-X", "123", "--bar", "--build-number=3", "--foox=barx"}, []string{"-Y", "--foo", "--foox"}, 5, 5, "barx"},
	}

	for _, test := range tests {
		actualFlagIndex, actualValueIndex, actualValue, err := FindFlagFirstMatch(test.flags, test.command)
		if err != nil {
			t.Error(err)
		}
		// Validate results.
		if actualValue != test.expectedValue {
			t.Errorf("Expected flag value of: %s, got: %s.", test.expectedValue, actualValue)
		}
		if actualValueIndex != test.expectedValueIndex {
			t.Errorf("Expected flag value index of: %d, got: %d.", test.expectedValueIndex, actualValueIndex)
		}
		if actualFlagIndex != test.expectedFlagIndex {
			t.Errorf("Expected flag index of: %d, got: %d.", test.expectedFlagIndex, actualFlagIndex)
		}
	}
}

func getFlagTestCases() []testCase {
	return []testCase{
		{"test1", []string{"-X", "GET", "/api/build/test1", "--server-id", "test1", "--foo", "bar"}, "--server-id", 3, "test1", 4, false},
		{"test2", []string{"-X", "GET", "/api/build/test2", "--server-idea", "foo", "--server-id=test2"}, "--server-id", 5, "test2", 5, false},
		{"test3", []string{"-XGET", "/api/build/test3", "--server-id="}, "--server-id", 2, "", -1, true},
		{"test4", []string{"-XGET", "/api/build/test4", "--build-name", "--foo", "bar"}, "--build-name", 2, "", -1, true},
		{"test5", []string{"-X", "GET", "api/build/test5", "--build-number", "4", "--foo", "bar"}, "--build-number", 3, "4", 4, false},
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

func TestParseArgs(t *testing.T) {
	// Check with no quote
	want := []string{`rt`}
	got, _ := ParseArgs([]string{`rt`})
	assert.EqualValues(t, got, want)
	// Check middle quote
	want = []string{`pom.xml`}
	got, _ = ParseArgs([]string{`"p"o"m.xml"`})
	assert.EqualValues(t, got, want)

	// Check with space and backslash
	want = []string{`-f a\b\pom.xml`}
	got, _ = ParseArgs([]string{`"-f a\b\pom.xml"`})
	assert.EqualValues(t, got, want)
	os.Setenv("JFROGPARSETEST", "jfrog")
	defer os.Unsetenv("JFROGPARSETEST")

	// Check env variable parse
	want = []string{`--build-name=jfrog`}
	got, _ = ParseArgs([]string{`"--build-name=$JFROGPARSETEST"`})
	assert.EqualValues(t, got, want)

	want = []string{`--build-name=jfrog jfrog`, "--build-number=1"}
	got, _ = ParseArgs([]string{`"--build-name=jfrog jfrog"`, "--build-number=1"})
	assert.EqualValues(t, got, want)
}
