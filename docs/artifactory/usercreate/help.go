package usercreate

var Usage = []string{"rt user-create <username> <password> <email>"}

func GetDescription() string {
	return "Create a new user"
}

func GetAIDescription() string {
	return `Create a single internal user on the configured Artifactory server. Useful for ad-hoc provisioning; for bulk creation use 'jf rt uc' with a CSV file.

When to use:
- One-off user creation from a script.

Prerequisites:
- A configured Artifactory server.
- Admin privileges.

Common patterns:
  $ jf rt user-create alice 'S3cret!' alice@example.com
  $ jf rt user-create alice 'S3cret!' alice@example.com --admin=true
  $ jf rt user-create alice 'S3cret!' alice@example.com --groups=readers,writers

Gotchas:
- Password is passed on the command line; consider 'jf rt uc' with a CSV or 'jf api' for safer handling.
- Email format is validated server-side; malformed addresses produce a 400.
- For external auth (LDAP/SAML), the user is created with the password but auth may still defer to the IdP.

Related: jf rt uc, jf rt udel, jf rt gau, jf rt gc`
}
