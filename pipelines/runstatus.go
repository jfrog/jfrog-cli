package pipelines

import (
	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	status "github.com/jfrog/jfrog-cli-core/v2/pipelines/commands"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	clientlog "github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/urfave/cli"
)

// fetchLatestPipelineRunStatus fetch pipeline run status based on flags
// supplied
func fetchLatestPipelineRunStatus(c *cli.Context) error {
	clientlog.Info(coreutils.PrintTitle("Fetching pipeline run status"))

	// Read flags for status command
	pipName := c.String("pipeline-name")
	notify := c.Bool("monitor")
	branch := c.String("branch")
	multiBranch := getMultiBranch(c)
	serviceDetails, servErr := createPipelinesDetailsByFlags(c)
	if servErr != nil {
		return servErr
	}
	statusCommand := status.NewStatusCommand()
	statusCommand.SetBranch(branch).
		SetPipeline(pipName).
		SetNotify(notify).
		SetMultiBranch(multiBranch)

	// Set server details
	statusCommand.SetServerDetails(serviceDetails)
	return commands.Exec(statusCommand)
}
