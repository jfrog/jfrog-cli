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
	assert.NoError(t, os.Remove(bashPath))
	assert.NoError(t, os.Remove(zshPath))
	defer func() {
		assert.NoError(t, os.Remove(bashPath))
		assert.NoError(t, os.Remove(zshPath))
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
