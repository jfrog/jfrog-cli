package pipelines

import (
	clientlog "github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/urfave/cli"
)

/*
 * getVersion version command handler
 */
func getVersion(c *cli.Context) error {
	err := writePipelinesVersion(c)
	if err != nil {
		return err
	}
	return nil
}

/*
 * writePipelinesVersion writes pipelines server version to console
 */
func writePipelinesVersion(c *cli.Context) error {
	serverID := c.String("server-id")
	c.Bool("monitor")
	serviceDetails, err2 := getServiceDetails(serverID)
	if err2 != nil {
		return err2
	}

	pipelinesMgr, err3 := getPipelinesManager(serviceDetails)
	if err3 != nil {
		return err3
	}
	p, err4 := pipelinesMgr.GetSystemInfo()
	if err4 != nil {
		return err4
	}
	clientlog.Output("Pipelines Server Version: ", p.Version)
	return nil
}
