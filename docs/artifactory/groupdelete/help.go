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

Related: jf rt gc, jf rt gau, jf rt ptu

QA:
Q: Could you guide me through the process of deleting a group in JFrog Artifactory considering that the group name is 'group1'?
A: jf rt gdel group1

Q: Could you elucidate the approach for eradicating a group in JFrog Artifactory provided the group name is 'group3'?
A: jf rt gdel group3

Q: Could you outline the procedure for discarding a group in JFrog Artifactory with the group name being 'group5'?
A: jf rt gdel group5
`
}
