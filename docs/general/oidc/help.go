package token

var Usage = []string{"eot --platformUrl <url> --oidc-token-id <id> --oidc-provider-name <name> [--oidc-audience <audience>] [--oidc-provider-type <type>] [--ApplicationKey <key>] [--Project <project>] [--repository <repository>]"}

func GetDescription() string {
	return `Exchanges a token ID from an OIDC provider with a JFrog server to a valid access token and returns the access token and the username.`
}

func GetArguments() string {
	return `	 --platformUrl (mandatory)
      The URL of the platform where the OIDC token exchange will take place.

     --oidc-token-id (mandatory)
      The ID of the OIDC token to be exchanged.

     --oidc-provider-name (mandatory)
      The provider name.

     --oidc-audience (optional)
      The audience for the OIDC token.

     --oidc-provider-type (optional default: "GitHub")
      The provider type e.g (GitHub, Azure, General OIDC...)

     --ApplicationKey (optional)
      The JFrog application key

     --Project (optional)
      The JFrog project key.

     --repository (optional)
      The source code repository name.
`
}
