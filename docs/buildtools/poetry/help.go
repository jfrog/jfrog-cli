package poetry

var Usage = []string{"poetry <poetry args> [command options]"}

func GetDescription() string {
	return "Run poetry command"
}

func GetArguments() string {
	return `	poetry sub-command
		Arguments and options for the poetry command.`
}
