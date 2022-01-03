package licenserelease

var Usage = []string{"mc lr [command options] <bucket id> <jpd id>"}

func GetDescription() string {
	return "Release a license from a JPD and return it to the specified bucket."
}

func GetArguments() string {
	return `	Bucket ID
		Bucket name or identifier to release license to.

	JPD ID
		If the license is used by a JPD, pass the JPD's ID. If the license was only acquired but is not used, pass the name it was acquired with.`
}
