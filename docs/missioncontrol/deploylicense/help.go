package deploylicense

const Description string = "Deploy a license from the specified bucket to an existing JPD. You may also deploy a number of licenses to an Artifactory 5.x HA cluster."

var Usage = []string{"jfrog mc dl [command options] <bucket id> <jpd id>"}

const Arguments string = `	Bucket id
		Bucket name or identifier to deploy licenses from.

	JPD Id
		An existing JPD's Id.`
