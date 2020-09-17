// +build ignore

// This program generates bash and zsh completion scripts.
// It can be invoked by running 'go generate ./completion/shells/...'
package main

import (
	"errors"
	"github.com/jfrog/jfrog-cli-core/utils/coreutils"
	"os"
	"strings"

	"github.com/jfrog/jfrog-cli-core/utils/log"
	"github.com/jfrog/jfrog-cli/completion/shells/bash"
	"github.com/jfrog/jfrog-cli/completion/shells/zsh"
)

func main() {
	log.SetDefaultLogger()
	dir, err := os.Getwd()
	coreutils.ExitOnErr(err)
	if strings.HasSuffix(dir, "bash") {
		writeScript(bash.BashAutocomplete)
	} else if strings.HasSuffix(dir, "zsh") {
		writeScript(zsh.ZshAutocomplete)
	} else {
		coreutils.ExitOnErr(errors.New("Unexpected script to create"))
	}
}

func writeScript(script string) {
	scriptFile, err := os.Create("jfrog")
	coreutils.ExitOnErr(err)
	defer scriptFile.Close()
	err = os.Chmod("jfrog", os.ModePerm)
	coreutils.ExitOnErr(err)
	_, err = scriptFile.WriteString(script)
	coreutils.ExitOnErr(err)
}
