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
	"github.com/jfrog/jfrog-cli/docs/pipelines/workspacedelete"
	"github.com/jfrog/jfrog-cli/docs/pipelines/workspacelist"
	"github.com/jfrog/jfrog-cli/docs/pipelines/workspacerun"
	"github.com/jfrog/jfrog-cli/docs/pipelines/workspacerunstatus"
	"github.com/jfrog/jfrog-cli/docs/pipelines/workspacesync"
	"github.com/jfrog/jfrog-cli/docs/pipelines/workspacesyncstatus"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
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
		{
			Name:         "validate",
			Flags:        cliutils.GetCommandFlags(cliutils.Validate),
			Aliases:      []string{"va"},
			Description:  "validate pipeline resources",
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action:       validatePipelineResources,
		},
		{
			Name:         "workspace-run",
			Flags:        cliutils.GetCommandFlags(cliutils.WorkspaceRun),
			Aliases:      []string{"wr"},
			Usage:        workspacerun.GetDescription(),
			HelpName:     corecommon.CreateUsage("pl workspace-run", workspacerun.GetDescription(), workspacerun.Usage),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action:       workspacePipelineRun,
		},
		{
			Name:         "workspace-run-status",
			Flags:        cliutils.GetCommandFlags(cliutils.WorkspaceRunStatus),
			Aliases:      []string{"wrs"},
			Usage:        workspacerunstatus.GetDescription(),
			HelpName:     corecommon.CreateUsage("pl workspace-run-status", workspacerunstatus.GetDescription(), workspacerunstatus.Usage),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action:       getWorkspacePipelinesRunStatus,
		},
		{
			Name:         "workspace-list",
			Flags:        cliutils.GetCommandFlags(cliutils.WorkspaceList),
			Aliases:      []string{"wl"},
			Usage:        workspacelist.GetDescription(),
			HelpName:     corecommon.CreateUsage("pl workspace-run", workspacelist.GetDescription(), workspacelist.Usage),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action:       listWorkspaces,
		},
		{
			Name:         "workspace-sync",
			Flags:        cliutils.GetCommandFlags(cliutils.WorkspaceSync),
			Aliases:      []string{"ws"},
			Usage:        workspacesync.GetDescription(),
			HelpName:     corecommon.CreateUsage("pl workspace-run", workspacesync.GetDescription(), workspacesync.Usage),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action:       doWorkspaceSync,
		},
		{
			Name:         "workspace-sync-status",
			Flags:        cliutils.GetCommandFlags(cliutils.WorkspaceSyncStatus),
			Aliases:      []string{"wss"},
			Usage:        workspacesyncstatus.GetDescription(),
			HelpName:     corecommon.CreateUsage("pl workspace-run", workspacesyncstatus.GetDescription(), workspacesyncstatus.Usage),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action:       getWorkspacePipelinesSyncStatus,
		},
		{
			Name:         "workspace-delete",
			Flags:        cliutils.GetCommandFlags(cliutils.WorkspaceDelete),
			Aliases:      []string{"wd"},
			Usage:        workspacedelete.GetDescription(),
			HelpName:     corecommon.CreateUsage("pl workspace-run", workspacedelete.GetDescription(), workspacedelete.Usage),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action:       deleteWorkspace,
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
	if plDetails.PipelinesUrl == "" {
		return nil, fmt.Errorf("the --pipelines-url option is mandatory")
	}
	return plDetails, nil
}

// fetchLatestPipelineRunStatus fetches pipeline run status and filters from pipeline-name and branch flags
func fetchLatestPipelineRunStatus(c *cli.Context) error {
	log.Info(coreutils.PrintTitle("Fetching pipeline run status"))

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
	log.Info("Triggering pipeline sync on repository:", repository, "branch:", branch)
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
	if branch == "" {
		return cliutils.PrintHelpAndReturnError("The --branch option is mandatory.", c)
	}
	repository := c.String("repository")
	if repository == "" {
		return cliutils.PrintHelpAndReturnError("The --repository option is mandatory.", c)
	}
	log.Info("Fetching pipeline sync status on repository:", repository, "branch:", branch)

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
	log.Info("Triggering on pipeline:", pipelineName, "for branch:", branch)

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

// validatePipelineResources validates pipelines definition files using validate api
func validatePipelineResources(c *cli.Context) error {
	pipelinesDefinitions := c.String("files")
	log.Info(coreutils.PrintTitle("Connecting to JFrog pipelines"))
	relativePathToPipelineDefinitions := c.String("directory")
	// Get service config details
	serviceDetails, err := createPipelinesDetailsByFlags(c)
	if err != nil {
		return err
	}
	validateCommand := pipelines.NewValidateCommand()
	validateCommand.SetServerDetails(serviceDetails).
		SetPipeResourceFiles(pipelinesDefinitions).
		SetDirectoryPath(relativePathToPipelineDefinitions).
		SetServerDetails(serviceDetails)
	return commands.Exec(validateCommand)
}

// workspacePipelineRun uses set of JFrog pipelines workspaces api
func workspacePipelineRun(c *cli.Context) error {
	files := c.String("files")
	projectKey := c.String("project")
	values := c.String("values")
	log.Info(coreutils.PrintTitle("Connecting to JFrog pipelines"))
	// Get service config details
	serviceDetails, err := createPipelinesDetailsByFlags(c)
	if err != nil {
		return err
	}
	workspaceCommand := pipelines.NewWorkspaceCommand()
	workspaceCommand.SetServerDetails(serviceDetails).
		SetPipeResourceFiles(files).
		SetProject(projectKey).
		SetValues(values)
	return commands.Exec(workspaceCommand)
}

// getWorkspaces retrieves all workspaces
func listWorkspaces(c *cli.Context) error {
	log.Info(coreutils.PrintTitle("Connecting to JFrog pipelines"))
	// Get service config details
	serviceDetails, err := createPipelinesDetailsByFlags(c)
	if err != nil {
		return err
	}
	workspaceCommand := pipelines.NewWorkspaceListCommand()
	workspaceCommand.SetServerDetails(serviceDetails)
	return commands.Exec(workspaceCommand)
}

func deleteWorkspace(c *cli.Context) error {
	log.Info(coreutils.PrintTitle("Connecting to JFrog pipelines"))
	projectKey := c.String("project")
	// Get service config details
	serviceDetails, err := createPipelinesDetailsByFlags(c)
	if err != nil {
		return err
	}
	workspaceCommand := pipelines.NewWorkspaceDeleteCommand()
	workspaceCommand.SetServerDetails(serviceDetails).
		SetProject(projectKey)
	return commands.Exec(workspaceCommand)
}

func getWorkspacePipelinesRunStatus(c *cli.Context) error {
	log.Info(coreutils.PrintTitle("Connecting to JFrog pipelines"))
	projectKey := c.String("project")
	log.Info("Operating on project: ", projectKey)
	// Get service config details
	serviceDetails, err := createPipelinesDetailsByFlags(c)
	if err != nil {
		return err
	}
	workspaceCommand := pipelines.NewWorkspaceRunStatusCommand()
	workspaceCommand.SetServerDetails(serviceDetails).
		SetProject(projectKey)
	return commands.Exec(workspaceCommand)
}

func getWorkspacePipelinesSyncStatus(c *cli.Context) error {
	log.Info(coreutils.PrintTitle("Connecting to JFrog pipelines"))
	projectKey := c.String("project")
	// Get service config details
	serviceDetails, err := createPipelinesDetailsByFlags(c)
	if err != nil {
		return err
	}
	workspaceCommand := pipelines.NewWorkspaceSyncStatusCommand()
	workspaceCommand.SetServerDetails(serviceDetails).
		SetProject(projectKey)
	return commands.Exec(workspaceCommand)
}

func doWorkspaceSync(c *cli.Context) error {
	log.Info(coreutils.PrintTitle("Connecting to JFrog pipelines"))
	projectKey := c.String("project")
	// Get service config details
	serviceDetails, err := createPipelinesDetailsByFlags(c)
	if err != nil {
		return err
	}
	workspaceCommand := pipelines.NewWorkspaceSyncCommand()
	workspaceCommand.SetServerDetails(serviceDetails).
		SetProject(projectKey)
	return commands.Exec(workspaceCommand)
}
