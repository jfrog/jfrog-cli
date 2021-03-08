package plugins

import (
	"github.com/codegangsta/cli"
	corecommon "github.com/jfrog/jfrog-cli-core/docs/common"
	"github.com/jfrog/jfrog-cli/docs/common"
	installdocs "github.com/jfrog/jfrog-cli/docs/plugin/install"
	publishdocs "github.com/jfrog/jfrog-cli/docs/plugin/publish"
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
			ArgsUsage:    common.CreateEnvVars(installdocs.EnvVar),
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
		{
			Name:         "publish",
			Aliases:      []string{"p"},
			Description:  publishdocs.Description,
			HelpName:     corecommon.CreateUsage("plugin publish", publishdocs.Description, publishdocs.Usage),
			UsageText:    publishdocs.Arguments,
			ArgsUsage:    common.CreateEnvVars(publishdocs.EnvVar),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return commands.PublishCmd(c)
			},
		},
	}
}
