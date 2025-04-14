package token

var Usage = []string{"eot  <oidc-provider-name> <oidc-token-id> [--platformUrl <platformUrl>] [--oidc-audience <audience>] [--oidc-provider-type <type>] [--application-key <key>] [--Project <project>] [--repository <repository>]"}

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
