package ruby

import "github.com/jfrog/jfrog-cli/docs/common"

var Usage = []string{"ruby <tasks and options> [command options]"}

var EnvVar = []string{common.JfrogCliReleasesRepo, common.JfrogCliDependenciesDir}

func GetDescription() string {
	return "Run Ruby build."
}

func GetArguments() string {
	return `	tasks and options
		Tasks and options to run with Ruby command. For example, -b path/to/GemFile.`
}
