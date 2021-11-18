package deleteprops

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description = "Delete properties on existing files in Artifactory."

var Usage = []string{cliutils.CliExecutableName + " rt delp [command options] <artifacts pattern> <artifact properties>",
	cliutils.CliExecutableName + " rt delp <artifact properties> --spec=<File Spec path> [command options]"}

const Arguments string = `	artifacts pattern
		Properties of artifacts that match this pattern will be removed.

	artifact properties
		The list of properties, in the form of key1,key2,..., to be removed from the matching artifacts.`
