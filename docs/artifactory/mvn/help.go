package mvn

var Usage = []string{"rt mvn <goals and options> [command options]"}

const EnvVar string = `	JFROG_CLI_EXTRACTORS_REMOTE
		Configured Artifactory server ID and repository name from which to download the jar needed by the mvn command.
		This environment variable's value format should be <server ID>/<repo name>. The repository should proxy https://releases.jfrog.io/artifactory/oss-release-local.

	JFROG_CLI_DEPENDENCIES_DIR
		[Default: $JFROG_CLI_HOME_DIR/dependencies]
		Defines the directory to which JFrog CLI's internal dependencies are downloaded.`

func GetDescription() string {
	return "Run Maven build."
}

func GetArguments() string {
	return `	goals and options
		Goals and options to run with mvn command. For example  -f path/to/pom.xml`
}
