package uvcommand

var Usage = []string{"uv <uv args> [command options]"}

func GetDescription() string {
	return "Run uv command"
}

func GetArguments() string {
	return `	uv sub-command
		Arguments and options for the uv command.`
}
