package jpddelete

var Usage = []string{"mc jd [command options] <jpd id>"}

func GetDescription() string {
	return "Delete a JPD from Mission Control."
}

func GetArguments() string {
	return `	JPD ID
		The ID of the JPD to be removed from Mission Control.`
}

func GetAIDescription() string {
	return `Remove a JFrog Platform Deployment (JPD) registration from Mission Control. The remote platform itself is not affected; only the Mission Control entry is removed.

When to use:
- Decommissioning a JPD that has been retired.
- Cleaning up a stale Mission Control entry.

Prerequisites:
- A configured Mission Control server (jf c add captures the mission control URL).
- Admin privileges on Mission Control.

Common patterns:
  $ jf mc jd my-jpd-id

Gotchas:
- No undo. Re-register with 'jf mc ja' if needed.
- Removing a JPD does NOT release acquired licenses; release them with 'jf mc lr' first if applicable.

Related: jf mc ja, jf mc lr`
}
