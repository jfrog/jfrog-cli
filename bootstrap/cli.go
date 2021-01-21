package bootstrap

import (
	"github.com/codegangsta/cli"
	corecommon "github.com/jfrog/jfrog-cli-core/docs/common"
	"github.com/jfrog/jfrog-cli/bootstrap/commands"
	"github.com/jfrog/jfrog-cli/docs/bootstrap/vcs"
	"github.com/jfrog/jfrog-cli/docs/common"
)

func GetCommands() []cli.Command {
	return []cli.Command{
		{
			Name:         "vcs",
			Usage:        vcs.Description,
			HelpName:     corecommon.CreateUsage("bootstrap vcs", vcs.Description, vcs.Usage),
			UsageText:    vcs.Arguments,
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return commands.VcsCmd(c)
			},
		},
	}
}
