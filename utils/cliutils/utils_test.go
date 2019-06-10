package cliutils

import (
	"encoding/json"
	"errors"
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
