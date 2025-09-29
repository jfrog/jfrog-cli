package statsDocs

var Usage = []string{"stats <product-name> [--ServerId <server-id>] [--Format <format>] [--AccessToken <access-token>]"}

func GetDescription() string {
	return `Returns all statistics related to a specific product for a given server.`
}

func GetArguments() string {
	return `	 
	Product (Mandatory)
	The Product name for which you want to display statistics for now, only artifactory(rt) is supported.

	--ServerId (optional) 
	The server id for which the product will be searched. If not provided, the default configured server value will be used.

	--Format (optional)
	The output format in which you want statistics to be shown. Currently, Json, Table and Text (default) are supported.

	--accessToken(optional)
	The access token using which you want statistics will be fetched from jfrog instance. By default, logged user access token is used. For some products, like JPDs, projects, user needs to provide admin token.
`
}
