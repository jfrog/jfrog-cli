package gopublish

var Usage = []string{"rt gp [command options] <project version>"}

func GetDescription() string {
	return "Publish go package and/or its dependencies to Artifactory."
}

func GetArguments() string {
	return `	project version
		Package version to be published.`
}
