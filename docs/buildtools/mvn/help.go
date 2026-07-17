package mvn

import "github.com/jfrog/jfrog-cli/docs/common"

var Usage = []string{"mvn <goals and options> [command options]"}

var EnvVar = []string{common.JfrogCliReleasesRepo, common.JfrogCliDependenciesDir}

func GetDescription() string {
	return "Run Maven build."
}

func GetArguments() string {
	return `	goals and options
		Goals and options to run with mvn command. For example  -f path/to/pom.xml`
}

func GetAIDescription() string {
	return `Run a Maven build with JFrog instrumentation: resolves dependencies and deploys artifacts through Artifactory, optionally collecting build-info for traceability. Wraps the local mvn binary; goals and flags after the command name are passed through.

When to use:
- Building Maven projects against Artifactory release/snapshot repositories.
- Producing a JFrog build-info record alongside the artifact deployment.

Prerequisites:
- A local mvn binary on PATH (this command does not install Maven).
- 'jf mvn-config' run once in the project directory to set the resolver and deployer repositories.
- A configured server (jf c add) referenced by 'jf mvn-config'.

Common patterns:
  $ jf mvn clean install
  $ jf mvn deploy -f path/to/pom.xml --build-name=my-build --build-number=1
  $ jf mvn package -DskipTests

Gotchas:
- 'jf mvn-config' must be run first; without it the resolution/deployment repos are unknown.
- --build-name and --build-number are required together for build-info collection; passing only one is silently ignored.
- All flags after 'mvn' are passed verbatim to Maven; jf-specific flags (--build-name, etc.) are consumed by jf, the rest go to mvn.
- Set JFROG_CLI_RELEASES_REPO to fetch the Maven extractor through a private repo (air-gapped builds).

Related: jf mvn-config, jf rt build-publish, jf rt build-add-deps`
}
