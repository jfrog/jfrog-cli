package cliutils

import (
	"testing"
	"reflect"
	"encoding/json"
)

func TestReformatRegexp(t *testing.T) {
	assertReformatRegexp("1(.*)234", "1hello234", "{1}", "hello", t)
	assertReformatRegexp("1234", "1hello234", "{1}", "{1}", t)
	assertReformatRegexp("1(2.*5)6", "123456", "{1}", "2345", t)
	assertReformatRegexp("(.*) somthing", "doing somthing", "{1} somthing else", "doing somthing else", t)
	assertReformatRegexp("(switch) (this)", "switch this", "{2} {1}", "this switch", t)
	assertReformatRegexp("before(.*)middle(.*)after", "before123middle456after", "{2}{1}{2}", "456123456", t)
	assertReformatRegexp("", "nothing should change", "nothing should change", "nothing should change", t)
}

func assertReformatRegexp(regexp, source, dest, expected string, t *testing.T) {
    result, err := ReformatRegexp(regexp, source, dest)
	if err != nil {
		t.Error(err.Error())
	}
	if expected != result {
		t.Error("Unexpected string built. Expected: `" + expected + "` Got `" + result + "`")
	}
}

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