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
	
	JFROG_CLI_SHOW_VERSION_UPDATE
		[Default: true]
		If true, JFrog CLI shows whether you're running the latest version.
		`
