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

Related: jf c add, jf c show, jf c use, jf c rm

QA:
Q: What's the JFrog CLI command to adjust a server configuration with the ID 'my-server'?
A: jf c edit my-server

Q: I want to modify the server configuration with the ID 'my-server' change the URL to 'https://new-jfrog-platform.com' and change the username to 'new-username'. What's the command for that?
A: jf c edit my-server --url='https://new-jfrog-platform.com' --user='new-username'

Q: What's the command to update the server configuration with the ID 'my-server' change the URL to 'https://new-jfrog-platform.com' change the username to 'new-username' change the password to 'new-password' and change the Artifactory URL to 'https://new-artifactory.com'?
A: jf c edit my-server --url='https://new-jfrog-platform.com' --user='new-username' --password='new-password' --artifactory-url='https://new-artifactory.com'
`
}
