package builddistribute

const Description = "This command is used to deploy builds from Artifactory to Bintray, and creates an entry in the corresponding Artifactory distribution repository specified."

var Usage = []string{"jfrog rt bd [command options] <build name> <build number> <target distribution repository>"}

const Arguments string =
`	build name
		Build name.

	build number
		Build number.

	target distribution repository
		Build distribution target repository.`