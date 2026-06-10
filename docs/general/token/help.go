package token

var Usage = []string{"atc", "atc <username>"}

func GetDescription() string {
	return `Creates an access token. By default, a user-scoped token is created. Administrator may provide the scope explicitly with '--scope', or implicitly with '--groups', '--grant-admin'.`
}

func GetArguments() string {
	return `	username
		The username for which this token is created. If not specified, the token will be created for the current user.`
}

func GetAIDescription() string {
	return `Mint a JFrog Platform access token. Defaults to a user-scoped token for the calling identity. Admins can broaden the scope with --scope (raw scope string) or --groups (comma-separated group names) plus --grant-admin for admin-scoped tokens.

When to use:
- Generating a long-lived service token for CI runners.
- Issuing a short-lived, narrowly scoped token for an agent or automation.
- Refreshing a token before it expires.

Prerequisites:
- A configured server (jf c add or jf login).
- Sufficient privileges to mint the requested scope (admin for cross-user / group scopes).

Common patterns:
  $ jf atc                                       # user-scoped token for current user
  $ jf atc alice --expiry=3600                   # 1-hour token for user alice
  $ jf atc --groups=readers,writers              # group-scoped (admin)
  $ jf atc --grant-admin --expiry=86400          # admin-scoped (24h)

Gotchas:
- --grant-admin requires admin privileges and may be disabled by platform policy.
- --refreshable produces a token that can be renewed; otherwise the token is one-shot.
- --expiry is in seconds. Platform policy may cap the maximum lifetime.
- The plaintext token is printed once; capture it immediately.

Related: jf eot, jf c add, jf login`
}
