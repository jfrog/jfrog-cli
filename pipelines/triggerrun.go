package pipelines

import (
	status "github.com/jfrog/jfrog-cli-core/v2/pipelines/commands"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	clientlog "github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/urfave/cli"
)

// triggerNewRun triggers a new run for supplied flag values
func triggerNewRun(c *cli.Context) error {
	// read flags for trigger run command
	b := c.String("branch")
	p := c.String("pipelineName")
	s := c.String("server-id")
	multiBranch := getMultiBranch(c)
	coreutils.PrintTitle("🐸🐸🐸 triggering pipeline run ")
	clientlog.Output("triggering pipelineName", p, "for branch ", b)

	// get service config details
	serviceDetails, servErr := getServiceDetails(s)
	if servErr != nil {
		return servErr
	}

	// trigger a pipeline run using branch name and pipeline name
	tc := status.NewTriggerCommand()
	tc.SetBranch(b).
		SetPipeline(p).
		SetServerDetails(serviceDetails).
		SetMultiBranch(multiBranch)
	runErr := tc.Run()
	if runErr != nil {
		return runErr
	}
	return nil
}
