package login

var Usage = []string{"login"}

func GetDescription() string {
	return "Log in to a JFrog Platform via your web browser. Available for Artifactory 7.64.0 and above"
}

func GetAIDescription() string {
	return `Authenticate to a JFrog Platform interactively via your default web browser. Returns an access token and persists a server configuration under ~/.jfrog/. Always uses a single JFrog Platform URL (no per-service paths). Works for Artifactory 7.64.0+. Headless environments (CI, agents) should prefer 'jf c add' with --access-token instead.

When to use:
- First-time setup on a developer workstation where a browser is available.
- Quickly authenticating without manually managing tokens.

Prerequisites:
- A default browser configured on the host.
- Network access to the platform URL.

Common patterns:
  $ jf login
  $ jf login --disable-token-refresh
  $ jf login --legacy

Gotchas:
- Requires Artifactory 7.64.0 or newer; older targets must use 'jf c add', or 'jf login --legacy'.
- --legacy is for Artifactory v6.x self-hosted, where Artifactory/Distribution/Xray/Mission Control/Pipelines don't share a single platform URL. It skips the browser-based web login and prompts for each service's URL and standard credentials instead.
- Does not work in CI/headless environments — no browser to open.
- The flow stores credentials locally under ~/.jfrog/.
- --disable-token-refresh persists in the saved server config; omitting the flag on a later 'jf login' re-run leaves the previously saved value untouched.

Related: jf c add, jf c show, jf eot, jf atc`
}
