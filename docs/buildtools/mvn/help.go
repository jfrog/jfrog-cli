package mvn

import "github.com/jfrog/jfrog-cli/docs/common"

var Usage = []string{"mvn <goals and options> [command options]"}

var EnvVar = []string{common.JfrogCliReleasesRepo, common.JfrogCliDependenciesDir}

func GetDescription() string {
	return "Package manager set me up command."
}

func GetArguments() string {
	return `	goals and options
		Goals and options to run with mvn command. For example  -f path/to/pom.xml`
}
