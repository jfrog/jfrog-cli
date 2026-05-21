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

  # Create a user (POST JSON from a file or -d/--data — not both). username, email, and password are required
  # unless internal_password_disabled is true (see Access API / Platform Administration — Create user).
  $ jf api /access/api/v2/users -X POST --input ./user.json -H "Content-Type: application/json"
  $ jf api /access/api/v2/users -X POST -d '{"username":"newuser","email":"newuser@example.com","password":"UseASecurePassword123"}' -H "Content-Type: application/json"

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

func GetAIDescription() string {
	return `Invoke any JFrog Platform REST endpoint using the configured server URL and credentials. Acts as an authenticated curl: paths are relative to the platform base URL, and the response body goes to stdout. Use this when no dedicated jf subcommand exists for the operation you need.

When to use:
- Calling REST APIs that don't have a first-class CLI wrapper (Access user/group APIs, repository CRUD, GraphQL, etc.).
- Scripting platform admin tasks where 'jf rt' or 'jf c' don't cover the operation.
- Debugging API responses with full visibility into status code and body.

Prerequisites:
- A configured server (jf c add or jf login) or explicit --url / --access-token / --server-id.
- The caller's identity must have the privileges the target endpoint requires.

Common patterns:
  $ jf api /access/api/v2/users
  $ jf api /artifactory/api/repositories -X POST -H "Content-Type: application/json" --input ./repo.json
  $ jf api /artifactory/api/repositories/my-repo -X DELETE
  $ jf api /access/api/v2/users -X POST -d '{"username":"newuser","email":"u@example.com","password":"S3cret!"}' -H "Content-Type: application/json"

Gotchas:
- The endpoint path must start with /; the platform base URL is prepended.
- -d/--data and --input are mutually exclusive.
- HTTP status goes to stderr (one line), body to stdout. Non-2xx exits with status 1 but still prints the body.
- Some APIs require trailing slashes or specific Accept headers; check the API reference before scripting.

Related: jf c add, jf rt, jf c show`
}
