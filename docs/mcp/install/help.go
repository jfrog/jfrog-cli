package install

var Usage = []string{"mcp install --agent <cursor|claude>"}

func GetDescription() string {
	return "Configure an AI agent (cursor or claude) to connect to the remote JFrog MCP server."
}

func GetAIDescription() string {
	return `Add the remote JFrog MCP server to an AI agent's configuration so the agent can call JFrog Platform tools. Writes the MCP entry for Cursor or Claude Code; authentication is completed via OAuth in the agent itself, so no credentials are written to the agent config.

When to use:
- Onboarding Cursor or Claude Code to the JFrog Platform MCP integration.
- Adding the MCP server at project scope (default) or user scope (--global).

Prerequisites:
- A configured JFrog Platform server (jf c add or jf login), or pass --server-id / --url.
- The target agent (Cursor or Claude Code) installed locally.

Common patterns:
  $ jf mcp install --agent=cursor
  $ jf mcp install --agent=claude --global
  $ jf mcp install --agent=cursor --dry-run

Gotchas:
- Before writing config, the command verifies the MCP server is reachable; pass --skip-check to bypass that probe.
- After installing, complete the OAuth authorization in the agent (e.g. run /mcp in Claude Code, or approve the server in Cursor) to activate the connection.
- Default scope is the current project; --global writes the user-level agent config instead.

Related: jf mcp show, jf mcp uninstall`
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
