package apidocsdescribe

var Usage = []string{"api docs describe <method> <path> [--format table|json]"}

func GetDescription() string {
	return "Return the full trimmed operation view (parameters, request body schema, response codes/descriptions, an example payload when available, and a ready-to-run 'jf api' command) for a single method+path from the OpenAPI operations embedded in this jf binary. Local and offline: no server configuration or network call is involved."
}

func GetArguments() string {
	return `	method
		HTTP method of the operation (GET, POST, PUT, DELETE, ...). Case-insensitive.

	path
		Exact endpoint path as declared in the catalog, e.g. /access/api/v2/users. Templated path segments (e.g. {workerKey}) must be passed literally, exactly as returned by 'jf api docs search'.

EXAMPLES
  # Describe a GET operation
  $ jf api docs describe GET /access/api/v2/users

  # Describe a POST operation, including its request body schema
  $ jf api docs describe POST /access/api/v2/users

  # A path with a templated segment, copied verbatim from 'jf api docs search'
  $ jf api docs describe DELETE /worker/api/v1/workers/{workerKey}

  # Human-readable table instead of the default JSON
  $ jf api docs describe GET /access/api/v2/users --format table

OUTPUT
  JSON by default (this command exists primarily for agent consumption); pass --format table for a human-readable table instead. The result includes the operation's method, path, summary, tags, "parameters" (path/query/header, required ones marked), "request_body" (top-level fields with name/type/required/description/default, plus an "example" payload when the spec declares one), "responses" (status code + description for each declared response), and a "jf_api" field with a ready-to-run 'jf api' invocation. When method+path isn't found in the embedded catalog, the command exits non-zero with an error naming the spec bundle searched and recommending 'jf api docs search' to find the exact method/path.`
}

func GetAIDescription() string {
	return `Return the full detail (parameters, request body schema, response codes, example payload, ready-to-run command) for one exact method+path from the OpenAPI operations embedded in this jf binary. Use this after 'jf api docs search' has narrowed down a candidate operation, to see its full shape before calling it with 'jf api'.

When to use:
- You already have a method+path (e.g. from 'jf api docs search' results) and need to know its parameters, request body shape, or possible response codes before calling it.
- Before calling a POST/PUT/PATCH endpoint whose exact required/optional body fields you don't already know.
- To confirm a path exists in the embedded catalog at all before guessing with 'jf api <path>'.

Prerequisites: none. This command is fully local/offline — no server configuration, credentials, or network call.

Typical flow: 'jf api docs search <query>' → pick a method+path from the results → 'jf api docs describe <method> <path>' → construct and run the real 'jf api' call.

Common patterns:
  $ jf api docs describe GET /access/api/v2/users
  $ jf api docs describe POST /access/api/v2/users --format json
  $ jf api docs describe DELETE /worker/api/v1/workers/{workerKey}

Gotchas:
- The embedded spec bundle may be a small "stub" subset in this build, not the full JFrog REST API surface — an unresolved lookup names spec_bundle so you know whether that's the likely cause. Set $JFROG_CLI_API_DOCS_REQUIRE_FULL_BUNDLE=true to fail fast on stub builds instead.
- Output is JSON by default (unconditionally, unlike most other jf commands' --ai-help-gated JSON defaults); pass --format table for a human-readable table instead.
- path must match the catalog exactly, including any literal {param} placeholders (e.g. "{workerKey}", not a real key) — copy it verbatim from 'jf api docs search' results rather than guessing.
- Not found (wrong method, wrong path, or the stub bundle lacks the operation) is a hard error (non-zero exit), unlike 'jf api docs search', which returns an empty match list with exit 0.
- request_body's "example" field is only present when the underlying spec declares one; its absence doesn't mean the operation has no valid payload — check "properties" either way.
- A request body property that is itself a nested object is reported by its type name (e.g. "PermissionResource") or "object" rather than being recursively flattened — only top-level fields are listed.

Related: jf api docs search, jf api, jf api --ai-help`
}
