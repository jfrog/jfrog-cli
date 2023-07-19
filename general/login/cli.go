package login

import (
	coreLogin "github.com/jfrog/jfrog-cli-core/v2/general/login"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/urfave/cli"
)

func LoginCmd(c *cli.Context) error {
	if c.NArg() > 0 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}
	return coreLogin.NewLoginCommand().Run()
}
