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

Related: jf rt user-create, jf rt uc, jf rt gdel`
}
