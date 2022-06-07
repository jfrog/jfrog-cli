package main

import (
	"github.com/jfrog/jfrog-cli/completion/shells/bash"
	"github.com/jfrog/jfrog-cli/completion/shells/zsh"
	clientTestUtils "github.com/jfrog/jfrog-client-go/utils/tests"
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
		clientTestUtils.RemoveAndAssert(t, bashPath)
	}
	if _, err := os.Stat(zshPath); err == nil {
		clientTestUtils.RemoveAndAssert(t, zshPath)
	}
	defer func() {
		if _, err := os.Stat(bashPath); err == nil {
			clientTestUtils.RemoveAndAssert(t, bashPath)
		}
		if _, err := os.Stat(zshPath); err == nil {
			clientTestUtils.RemoveAndAssert(t, zshPath)
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
	assert.NoError(t, err)
	assert.Equal(t, bash.BashAutocomplete, string(b))

	// Check zsh completion script
	zshFile, err := os.Open(zshPath)
	defer func() {
		assert.NoError(t, zshFile.Close())
	}()
	assert.NoError(t, err)
	b, err = ioutil.ReadAll(zshFile)
	assert.NoError(t, err)
	assert.Equal(t, zsh.ZshAutocomplete, string(b))
}
