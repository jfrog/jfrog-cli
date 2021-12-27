package terraformpublish

var Usage = []string{"rt tp"}

func GetDescription() string {
	return "Publish terraform modules to Artifactory"
}

func GetArguments() string {
	return `	project version
		Package version to be published.`
}
