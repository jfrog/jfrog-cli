package uninstall

var Usage = []string{"mcp uninstall --agent <cursor|claude>"}

func GetDescription() string {
	return "Remove the JFrog MCP server entry previously added to an AI agent's configuration."
}

func GetAIDescription() string {
	return `Remove the JFrog MCP server entry previously added to an AI agent's configuration (Cursor or Claude Code). Only the JFrog entry is removed; other configured MCP servers are left untouched.

When to use:
- Disconnecting an agent from the JFrog Platform MCP integration.
- Cleaning up before re-installing against a different platform or scope.

Prerequisites:
- The agent (Cursor or Claude Code) with a JFrog MCP entry previously added by 'jf mcp install'.

Common patterns:
  $ jf mcp uninstall --agent=cursor
  $ jf mcp uninstall --agent=claude --global

Gotchas:
- Match the scope used at install time: a --global install is removed with --global; a project install is removed from the project config.
- If the entry was installed under a non-default name, pass --name to target it.

Related: jf mcp install, jf mcp show`
}

func GetArguments() string {
	return `	EXAMPLES
  # Remove the JFrog MCP server from Cursor (current project)
  $ jf mcp uninstall --agent cursor

  # Remove it from Claude Code's global configuration
  $ jf mcp uninstall --agent claude --global

NOTES
  Only the JFrog MCP server entry is removed; other configured MCP servers are
  left untouched. Use --name to target an entry written under a non-default name.`
}
