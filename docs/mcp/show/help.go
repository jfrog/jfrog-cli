package show

var Usage = []string{"mcp show"}

func GetDescription() string {
	return "Print the remote MCP (Model Context Protocol) server endpoint for the configured JFrog Platform."
}

func GetAIDescription() string {
	return `Print the remote MCP (Model Context Protocol) server endpoint for the configured JFrog Platform, derived as <platform-url>/mcp. Use this to discover the URL an AI agent should connect to, or to confirm which platform the MCP integration points at.

When to use:
- Looking up the MCP endpoint to configure an agent manually (rather than via 'jf mcp install').
- Verifying which platform / server the MCP integration resolves to before installing.

Prerequisites:
- A configured JFrog Platform server (jf c add or jf login), or pass --server-id / --url.

Common patterns:
  $ jf mcp show
  $ jf mcp show --server-id=my-platform --format=json

Gotchas:
- The endpoint defaults to <platform-url>/mcp; override it with --mcp-url or the JFROG_CLI_MCP_URL environment variable.
- This is the remote, platform-hosted MCP server. It is unrelated to 'jf source-mcp', which runs a local MCP server for source-code analysis.

Related: jf mcp install, jf mcp uninstall`
}

func GetArguments() string {
	return `	EXAMPLES
  # Print the MCP endpoint for the default server
  $ jf mcp show

  # Print the MCP endpoint for a specific server, as JSON
  $ jf mcp show --server-id my-platform --format json

NOTES
  The endpoint is derived as <platform-url>/mcp. It can be overridden with the
  --mcp-url flag or the JFROG_CLI_MCP_URL environment variable.

  This command surfaces the remote, platform-hosted MCP server. It is unrelated
  to 'jf source-mcp', which runs a local MCP server for source-code analysis.`
}
