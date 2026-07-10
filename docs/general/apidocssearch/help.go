package apidocssearch

var Usage = []string{"api docs search <query> [--tag <product>] [--method <http-method>] [--limit <n>] [--format table|json]"}

func GetDescription() string {
	return "Search the OpenAPI operations embedded in this jf binary for a keyword, and print ranked matches with a ready-to-run 'jf api' command for each. Local and offline: no server configuration or network call is involved."
}

func GetArguments() string {
	return `	query
		Keyword to search for. Matched (case-insensitive, contains or fuzzy) against each operation's path, summary, tags, and operationId.

EXAMPLES
  # Find operations related to users
  $ jf api docs search user

  # Narrow to a specific product/tag
  $ jf api docs search token --tag Users

  # Narrow to a specific HTTP method
  $ jf api docs search worker --method DELETE

  # Cap the number of results
  $ jf api docs search repository --limit 3

  # Human-readable table instead of the default JSON
  $ jf api docs search user --format table

OUTPUT
  JSON by default (this command exists primarily for agent consumption); pass --format table for a human-readable table instead. Each match includes the operation's method, path, summary, tags, a relevance score, and a "jf_api" field with a ready-to-run 'jf api' invocation for that operation. When the operation takes path/query parameters, they're listed under "parameters" (required ones marked). When it takes a JSON request body, its top-level fields (name, type, required, description, default) are listed under "request_body", and "jf_api" already includes a minimal -d '{...}' skeleton covering just the required fields — fill in real values before running it. Table view shows this as compact PARAMS/BODY columns ("*" marks a required field). An empty result set still reports which spec bundle was searched (spec_bundle) — a "stub" bundle may simply be missing the operation. Exits 0 even when no matches are found.`
}

func GetAIDescription() string {
	return `Search the OpenAPI operations embedded in this jf binary and return ranked matches with a ready-to-run 'jf api' command for each — use this before guessing at a 'jf api <path>' invocation.

When to use:
- You know roughly what you want to do (e.g. "list users", "delete a worker") but not the exact REST path/method.
- Before calling 'jf api <path>' with a path you're not fully sure exists.
- Before calling a POST/PUT/PATCH endpoint whose payload shape you don't already know.

Prerequisites: none. This command is fully local/offline — no server configuration, credentials, or network call.

Common patterns:
  $ jf api docs search user
  $ jf api docs search token --tag Users --method GET
  $ jf api docs search repository --limit 3 --format json

Gotchas:
- The embedded spec bundle may be a small "stub" subset in this build, not the full JFrog REST API surface. An empty match list includes spec_bundle so you know whether that's the likely cause.
- Output is JSON by default (unconditionally, unlike most other jf commands' --ai-help-gated JSON defaults); pass --format table for a human-readable table instead.
- Filters (--tag, --method) are hard excludes, applied before ranking/scoring.
- A query with no contains-match anywhere falls back to fuzzy (typo-tolerant) matching, gated by a similarity floor to avoid coincidental false positives (e.g. "evidence" vs "environments"). Advanced: override the floor (0-1, default 0.6) with $JFROG_CLI_API_DOCS_SEARCH_FUZZY_MIN.
- A match's "jf_api" one-liner only fills in required request-body fields with type-appropriate placeholders (e.g. "" for string, false for boolean) -- inspect the full "request_body"/"parameters" fields for optional ones, descriptions, and defaults before running it for real.
- A request body property that is itself a nested object is reported by its type name (e.g. "PermissionResource") or "object" rather than being recursively flattened -- only top-level fields are listed.

Related: jf api, jf api --ai-help`
}
