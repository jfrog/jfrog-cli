package gradle

import "github.com/jfrog/jfrog-cli/docs/common"

var Usage = []string{"rt gradle <tasks and options> [command options]"}

var EnvVar = []string{common.JfrogCliExtractorsRemote, common.JfrogCliDependenciesDir}

func GetDescription() string {
	return "Run Gradle build."
}

func GetArguments() string {
	return `	tasks and options
		Tasks and options to run with gradle command. For example, -b path/to/build.gradle.`
}
