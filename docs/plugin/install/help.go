package install

var Usage = []string{"plugin install <plugin name and version>"}

const EnvVar string = `	JFROG_CLI_PLUGINS_SERVER
		[Default: Official JFrog CLI Plugins registry]
		Configured Artifactory server ID from which to download JFrog CLI Plugins.

	JFROG_CLI_PLUGINS_REPO
		[Default: 'jfrog-cli-plugins']
		Can be optionally used with the JFROG_CLI_PLUGINS_SERVER environment variable.
		Determines the name of the local repository to use.`

func GetDescription() string {
	return "Install or upgrade a JFrog CLI plugin."
}

func GetArguments() string {
	return `	plugin name and version
		Specifies the name and version of the JFrog CLI Plugin you wish to install or upgrade from the plugins registry.
		The version should be specified after a '@' separator, such as: 'hello-frog@1.0.0'. 
		To download the latest version, specify the plugin name only.`
}
