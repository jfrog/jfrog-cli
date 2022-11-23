package pipelines

import (
	corecommon "github.com/jfrog/jfrog-cli-core/v2/docs/common"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/urfave/cli"
)

func GetCommands() []cli.Command {
	return cliutils.GetSortedCommands(cli.CommandsByName{
		{
			Name:         cliutils.Status,
			Flags:        cliutils.GetCommandFlags(cliutils.Status),
			Aliases:      []string{"s"},
			Description:  "gets status of latest run of pipeline",
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return fetchLatestPipelineRunStatus(c, c.String("branch"))
			},
		},
		{
			Name:         "trigger",
			Flags:        cliutils.GetCommandFlags(cliutils.Trigger),
			Aliases:      []string{"t"},
			Description:  "trigger a run for pipeline",
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return triggerNewRun(c)
			},
		},
		{
			Name:         "version",
			Flags:        cliutils.GetCommandFlags("version"),
			Aliases:      []string{"v"},
			Description:  "get pipeline version on server",
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return getVersion(c)
			},
		},
	})
}
