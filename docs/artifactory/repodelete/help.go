package repodelete

var Usage = []string{"rt rdel <repository pattern>"}

func GetDescription() string {
	return "Permanently delete repositories with all of their content from Artifactory."
}

func GetArguments() string {
	return `	repository pattern
		Specifies the repositories that should be removed. You can use wildcards to specify multiple repositories.`
}
