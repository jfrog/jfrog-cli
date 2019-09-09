package common

const GlobalEnvVars string = `	JFROG_CLI_LOG_LEVEL
		[Default: INFO]
		This variable determines the log level of the JFrog CLI.
		Possible values are: INFO, ERROR, and DEBUG.
		If set to ERROR, JFrog CLI logs error messages only.
		It is useful when you wish to read or parse the JFrog CLI output and do not want any other information logged.

	JFROG_CLI_OFFER_CONFIG
		[Default: true]
		If true, JFrog CLI prompts for product server details and saves them in its config file.
		To avoid having automation scripts interrupted, set this value to false, and instead,
		provide product server details using the config command.

	JFROG_CLI_HOME_DIR
		[Default: ~/.jfrog]
		Defines the JFrog CLI home directory path.

	JFROG_CLI_TEMP_DIR
		[Default: The operating system's temp directory]
		Defines the temp directory used by JFrog CLI.

	JFROG_CLI_BUILD_NAME
		Build name to use in build releated commands. This environment variable will be used in case of absence of --build-name flag and <build name> argument.
	
	JFROG_CLI_BUILD_NUMBER
		Build number to use in build releated commands. This environment variable will be used in case of absence of '--build-number' flag and '<build number>' argument.

	CI
		[Default: false]
		If true, disables progress bar on the supporting commands.
		`
