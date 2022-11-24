package pipelines

import (
	status "github.com/jfrog/jfrog-cli-core/v2/pipelines/commands"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	clientlog "github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/urfave/cli"
)

/*
 * fetchLatestPipelineRunStatus fetch pipeline run status based on flags
 * supplied
 */
func fetchLatestPipelineRunStatus(c *cli.Context, branch string) error {
	clientlog.Info(coreutils.PrintTitle("🐸🐸🐸 fetching pipeline run status"))

	serverID := c.String("server-id")
	pipName := c.String("name")
	notify := c.Bool("monitor")
	serviceDetails, err2 := getServiceDetails(serverID)
	if err2 != nil {
		return err2
	}

	sc := status.NewStatusCommand()
	sc.SetBranch(branch).
		SetPipeline(pipName).
		SetNotify(notify)

	sc.SetServerDetails(serviceDetails)
	op, err4 := sc.Run()
	if err4 != nil {
		return err4
	}
	clientlog.Output(op)

	return nil
}
