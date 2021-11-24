package freetiersetup

import (
	"github.com/jfrog/jfrog-cli-core/v2/general/freetiersetup"
	"github.com/jfrog/jfrog-cli/utils/progressbar"
)

const (
	registrationPageURL = "https://jfrog.info/start-free/cli/"
)

func RunFreeTierSetupCmd() error {

	setupCmd := freetiersetup.NewFreeTierSetupCommand(registrationPageURL)
	return progressbar.ExecWithProgress(setupCmd)
}
