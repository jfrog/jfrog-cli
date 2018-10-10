package utils

import "testing"
import "reflect"

func TestReformatRegexp(t *testing.T) {
	assertReformatDestByPaths("1(*)234", "1hello234", "{1}", "hello", true, t)
	assertReformatDestByPaths("1234", "1hello234", "{1}", "{1}", true, t)
	assertReformatDestByPaths("1(2*5)6", "123456", "{1}", "2345", true, t)
	assertReformatDestByPaths("(*) somthing", "doing somthing", "{1} somthing else", "doing somthing else", true, t)
	assertReformatDestByPaths("(switch) (this)", "switch this", "{2} {1}", "this switch", true, t)
	assertReformatDestByPaths("before(*)middle(*)after", "before123middle456after", "{2}{1}{2}", "456123456", true, t)
	assertReformatDestByPaths("foo/before(*)middle(*)after", "foo/before123middle456after", "{2}{1}{2}", "456123456", true, t)
	assertReformatDestByPaths("foo/before(*)middle(*)after", "bar/before123middle456after", "{2}{1}{2}", "456123456", true, t)
	assertReformatDestByPaths("foo/before(*)middle(*)after", "bar/before123middle456after", "{2}{1}{2}", "{2}{1}{2}", false, t)
	assertReformatDestByPaths("", "nothing should change", "nothing should change", "nothing should change", true, t)
}

func assertReformatDestByPaths(regexp, source, dest, expected string, ignoreRepo bool, t *testing.T) {
	result, err := ReformatDestByPaths(regexp, source, dest, ignoreRepo)
	if err != nil {
		t.Error(err.Error())
	}
	if expected != result {
		t.Error("Unexpected string built. Expected: `" + expected + "` Got `" + result + "`")
	}
}

func TestSplitWithEscape(t *testing.T) {
	assertSplitWithEscape("", []string{""}, t)
	assertSplitWithEscape("a", []string{"a"}, t)
	assertSplitWithEscape("a/b", []string{"a", "b"}, t)
	assertSplitWithEscape("a/b/c", []string{"a", "b", "c"}, t)
	assertSplitWithEscape("a/b\\5/c", []string{"a", "b5", "c"}, t)
	assertSplitWithEscape("a/b\\\\5.2/c", []string{"a", "b\\5.2", "c"}, t)
	assertSplitWithEscape("a\\8/b\\5/c", []string{"a8", "b5", "c"}, t)
	assertSplitWithEscape("a\\\\8/b\\\\5.2/c", []string{"a\\8", "b\\5.2", "c"}, t)
	assertSplitWithEscape("a/b\\5/c\\0", []string{"a", "b5", "c0"}, t)
	assertSplitWithEscape("a/b\\\\5.2/c\\\\0", []string{"a", "b\\5.2", "c\\0"}, t)
}

func assertSplitWithEscape(str string, expected []string, t *testing.T) {
	result := SplitWithEscape(str, '/')
	if !reflect.DeepEqual(result, expected) {
		t.Error("Unexpected string array built. Expected: `", expected, "` Got `", result, "`")
	}
}
