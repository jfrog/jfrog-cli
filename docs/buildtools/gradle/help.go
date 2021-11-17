package gradle

const Description = "Run Gradle build."

var Usage = []string{`jfrog gradle <tasks and options> [command options]`}

const Arguments string = `	tasks and options
		Tasks and options to run with gradle command. For example, -b path/to/build.gradle.`

const EnvVar string = `	JFROG_CLI_EXTRACTORS_REMOTE
		Configured Artifactory server ID and repository name from which to download the jar needed by the gradle command.
		This environemt variable’s value format should be <server ID>/<repo name>. The repository should proxy https://oss.jfrog.org/artifactory/oss-release-local.

	JFROG_CLI_DEPENDENCIES_DIR
		[Default: $JFROG_CLI_HOME_DIR/dependencies]
		Defines the directory to which JFrog CLI's internal dependencies are downloaded.`
