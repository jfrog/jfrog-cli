package gocommand

var Usage = []string{"rt go <go arguments> [command options]"}

func GetDescription() string {
	return "Runs go."
}

func GetArguments() string {
	return `	go commands
		Arguments and options for the go command.`
}
