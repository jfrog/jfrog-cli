package uninstall

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description = "Uninstall a JFrog CLI plugin."

var Usage = []string{cliutils.CliExecutableName + " plugin uninstall <plugin name>"}

const Arguments string = `	plugin name
		Specifies the name of the JFrog CLI Plugin you wish to uninstall from the local plugins pool.`
