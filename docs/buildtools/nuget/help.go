package nuget

var Usage = []string{"nuget <nuget args> [command options]"}

func GetDescription() string {
	return "Run NuGet."
}

func GetArguments() string {
	return `	nuget command
		The nuget command to run. For example, restore.`
}
