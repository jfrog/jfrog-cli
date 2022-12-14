package pipelines

import (
	status "github.com/jfrog/jfrog-cli-core/v2/pipelines/commands"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	clientlog "github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/urfave/cli"
)

// fetchLatestPipelineRunStatus fetch pipeline run status based on flags
// supplied
func fetchLatestPipelineRunStatus(c *cli.Context, branch string) error {
	clientlog.Info(coreutils.PrintTitle("🐸🐸🐸 fetching pipeline run status"))

	serverID := c.String("server-id")
	pipName := c.String("name")
	notify := c.Bool("monitor")
	multiBranch := getMultiBranch(c)
	clientlog.Output(serverID, pipName, notify, multiBranch)

	serviceDetails, servErr := getServiceDetails(serverID)
	if servErr != nil {
		return errorutils.CheckError(servErr)
	}

	sc := status.NewStatusCommand()
	sc.SetBranch(branch).
		SetPipeline(pipName).
		SetNotify(notify).
		SetMultiBranch(multiBranch)

	sc.SetServerDetails(serviceDetails)
	op, runErr := sc.Run()
	if runErr != nil {
		return errorutils.CheckError(runErr)
	}
	clientlog.Output(op)

	return nil
}
