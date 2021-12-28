package terraformdocs

var Usage = []string{"rt terraform <terraform arguments> [command options]"}

func GetDescription() string {
	return "Runs terraform "
}

func GetArguments() string {
	return `	terraform commands
		Arguments and options for the terraform command.`
}
