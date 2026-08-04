package version

var Usage = []string{"pl version"}

func GetDescription() string {
	return "Show the version of JFrog Pipelines."
}

func GetAIDescription() string {
	return `Print the version string reported by the configured JFrog Pipelines server. Useful as a quick connectivity and auth smoke test before running other 'pl' commands.

When to use:
- Verifying the pipelines URL and credentials are correct after configuration changes.
- Capturing the server version for compatibility checks.

Prerequisites:
- A configured server with the pipelines URL set.

Common patterns:
  $ jf pl version

Gotchas:
- Returns the Pipelines service version, NOT the JFrog Platform version.

Related: jf pl status, jf c show`
}
