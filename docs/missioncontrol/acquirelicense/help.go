package acquirelicense

const Description string = "Acquire a license from the specified bucket and mark it as taken by provided name."

var Usage = []string{"jfrog mc al [command options] <bucket id> <name>"}

const Arguments string = `	Bucket id
		Bucket name or identifier to acquire license from.

	Name
		Custom name to mark the license as taken by. Can be a JPD name or a temporary name. If the license does not end up being used by a JPD, this name should be used if a license release is needed.`
