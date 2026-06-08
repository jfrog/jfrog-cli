package install

import "github.com/jfrog/jfrog-cli/docs/common"

var Usage = []string{"plugin install <plugin name and version>"}

var EnvVar = []string{common.JfrogCliPluginsServer, common.JfrogCliPluginsRepo}

func GetDescription() string {
	return "Install or upgrade a JFrog CLI plugin."
}

func GetArguments() string {
	return `	plugin name and version
		Specifies the name and version of the JFrog CLI Plugin you wish to install or upgrade from the plugins registry.
		The version should be specified after a '@' separator, such as: 'hello-frog@1.0.0'.
		To download the latest version, specify the plugin name only.`
}

func GetAIDescription() string {
	return `Install or upgrade a JFrog CLI plugin from the configured plugins registry into the local plugins pool (~/.jfrog/plugins/). Plugins extend the jf binary with custom subcommands. By default, the official JFrog plugins server is used; private registries can be set via JFROG_CLI_PLUGINS_SERVER / JFROG_CLI_PLUGINS_REPO.

When to use:
- Adding a JFrog community or custom plugin (for example 'hello-frog').
- Upgrading an installed plugin to a newer version.

Prerequisites:
- Network access to the plugins registry.
- For private registries: JFROG_CLI_PLUGINS_SERVER (server ID) and JFROG_CLI_PLUGINS_REPO (repo key).

Common patterns:
  $ jf plugin install hello-frog            # latest version
  $ jf plugin install hello-frog@1.0.0      # pinned version

Gotchas:
- Installing the same name without @ upgrades to latest, overwriting the previous binary.
- The plugin binary must match the host OS/arch; older plugins may not be available for all platforms.
- Plugins run in-process under the jf binary; they inherit the active server config.

Related: jf plugin uninstall, jf plugin publish`
}
