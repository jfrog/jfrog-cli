package use

var Usage = []string{"config use <server ID>"}

func GetDescription() string {
	return "Set the active server by its ID."
}

func GetAIDescription() string {
	return `Mark a previously configured server as the active default. Commands that do not specify --server-id will target this server.

When to use:
- Switching between staging and production targets in a shell session.
- Setting the default after 'jf c add' so subsequent commands need no --server-id.

Prerequisites:
- A previously added server ID (see 'jf c show' for the list).

Common patterns:
  $ jf c use my-server

Gotchas:
- Fails silently from the user's perspective if the server ID does not exist; check with 'jf c show' first.
- Setting an active server is per-machine, not per-shell; affects all subsequent jf invocations.

Related: jf c add, jf c show`
}
