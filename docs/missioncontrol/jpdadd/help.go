package jpdadd

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description string = "Add a JPD to Mission Control."

var Usage = []string{cliutils.CliExecutableName + " mc ja [command options] <config>"}

const Arguments string = `	Config
		Path to a JSON configuration file containing the JPD details.`
