package pipelines

import (
	"fmt"
	coreConfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/urfave/cli"
)

// getMultiBranch parses singleBranch flag and computes whether multiBranch is set to true/false
func getMultiBranch(c *cli.Context) bool {
	singleBranch := c.Bool("single-branch")
	if singleBranch {
		return false
	}
	return true
}

// createPipelinesDetailsByFlags creates pipelines configuration details
func createPipelinesDetailsByFlags(c *cli.Context) (*coreConfig.ServerDetails, error) {
	plDetails, err := cliutils.CreateServerDetailsWithConfigOffer(c, true, cliutils.CmdPipelines)
	if err != nil {
		return nil, err
	}
	if plDetails.DistributionUrl == "" {
		return nil, fmt.Errorf("the --pipelines-url option is mandatory")
	}
	return plDetails, nil
}
