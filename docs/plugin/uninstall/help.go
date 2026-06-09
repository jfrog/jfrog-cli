package uninstall

var Usage = []string{"plugin uninstall <plugin name>"}

func GetDescription() string {
	return "Uninstall a JFrog CLI plugin."
}

func GetArguments() string {
	return `	plugin name
		Specifies the name of the JFrog CLI Plugin you wish to uninstall from the local plugins pool.`
}

func GetAIDescription() string {
	return `Remove a previously installed JFrog CLI plugin from the local plugins pool (~/.jfrog/plugins/).

When to use:
- Cleaning up plugins that are no longer needed.
- Removing a broken or incompatible plugin before reinstalling.

Prerequisites:
- The plugin must already be installed locally.

Common patterns:
  $ jf plugin uninstall hello-frog

Gotchas:
- No confirmation prompt; the plugin binary is removed immediately.
- Reinstalling requires another 'jf plugin install' against the registry.

Related: jf plugin install, jf plugin publish`
}
