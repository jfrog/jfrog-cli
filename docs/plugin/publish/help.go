package publish

const Description = "Publishing a JFrog CLI plugin."

var Usage = []string{"jfrog plugin publish <plugin name> <plugin version>"}

const Arguments string = `	plugin name
		Specifies the name of the JFrog CLI Plugin you wish to publish. You should run this command from the plugin's directory.

	plugin version
		Specifies the version of the JFrog CLI Plugin you wish to publish.`

const EnvVar string = `	JFROG_CLI_PLUGINS_SERVER
		[Mandatory]
		Configured Artifactory server ID to publish the JFrog CLI Plugin to.
		The Artifactory server should include a local repository corresponding to the JFROG_CLI_PLUGINS_REPO environment variable.

	JFROG_CLI_PLUGINS_REPO
		[Default: 'jfrog-cli-plugins']
		Can be optionally used with the JFROG_CLI_PLUGINS_SERVER environment variable.
		Determines the name of the local repository to use.`
