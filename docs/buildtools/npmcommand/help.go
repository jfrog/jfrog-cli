package npmcommand

var Usage = []string{"npm <go arguments> [command options]"}

func GetDescription() string {
	return "Run npm command."
}

func GetArguments() string {
	return `	npm commands
		Arguments and options for the npm command.`
}
