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
	The access token using which you want statistics will be fetched from jfrog instance. By default, logged user access token is used. For some products, like JFrog Platform Deployments and projects, you need to provide an admin token.
`
}

func GetAIDescription() string {
	return `Display platform statistics for a given JFrog product on the configured server. Currently only the 'rt' (Artifactory) product is supported. Useful for inspecting storage, repository counts, and similar metrics.

When to use:
- Auditing platform usage from a script.
- Capturing structured metrics for dashboards (--format=json).

Prerequisites:
- A configured server (jf c add or jf login) or --server-id, --access-token.
- For some metrics (JPDs, projects), an admin-scoped token.

Common patterns:
  $ jf st rt
  $ jf st rt --format=json
  $ jf st rt --server-id=my-prod --access-token=eyJ...

Gotchas:
- Only 'rt' is accepted as the product argument today.
- Some metrics require admin privileges; a user-scoped token returns partial or empty data.
- Default output is text; pass --format for json or table.

Related: jf api, jf c show`
}
