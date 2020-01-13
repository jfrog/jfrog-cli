package licensedeploy

const Description string = "Deploy a license from the specified bucket to an existing JPD. You may also deploy a number of licenses to an Artifactory HA."

var Usage = []string{"jfrog mc ld [command options] <bucket id> <jpd id>"}

const Arguments string = `	Bucket ID
		Bucket name or identifier to deploy licenses from.

	JPD ID
		An existing JPD's ID.`
