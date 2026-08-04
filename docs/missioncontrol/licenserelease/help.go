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

func GetAIDescription() string {
	return `Return a license from a JPD (or acquisition placeholder) back to its Mission Control bucket. Use the JPD ID if the license is deployed, or the acquisition name from 'jf mc la' if it was only reserved.

When to use:
- Decommissioning a JPD and reclaiming its license.
- Cleaning up abandoned 'jf mc la' acquisitions.

Prerequisites:
- A configured Mission Control server.
- Admin privileges on Mission Control.

Common patterns:
  $ jf mc lr my-bucket my-jpd
  $ jf mc lr my-bucket placeholder-1   # release an unused acquisition

Gotchas:
- The second argument is the holder identifier: JPD ID for deployed licenses, acquisition name for unused ones.
- Releasing a license used by a running JPD will invalidate that JPD's licensing.

Related: jf mc la, jf mc ld`
}
