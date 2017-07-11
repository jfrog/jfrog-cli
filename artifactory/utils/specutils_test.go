package utils

import (
	"testing"
	"bytes"
)

func TestReplaceSpecVars(t *testing.T) {
	var actual []byte
	actual = replaceSpecVars([]byte("${foo}aa"), map[string]string{"a": "k", "foo": "bar"})
	assertVariablesMap([]byte("baraa"), actual, t)

	actual = replaceSpecVars([]byte("a${foo}a"), map[string]string{"foo": "bar"})
	assertVariablesMap([]byte("abara"), actual, t)

	actual = replaceSpecVars([]byte("aa${foo}"), map[string]string{"foo": "bar"})
	assertVariablesMap([]byte("aabar"), actual, t)

	actual = replaceSpecVars([]byte("${foo}${foo}${foo}"), map[string]string{"foo": "bar"})
	assertVariablesMap([]byte("barbarbar"), actual, t)

	actual = replaceSpecVars([]byte("${talk}-${broh}-${foo}"), map[string]string{"foo": "bar", "talk": "speak", "broh": "sroh"})
	assertVariablesMap([]byte("speak-sroh-bar"), actual, t)

	actual = replaceSpecVars([]byte("a${foo}a"), map[string]string{"foo": ""})
	assertVariablesMap([]byte("aa"), actual, t)

	actual = replaceSpecVars([]byte("a${foo}a"), map[string]string{"a": "k", "f": "a"})
	assertVariablesMap([]byte("a${foo}a"), actual, t)

	actual = replaceSpecVars([]byte("a${foo}a"), map[string]string{})
	assertVariablesMap([]byte("a${foo}a"), actual, t)

	actual = replaceSpecVars(nil, nil)
	assertVariablesMap([]byte(""), actual, t)
}

func assertVariablesMap(expected, actual []byte, t *testing.T)  {
	if 0 != bytes.Compare(expected, actual) {
		t.Error("Wrong matching expected: `" + string(expected) + "` Got `" + string(actual) + "`")
	}
}