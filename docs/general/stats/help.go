package statsDocs

var Usage = []string{"stats [--Product <product>] [--Output <output>] [--AccessToken <access-token>]"}

func GetDescription() string {
	return `Returns all statistics related to a specific product or of all products.`
}

func GetArguments() string {
	return `	 

     --Product (optional)
      The product name for which you want statistics Default value is all products. Currently, supported products are: Artifactory(rt), Xray(xr), JPDs, repositories, projects, release-bundle(rb)

     --Output (optional)
      The output format in which you want statistics to be shown. Currently, Json, Table and Console Text (default) are supported.

	 --accessToken(optional)
		The access token using which you want statistics will be fetched from jfrog instance. By default, logged user access token is used. For some products, like JPDs, projects, user needs to provide admin token.

`
}
