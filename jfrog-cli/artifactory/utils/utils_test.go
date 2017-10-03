package utils

import (
	"testing"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"
	"path/filepath"
)

func TestGetHomeDir(t *testing.T) {
	path, err := GetJfrogSecurityDir()
	if err != nil {
		t.Error(err.Error())
	}
	homeDir, err := config.GetJfrogHomeDir()
	if err != nil {
		t.Error(err.Error())
	}
	expectedPath := filepath.Join(homeDir, "security")
	if path != expectedPath {
		t.Error("Expecting", expectedPath, "got:", path)
	}
}
