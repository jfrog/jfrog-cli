package pipelines

import (
	utilsconfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	clientlog "github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/urfave/cli"
	"strconv"
)

// getServiceDetails returns server details based on serverID
// if serverID is empty returns default config otherwise error
func getServiceDetails(serverID string) (*utilsconfig.ServerDetails, error) {
	if serverID == "" {
		conf, err := utilsconfig.GetDefaultServerConf()
		if err != nil {
			clientlog.Error("unable to find server configuration exiting")
			return nil, errorutils.CheckError(err)
		}
		serverID = conf.ServerId
	}
	serviceDetails, err := utilsconfig.GetSpecificConfig(serverID, false, false)
	if err != nil {
		clientlog.Error(err)
		return nil, errorutils.CheckError(err)
	}
	return serviceDetails, err
}

// getMultiBranch parses multibranch flag to bool
func getMultiBranch(c *cli.Context) bool {
	multiBranch := c.String("multiBranch")
	if multiBranch == "" {
		return true
	} else {
		multiBranch, err := strconv.ParseBool(multiBranch)
		if err != nil {
			clientlog.Warn("multiBranch should be one of these values: [1, t, T, TRUE, true, True, 0, f, F, FALSE, false, False]")
			clientlog.Warn("setting multiBranch to true")
			return true
		}
		return multiBranch
	}
}
