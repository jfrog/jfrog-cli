package cliutils

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/jfrog/jfrog-cli-go/utils/log"
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
	log.SetDefaultLogger()
	var actual []byte
	actual = ReplaceSpecVars([]byte("${foo}aa"), map[string]string{"a": "k", "foo": "bar"})
	assertVariablesMap([]byte("baraa"), actual, t)

	actual = ReplaceSpecVars([]byte("a${foo}a"), map[string]string{"foo": "bar"})
	assertVariablesMap([]byte("abara"), actual, t)

	actual = ReplaceSpecVars([]byte("aa${foo}"), map[string]string{"foo": "bar"})
	assertVariablesMap([]byte("aabar"), actual, t)

	actual = ReplaceSpecVars([]byte("${foo}${foo}${foo}"), map[string]string{"foo": "bar"})
	assertVariablesMap([]byte("barbarbar"), actual, t)

	actual = ReplaceSpecVars([]byte("${talk}-${broh}-${foo}"), map[string]string{"foo": "bar", "talk": "speak", "broh": "sroh"})
	assertVariablesMap([]byte("speak-sroh-bar"), actual, t)

	actual = ReplaceSpecVars([]byte("a${foo}a"), map[string]string{"foo": ""})
	assertVariablesMap([]byte("aa"), actual, t)

	actual = ReplaceSpecVars([]byte("a${foo}a"), map[string]string{"a": "k", "f": "a"})
	assertVariablesMap([]byte("a${foo}a"), actual, t)

	actual = ReplaceSpecVars([]byte("a${foo}a"), map[string]string{})
	assertVariablesMap([]byte("a${foo}a"), actual, t)

	actual = ReplaceSpecVars(nil, nil)
	assertVariablesMap([]byte(""), actual, t)
}

func assertVariablesMap(expected, actual []byte, t *testing.T) {
	if 0 != bytes.Compare(expected, actual) {
		t.Error("Wrong matching expected: `" + string(expected) + "` Got `" + string(actual) + "`")
	}
}
