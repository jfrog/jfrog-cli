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
	expected := map[checksum.Algorithm]string{checksum.MD5: "ce70d45d713edafc0bab1709eb3c8f6c", checksum.SHA1: "d59f12054456491d58c8576f76fe9a5bdebf4e9c"}
	actual, err := checksum.Calc(buff)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expecting: %v, Got: %v", expected, actual)
	}
	tests.RenamePath(dotGitPath, filepath.Join(baseDir, originalFolder), t)
}
