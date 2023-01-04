package pipelines

import (
	status "github.com/jfrog/jfrog-cli-core/v2/pipelines/commands"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	clientlog "github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/urfave/cli"
)

// triggerNewRun triggers a new run for supplied flag values
func triggerNewRun(c *cli.Context) error {
	// Read flags for trigger run command
	branch := c.String("branch")
	pipelineName := c.String("pipelineName")
	serverID := c.String("server-id")
	multiBranch := getMultiBranch(c)
	coreutils.PrintTitle("🐸🐸🐸 Triggering pipeline run ")
	clientlog.Output("Triggering pipelineName", pipelineName, "for branch ", branch)

	// Get service config details
	serviceDetails, servErr := getServiceDetails(serverID)
	if servErr != nil {
		return servErr
	}

	// Trigger a pipeline run using branch name and pipeline name
	tc := status.NewTriggerCommand()
	tc.SetBranch(branch).
		SetPipeline(pipelineName).
		SetServerDetails(serviceDetails).
		SetMultiBranch(multiBranch)
	runErr := tc.Run()
	if runErr != nil {
		return runErr
	}
	return nil
}
