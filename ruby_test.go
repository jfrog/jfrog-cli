package main

import (
	"os/exec"
	"testing"

	coretests "github.com/jfrog/jfrog-cli-core/v2/utils/tests"
	"github.com/stretchr/testify/assert"
)

// These tests exercise the `jf ruby <gem|bundle>` command wiring end-to-end without
// requiring a live Artifactory: help/version sub-commands bypass auth injection and
// Artifactory calls, so they validate the dispatcher, flag extraction and native exec.

func rubyToolAvailable(t *testing.T, tool string) {
	if _, err := exec.LookPath(tool); err != nil {
		t.Skipf("'%s' is not installed; skipping ruby command test", tool)
	}
}

func TestRubyGemVersionPassthrough(t *testing.T) {
	rubyToolAvailable(t, "gem")
	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	// `jf ruby gem --version` must pass straight through to the gem binary.
	assert.NoError(t, jfrogCli.Exec("ruby", "gem", "--version"))
}

func TestRubyBundleVersionPassthrough(t *testing.T) {
	rubyToolAvailable(t, "bundle")
	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	assert.NoError(t, jfrogCli.Exec("ruby", "bundle", "--version"))
}

func TestRubyGemHelpPassthrough(t *testing.T) {
	rubyToolAvailable(t, "gem")
	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	// `help` sub-command bypasses auth and must not error.
	assert.NoError(t, jfrogCli.Exec("ruby", "gem", "help"))
}

func TestRubyUnsupportedToolErrors(t *testing.T) {
	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	// An unsupported native tool must be rejected by RubyCommand.Run.
	err := jfrogCli.Exec("ruby", "npm", "install")
	assert.Error(t, err)
}

func TestRubyNoArgsErrors(t *testing.T) {
	jfrogCli := coretests.NewJfrogCli(execMain, "jfrog", "")
	// `jf ruby` with no native tool must report a usage error, not panic.
	err := jfrogCli.Exec("ruby")
	assert.Error(t, err)
}
