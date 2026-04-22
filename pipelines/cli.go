package pipelines

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	coreformat "github.com/jfrog/jfrog-cli-core/v2/common/format"
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
	clientUtils "github.com/jfrog/jfrog-client-go/utils"
	clientlog "github.com/jfrog/jfrog-client-go/utils/log"
	plservices "github.com/jfrog/jfrog-client-go/pipelines/services"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
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
	if plDetails.PipelinesUrl == "" {
		return nil, errors.New("no JFrog Pipelines URL specified as part of the server configuration")
	}
	return plDetails, nil
}

// fetchLatestPipelineRunStatus fetches pipeline run status and filters from pipeline-name and branch flags
func fetchLatestPipelineRunStatus(c *cli.Context) error {
	clientlog.Info(coreutils.PrintTitle("Fetching pipeline run status"))

	pipName := c.String("pipeline-name")
	notify := c.Bool("monitor")
	branch := c.String("branch")
	multiBranch := getMultiBranch(c)
	serviceDetails, err := createPipelinesDetailsByFlags(c)
	if err != nil {
		return err
	}

	outputFormat, err := getPipelineStatusOutputFormat(c)
	if err != nil {
		return err
	}

	statusCommand := pipelines.NewStatusCommand()
	statusCommand.SetBranch(branch).
		SetPipeline(pipName).
		SetNotify(notify).
		SetMultiBranch(multiBranch).
		SetServerDetails(serviceDetails)

	if c.IsSet(cliutils.Format) {
		statusCommand.SetSuppressOutput(true)
	}

	if err = commands.Exec(statusCommand); err != nil {
		return err
	}

	if !c.IsSet(cliutils.Format) {
		return nil
	}
	return printPipelineStatusResponse(statusCommand.Response(), outputFormat, os.Stdout)
}

// getPipelineStatusOutputFormat defaults to table.
func getPipelineStatusOutputFormat(c *cli.Context) (coreformat.OutputFormat, error) {
	if !c.IsSet(cliutils.Format) {
		return coreformat.Table, nil
	}
	return coreformat.GetOutputFormat(c.String(cliutils.Format))
}

func printPipelineStatusResponse(resp *plservices.PipelineRunStatusResponse, outputFormat coreformat.OutputFormat, w io.Writer) error {
	switch outputFormat {
	case coreformat.Json:
		data, err := json.Marshal(resp)
		if err != nil {
			return errorutils.CheckErrorf("failed to marshal pipeline status response: %s", err.Error())
		}
		clientlog.Output(clientUtils.IndentJson(data))
		return nil
	case coreformat.Table:
		return printPipelineStatusTable(resp, w)
	default:
		return errorutils.CheckErrorf("unsupported format '%s' for pl status. Accepted values: table, json", outputFormat)
	}
}

func printPipelineStatusTable(resp *plservices.PipelineRunStatusResponse, w io.Writer) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "PIPELINE\tBRANCH\tRUN\tSTATUS\tDURATION")
	for _, pipe := range resp.Pipelines {
		if pipe.LatestRunID == 0 {
			continue
		}
		_, _ = fmt.Fprintf(tw, "%s\t%s\t%d\t%d\t%ds\n",
			pipe.Name,
			pipe.PipelineSourceBranch,
			pipe.Run.RunNumber,
			pipe.Run.StatusCode,
			pipe.Run.DurationSeconds,
		)
	}
	return tw.Flush()
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
	if branch == "" {
		return cliutils.PrintHelpAndReturnError("The --branch option is mandatory.", c)
	}
	repository := c.String("repository")
	if repository == "" {
		return cliutils.PrintHelpAndReturnError("The --repository option is mandatory.", c)
	}
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
	pipelineName := c.Args().Get(0)
	branch := c.Args().Get(1)
	multiBranch := getMultiBranch(c)
	coreutils.PrintTitle("Triggering pipeline run ")
	clientlog.Info("Triggering on pipeline:", pipelineName, "for branch:", branch)

	serviceDetails, err := createPipelinesDetailsByFlags(c)
	if err != nil {
		return err
	}

	triggerCommand := pipelines.NewTriggerCommand()
	triggerCommand.SetBranch(branch).
		SetPipelineName(pipelineName).
		SetServerDetails(serviceDetails).
		SetMultiBranch(multiBranch)

	if err = commands.Exec(triggerCommand); err != nil {
		return err
	}

	// error == nil guarantees the server responded with 200.
	// The client layer discards the body, so we pass nil and let the helper
	// synthesize {"status_code": 200, "message": "OK"}.
	if c.IsSet(cliutils.Format) {
		outputFormat, fmtErr := coreformat.GetOutputFormat(c.String(cliutils.Format))
		if fmtErr != nil {
			return fmtErr
		}
		if outputFormat == coreformat.Json {
			cliutils.FormatHTTPResponseJSON(nil, 200)
		}
	}
	return nil
}
