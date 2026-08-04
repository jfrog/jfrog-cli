package remove

var Usage = []string{"config rm",
	"config rm <server ID>"}

func GetDescription() string {
	return `Removes the stored configuration. In case this argument is followed by a configured server ID, then only this server's configurations is removed.`
}

func GetArguments() string {
	return `	server ID
		A unique ID for an existing JFrog configuration.`
}

func GetAIDescription() string {
	return `Delete a stored server configuration from ~/.jfrog/. With a server ID, removes only that entry. Without one, asks to clear ALL configurations (destructive).

When to use:
- Decommissioning a server profile that is no longer needed.
- Cleaning up CI runner state at job end.
- Wiping all credentials before reinstalling.

Prerequisites:
- A previously added server ID (omit to clear all).

Common patterns:
  $ jf c rm my-server
  $ jf c rm my-server --quiet
  $ jf c rm --quiet

Gotchas:
- Without a server ID, the command prompts to clear ALL configurations. Use --quiet to skip the prompt (still destructive).
- No undo; re-add via 'jf c add' if removed by mistake.
- Removing the active server leaves no default until you run 'jf c use'.

Related: jf c add, jf c show, jf c use`
}
