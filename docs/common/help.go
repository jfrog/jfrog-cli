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
		Build name to use in build related commands. This environment variable will be used in case of omitted --build-name flag and <build name> argument.
	
	JFROG_CLI_BUILD_NUMBER
		Build number to use in build related commands. This environment variable will be used in case of omitted '--build-number' flag and '<build number>' argument.

	JFROG_CLI_BUILD_URL
		Can be used for setting the CI server build URL in the build-info. This environment variable will be used in case of omitted '--build-url' flag.
	
	JFROG_CLI_ENV_EXCLUDE
		[Default: *password*;*secret*;*key*;*token*] 
		List of case insensitive patterns in the form of "value1;value2;...". Environment variables match those patterns will be excluded. This environment variable will be used in case of omitted '--env-exclude' flag.

	CI
		[Default: false]
		If true, disables progress bar on the supporting commands.
		`
