package accesstoken

import (
	coreConfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/urfave/cli"
)

// If the username is provided as an argument, then it is used when creating the token.
// If not, then the configured username (or the value of the --user option) is used.
func GetSubjectUsername(c *cli.Context, serverDetails *coreConfig.ServerDetails) string {
	if c.NArg() > 0 {
		return c.Args().Get(0)
	}
	return serverDetails.GetUser()
}
