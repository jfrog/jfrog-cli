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

  # Force JSON output (also the default when --ai-help or JFROG_CLI_AI_HELP=true)
  $ jf api docs search user --format json

OUTPUT
  Each match includes the operation's method, path, summary, tags, a relevance score, and a "jf_api" field with the ready-to-run 'jf api' invocation for that operation. An empty result set still reports which spec bundle was searched (spec_bundle) — a "stub" bundle may simply be missing the operation. Exits 0 even when no matches are found.`
}

func GetAIDescription() string {
	return `Search the OpenAPI operations embedded in this jf binary and return ranked matches with a ready-to-run 'jf api' command for each — use this before guessing at a 'jf api <path>' invocation.

When to use:
- You know roughly what you want to do (e.g. "list users", "delete a worker") but not the exact REST path/method.
- Before calling 'jf api <path>' with a path you're not fully sure exists.

Prerequisites: none. This command is fully local/offline — no server configuration, credentials, or network call.

Common patterns:
  $ jf api docs search user
  $ jf api docs search token --tag Users --method GET
  $ jf api docs search repository --limit 3 --format json

Gotchas:
- The embedded spec bundle may be a small "stub" subset in this build, not the full JFrog REST API surface. An empty match list includes spec_bundle so you know whether that's the likely cause.
- Output is JSON by default when --ai-help/$JFROG_CLI_AI_HELP=true is set, table otherwise; --format overrides either way.
- Filters (--tag, --method) are hard excludes, applied before ranking/scoring.
- A query with no contains-match anywhere falls back to fuzzy (typo-tolerant) matching, gated by a similarity floor to avoid coincidental false positives (e.g. "evidence" vs "environments"). Advanced: override the floor (0-1, default 0.6) with $JFROG_CLI_API_DOCS_SEARCH_FUZZY_MIN.

Related: jf api, jf api --ai-help`
}
