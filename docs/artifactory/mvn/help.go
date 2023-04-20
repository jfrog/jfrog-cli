package mvn

import "github.com/jfrog/jfrog-cli/docs/common"

var Usage = []string{"rt mvn <goals and options> [command options]"}

var EnvVar = []string{common.JfrogCliExtractorsRemote, common.JfrogCliDependenciesDir}

func GetDescription() string {
	return "Run Maven build."
}

func GetArguments() string {
	return `	goals and options
		Goals and options to run with mvn command. For example  -f path/to/pom.xml`
}
