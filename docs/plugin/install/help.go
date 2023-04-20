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
