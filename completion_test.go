package main

import (
	"testing"

	"github.com/jfrog/jfrog-cli/completion/shells/bash"
	"github.com/jfrog/jfrog-cli/completion/shells/zsh"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/stretchr/testify/assert"
)

func TestBashCompletion(t *testing.T) {
	jfrogCli := tests.NewJfrogCli(execMain, "jfrog", "")
	content, err := tests.GetCmdOutput(t, jfrogCli, "completion", "bash")
	assert.NoError(t, err)
	assert.Equal(t, bash.BashAutocomplete, string(content))
}

func TestZshCompletion(t *testing.T) {
	jfrogCli := tests.NewJfrogCli(execMain, "jfrog", "")
	content, err := tests.GetCmdOutput(t, jfrogCli, "completion", "zsh")
	assert.NoError(t, err)
	assert.Equal(t, zsh.ZshAutocomplete, string(content))
}

func TestFishCompletion(t *testing.T) {
	jfrogCli := tests.NewJfrogCli(execMain, "jfrog", "")
	content, err := tests.GetCmdOutput(t, jfrogCli, "completion", "fish")
	assert.NoError(t, err)
	assert.Contains(t, string(content), "complete -c")
}
