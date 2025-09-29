package main

import (
	"testing"

	coreTests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	"github.com/jfrog/jfrog-cli/completion/shells/bash"
	"github.com/jfrog/jfrog-cli/completion/shells/zsh"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/stretchr/testify/assert"
)

func TestBashCompletionSetup(t *testing.T) {
	jfrogCli := coreTests.NewJfrogCli(execMain, "jfrog", "")
	content, _, err := tests.GetCmdOutput(t, jfrogCli, "completion", "bash")
	assert.NoError(t, err)
	assert.Equal(t, bash.BashAutocomplete, string(content))
}

func TestZshCompletionSetup(t *testing.T) {
	jfrogCli := coreTests.NewJfrogCli(execMain, "jfrog", "")
	content, _, err := tests.GetCmdOutput(t, jfrogCli, "completion", "zsh")
	assert.NoError(t, err)
	assert.Equal(t, zsh.ZshAutocomplete, string(content))
}

func TestFishCompletionSetup(t *testing.T) {
	jfrogCli := coreTests.NewJfrogCli(execMain, "jfrog", "")
	content, _, err := tests.GetCmdOutput(t, jfrogCli, "completion", "fish")
	assert.NoError(t, err)
	assert.Contains(t, string(content), "complete -c")
}

func TestBashCompletionRt(t *testing.T) {
	jfrogCli := coreTests.NewJfrogCli(execMain, "jfrog", "")
	content, errContent, err := tests.GetCmdOutput(t, jfrogCli, "rt", "--generate-bash-completion")
	assert.NoError(t, err)
	assert.Contains(t, string(content), "permission-target-create")
	assert.Empty(t, string(errContent))
}

func TestBashCompletionJf(t *testing.T) {
	jfrogCli := coreTests.NewJfrogCli(execMain, "jfrog", "")
	content, errContent, err := tests.GetCmdOutput(t, jfrogCli, "", "--generate-bash-completion")
	assert.NoError(t, err)
	assert.Contains(t, string(content), "rt")
	assert.Empty(t, string(errContent))
}
