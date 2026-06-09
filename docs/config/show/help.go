package show

var Usage = []string{"config show <server ID>"}

func GetDescription() string {
	return `Shows the stored configuration. In case this argument is followed by a configured server ID, then only this server's configurations is shown.`
}

func GetAIDescription() string {
	return `Print stored server configurations from ~/.jfrog/. With a server ID, prints only that entry; without one, prints all. Sensitive fields (password, access token, refresh token, SSH passphrase) are masked as "***" in both table and JSON output.

When to use:
- Listing configured server IDs before picking one with --server-id.
- Confirming which server is the default after 'jf c use'.
- Inspecting URLs and which auth method is configured for a given environment.

Prerequisites:
- At least one server added via 'jf c add' or 'jf login'.

Common patterns:
  $ jf c show
  $ jf c show my-server
  $ jf c show --format=json
  $ jf c show my-server --format=table

Gotchas:
- Without --format, output uses the legacy format (different from table/json).
- Credentials are always masked; this command cannot dump real secrets.
- Returns nothing silently if no servers are configured.

Related: jf c add, jf c edit, jf c use, jf c export`
}
