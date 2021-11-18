package buildpromote

var Usage = []string{"rt bpr [command options] <build name> <build number> <target repository>"}

func GetDescription() string {
	return "This command is used to promote build in Artifactory."
}

func GetArguments() string {
	return `	build name
		Build name.

	build number
		Build number.

	target repository
		Build promotion target repository.`
}
