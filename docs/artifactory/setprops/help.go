package setprops

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description = "Set properties on existing files in Artifactory."

var Usage = []string{cliutils.CliExecutableName + " rt sp [command options] <artifacts pattern> <artifact properties>",
	cliutils.CliExecutableName + " rt sp <artifact properties> --spec=<File Spec path> [command options]"}

const Arguments string = `	artifacts pattern
		Artifacts that match the pattern will be set with the specified properties.

	artifact properties
		The list of properties, in the form of key1=value1;key2=value2,..., to be set on the matching artifacts.`
