package login

import (
	coreLogin "github.com/jfrog/jfrog-cli-core/v2/general/login"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/urfave/cli"
)

const disableTokenRefreshFlag = "disable-token-refresh"

func LoginCmd(c *cli.Context) error {
	if c.NArg() > 0 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}
	loginCmd := coreLogin.NewLoginCommand().SetServerId(c.String("server-id"))
	if c.IsSet(disableTokenRefreshFlag) {
		disableTokenRefresh := c.Bool(disableTokenRefreshFlag)
		loginCmd.SetDisableTokenRefresh(&disableTokenRefresh)
	}
	return loginCmd.Run()
}
