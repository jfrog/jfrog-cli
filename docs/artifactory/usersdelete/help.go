package usersdelete

var Usage = []string{"rt udel <users list>", "rt udel --csv <users details file path>"}

func GetDescription() string {
	return "Delete users."
}

func GetArguments() string {
	return `	users list
		Comma-separated(,) list of usernames to delete in the form of user1,user2,....`
}

func GetAIDescription() string {
	return `Delete one or more users from Artifactory. Provide a comma-separated list inline or via a CSV file.

When to use:
- Offboarding multiple users in one operation.

Prerequisites:
- A configured Artifactory server.
- Admin privileges.

Common patterns:
  $ jf rt udel alice,bob,carol
  $ jf rt udel --csv=./users.csv
  $ jf rt udel alice --quiet

Gotchas:
- No undo. Audit beforehand with 'jf api /artifactory/api/security/users'.
- Deleting a user does NOT revoke long-lived access tokens issued to that user; rotate or revoke those separately.

Related: jf rt user-create, jf rt uc, jf rt gdel

QA:
Q: Could you guide me through the process of deleting users in JFrog Artifactory considering that the usernames are 'user1,user2,user3'?
A: jf rt udel 'user1,user2,user3'

Q: Could you elucidate the approach for removing users in JFrog Artifactory provided the usernames are 'user7,user8,user9'?
A: jf rt udel 'user7,user8,user9'

Q: Could you outline the procedure for discarding users in JFrog Artifactory with the usernames being 'user13,user14,user15'?
A: jf rt udel 'user13,user14,user15'
`
}
