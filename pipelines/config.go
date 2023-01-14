package pipelines

import (
	"fmt"
	coreConfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	clientlog "github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/urfave/cli"
	"strconv"
)

// getMultiBranch parses multibranch flag to bool
func getMultiBranch(c *cli.Context) bool {
	multiBranch := c.String("multiBranch")
	if multiBranch == "" {
		return true
	} else {
		multiBranch, err := strconv.ParseBool(multiBranch)
		if err != nil {
			clientlog.Warn("MultiBranch flag can parse these values: [1, t, T, TRUE, true, True, 0, f, F, FALSE, false, False]")
			clientlog.Warn("Setting multiBranch to true")
			return true
		}
		return multiBranch
	}
}

// createPipelinesDetailsByFlags creates pipelines configuration details
func createPipelinesDetailsByFlags(c *cli.Context) (*coreConfig.ServerDetails, error) {
	plDetails, err := cliutils.CreateServerDetailsWithConfigOffer(c, true, cliutils.CmdPipelines)
	if err != nil {
		return nil, err
	}
	if plDetails.DistributionUrl == "" {
		return nil, fmt.Errorf("The --pipelines-url option is mandatory")
	}
	return plDetails, nil
}
