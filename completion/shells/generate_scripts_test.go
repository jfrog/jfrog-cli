package main

import (
	"github.com/jfrog/jfrog-cli-go/completion/shells/bash"
	"github.com/jfrog/jfrog-cli-go/completion/shells/zsh"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestGenrateScripts(t *testing.T) {
	bashPath := filepath.Join("bash", "jfrog")
	zshPath := filepath.Join("zsh", "jfrog")

	// Make sure test environment is clean before and after test
	os.Remove(bashPath)
	os.Remove(zshPath)
	defer os.Remove(bashPath)
	defer os.Remove(zshPath)

	// Run go generate ./...
	cmd := exec.Command("go", "generate", "./...")
	err := cmd.Run()
	assert.NoError(t, err)

	// Check bash completion script
	bashFile, err := os.Open(bashPath)
	defer bashFile.Close()
	assert.NoError(t, err)
	b, err := ioutil.ReadAll(bashFile)
	assert.Equal(t, bash.BashAutocomplete, string(b))

	// Check zsh completion script
	zshFile, err := os.Open(zshPath)
	defer zshFile.Close()
	assert.NoError(t, err)
	b, err = ioutil.ReadAll(zshFile)
	assert.Equal(t, zsh.ZshAutocomplete, string(b))
}
