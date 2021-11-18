package mvn

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description = "Run Maven build."

var Usage = []string{cliutils.CliExecutableName + " mvn <goals and options> [command options]"}

const Arguments string = `	goals and options
		Goals and options to run with mvn command. For example  -f path/to/pom.xml`

const EnvVar string = `	JFROG_CLI_EXTRACTORS_REMOTE
		Configured Artifactory server ID and repository name from which to download the jar needed by the mvn command.
		This environemt variable’s value format should be <server ID>/<repo name>. The repository should proxy https://oss.jfrog.org/artifactory/oss-release-local.

	JFROG_CLI_DEPENDENCIES_DIR
		[Default: $JFROG_CLI_HOME_DIR/dependencies]
		Defines the directory to which JFrog CLI's internal dependencies are downloaded.`
