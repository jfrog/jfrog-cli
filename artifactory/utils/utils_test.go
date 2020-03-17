package utils

import (
	"github.com/jfrog/jfrog-cli-go/utils/cliutils"
	"path/filepath"
	"testing"
)

func TestGetHomeDir(t *testing.T) {
	path, err := cliutils.GetJfrogSecurityDir()
	if err != nil {
		t.Error(err.Error())
	}
	homeDir, err := cliutils.GetJfrogHomeDir()
	if err != nil {
		t.Error(err.Error())
	}
	expectedPath := filepath.Join(homeDir, "security")
	if path != expectedPath {
		t.Error("Expecting", expectedPath, "got:", path)
	}
}
