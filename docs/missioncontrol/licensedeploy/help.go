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

func GetAIDescription() string {
	return `Push one or more licenses from a Mission Control bucket onto an existing JPD. For Artifactory HA deployments, use --license-count to seat multiple licenses at once.

When to use:
- Activating a freshly registered JPD with a license.
- Adding HA seats by deploying additional licenses to the same JPD.

Prerequisites:
- A configured Mission Control server.
- A bucket with sufficient available licenses.
- The target JPD must already be registered (see 'jf mc ja').

Common patterns:
  $ jf mc ld my-bucket my-jpd
  $ jf mc ld my-bucket my-ha-jpd --license-count=3
  $ jf mc ld my-bucket my-jpd --format=json

Gotchas:
- --license-count must be at least 1; the default is 1 if unset.
- Deploying more licenses than available in the bucket fails fast.

Related: jf mc la, jf mc lr, jf mc ja`
}
