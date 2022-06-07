package gradle

var Usage = []string{"gradle <tasks and options> [command options]"}

const EnvVar string = `	JFROG_CLI_EXTRACTORS_REMOTE
		Configured Artifactory server ID and repository name from which to download the jar needed by the gradle command.
		This environment variable's value format should be <server ID>/<repo name>. The repository should proxy https://releases.jfrog.io/artifactory/oss-release-local.

	JFROG_CLI_DEPENDENCIES_DIR
		[Default: $JFROG_CLI_HOME_DIR/dependencies]
		Defines the directory to which JFrog CLI's internal dependencies are downloaded.`

func GetDescription() string {
	return "Run Gradle build."
}

func GetArguments() string {
	return `	tasks and options
		Tasks and options to run with gradle command. For example, -b path/to/build.gradle.`
}
