package licenseacquire

const Description string = "Acquire a license from the specified bucket and mark it as taken by the provided name."

var Usage = []string{"jfrog mc la [command options] <bucket id> <name>"}

const Arguments string = `	Bucket ID
		Bucket name or identifier to acquire license from.

	Name
		A custom name used to mark the license as taken. Can be a JPD ID or a temporary name. If the license does not end up being used by a JPD, this is the name that should be used to release the license.`
