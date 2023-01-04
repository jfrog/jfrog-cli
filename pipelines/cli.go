package pipelines

import (
	corecommon "github.com/jfrog/jfrog-cli-core/v2/docs/common"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/urfave/cli"
)

func GetCommands() []cli.Command {
	return cliutils.GetSortedCommands(cli.CommandsByName{
		{
			Name:         "status",
			Flags:        cliutils.GetCommandFlags(cliutils.Status),
			Aliases:      []string{"s"},
			Description:  "Get status of latest run of pipeline",
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return fetchLatestPipelineRunStatus(c)
			},
		},
		{
			Name:         "trigger",
			Flags:        cliutils.GetCommandFlags(cliutils.Trigger),
			Aliases:      []string{"t"},
			Description:  "Trigger a run for given pipeline and branch",
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return triggerNewRun(c)
			},
		},
		{
			Name:         "version",
			Flags:        cliutils.GetCommandFlags(cliutils.Version),
			Aliases:      []string{"v"},
			Description:  "Get pipeline version on server",
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return getVersion(c)
			},
		},
		{
			Name:         "sync",
			Flags:        cliutils.GetCommandFlags(cliutils.Sync),
			Aliases:      []string{"sy"},
			Description:  "Trigger pipeline resource sync",
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return syncPipelineResources(c)
			},
		},
		{
			Name:         "syncstatus",
			Flags:        cliutils.GetCommandFlags(cliutils.SyncStatus),
			Aliases:      []string{"ss"},
			Description:  "Get pipeline resource sync status",
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return getSyncPipelineResourcesStatus(c)
			},
		},
	})
}
