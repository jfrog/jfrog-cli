package api

import "github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"

var Usage = []string{"api <endpoint-path>"}

func GetDescription() string {
	return "Invoke a JFrog Platform HTTP API using the configured server URL and credentials (hostname and token are not passed manually; use 'jf config' or --url / --access-token / --server-id as usual). REST API reference: " + coreutils.JFrogHelpUrl + "jfrog-platform-documentation/rest-apis (OpenAPI bundles are not shipped with the CLI)."
}

func GetArguments() string {
	return `	endpoint-path
		API path on the platform host (for example: /access/api/v1/users or /artifactory/api/repositories). The configured platform base URL is prepended automatically.

EXAMPLES
  # List users (Access API)
  $ jf api /access/api/v2/users

  # Get a user by name
  $ jf api /access/api/v2/users/admin

  # Create a user (POST JSON body from a file, or use -d/--data for an inline body — not both)
  $ jf api /access/api/v2/users -X POST --input ./user.json -H "Content-Type: application/json" 
  $ jf api /access/api/v2/users -X POST -d '{"name":"admin"}' -H "Content-Type: application/json" 

  # Replace a user
  $ jf api /access/api/v2/users/newuser -X PUT --input ./user.json -H "Content-Type: application/json"

  # Delete a user
  $ jf api -X DELETE /access/api/v2/users/tempuser

  # List local repositories (Artifactory REST)
  $ jf api /artifactory/api/repositories

  # Create a local Maven repository
  $ jf api /artifactory/api/repositories/my-maven-local -X PUT -H "Content-Type: application/json" --input ./repo-maven-local.json

  # Update repository configuration
  $ jf api /artifactory/api/repositories/libs-release -X POST -H "Content-Type: application/json" --input ./repo-config.json

  # Delete a repository
  $ jf api /artifactory/api/repositories/old-repo -X DELETE

  # One Model GraphQL
  $ jf api /onemodel/api/v1/graphql -X POST -H "Content-Type: application/json" --input ./graphql-query.json

  # Set a request timeout (seconds)
  $ jf api /artifactory/api/repositories --timeout 10

OUTPUT
  The response body is written to standard output. The HTTP status code is written to standard error as a single line. Non-2xx responses still print the body and exit with status 1.

REFERENCES
   Binary Management (Artifactory):  https://docs.jfrog.com/artifactory/reference/
   JFrog Security                 :  https://docs.jfrog.com/security/reference/
   Governance (AppTrust)          :  https://docs.jfrog.com/governance/reference/
   Integrations                   :  https://docs.jfrog.com/integrations/reference/
   Project Management             :  https://docs.jfrog.com/projects/reference/
   Platform Administration        :  https://docs.jfrog.com/administration/reference/
SEE ALSO
  Use 'jf config' to add or select a server. Use --server-id to target a specific configuration.`
}
