package pipelines

import (
	"errors"
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
	"github.com/jfrog/jfrog-cli/docs/pipelines/validatesignedpipelines"
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
		{
			Name:         "validate-signed-pipelines",
			Flags:        cliutils.GetCommandFlags(cliutils.ValidateSignedPipelines),
			Aliases:      []string{"vsp"},
			Usage:        validatesignedpipelines.GetDescription(),
			HelpName:     corecommon.CreateUsage("pl validate-signed-pipelines", validatesignedpipelines.GetDescription(), validatesignedpipelines.Usage),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action:       performSignedPipelinesValidation,
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

func performSignedPipelinesValidation(c *cli.Context) error {
	artifactType := c.Args().Get(0)
	coreutils.PrintTitle("Preparing signed pipelines validation")
	// Get service config details
	serviceDetails, err := createPipelinesDetailsByFlags(c)
	if err != nil {
		return err
	}
	validateSignedPipelinesCommand := pipelines.NewValidateSignedPipelinesCommand()
	switch artifactType {
	case "buildInfo":
		buildName, buildNumber, projectKey, err := getBuildInfoArtifactTypeParams(c)
		if err != nil {
			return err
		}
		validateSignedPipelinesCommand.SetArtifactType(artifactType).
			SetBuildName(buildName).
			SetBuildNumber(buildNumber).
			SetProjectKey(projectKey).
			SetServerDetails(serviceDetails)
	case "artifact":
		artifactPath, err := getArtifactArtifactTypeParams(c)
		if err != nil {
			return err
		}
		validateSignedPipelinesCommand.SetArtifactType(artifactType).
			SetArtifactPath(artifactPath).
			SetServerDetails(serviceDetails)
	case "releaseBundle":
		releaseBundleName, releaseBundleVersion, err := getReleaseBundleArtifactTypeParams(c)
		if err != nil {
			return nil
		}
		validateSignedPipelinesCommand.SetArtifactType(artifactType).
			SetReleaseBundleName(releaseBundleName).
			SetReleaseBundleVersion(releaseBundleVersion).
			SetServerDetails(serviceDetails)
	default:
		return errors.New("Allowed artifactType is buildInfo, artifact, releaseBundle")
	}
	return commands.Exec(validateSignedPipelinesCommand)
}

func getBuildInfoArtifactTypeParams(c *cli.Context) (string, string, string, error) {
	buildName := c.String("build-name")
	if buildName == "" {
		return "", "", "", cliutils.PrintHelpAndReturnError("The --build-name option is mandatory.", c)
	}
	buildNumber := c.String("build-number")
	if buildNumber == "" {
		return "", "", "", cliutils.PrintHelpAndReturnError("The --build-number option is mandatory.", c)
	}
	projectKey := c.String("project-key")
	if projectKey == "" {
		return "", "", "", cliutils.PrintHelpAndReturnError("The --project-key option is mandatory.", c)
	}
	return buildName, buildNumber, projectKey, nil
}

func getArtifactArtifactTypeParams(c *cli.Context) (string, error) {
	artifactPath := c.String("artifact-path")
	if artifactPath == "" {
		return "", cliutils.PrintHelpAndReturnError("The --artifact-path option is mandatory.", c)
	}
	return artifactPath, nil
}

func getReleaseBundleArtifactTypeParams(c *cli.Context) (string, string, error) {
	releaseBundleName := c.String("release-bundle-name")
	if releaseBundleName == "" {
		return "", "", cliutils.PrintHelpAndReturnError("The --release-bundle-name option is mandatory.", c)
	}
	releaseBundleVersion := c.String("release-bundle-version")
	if releaseBundleVersion == "" {
		return "", "", cliutils.PrintHelpAndReturnError("The --release-bundle-version option is mandatory.", c)
	}
	return releaseBundleName, releaseBundleVersion, nil
}
