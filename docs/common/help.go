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
		Build name to be used by commands which expect a build name, unless sent as a command argument or option.
	
	JFROG_CLI_BUILD_NUMBER
		Build number to be used by commands which expect a build number, unless sent as a command argument or option.

	JFROG_CLI_BUILD_URL
		Sets the CI server build URL in the build-info. The "jfrog rt build-publish" command uses the value of this environment variable, unless the --build-url command option is sent.
	
	JFROG_CLI_ENV_EXCLUDE
		[Default: *password*;*psw*;*secret*;*key*;*token*] 
		List of case insensitive patterns in the form of "value1;value2;...". Environment variables match those patterns will be excluded. This environment variable is used by the "jfrog rt build-publish" command, in case the --env-exclude command option is not sent.

	CI
		[Default: false]
		If true, disables interactive prompts and progress bar.
		`
