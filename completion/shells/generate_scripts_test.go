package main

import (
	"github.com/jfrog/jfrog-cli/completion/shells/bash"
	"github.com/jfrog/jfrog-cli/completion/shells/zsh"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestGenerateScripts(t *testing.T) {
	bashPath := filepath.Join("bash", "jfrog")
	zshPath := filepath.Join("zsh", "jfrog")

	// Make sure test environment is clean before and after test
	if _, err := os.Stat(bashPath); err == nil {
		assert.NoError(t, os.Remove(bashPath))
	}
	if _, err := os.Stat(zshPath); err == nil {
		assert.NoError(t, os.Remove(bashPath))
	}
	defer func() {
		if _, err := os.Stat(bashPath); err == nil {
			assert.NoError(t, os.Remove(bashPath))
		}
		if _, err := os.Stat(zshPath); err == nil {
			assert.NoError(t, os.Remove(bashPath))
		}
	}()

	// Run go generate ./...
	cmd := exec.Command("go", "generate", "./...")
	err := cmd.Run()
	assert.NoError(t, err)

	// Check bash completion script
	bashFile, err := os.Open(bashPath)
	defer func() {
		assert.NoError(t, bashFile.Close())
	}()
	assert.NoError(t, err)
	b, err := ioutil.ReadAll(bashFile)
	assert.Equal(t, bash.BashAutocomplete, string(b))

	// Check zsh completion script
	zshFile, err := os.Open(zshPath)
	defer func() {
		assert.NoError(t, zshFile.Close())
	}()
	assert.NoError(t, err)
	b, err = ioutil.ReadAll(zshFile)
	assert.Equal(t, zsh.ZshAutocomplete, string(b))
}
