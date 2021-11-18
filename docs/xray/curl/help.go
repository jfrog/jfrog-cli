package curl

var Usage = []string{"xr curl [command options] <curl command>"}

func GetDescription() string {
	return "Execute a cUrl command, using the configured Xray details."
}

func GetArguments() string {
	return `	curl command
		cUrl command to run.`
}
