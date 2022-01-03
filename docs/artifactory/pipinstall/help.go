package pipinstall

var Usage = []string{"rt pipi <pip sub-command>"}

func GetDescription() string {
	return "Run pip install."
}

func GetArguments() string {
	return `	pip sub-command
		Arguments and options for the pip command.`
}
