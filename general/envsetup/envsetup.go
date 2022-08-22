package envsetup

import (
	"github.com/jfrog/jfrog-cli-core/v2/general/envsetup"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-cli/utils/progressbar"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/urfave/cli"
)

const (
	registrationPageURL = "https://jfrog.com/start-free/cli/"
)

func RunEnvSetupCmd(c *cli.Context, outputFormat envsetup.OutputFormat) error {
	base64Credentials := ""
	if outputFormat == envsetup.Human {
		if c.NArg() > 1 {
			return cliutils.WrongNumberOfArgumentsHandler(c)
		}
		if c.NArg() == 1 {
			base64Credentials = c.Args().Get(0)
		} else {
			// Setup new user
			log.Output(coreutils.PrintTitle("We'll now set up a FREE JFrog environment in the cloud for you, and configure your local machine to use it."))
			log.Output("Your environment will be ready in less than a minute.")
		}
	}
	setupCmd := envsetup.NewEnvSetupCommand().SetRegistrationURL(registrationPageURL).SetEncodedConnectionDetails(base64Credentials).SetOutputFormat(outputFormat)
	return progressbar.ExecWithProgress(setupCmd)
}
