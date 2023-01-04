package pipelines

import (
	status "github.com/jfrog/jfrog-cli-core/v2/pipelines/commands"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	clientlog "github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/urfave/cli"
)

// fetchLatestPipelineRunStatus fetch pipeline run status based on flags
// supplied
func fetchLatestPipelineRunStatus(c *cli.Context) error {
	clientlog.Info(coreutils.PrintTitle("🐸🐸🐸 Fetching pipeline run status"))

	// read flags for status command
	serverID := c.String("server-id")
	pipName := c.String("name")
	notify := c.Bool("monitor")
	branch := c.String("branch")
	multiBranch := getMultiBranch(c)
	serviceDetails, servErr := getServiceDetails(serverID)
	if servErr != nil {
		return servErr
	}
	sc := status.NewStatusCommand()
	sc.SetBranch(branch).
		SetPipeline(pipName).
		SetNotify(notify).
		SetMultiBranch(multiBranch)

	// set server details
	sc.SetServerDetails(serviceDetails)
	op, runErr := sc.Run()
	if runErr != nil {
		return runErr
	}
	clientlog.Output(op)
	return nil
}
