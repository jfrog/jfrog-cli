package login

var Usage = []string{"login"}

func GetDescription() string {
	return "Log in to a JFrog Platform via your web browser. Available for Artifactory 7.64.0 and above"
}

func GetAIDescription() string {
	return `Authenticate to a JFrog Platform interactively via your default web browser. Returns an access token and persists a server configuration under ~/.jfrog/. Works for Artifactory 7.64.0+. Headless environments (CI, agents) should prefer 'jf c add' with --access-token instead.

When to use:
- First-time setup on a developer workstation where a browser is available.
- Quickly authenticating without manually managing tokens.

Prerequisites:
- A default browser configured on the host.
- Network access to the platform URL.

Common patterns:
  $ jf login

Gotchas:
- Requires Artifactory 7.64.0 or newer; older targets must use 'jf c add'.
- Does not work in CI/headless environments — no browser to open.
- The flow stores credentials locally under ~/.jfrog/.

Related: jf c add, jf c show, jf eot, jf atc`
}
