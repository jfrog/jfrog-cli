package tokenverify

var Usage = []string{"atv <token>", "atv --token-file=<path>", "atv"}

func GetDescription() string {
	return `Verifies a JFrog access token: checks its format and signature, checks that it has not expired, asks the configured JFrog server whether it is still accepted, and reports its type.`
}

func GetArguments() string {
	return `	token
		[Optional] The access token to verify, or the JSON output of 'jf atc'. If neither this argument nor --token-file is set, the token is read from stdin.`
}

func GetAIDescription() string {
	return `Verify a JFrog Platform access token and report what kind of token it is. The token's signature is checked against the configured server's root certificate, its claims and expiry are validated, and the server's Access readiness endpoint is asked whether it still accepts the token.

When to use:
- Checking whether a token is still valid before using it in CI.
- Finding out what kind of token you were given (identity, admin, groups or scoped).
- Debugging authentication failures caused by an expired, revoked or foreign token.

Prerequisites:
- A server configured with jf config add or jf login. Its credentials are not used.

Common patterns:
  $ jf atv "$TOKEN"                                        # verify a token passed as an argument
  $ jf atc | jf atv                                        # verify a freshly minted token
  $ jf atv --token-file=token.txt                          # read the token from a file
  $ jf atv --token-file=token.txt --format=json | jq .types

Gotchas:
- With no argument and no --token-file, the token is read from stdin. In a terminal, it prompts and waits.
- JSON input is read for its access_token value.
- Only configured servers can be used. There's no --insecure-tls: trust a private CA by adding it to ~/.jfrog/security/certs. http:// servers work, but print a warning.
- Reference tokens can't be verified; verify the JWT access token instead.
- Internal tokens are reported as Internal Token.
- The token is sent to the server only after its signature checks out.
- The exit code is 1 when the token is invalid. The payload is printed even when the token is invalid.

Related: jf atc, jf eot

QA:
Q: How do I check whether a JFrog access token is still valid?
A: jf atv <token>

Q: How do I verify the token that jf atc just created and get JSON output?
A: jf atc | jf atv --format=json
`
}
