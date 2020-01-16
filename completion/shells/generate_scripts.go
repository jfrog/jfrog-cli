// +build ignore

// This program generates bash and zsh completion scripts.
// It can be invoked by running 'go generate ./completion/shells/...'
package main

import (
	"errors"
	"os"
	"strings"

	"github.com/jfrog/jfrog-cli-go/completion/shells/bash"
	"github.com/jfrog/jfrog-cli-go/completion/shells/zsh"
	"github.com/jfrog/jfrog-cli-go/utils/cliutils"
	"github.com/jfrog/jfrog-cli-go/utils/log"
)

func main() {
	log.SetDefaultLogger()
	dir, err := os.Getwd()
	cliutils.ExitOnErr(err)
	if strings.HasSuffix(dir, "bash") {
		writeScript(bash.BashAutocomplete)
	} else if strings.HasSuffix(dir, "zsh") {
		writeScript(zsh.ZshAutocomplete)
	} else {
		cliutils.ExitOnErr(errors.New("Unexpected script to create"))
	}
}

func writeScript(script string) {
	scriptFile, err := os.Create("jfrog")
	cliutils.ExitOnErr(err)
	defer scriptFile.Close()
	err = os.Chmod("jfrog", os.ModePerm)
	cliutils.ExitOnErr(err)
	_, err = scriptFile.WriteString(script)
	cliutils.ExitOnErr(err)
}
