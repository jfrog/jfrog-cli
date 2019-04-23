package entitlements

const Description string = "Manage Entitlements."

var Usage = []string{"jfrog bt ent [command options] <scope>",
	"jfrog bt ent [command options] <action> <action scope>"}

const Arguments string = `	scope
		A scope can be specified as one of the following:
			- subject/repository.
			- subject/repository/package.
			- subject/repository/package/version.

	action
		Action can be one of:
			create: Creates a new entitlement with access as specified in the --access option.
			show: Provides details of the entitlement specified in the --id option.
			update: Updates the entitlement specified in the --id option.
			delete: Deletes the entitlement specified in the --id option.

	action scope
		The scope on which the action is applied as described above.
		In addition, for create, the --access option must be specified.
		For show, update, or delete, the --id option must be specified.`
