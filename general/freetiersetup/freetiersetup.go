package freetiersetup

import (
	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	"github.com/jfrog/jfrog-cli-core/v2/general/freetiersetup"
)

const (
	registrationPageURL = "https://google.com"
)

func RunFreeTierSetupCmd() error {

	setupCmd := freetiersetup.NewFreeTierSetupCommand(registrationPageURL)
	return commands.ExecWithProgress(setupCmd)
}
