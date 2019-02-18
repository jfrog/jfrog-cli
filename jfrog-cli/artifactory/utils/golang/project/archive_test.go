package project

import (
	"bytes"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/tests"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils/checksum"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestArchiveProject(t *testing.T) {
	if cliutils.IsWindows() {
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
	baseDir, dotGitPath := tests.PrepareDotGitDir(t, originalFolder, "testdata", false)
	err = archiveProject(buff, filepath.Join(pwd, "testdata"), "my/module/name", "1.0.0", regex)
	if err != nil {
		t.Error(err)
	}
	expected := map[checksum.Algorithm]string{checksum.MD5: "24d14818afe0cf0a7e20ffa36518043e", checksum.SHA1: "97dca6a68810f07ac0f481da6080d799c11b1c4c"}
	actual, err := checksum.Calc(buff)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expecting: %v, Got: %v", expected, actual)
	}
	tests.RenamePath(dotGitPath, filepath.Join(baseDir, originalFolder), t)
}
