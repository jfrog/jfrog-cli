package dotnet

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description = "Run .NET Core CLI"

var Usage = []string{cliutils.CliExecutableName + " dotnet <dotnet sub-command> [command options]"}

const Arguments string = `	dotnet sub-command
		 Arguments and options for the dotnet command.`
