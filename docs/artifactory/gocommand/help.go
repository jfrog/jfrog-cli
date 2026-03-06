package gocommand

var Usage = []string{"rt go <go arguments> [command options]"}

func GetDescription() string {
	return "Run go"
}

func GetArguments() string {
	return `	go commands
		Arguments and options for the go command.`
}
