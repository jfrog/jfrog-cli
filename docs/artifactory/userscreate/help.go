package userscreate

var Usage = []string{"rt uc --csv <users details file path>"}

func GetDescription() string {
	return "Create new users"
}

func GetAIDescription() string {
	return `Bulk-create users from a CSV file. The CSV header must be 'username,password,email[,admin,groups]' and one user per row.

When to use:
- Onboarding many users at once (organization migration, team scale-out).

Prerequisites:
- A configured Artifactory server.
- Admin privileges.
- A CSV with the expected columns.

Common patterns:
  $ jf rt uc --csv=./users.csv
  $ jf rt uc --csv=./users.csv --replace=true

Gotchas:
- The CSV is read line-by-line; one malformed row aborts the batch in default mode.
- --replace=true overwrites existing users with the same username (destructive).
- Passwords sit in plaintext in the CSV; protect the file and delete after use.

Related: jf rt user-create, jf rt udel, jf rt gau

QA:
Q: Could you guide me through the process of creating users in JFrog Artifactory considering that the details for the users are in the present CSV file named 'users.csv'?
A: jf rt uc --csv users.csv

Q: Could you elucidate the approach for creating users in JFrog Artifactory provided the user details exist within a CSV file named 'team.csv'?
A: jf rt uc --csv team.csv

Q: Could you outline the procedure for creating users in JFrog Artifactory with the user details situated within a CSV file labeled 'employees.csv'?
A: jf rt uc --csv employees.csv
`
}
