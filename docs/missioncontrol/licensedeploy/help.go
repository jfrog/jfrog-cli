package licensedeploy

var Usage = []string{"mc ld [command options] <bucket id> <jpd id>"}

func GetDescription() string {
	return "Deploy a license from the specified bucket to an existing JPD. You may also deploy a number of licenses to an Artifactory HA."
}

func GetArguments() string {
	return `	Bucket ID
		Bucket name or identifier to deploy licenses from.

	JPD ID
		An existing JPD's ID.`
}
