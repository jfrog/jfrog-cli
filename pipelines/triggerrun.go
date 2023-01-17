package pipelines

import (
	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	status "github.com/jfrog/jfrog-cli-core/v2/pipelines/commands"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	clientlog "github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/urfave/cli"
)

// triggerNewRun triggers a new run for supplied flag values
func triggerNewRun(c *cli.Context) error {
	// Read arguments pipeline name and branch to trigger pipeline run
	pipelineName := c.Args().Get(0)
	branch := c.Args().Get(1)
	multiBranch := getMultiBranch(c)
	coreutils.PrintTitle("Triggering pipeline run ")
	clientlog.Info("Triggering pipelineName", pipelineName, "for branch ", branch)

	// Get service config details
	serviceDetails, servErr := createPipelinesDetailsByFlags(c)
	if servErr != nil {
		return servErr
	}

	// Trigger a pipeline run using branch name and pipeline name
	triggerCommand := status.NewTriggerCommand()
	triggerCommand.SetBranch(branch).
		SetPipeline(pipelineName).
		SetServerDetails(serviceDetails).
		SetMultiBranch(multiBranch)
	return commands.Exec(triggerCommand)
}
