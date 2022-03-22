package envsetup

import (
	"fmt"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/urfave/cli"

	"github.com/jfrog/jfrog-cli-core/v2/general/envsetup"
	"github.com/jfrog/jfrog-cli/utils/progressbar"
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
	fmt.Println("Thank you for installing JFrog CLI! 🐸")
	setupCmd := envsetup.NewEnvSetupCommand().SetRegistrationURL(registrationPageURL).SetBase64Credentials(base64Credentials)
	return progressbar.ExecWithProgress(setupCmd, false)
}
