package install

var Usage = []string{"mcp install --agent <cursor|claude>"}

func GetDescription() string {
	return "Configure an AI agent (cursor or claude) to connect to the remote JFrog MCP server."
}

func GetArguments() string {
	return `	EXAMPLES
  # Configure Cursor for the current project
  $ jf mcp install --agent cursor

  # Configure Claude Code globally (user-level)
  $ jf mcp install --agent claude --global

  # Preview the configuration without writing it
  $ jf mcp install --agent cursor --dry-run

NOTES
  Before writing any configuration, the command verifies that the MCP server is
  reachable. Use --skip-check to bypass this verification.

  Authentication is completed via OAuth in the agent itself; no credentials are
  written to the agent configuration. After installation, complete the OAuth
  authorization as required by the agent (for example, run /mcp in Claude Code,
  or approve the server in Cursor) to activate the connection.`
}
