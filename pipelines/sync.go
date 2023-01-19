package pipelines

import (
	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	syncPipeRes "github.com/jfrog/jfrog-cli-core/v2/pipelines/commands"
	clientlog "github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/urfave/cli"
)

// syncPipelineResources sync pipelines resource
func syncPipelineResources(c *cli.Context) error {
	// Get arguments repository name and branch name
	repository := c.Args().Get(0)
	branch := c.Args().Get(1)
	clientlog.Info("Triggering pipeline sync on repository ", repository, "branch", branch)
	serviceDetails, err := createPipelinesDetailsByFlags(c)
	if err != nil {
		return err
	}

	// Create new sync command and add filters
	syncCommand := syncPipeRes.NewSyncCommand()
	syncCommand.SetBranch(branch)
	syncCommand.SetRepositoryFullName(repository)
	syncCommand.SetServerDetails(serviceDetails)
	return commands.Exec(syncCommand)
}

// getSyncPipelineResourcesStatus fetch sync status for a given repository path and branch name
func getSyncPipelineResourcesStatus(c *cli.Context) error {
	branch := c.String("branch")
	repository := c.String("repository")
	clientlog.Info("Fetching pipeline sync status on repository ", repository, "branch", branch)

	// Fetch service details for authentication
	serviceDetails, err := createPipelinesDetailsByFlags(c)
	if err != nil {
		return err
	}

	// Create sync status command and add filter params
	syncStatusCommand := syncPipeRes.NewSyncStatusCommand()
	syncStatusCommand.SetBranch(branch)
	syncStatusCommand.SetRepoPath(repository)
	syncStatusCommand.SetServerDetails(serviceDetails)
	return commands.Exec(syncStatusCommand)
}
