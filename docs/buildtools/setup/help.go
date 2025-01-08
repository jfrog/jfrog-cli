package setup

import (
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/setup"
	"strings"
)

var Usage = []string{"setup [command options]",
	"setup <package manager> [command options]"}

func GetDescription() string {
	return "An interactive command to configure your local package manager (e.g., npm, pip) to work with JFrog Artifactory."
}

func GetArguments() string {
	return `	package manager
		The package manager to configure. Supported package managers are: ` + strings.Join(setup.GetSupportedPackageManagersList(), ", ") + "."
}
