package token

var Usage = []string{"eot  <oidc-provider-name> <oidc-token-id> [--url <url>] [--oidc-audience <audience>] [--oidc-provider-type <type>] [--application-key <key>] [--Project <project>] [--repository <repository>]"}

func GetDescription() string {
	return `Exchanges a token ID from an OIDC provider with a JFrog server to a valid access token and returns the access token and the username.`
}

func GetArguments() string {
	return `

     --oidc-provider-name (mandatory)
      The provider name.

     --oidc-token-id (mandatory)
      The OIDC token ID to be exchanged for an access token.

`
}

func GetAIDescription() string {
	return `Exchange an OIDC ID token from a trusted external provider (GitHub Actions, GitLab CI, generic OIDC) for a JFrog Platform access token. Removes the need to store long-lived JFrog credentials in CI secrets.

When to use:
- Authenticating CI/CD jobs (GitHub Actions, GitLab pipelines) without storing static JFrog tokens.
- Federated identity setups where the OIDC provider issues short-lived tokens.

Prerequisites:
- An OIDC identity mapping configured in the JFrog Platform (Admin > User Management > OIDC).
- An OIDC ID token from the provider (usually injected by the CI environment via ACTIONS_ID_TOKEN_REQUEST_TOKEN or equivalent).
- --url or a configured server.

Common patterns:
  $ jf eot --oidc-provider-name=my-gh-provider --oidc-token-id=$ACTIONS_ID_TOKEN --url=https://mycorp.jfrog.io
  $ jf eot --oidc-provider-name=my-provider --oidc-token-id=$ID_TOKEN --oidc-provider-type=github --application-key=my-app

Gotchas:
- The OIDC provider mapping must exist on the platform first; the exchange fails otherwise.
- The returned access token is short-lived; do not cache it across jobs.
- Some CI environments (GitHub Actions) auto-inject the token id when JFROG_CLI_OIDC_EXCHANGE_TOKEN_ID is set.
- --oidc-audience is required when the provider mapping enforces an audience claim.

Related: jf c add (--oidc-provider), jf atc, jf login`
}
