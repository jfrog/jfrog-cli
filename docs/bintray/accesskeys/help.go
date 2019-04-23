package accesskeys

const Description string = "Manage Access Keys."

var Usage = []string{"jfrog bt acc-keys [command options]",
	"jfrog bt acc-keys [command options] <action> <key ID>"}

const Arguments string = `	none
		If no arguments are provided, the command provides a list of all access keys active for the subject.

	action
		Action can be one of:
			create: Creates a new access key
			show: Provides details of the entitlement specified in the --id option.
			update: Updates the entitlement specified in the --id option.
			delete: Deletes the entitlement specified in the --id option.

	key ID
		The ID of the key on which the action should be performed.`
