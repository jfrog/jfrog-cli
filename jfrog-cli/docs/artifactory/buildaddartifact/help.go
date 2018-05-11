package buildaddartifact

const Description = "This command is used to add a priorly uploaded artifact to a build."

var Usage = []string{"jfrog rt baa [command options] <build name> <build number> <artifact>"}

const Arguments string =
`	build name
		Build name.

	build number
		Build number.

	artifact
		Existing artifact to add to build, specified as <repo-key>/<file-path>.`