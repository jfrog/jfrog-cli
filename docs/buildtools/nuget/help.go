package nuget

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description = "Run NuGet."

var Usage = []string{cliutils.CliExecutableName + " nuget <nuget args> [command options]"}

const Arguments string = `	nuget command
		The nuget command to run. For example, restore.`
