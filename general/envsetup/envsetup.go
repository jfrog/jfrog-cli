package envsetup

import (
	"github.com/jfrog/jfrog-cli-core/v2/general/envsetup"
	"github.com/jfrog/jfrog-cli/utils/progressbar"
)

const (
	registrationPageURL = "https://jfrog.com/start-free/cli/"
)

func RunEnvSetupCmd() error {

	setupCmd := envsetup.NewEnvSetupCommand(registrationPageURL)
	return progressbar.ExecWithProgress(setupCmd)
}
