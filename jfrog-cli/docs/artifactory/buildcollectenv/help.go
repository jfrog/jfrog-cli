package buildcollectenv

const Description = "Collect environment variables. Environment variables can be excluded using the build-publish command."

var Usage = []string{"jfrog rt bce <build name> <build number>"}

const Arguments string = `	build name
		Build name.

	build number
		Build number.`
