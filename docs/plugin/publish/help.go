package publish

import "github.com/jfrog/jfrog-cli/docs/common"

var Usage = []string{"plugin publish <plugin name> <plugin version>"}

var EnvVar = []string{common.JfrogCliPluginsServer, common.JfrogCliPluginsRepo}

func GetDescription() string {
	return "Publishing a JFrog CLI plugin."
}

func GetArguments() string {
	return `	plugin name
		Specifies the name of the JFrog CLI Plugin you wish to publish. You should run this command from the plugin's directory.

	plugin version
		Specifies the version of the JFrog CLI Plugin you wish to publish.`
}
