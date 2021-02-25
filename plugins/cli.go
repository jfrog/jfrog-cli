package plugins

import (
	"github.com/codegangsta/cli"
	corecommon "github.com/jfrog/jfrog-cli-core/docs/common"
	"github.com/jfrog/jfrog-cli/docs/common"
	installdocs "github.com/jfrog/jfrog-cli/docs/plugin/install"
	uninstalldocs "github.com/jfrog/jfrog-cli/docs/plugin/uninstall"
	"github.com/jfrog/jfrog-cli/plugins/commands"
)

func GetCommands() []cli.Command {
	return []cli.Command{
		{
			Name:         "install",
			Aliases:      []string{"i"},
			Description:  installdocs.Description,
			HelpName:     corecommon.CreateUsage("plugin install", installdocs.Description, installdocs.Usage),
			UsageText:    installdocs.Arguments,
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return commands.InstallCmd(c)
			},
		},
		{
			Name:         "uninstall",
			Aliases:      []string{"ui"},
			Description:  uninstalldocs.Description,
			HelpName:     corecommon.CreateUsage("plugin uninstall", uninstalldocs.Description, uninstalldocs.Usage),
			UsageText:    uninstalldocs.Arguments,
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return commands.UninstallCmd(c)
			},
		},
	}
}
