package project

import (
	"bytes"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/tests"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/io/fileutils/checksum"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

func TestArchiveProject(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping archive test...")
	}
	pwd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}

	buff := &bytes.Buffer{}
	regex, err := getPathExclusionRegExp()
	if err != nil {
		t.Error(err)
	}
	originalFolder := "test_.git_suffix"
	baseDir, dotGitPath := tests.PrepareDotGitDir(t, originalFolder, false)
	err = archiveProject(buff, filepath.Join(pwd, "testdata"), "my/module/name", "1.0.0", regex)
	if err != nil {
		t.Error(err)
	}
	expected := map[checksum.Algorithm]string{checksum.MD5: "5f3b3609258f05c1b2d52a66e8d54e2a", checksum.SHA1: "2f2ccb2e42c0d3abd351a8d0b9ba488e1277cf23"}
	actual, err := checksum.Calc(buff)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expecting: %v, Got: %v", expected, actual)
	}
	tests.RenamePath(dotGitPath, filepath.Join(baseDir, originalFolder), t)
}
