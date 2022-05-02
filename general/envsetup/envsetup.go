package envsetup

import (
	"fmt"
	"github.com/jfrog/jfrog-cli-core/v2/general/envsetup"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-cli/utils/progressbar"
	"github.com/urfave/cli"
	"github.com/jfrog/jfrog-client-go/utils/log"
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
	log.Output()
	log.Output()
	log.Output(coreutils.PrintTitle("Thank you for installing JFrog CLI! 🐸"))
	log.Output(coreutils.PrintTitle("We'll now set up a FREE JFrog environment in the cloud for you, and configure your local machine to use it."))
	log.Output("Your environment will be ready in less than a minute.")
	setupCmd := envsetup.NewEnvSetupCommand().SetRegistrationURL(registrationPageURL).SetBase64Credentials(base64Credentials)	return progressbar.ExecWithProgress(setupCmd, false)
}
