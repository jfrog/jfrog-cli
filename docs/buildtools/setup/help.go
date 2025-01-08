package setup

import (
	"fmt"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/setup"
	"strings"
)

var Usage = []string{"setup [command options]",
	"setup <package manager> [command options]"}

func GetDescription() string {
	return fmt.Sprintf(
		`An interactive command to set up your local package manager to work with JFrog Artifactory.
						 Supported package managers are: %v`,
		strings.Join(setup.GetSupportedPackageManagersList(), ", "))
}
