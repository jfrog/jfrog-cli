package groupaddusers

var Usage = []string{"rt gau <group name> <users list>"}

func GetDescription() string {
	return "Add a list of users to a group."
}

func GetArguments() string {
	return `	group name
		The name of the group.

	users list
		Specifies the usernames to add to the specified group.
		The list should be comma-separated(,) in the form of user1,user2,...
	`
}

func GetAIDescription() string {
	return `Add one or more existing users to an existing group on the configured Artifactory server.

When to use:
- Onboarding users into a team's group after they have been created.
- Adding service accounts to a release-manager group.

Prerequisites:
- A configured Artifactory server.
- Admin privileges.
- The group and the users must already exist.

Common patterns:
  $ jf rt gau readers alice,bob,carol

Gotchas:
- Users not already created on the server are not added; the command may succeed with a partial set.
- Existing memberships are preserved; this is additive, not a full replacement.

Related: jf rt gc, jf rt gdel, jf rt user-create, jf rt uc

QA:
Q: Could you guide me through the process of adding users to a group in JFrog Artifactory considering that the group name is 'group1' and the users are 'user1,user2,user3'?
A: jf rt gau group1 'user1,user2,user3'

Q: Could you elucidate the approach for inserting users into a group in JFrog Artifactory provided the group name is 'group3' and the users are 'user7,user8,user9'?
A: jf rt gau group3 'user7,user8,user9'

Q: Could you outline the procedure for adding users to a group in JFrog Artifactory with the group name being 'group5' and the users are 'user13,user14,user15'?
A: jf rt gau group5 'user13,user14,user15'
`
}
