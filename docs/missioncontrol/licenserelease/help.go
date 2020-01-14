package licenserelease

const Description string = "Release a license from a JPD and return it to the specified bucket."

var Usage = []string{"jfrog mc lr [command options] <bucket id> <jpd id>"}

const Arguments string = `	Bucket ID
		Bucket name or identifier to release license to.

	JPD ID
		If the license is used by a JPD, pass the JPD's ID. If the license was only acquired but is not used, pass the name it was acquired with.`
