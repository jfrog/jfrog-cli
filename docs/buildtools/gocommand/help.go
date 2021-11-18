package gocommand

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description = "Runs go"

var Usage = []string{cliutils.CliExecutableName + " go <go arguments> [command options]"}

const Arguments string = `	go commands
		Arguments and options for the go command.`
