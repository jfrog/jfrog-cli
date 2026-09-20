package mvnw

import "github.com/jfrog/jfrog-cli/docs/common"

var Usage = []string{"mvnw <goals and options> [command options]"}

var EnvVar = []string{common.JfrogCliReleasesRepo, common.JfrogCliDependenciesDir}

func GetDescription() string {
	return "Run a Maven build using the project's Maven Wrapper (mvnw/mvnw.cmd)."
}

func GetArguments() string {
	return `	goals and options
		Goals and options to run with the mvnw command. For example  -f path/to/pom.xml`
}

func GetAIDescription() string {
	return `Run a Maven build through the project's Maven Wrapper (mvnw on Mac/Linux, mvnw.cmd on Windows) with JFrog instrumentation: resolves dependencies and deploys artifacts through Artifactory, optionally collecting build-info for traceability. Behaves like 'jf mvn' but requires a Maven Wrapper; it does not fall back to a system-installed mvn.

When to use:
- The project pins its Maven version via a checked-in Maven Wrapper and CI/local builds must use that exact version instead of whatever 'mvn' resolves to on PATH.

Prerequisites:
- A Maven Wrapper (mvnw/mvnw.cmd plus .mvn/wrapper/maven-wrapper.properties) checked into the project, in the current directory or a parent directory.
- Only meaningful in native (FlexPack) mode; if a Maven config file exists for the project, this command behaves exactly like 'jf mvn'.

Common patterns:
  $ jf mvnw clean install
  $ jf mvnw package -DskipTests

Gotchas:
- If no Maven Wrapper is found upward from the working directory, the command fails rather than silently falling back to a system mvn.
- Unlike 'jf mvn', this command never runs a system-installed mvn when a wrapper is present or required.

Related: jf mvn, jf mvn-config`
}
