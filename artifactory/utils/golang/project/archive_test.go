package project

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/jfrog/jfrog-cli-go/utils/cliutils"
	"github.com/jfrog/jfrog-cli-go/utils/log"
	"github.com/jfrog/jfrog-cli-go/utils/tests"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils/checksum"
)

func TestArchiveProject(t *testing.T) {
	log.SetDefaultLogger()
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
	baseDir, dotGitPath := tests.PrepareDotGitDir(t, originalFolder, "testdata")
	err = archiveProject(buff, filepath.Join(pwd, "testdata"), "my/module/name", "1.0.0", regex)
	if err != nil {
		t.Error(err)
	}
	expected := map[checksum.Algorithm]string{checksum.MD5: "6d5616e2cbfbf0ec9d88bd3ca0fed938", checksum.SHA1: "0b9794f8805d56e887ef1c58e7447930b6140568"}
	actual, err := checksum.Calc(buff)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expecting: %v, Got: %v", expected, actual)
	}
	tests.RenamePath(dotGitPath, filepath.Join(baseDir, originalFolder), t)
}
