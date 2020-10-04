package install

const Description = "Install a JFrog CLI plugin."

var Usage = []string{"jfrog plugin install <plugin name and version>"}

const Arguments string = `	plugin name and version
		Specifies the name and version of the JFrog CLI Plugin you wish to install from the plugins registry.
		The version should be specified after a '@' separator, such as: 'hello-frog@1.0.0'. 
		To download the latest version, specify the plugin name only.`
