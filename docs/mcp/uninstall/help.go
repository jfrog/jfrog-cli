package uninstall

var Usage = []string{"mcp uninstall --agent <cursor|claude>"}

func GetDescription() string {
	return "Remove the JFrog MCP server entry previously added to an AI agent's configuration."
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
