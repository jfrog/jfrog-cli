package pipinstall

var Usage = []string{"pip <pip arguments> [command options]"}

func GetDescription() string {
	return "Run pip install."
}

func GetArguments() string {
	return `	pip sub-command
		Arguments and options for the pip command.`
}
