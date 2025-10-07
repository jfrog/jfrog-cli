package statsDocs

var Usage = []string{"stats <product-name> [--server-id <server-id>] [--format <format>] [--access-token <access-token>]",
"jf stats rt",
"jf stats rt --format json",
"jf stats rt --server-id <JFROG_SERVER_ID>"}

func GetDescription() string {
	return `Returns all statistics related to a specific product for a given server.`
}

func GetArguments() string {
	return `	 
	Product (Mandatory)
	The Product name for which you want to display statistics for now, only artifactory(rt) is supported.

    --server-id (optional)
	The server id for which the product will be searched. If not provided, the default configured server value will be used.

	--format (optional)
	The output format in which you want statistics to be shown. Currently, Json, Table and Text (default) are supported.

	--access-token(optional)
	The access token using which you want statistics will be fetched from jfrog instance. By default, logged user access token is used. For some products, like JPDs, projects, user needs to provide admin token.
`
}
