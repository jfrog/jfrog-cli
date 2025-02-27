package common

import (
	"strings"

	"github.com/jfrog/jfrog-client-go/artifactory/services"

	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
)

const (
	JfrogCliLogLevel = `	JFROG_CLI_LOG_LEVEL
		[Default: INFO]
		This variable determines the log level of the JFrog CLI.
		Possible values are: DEBUG, INFO, WARN and ERROR.
		If set to ERROR, JFrog CLI logs error messages only.
		It is useful when you wish to read or parse the JFrog CLI output and do not want any other information logged.`

	JfrogCliLogTimestamp = `	JFROG_CLI_LOG_TIMESTAMP
		[Default: TIME]
		Controls the log messages timestamp format.
		Possible values are: TIME, DATE_AND_TIME, and OFF.`

	JfrogCliHomeDir = `	JFROG_CLI_HOME_DIR
		[Default: ~/.jfrog]
		Defines the JFrog CLI home directory path.`

	JfrogCliTempDir = `	JFROG_CLI_TEMP_DIR
		[Default: The operating system's temp directory]
		Defines the temp directory used by JFrog CLI.`

	JfrogCliBuildName = `	JFROG_CLI_BUILD_NAME
		Build name to be used by commands which expect a build name, unless sent as a command argument or option.`

	JfrogCliBuildNumber = `	JFROG_CLI_BUILD_NUMBER
		Build number to be used by commands which expect a build number, unless sent as a command argument or option.`

	JfrogCliBuildProject = `	JFROG_CLI_BUILD_PROJECT
		Artifactory project key.`

	JfrogCliServerID = `	JFROG_CLI_SERVER_ID
		Server ID configured using the config command.`

	Ci = `	CI
		[Default: false]
		If true, disables interactive prompts and progress bar.`

	JfrogCliPluginsServer = `	JFROG_CLI_PLUGINS_SERVER
		[Default: Official JFrog CLI Plugins registry]
		Configured Artifactory server ID from which to download JFrog CLI Plugins.`

	JfrogCliPluginsRepo = `	JFROG_CLI_PLUGINS_REPO
		[Default: 'jfrog-cli-plugins']
		Can be optionally used with the JFROG_CLI_PLUGINS_SERVER environment variable.
		Determines the name of the local repository to use.`

	JfrogCliTransitiveDownload = `	JFROG_CLI_TRANSITIVE_DOWNLOAD
		[Default: false]
		Set this option to true to include remote repositories in artifact searches when using the 'rt download' command. 
		The search will target the first five remote repositories within the virtual repository. 
		This feature is available starting from Artifactory version 7.17.0.
		NOTE: Enabling this option may increase the load on Artifactory instances that are proxied by multiple remote repositories. `

	JfrogCliReleasesRepo = `	JFROG_CLI_RELEASES_REPO
		Configured Artifactory repository name from which to download the jar needed by the mvn/gradle command.
		This environment variable's value format should be <server ID configured by the 'jf c add' command>/<repo name>.

		The repository should proxy https://releases.jfrog.io.
		This environment variable is used by the 'jf mvn' and 'jf gradle' commands, and also by the 'jf audit' command, when used for maven or gradle projects.`

	JfrogCliDependenciesDir = `	JFROG_CLI_DEPENDENCIES_DIR
		[Default: $JFROG_CLI_HOME_DIR/dependencies]
		Defines the directory to which JFrog CLI's internal dependencies are downloaded.
		Support by the following commands: maven and gradle`

	JfrogCliMinChecksumDeploySizeKb = `	JFROG_CLI_MIN_CHECKSUM_DEPLOY_SIZE_KB
		[Default: 10]
		Minimum file size in KB for which JFrog CLI performs checksum deploy optimization.
		Supported by the upload command`

	JfrogCliFailNoOp = `	JFROG_CLI_FAIL_NO_OP
		[Default: false]
		Set to true if you'd like the command to return exit code 2 in case of no files are affected.
		Support by the following commands: copy, delete, delete-props, set-props, download, move, search and upload`

	JfrogCliUploadEmptyArchive = `	` + services.JfrogCliUploadEmptyArchiveEnv + `
		[Default: false]
		Set to true if you'd like to upload an empty archive when '--archive' is set but all files were excluded by exclusions pattern.
		Supported by the upload command`

	JfrogCliEncryptionKey = `   	JFROG_CLI_ENCRYPTION_KEY
		If provided, encrypt the sensitive data stored in the config with the provided key. Must be exactly 32 characters.`

	JfrogCliAvoidNewVersionWarning = `   	JFROG_CLI_AVOID_NEW_VERSION_WARNING
		[Default: false]
		Set to true if you'd like to avoid checking the latest available JFrog CLI version and printing warning when it newer than the current one. `

	JfrogCliCommandSummaryOutputDirectory = `    JFROG_CLI_COMMAND_SUMMARY_OUTPUT_DIR
		Defines the directory path where the command summaries data is stored.
		Every command will have its own individual directory within this base directory.`

	JfrogSecurityCliAnalyzerManagerVersion = `    JFROG_CLI_ANALYZER_MANAGER_VERSION
		Specifies the version of Analyzer Manager to be used for security commands, provided in semantic versioning (e.g 1.13.4) format. 
		By default, the latest stable version is used. `
)

var (
	JfrogCliBuildUrl = `	JFROG_CLI_BUILD_URL
		Sets the CI server build URL in the build-info.
		The "` + coreutils.GetCliExecutableName() + ` rt build-publish" command uses the value of this environment variable,
		unless the --build-url command option is sent.`

	JfrogCliEnvExclude = `	JFROG_CLI_ENV_EXCLUDE
		[Default: *password*;*psw*;*secret*;*key*;*token*;*auth*]
		List of case insensitive semicolon-separated(;) patterns in the form of "value1;value2;...".
		Environment variables match those patterns will be excluded.
		This environment variable is used by the "` + coreutils.GetCliExecutableName() + ` rt build-publish" command,
		in case the --env-exclude command option is not sent.`
)

func GetGlobalEnvVars() string {
	return CreateEnvVars(
		JfrogCliLogLevel,
		JfrogCliLogTimestamp,
		JfrogCliHomeDir,
		JfrogCliTempDir,
		JfrogCliBuildName,
		JfrogCliBuildNumber,
		JfrogCliBuildProject,
		JfrogCliServerID,
		Ci,
		JfrogCliPluginsServer,
		JfrogCliPluginsRepo,
		JfrogCliTransitiveDownload,
		JfrogCliReleasesRepo,
		JfrogCliDependenciesDir,
		JfrogCliMinChecksumDeploySizeKb,
		JfrogCliUploadEmptyArchive,
		JfrogCliBuildUrl,
		JfrogCliEnvExclude,
		JfrogCliFailNoOp,
		JfrogCliEncryptionKey,
		JfrogCliAvoidNewVersionWarning,
		JfrogCliCommandSummaryOutputDirectory,
		JfrogSecurityCliAnalyzerManagerVersion)
}

func CreateEnvVars(envVars ...string) string {
	var s []string
	s = append(s, envVars...)
	return strings.Join(s, "\n\n")
}
