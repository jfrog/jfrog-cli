package releaselicense

const Description string = "Release a license from a JPD and returns it to the specified bucket."

var Usage = []string{"jfrog mc rl [command options] <bucket id> <jpd id>"}

const Arguments string = `	Bucket id
		Bucket name or identifier to release license to.

	JPD Id
		If the license is used by a JPD, pass the JPD's Id. If the license was only acquired but is not used, pass the name it was acquired with.`
