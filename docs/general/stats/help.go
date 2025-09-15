package statsDocs

var Usage = []string{"stats [--ServerId <server-id>] [--Product <product>] [--Output <output>] [--AccessToken <access-token>]"}

func GetDescription() string {
	return `Returns all statistics related to a specific product or of all products for a given server.`
}

func GetArguments() string {
	return `	 
	--ServerId (optional) 
	The server id for which the product will be searched. If not provided, the default configured server value will be used.

	--Product (optional)
	The product name for which you want statistics Default value is all products. Currently, supported products are: Artifactory(rt), JPDs, repositories, projects, release-bundle(rb)
	Only abbrevation needs to be given, i.e., rt, jpd, pj, rb

	--Output (optional)
	The output format in which you want statistics to be shown. Currently, Json, Table and Console Text (default) are supported.

	--accessToken(optional)
	The access token using which you want statistics will be fetched from jfrog instance. By default, logged user access token is used. For some products, like JPDs, projects, user needs to provide admin token.

`
}
