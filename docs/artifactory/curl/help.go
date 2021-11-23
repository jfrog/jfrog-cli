package curl

var Usage = []string{"rt curl [command options] <curl command>"}

func GetDescription() string {
	return "Execute a cUrl command, using the configured Artifactory details."
}

func GetArguments() string {
	return `	curl command
		cUrl command to run.`
}
