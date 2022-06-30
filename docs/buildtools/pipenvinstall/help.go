package pipenvinstall

var Usage = []string{"pipenv <pipenv arguments> [command options]"}

func GetDescription() string {
	return "Run pipenv install."
}

func GetArguments() string {
	return `	pipenv sub-command
		Arguments and options for the pipenv command.`
}
