package jpddelete

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description string = "Delete a JPD from Mission Control."

var Usage = []string{cliutils.CliExecutableName + " mc jd [command options] <jpd id>"}

const Arguments string = `	JPD ID
		The ID of the JPD to be removed from Mission Control.`
