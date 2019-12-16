package gradle

const Description = "Run Gradle build."

var Usage = []string{`jfrog rt gradle <tasks and options> [command options]`}

const Arguments string = `	tasks and options
		Tasks and options to run with gradle command. For example, -b path/to/build.gradle.`

const EnvVar string = `	JFROG_CLI_JCENTER_REMOTE_SERVER
		Configured Artifactory server ID from which to download the jar needed by the gradle command.
		The Artifactory server should include a remote maven repository named jcenter, which proxies jcenter.

	JFROG_CLI_JCENTER_REMOTE_REPO
		[Default: jcenter]
		Can be optionally used with the JFROG_CLI_JCENTER_REMOTE_SERVER environment variable.
		Determines the name of the remote repository to use.

	JFROG_CLI_DEPENDENCIES_DIR
		[Default: $JFROG_CLI_HOME_DIR/dependencies]
		Defines the directory to which JFrog CLI's internal dependencies are downloaded.`
