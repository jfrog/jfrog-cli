package show

var Usage = []string{"mcp show"}

func GetDescription() string {
	return "Print the remote MCP (Model Context Protocol) server endpoint for the configured JFrog Platform."
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
