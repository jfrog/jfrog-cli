package edit

var Usage = []string{"config edit <server ID>"}

func GetDescription() string {
	return `Edits a server configuration.`
}

func GetAIDescription() string {
	return `Update an existing server configuration in place. The server ID is required and must already exist; this command does not create new entries.

When to use:
- Rotating credentials (new access token, new password).
- Changing the platform URL after a migration.
- Toggling --basic-auth-only or updating client cert paths.

Prerequisites:
- A previously added server ID (see 'jf c show' to list them).

Common patterns:
  $ jf c edit my-server --access-token=eyJ... --interactive=false
  $ jf c edit my-server --user=newadmin --password=newsecret --interactive=false

Gotchas:
- Errors if the server ID does not exist; use 'jf c add' for new entries.
- Interactive mode is on by default. Use --interactive=false for scripts.
- Only the fields you pass are updated; omitted fields keep their previous values.

Related: jf c add, jf c show, jf c use, jf c rm`
}
