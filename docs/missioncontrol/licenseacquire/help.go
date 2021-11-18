package licenseacquire

var Usage = []string{"mc la [command options] <bucket id> <name>"}

func GetDescription() string {
	return "Acquire a license from the specified bucket and mark it as taken by the provided name."
}

func GetArguments() string {
	return `	Bucket ID
		Bucket name or identifier to acquire license from.

	Name
		A custom name used to mark the license as taken. Can be a JPD ID or a temporary name. If the license does not end up being used by a JPD, this is the name that should be used to release the license.`
}
