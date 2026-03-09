package terraformdocs

var Usage = []string{"terraform <terraform arguments> [command options]"}

func GetDescription() string {
	return "Run Terraform"
}

func GetArguments() string {
	return `	terraform commands
		Arguments and options for the terraform command.`
}
