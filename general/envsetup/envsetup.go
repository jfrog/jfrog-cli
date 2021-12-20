package envsetup

import (
	"fmt"

	"github.com/jfrog/jfrog-cli-core/v2/general/envsetup"
	"github.com/jfrog/jfrog-cli/utils/progressbar"
)

const (
	registrationPageURL = "https://jfrog.com/start-free/cli/"
)

func RunEnvSetupCmd() error {
	fmt.Println("Thank you for installing JFrog CLI!")
	setupCmd := envsetup.NewEnvSetupCommand(registrationPageURL)
	return progressbar.ExecWithProgress(setupCmd)
}
