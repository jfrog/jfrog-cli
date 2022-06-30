package dotnet

var Usage = []string{"rt dotnet <dotnet sub-command> [command options]"}

func GetDescription() string {
	return "Run .NET Core CLI."
}

func GetArguments() string {
	return `	dotnet sub-command
		 Arguments and options for the dotnet command.`
}
