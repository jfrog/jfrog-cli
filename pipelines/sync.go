package pipelines

import (
	syncPipeRes "github.com/jfrog/jfrog-cli-core/v2/pipelines/commands"
	clientlog "github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/urfave/cli"
)

// syncPipelineResources sync pipelines resource
func syncPipelineResources(c *cli.Context) error {
	b := c.String("branch")
	r := c.String("repository")
	s := c.String("server-id")
	clientlog.Info("🐸🐸🐸 triggering pipeline sync on repository ", r, "branch", b)
	serviceDetails, servErr := getServiceDetails(s)
	if servErr != nil {
		return servErr
	}

	// create new sync command and add filters
	sp := syncPipeRes.NewSyncCommand()
	sp.SetBranch(b)
	sp.SetRepositoryFullName(r)
	sp.SetServerDetails(serviceDetails)
	err := sp.Run()
	if err != nil {
		return err
	}
	return nil
}

// getSyncPipelineResourcesStatus fetch sync status for a given repository path and branch name
func getSyncPipelineResourcesStatus(c *cli.Context) error {
	b := c.String("branch")
	r := c.String("repository")
	s := c.String("server-id")
	clientlog.Info("🐸🐸🐸 fetching pipeline sync status on repository ", r, "branch", b)

	// fetch service details for authentication
	serviceDetails, servErr := getServiceDetails(s)
	if servErr != nil {
		return servErr
	}

	// create sync status command and add filter params
	sp := syncPipeRes.NewSyncStatusCommand()
	sp.SetBranch(b)
	sp.SetRepoPath(r)
	sp.SetServerDetails(serviceDetails)
	statusOutput, err := sp.Run()
	if err != nil {
		return err
	}
	clientlog.Output(statusOutput)
	return nil
}
