package envsetup

import (
	"fmt"
	"github.com/jfrog/jfrog-cli-core/v2/general/envsetup"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-cli/utils/progressbar"
	"github.com/urfave/cli"
)

const (
	registrationPageURL = "https://jfrog.com/start-free/cli/"
)

func RunEnvSetupCmd(c *cli.Context) error {
	if c.NArg() > 1 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}
	base64Credentials := ""
	if c.NArg() == 1 {
		base64Credentials = c.Args().Get(0)
	}
	fmt.Println()
	fmt.Println()
	fmt.Println(coreutils.PrintTitle("Thank you for installing JFrog CLI! 🐸"))
	fmt.Println(coreutils.PrintTitle("We'll now set up a FREE JFrog environment in the cloud for you, and configure your local machine to use it."))
	fmt.Println("Your environment will be ready in less than a minute.")
	setupCmd := envsetup.NewEnvSetupCommand().SetRegistrationURL(registrationPageURL).SetBase64Credentials(base64Credentials)
	return progressbar.ExecWithProgress(setupCmd, false)
}
