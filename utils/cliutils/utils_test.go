package cliutils

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/magiconair/properties/assert"
	"reflect"
	"testing"
)

func TestSpecVarsStringToMap(t *testing.T) {
	var actual map[string]string
	actual = SpecVarsStringToMap("")
	assertSpecVars(nil, actual, t)

	actual = SpecVarsStringToMap("foo=bar")
	assertSpecVars(map[string]string{"foo": "bar"}, actual, t)

	actual = SpecVarsStringToMap("foo=bar;bar=foo")
	assertSpecVars(map[string]string{"foo": "bar", "bar": "foo"}, actual, t)

	actual = SpecVarsStringToMap("foo=bar\\;bar=foo")
	assertSpecVars(map[string]string{"foo": "bar;bar=foo"}, actual, t)

	actual = SpecVarsStringToMap("a=b;foo=foo=bar\\;bar=foo")
	assertSpecVars(map[string]string{"foo": "foo=bar;bar=foo", "a": "b"}, actual, t)

	actual = SpecVarsStringToMap("foo=bar;foo=bar")
	assertSpecVars(map[string]string{"foo": "bar"}, actual, t)
}

func assertSpecVars(expected, actual map[string]string, t *testing.T) {
	if !reflect.DeepEqual(expected, actual) {
		expectedMap, _ := json.Marshal(expected)
		actualMap, _ := json.Marshal(actual)
		t.Error("Wrong matching expected: `" + string(expectedMap) + "` Got `" + string(actualMap) + "`")
	}
}

func TestGetExitCode(t *testing.T) {
	// No error
	exitCode := GetExitCode(nil, 0, 0, false)
	checkExitCode(t, ExitCodeNoError, exitCode)

	// Empty error
	exitCode = GetExitCode(errors.New(""), 0, 0, false)
	checkExitCode(t, ExitCodeError, exitCode)

	// Regular error
	exitCode = GetExitCode(errors.New("Error"), 0, 0, false)
	checkExitCode(t, ExitCodeError, exitCode)

	// Fail-no-op true without success
	exitCode = GetExitCode(nil, 0, 0, true)
	checkExitCode(t, ExitCodeFailNoOp, exitCode)

	// Fail-no-op true with success
	exitCode = GetExitCode(nil, 1, 0, true)
	checkExitCode(t, ExitCodeNoError, exitCode)

	// Fail-no-op false
	exitCode = GetExitCode(nil, 0, 0, false)
	checkExitCode(t, ExitCodeNoError, exitCode)
}

func checkExitCode(t *testing.T, expected, actual ExitCode) {
	if expected != actual {
		t.Errorf("Exit code expected %v, got %v", expected, actual)
	}
}

func TestReplaceSpecVars(t *testing.T) {
	log.SetLogger(log.NewLogger(log.INFO, nil))
	var actual []byte
	actual = ReplaceVars([]byte("${foo}aa"), map[string]string{"a": "k", "foo": "bar"})
	assertVariablesMap([]byte("baraa"), actual, t)

	actual = ReplaceVars([]byte("a${foo}a"), map[string]string{"foo": "bar"})
	assertVariablesMap([]byte("abara"), actual, t)

	actual = ReplaceVars([]byte("aa${foo}"), map[string]string{"foo": "bar"})
	assertVariablesMap([]byte("aabar"), actual, t)

	actual = ReplaceVars([]byte("${foo}${foo}${foo}"), map[string]string{"foo": "bar"})
	assertVariablesMap([]byte("barbarbar"), actual, t)

	actual = ReplaceVars([]byte("${talk}-${broh}-${foo}"), map[string]string{"foo": "bar", "talk": "speak", "broh": "sroh"})
	assertVariablesMap([]byte("speak-sroh-bar"), actual, t)

	actual = ReplaceVars([]byte("a${foo}a"), map[string]string{"foo": ""})
	assertVariablesMap([]byte("aa"), actual, t)

	actual = ReplaceVars([]byte("a${foo}a"), map[string]string{"a": "k", "f": "a"})
	assertVariablesMap([]byte("a${foo}a"), actual, t)

	actual = ReplaceVars([]byte("a${foo}a"), map[string]string{})
	assertVariablesMap([]byte("a${foo}a"), actual, t)

	actual = ReplaceVars(nil, nil)
	assertVariablesMap([]byte(""), actual, t)
}

func assertVariablesMap(expected, actual []byte, t *testing.T) {
	if 0 != bytes.Compare(expected, actual) {
		t.Error("Wrong matching expected: `" + string(expected) + "` Got `" + string(actual) + "`")
	}
}

type yesNoCases struct {
	ans string
	def bool
	expectedParsed bool
	expectedValid bool
	testName string
}
func TestParseYesNo(t *testing.T) {
	cases := []yesNoCases{
		// Positive answer.
		{"yes", true, true, true, "yes"},
		{"y", true, true, true, "y"},
		{"Y", true, true, true, "y capital"},
		{"YES", true, true, true, "yes capital"},
		{"y", false, true, true, "positive with different default"},

		// Negative answer.
		{"no", true, false, true, "no"},
		{"n", true, false, true, "n"},
		{"N", true, false, true, "n capital"},
		{"NO", true, false, true, "no capital"},
		{"n", false, false, true, "negative with different default"},

		// Default answer.
		{"", true, true, true, "empty with positive default"},
		{"", false, false, true, "empty with negative default"},
		{" ", false, false, true, "empty with space"},

		// Spaces.
		{" y", false, true, true, "space before"},
		{"yes ", true, true, true, "space after"},

		// Invalid answer.
		{"notvalid", false, false, false, "invalid all lower"},
		{"yNOVALIDyes", false, false, false, "invalid changing"},
	}

	for _, c := range cases {
		t.Run(c.testName, func(t *testing.T) {
			actualParsed, actualValid := parseYesNo(c.ans, c.def)
			assert.Equal(t, actualValid, c.expectedValid)
			if actualValid {
				assert.Equal(t, actualParsed, c.expectedParsed)
			}
		})
	}
}