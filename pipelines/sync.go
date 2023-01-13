package pipelines

import (
	syncPipeRes "github.com/jfrog/jfrog-cli-core/v2/pipelines/commands"
	clientlog "github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/urfave/cli"
)

// syncPipelineResources sync pipelines resource
func syncPipelineResources(c *cli.Context) error {
	repository := c.Args().Get(0)
	branch := c.Args().Get(1)
	clientlog.Info("Triggering pipeline sync on repository ", repository, "branch", branch)
	serviceDetails, servErr := createPipelinesDetailsByFlags(c)
	if servErr != nil {
		return servErr
	}

	// create new sync command and add filters
	syncCommand := syncPipeRes.NewSyncCommand()
	syncCommand.SetBranch(branch)
	syncCommand.SetRepositoryFullName(repository)
	syncCommand.SetServerDetails(serviceDetails)
	err := syncCommand.Run()
	if err != nil {
		return err
	}
	return nil
}

// getSyncPipelineResourcesStatus fetch sync status for a given repository path and branch name
func getSyncPipelineResourcesStatus(c *cli.Context) error {
	branch := c.String("branch")
	repository := c.String("repository")
	clientlog.Info("Fetching pipeline sync status on repository ", repository, "branch", branch)

	// fetch service details for authentication
	serviceDetails, servErr := createPipelinesDetailsByFlags(c)
	if servErr != nil {
		return servErr
	}

	// create sync status command and add filter params
	syncStatusCommand := syncPipeRes.NewSyncStatusCommand()
	syncStatusCommand.SetBranch(branch)
	syncStatusCommand.SetRepoPath(repository)
	syncStatusCommand.SetServerDetails(serviceDetails)
	err := syncStatusCommand.Run()
	if err != nil {
		return err
	}
	return nil
}
