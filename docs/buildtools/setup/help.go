package setup

import (
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/setup"
	"strings"
)

var Usage = []string{"setup [command options]",
	"setup <package manager> [command options]"}

func GetDescription() string {
	return "An interactive command to set up your local package manager to work with JFrog Artifactory. Supported package managers are: " +
		strings.Join(setup.GetSupportedPackageManagersList(), ", ")
}

func GetArguments() string {
	return `	package manager
		The package manager to set up. Supported package managers are: ` + strings.Join(setup.GetSupportedPackageManagersList(), ", ")
}
