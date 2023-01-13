package pipelines

import (
	corecommon "github.com/jfrog/jfrog-cli-core/v2/docs/common"
	"github.com/jfrog/jfrog-cli/docs/common"
	"github.com/jfrog/jfrog-cli/docs/pipelines/status"
	"github.com/jfrog/jfrog-cli/docs/pipelines/sync"
	"github.com/jfrog/jfrog-cli/docs/pipelines/syncstatus"
	"github.com/jfrog/jfrog-cli/docs/pipelines/trigger"
	"github.com/jfrog/jfrog-cli/docs/pipelines/version"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/urfave/cli"
)

func GetCommands() []cli.Command {
	return cliutils.GetSortedCommands(cli.CommandsByName{
		{
			Name:         "status",
			Flags:        cliutils.GetCommandFlags(cliutils.Status),
			Aliases:      []string{"s"},
			Usage:        status.GetDescription(),
			HelpName:     corecommon.CreateUsage("pl status", status.GetDescription(), status.Usage),
			UsageText:    status.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return fetchLatestPipelineRunStatus(c)
			},
		},
		{
			Name:         "trigger",
			Flags:        cliutils.GetCommandFlags(cliutils.Trigger),
			Aliases:      []string{"t"},
			Usage:        trigger.GetDescription(),
			HelpName:     corecommon.CreateUsage("pl trigger", trigger.GetDescription(), trigger.Usage),
			UsageText:    trigger.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return triggerNewRun(c)
			},
		},
		{
			Name:         "version",
			Flags:        cliutils.GetCommandFlags(cliutils.Version),
			Aliases:      []string{"v"},
			Usage:        version.GetDescription(),
			HelpName:     corecommon.CreateUsage("pl version", version.GetDescription(), version.Usage),
			UsageText:    version.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return getVersion(c)
			},
		},
		{
			Name:         "sync",
			Flags:        cliutils.GetCommandFlags(cliutils.Sync),
			Aliases:      []string{"sy"},
			Usage:        sync.GetDescription(),
			HelpName:     corecommon.CreateUsage("pl sync", sync.GetDescription(), sync.Usage),
			UsageText:    sync.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return syncPipelineResources(c)
			},
		},
		{
			Name:         "syncstatus",
			Flags:        cliutils.GetCommandFlags(cliutils.SyncStatus),
			Aliases:      []string{"ss"},
			Usage:        syncstatus.GetDescription(),
			HelpName:     corecommon.CreateUsage("pl syncstatus", syncstatus.GetDescription(), syncstatus.Usage),
			UsageText:    syncstatus.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return getSyncPipelineResourcesStatus(c)
			},
		},
	})
}
