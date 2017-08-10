package cliutils

import (
	"testing"
	"reflect"
	"encoding/json"
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