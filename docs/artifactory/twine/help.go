package twinedocs

var Usage = []string{"twine <twine arguments> [command options]"}

func GetDescription() string {
	return "Runs twine "
}

func GetArguments() string {
	return `	twine commands
		Arguments and options for the twine command.`
}
