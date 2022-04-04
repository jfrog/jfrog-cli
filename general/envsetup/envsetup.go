package envsetup

import (
	"fmt"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"

	"github.com/jfrog/jfrog-cli-core/v2/general/envsetup"
	"github.com/jfrog/jfrog-cli/utils/progressbar"
)

const (
	registrationPageURL = "https://jfrog.com/start-free/cli/"
)

func RunEnvSetupCmd() error {
	fmt.Println()
	fmt.Println()
	fmt.Println(coreutils.PrintTitle("Thank you for installing JFrog CLI! 🐸"))
	fmt.Println(coreutils.PrintTitle("We'll now set up a FREE JFrog environment in the cloud for you, and configure your local machine to use it."))
	fmt.Println("Your environment will be ready in less than a minute.")
	setupCmd := envsetup.NewEnvSetupCommand(registrationPageURL)
	return progressbar.ExecWithProgress(setupCmd, false)
}
