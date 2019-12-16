package mvn

const Description = "Run Maven build."

var Usage = []string{`jfrog rt mvn <goals and options> [command options]`}

const Arguments string = `	goals and options
		Goals and options to run with mvn command. For example  -f path/to/pom.xml`

const EnvVar string = `	JFROG_CLI_JCENTER_REMOTE_SERVER
		Configured Artifactory server ID from which to download the jar needed by the mvn command.
		The Artifactory server should include a remote maven repository named jcenter, which proxies jcenter.

	JFROG_CLI_JCENTER_REMOTE_REPO
		[Default: jcenter]
		Can be optionally used with the JFROG_CLI_JCENTER_REMOTE_SERVER environment variable.
		Determines the name of the remote repository to use.

	JFROG_CLI_DEPENDENCIES_DIR
		[Default: $JFROG_CLI_HOME_DIR/dependencies]
		Defines the directory to which JFrog CLI's internal dependencies are downloaded.`
