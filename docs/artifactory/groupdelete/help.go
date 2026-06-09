package groupdelete

var Usage = []string{"rt gdel <group name>"}

func GetDescription() string {
	return "Delete a user group"
}

func GetArguments() string {
	return `	group name
		Group name to be deleted.`
}

func GetAIDescription() string {
	return `Delete a user group from Artifactory. Users in the group are not deleted; they simply lose this group membership.

When to use:
- Removing an obsolete team group during cleanup.

Prerequisites:
- A configured Artifactory server.
- Admin privileges.

Common patterns:
  $ jf rt gdel old-team

Gotchas:
- No undo; recreate via 'jf rt gc' and re-add users with 'jf rt gau' if needed.
- Deleting a group does not delete permission targets that reference it; those will lose the group as a principal silently.

Related: jf rt gc, jf rt gau, jf rt ptu`
}
