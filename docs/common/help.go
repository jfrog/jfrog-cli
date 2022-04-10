package common

import "github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"

func GetGlobalEnvVars() string {
	return `
	JFROG_CLI_LOG_LEVEL
		[Default: INFO]
		This variable determines the log level of the JFrog CLI.
		Possible values are: INFO, ERROR, and DEBUG.
		If set to ERROR, JFrog CLI logs error messages only.
		It is useful when you wish to read or parse the JFrog CLI output and do not want any other information logged.

	JFROG_CLI_LOG_TIMESTAMP
		[Default: TIME]
		Controls the log messages timestamp format.
		Possible values are: TIME, DATE_AND_TIME, and OFF.

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

	JFROG_CLI_BUILD_PROJECT
		Artifactory project key.

	JFROG_CLI_BUILD_URL
		Sets the CI server build URL in the build-info. The "` + coreutils.GetCliExecutableName() + ` rt build-publish" command uses the value of this environment variable, unless the --build-url command option is sent.
	
	JFROG_CLI_ENV_EXCLUDE
		[Default: *password*;*psw*;*secret*;*key*;*token*;*auth*] 
		List of case insensitive patterns in the form of "value1;value2;...". Environment variables match those patterns will be excluded. This environment variable is used by the "` + coreutils.GetCliExecutableName() + ` rt build-publish" command, in case the --env-exclude command option is not sent.

	CI
		[Default: false]
		If true, disables interactive prompts and progress bar.
	
	JFROG_CLI_PLUGINS_SERVER
		[Default: Official JFrog CLI Plugins registry]
		Configured Artifactory server ID from which to download JFrog CLI Plugins.
	
	JFROG_CLI_PLUGINS_REPO
		[Default: 'jfrog-cli-plugins']
		Can be optionally used with the JFROG_CLI_PLUGINS_SERVER environment variable.
		Determines the name of the local repository to use.
	
	JFROG_CLI_TRANSITIVE_DOWNLOAD_EXPERIMENTAL
		[Default: false]
		Set to true to look for artifacts also in remote repositories when using the 'rt download' command. The search will run on the first five remote repositories within the virtual repository. This feature is experimental and available on Artifactory version 7.17.0 or higher.
	
	JFROG_CLI_EXTRACTORS_REMOTE
		Configured Artifactory server ID and repository name from which to download the jar needed by the mvn/gradle command.
		This environment variable's value format should be <server ID>/<repo name>. The repository should proxy https://releases.jfrog.io/artifactory/oss-release-local.
	
	JFROG_CLI_DEPENDENCIES_DIR
		[Default: $JFROG_CLI_HOME_DIR/dependencies]
		Defines the directory to which JFrog CLI's internal dependencies are downloaded.
	
	JFROG_CLI_MIN_CHECKSUM_DEPLOY_SIZE_KB
		[Default: 10]
		Minimum file size in KB for which JFrog CLI performs checksum deploy optimization.
		`
}
