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

func GetAIDescription() string {
	return `Reserve a license from a Mission Control bucket and tag it with a holder name. The license is checked out but not yet deployed to a JPD; use 'jf mc ld' for that.

When to use:
- Pre-allocating a license before standing up a new JPD.
- Reserving a license for a specific deployment workflow.

Prerequisites:
- A configured Mission Control server.
- A bucket containing available licenses.
- Admin privileges on Mission Control.

Common patterns:
  $ jf mc la my-bucket my-new-jpd
  $ jf mc la my-bucket placeholder-1 --format=json

Gotchas:
- The license remains tagged until 'jf mc lr' releases it; abandoned acquisitions leak licenses.
- The 'name' is a free-form label, not necessarily a real JPD ID. Use 'jf mc lr <bucket> <name>' to return it.

Related: jf mc ld, jf mc lr, jf mc ja`
}
