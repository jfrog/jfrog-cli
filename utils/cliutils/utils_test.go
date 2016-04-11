package cliutils

import "testing"

func TestReformatRegexp(t *testing.T) {
	assertReformatRegexp("1(.*)234", "1hello234", "{1}", "hello", t)
	assertReformatRegexp("1234", "1hello234", "{1}", "{1}", t)
	assertReformatRegexp("1(2.*5)6", "123456", "{1}", "2345", t)
	assertReformatRegexp("(.*) somthing", "doing somthing", "{1} somthing else", "doing somthing else", t)
	assertReformatRegexp("(switch) (this)", "switch this", "{2} {1}", "this switch", t)
	assertReformatRegexp("before(.*)middle(.*)after", "before123middle456after", "{2}{1}{2}", "456123456", t)
}

func assertReformatRegexp(regexp, source, dest, expected string, t *testing.T) {
	if ( expected != ReformatRegexp(regexp, source, dest)) {
		t.Error("Unexpected string built. Expected: `" + expected + "` Got `" + ReformatRegexp(regexp, source, dest) + "`")
	}
}
