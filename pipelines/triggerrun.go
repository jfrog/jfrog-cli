package pipelines

import (
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	clientlog "github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/urfave/cli"
)

/*
triggerNewRun triggers a new run for supplied flag values
*/
func triggerNewRun(c *cli.Context) error {
	b := c.String("branch")
	p := c.String("pipelineName")
	s := c.String("server-id")
	clientlog.Info(coreutils.PrintTitle("🐸🐸🐸 triggering pipeline run "))
	clientlog.Output("triggering pipelineName", p, "for branch ", b)

	serviceDetails, err2 := getServiceDetails(s)
	if err2 != nil {
		return err2
	}

	pipelinesMgr, err3 := getPipelinesManager(serviceDetails)
	if err3 != nil {
		return err3
	}

	run, err := pipelinesMgr.TriggerPipelineRun(b, p)
	if err != nil {
		return err
	}
	clientlog.Output(run)
	return nil
}
