package add

var Usage = []string{"config add",
	"config add <server ID>"}

func GetDescription() string {
	return `Adds a server configuration.`
}

func GetAIDescription() string {
	return `Add a named JFrog Platform server configuration to the local config store under ~/.jfrog/. Use this once per platform deployment before running commands that need credentials. The server ID is a stable label you reference later via --server-id or as the active default.

When to use:
- First-time setup on a new machine or CI runner before running any jf command that targets the platform.
- Registering an additional environment (for example a staging server alongside production).

Prerequisites:
- A JFrog Platform URL (the platform base URL, not the Artifactory subpath).
- One of: access token, username and password, or interactive prompts.
- Write access to ~/.jfrog/ on the local machine.

Common patterns:
  $ jf c add my-server --url=https://mycorp.jfrog.io --user=admin --password=secret --interactive=false
  $ jf c add my-server --url=https://mycorp.jfrog.io --access-token=eyJ... --interactive=false
  $ jf c add my-server --overwrite

Gotchas:
- The command is interactive by default. Pass --interactive=false in scripts and CI.
- Re-adding an existing server ID fails unless --overwrite is set; use 'jf c edit' to modify in place.
- Server IDs cannot be "delete", "use", "show", or "clear" (reserved names).
- --basic-auth-only is incompatible with --access-token.

Related: jf c edit, jf c show, jf c use, jf login

QA:
Q: What's the JFrog CLI command to add a new server configuration to the JFrog Platform?
A: jf c add

Q: I want to add a new server configuration with the URL 'https://my-jfrog-platform.com' and the username 'my-username'. What's the command for that?
A: jf c add --url=https://my-jfrog-platform.com --user=my-username

Q: What's the command to add a new server configuration with the URL 'https://my-jfrog-platform.com' the username 'my-username' the password 'my-password' and the Artifactory URL 'https://my-artifactory.com'?
A: jf c add --url='https://my-jfrog-platform.com' --user=my-username --password='my-password' --artifactory-url='https://my-artifactory.com'
`
}
