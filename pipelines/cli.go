package pipelines

import (
	"fmt"
	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	corecommon "github.com/jfrog/jfrog-cli-core/v2/docs/common"
	pipelines "github.com/jfrog/jfrog-cli-core/v2/pipelines/commands"
	coreConfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-cli/docs/common"
	"github.com/jfrog/jfrog-cli/docs/pipelines/status"
	"github.com/jfrog/jfrog-cli/docs/pipelines/sync"
	"github.com/jfrog/jfrog-cli/docs/pipelines/syncstatus"
	"github.com/jfrog/jfrog-cli/docs/pipelines/trigger"
	"github.com/jfrog/jfrog-cli/docs/pipelines/version"

	"github.com/jfrog/jfrog-cli/utils/cliutils"
	clientlog "github.com/jfrog/jfrog-client-go/utils/log"
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
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action:       fetchLatestPipelineRunStatus,
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
			Action:       triggerNewRun,
		},
		{
			Name:         "version",
			Flags:        cliutils.GetCommandFlags(cliutils.Version),
			Aliases:      []string{"v"},
			Usage:        version.GetDescription(),
			HelpName:     corecommon.CreateUsage("pl version", version.GetDescription(), version.Usage),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action:       getVersion,
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
			Action:       syncPipelineResources,
		},
		{
			Name:         "sync-status",
			Flags:        cliutils.GetCommandFlags(cliutils.SyncStatus),
			Aliases:      []string{"ss"},
			Usage:        syncstatus.GetDescription(),
			HelpName:     corecommon.CreateUsage("pl sync-status", syncstatus.GetDescription(), syncstatus.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action:       getSyncPipelineResourcesStatus,
		},
	})
}

// getMultiBranch parses singleBranch flag and computes whether multiBranch is set to true/false
func getMultiBranch(c *cli.Context) bool {
	return !c.Bool("single-branch")
}

// createPipelinesDetailsByFlags creates pipelines configuration details
func createPipelinesDetailsByFlags(c *cli.Context) (*coreConfig.ServerDetails, error) {
	plDetails, err := cliutils.CreateServerDetailsWithConfigOffer(c, true, cliutils.CmdPipelines)
	if err != nil {
		return nil, err
	}
	if plDetails.DistributionUrl == "" {
		return nil, fmt.Errorf("the --pipelines-url option is mandatory")
	}
	return plDetails, nil
}

// fetchLatestPipelineRunStatus fetches pipeline run status and filters from pipeline-name and branch flags
func fetchLatestPipelineRunStatus(c *cli.Context) error {
	clientlog.Info(coreutils.PrintTitle("Fetching pipeline run status"))

	// Read flags for status command
	pipName := c.String("pipeline-name")
	notify := c.Bool("monitor")
	branch := c.String("branch")
	multiBranch := getMultiBranch(c)
	serviceDetails, err := createPipelinesDetailsByFlags(c)
	if err != nil {
		return err
	}
	statusCommand := pipelines.NewStatusCommand()
	statusCommand.SetBranch(branch).
		SetPipeline(pipName).
		SetNotify(notify).
		SetMultiBranch(multiBranch)

	// Set server details
	statusCommand.SetServerDetails(serviceDetails)
	return commands.Exec(statusCommand)
}

// syncPipelineResources sync pipelines resource
func syncPipelineResources(c *cli.Context) error {
	// Get arguments repository name and branch name
	repository := c.Args().Get(0)
	branch := c.Args().Get(1)
	clientlog.Info("Triggering pipeline sync on repository:", repository, "branch:", branch)
	serviceDetails, err := createPipelinesDetailsByFlags(c)
	if err != nil {
		return err
	}

	// Create new sync command and add filters
	syncCommand := pipelines.NewSyncCommand()
	syncCommand.SetBranch(branch)
	syncCommand.SetRepositoryFullName(repository)
	syncCommand.SetServerDetails(serviceDetails)
	return commands.Exec(syncCommand)
}

// getSyncPipelineResourcesStatus fetch sync status for a given repository path and branch name
func getSyncPipelineResourcesStatus(c *cli.Context) error {
	branch := c.String("branch")
	repository := c.String("repository")
	clientlog.Info("Fetching pipeline sync status on repository:", repository, "branch:", branch)

	// Fetch service details for authentication
	serviceDetails, err := createPipelinesDetailsByFlags(c)
	if err != nil {
		return err
	}

	// Create sync status command and add filter params
	syncStatusCommand := pipelines.NewSyncStatusCommand()
	syncStatusCommand.SetBranch(branch)
	syncStatusCommand.SetRepoPath(repository)
	syncStatusCommand.SetServerDetails(serviceDetails)
	return commands.Exec(syncStatusCommand)
}

// getVersion version command handler
func getVersion(c *cli.Context) error {
	serviceDetails, err := createPipelinesDetailsByFlags(c)
	if err != nil {
		return err
	}
	versionCommand := pipelines.NewVersionCommand()
	versionCommand.SetServerDetails(serviceDetails)
	return commands.Exec(versionCommand)
}

// triggerNewRun triggers a new run for supplied flag values
func triggerNewRun(c *cli.Context) error {
	// Read arguments pipeline name and branch to trigger pipeline run
	pipelineName := c.Args().Get(0)
	branch := c.Args().Get(1)
	multiBranch := getMultiBranch(c)
	coreutils.PrintTitle("Triggering pipeline run ")
	clientlog.Info("Triggering on pipeline:", pipelineName, "for branch:", branch)

	// Get service config details
	serviceDetails, err := createPipelinesDetailsByFlags(c)
	if err != nil {
		return err
	}

	// Trigger a pipeline run using branch name and pipeline name
	triggerCommand := pipelines.NewTriggerCommand()
	triggerCommand.SetBranch(branch).
		SetPipelineName(pipelineName).
		SetServerDetails(serviceDetails).
		SetMultiBranch(multiBranch)
	return commands.Exec(triggerCommand)
}
