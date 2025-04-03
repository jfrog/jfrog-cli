package token

var Usage = []string{"eot <platformUrl>  <oidc-provider-name> <oidc-token-id> [--oidc-audience <audience>] [--oidc-provider-type <type>] [--application-key <key>] [--Project <project>] [--repository <repository>]"}

func GetDescription() string {
	return `Exchanges a token ID from an OIDC provider with a JFrog server to a valid access token and returns the access token and the username.`
}

func GetArguments() string {
	return `	 --platformUrl (mandatory)   The JFrog platform base URL to which the OIDC token will be exchanged for an access token (e.g., https://mycompany.jfrog.io)."

     --oidc-provider-name (mandatory)
      The provider name.

     --oidc-token-id (mandatory)
      The OIDC token ID to be exchanged for an access token.

`
}
