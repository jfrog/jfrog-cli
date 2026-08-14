package publish

import "github.com/jfrog/jfrog-cli/docs/common"

var Usage = []string{"plugin publish <plugin name> <plugin version>"}

var EnvVar = []string{common.JfrogCliPluginsServer, common.JfrogCliPluginsRepo}

func GetDescription() string {
	return "Publish a JFrog CLI plugin"
}

func GetArguments() string {
	return `	plugin name
		Specifies the name of the JFrog CLI Plugin you wish to publish. You should run this command from the plugin's directory.

	plugin version
		Specifies the version of the JFrog CLI Plugin you wish to publish.`
}

func GetAIDescription() string {
	return `Build and upload the current directory's JFrog CLI plugin to a plugins registry. Run from the plugin's source directory; the command compiles for multiple OS/arch targets and uploads the resulting binaries.

When to use:
- Releasing a new version of a custom plugin to a private or public registry.

Prerequisites:
- A Go toolchain installed locally.
- A configured plugins registry: JFROG_CLI_PLUGINS_SERVER (server ID) and JFROG_CLI_PLUGINS_REPO (target repo key).
- Write access to the target repo.

Common patterns:
  $ jf plugin publish my-plugin 1.2.3

Gotchas:
- Must be run from the plugin's repository root (where main.go lives).
- The version string is used as the artifact path; reusing an existing version may overwrite or fail depending on repo policy.
- Cross-compilation requires CGO-free builds; native deps may fail to cross-compile.

Related: jf plugin install, jf plugin uninstall`
}
