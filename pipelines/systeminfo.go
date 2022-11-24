package pipelines

import (
	status "github.com/jfrog/jfrog-cli-core/v2/pipelines/commands"
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

	vc := status.NewVersionCommand()
	vc.SetServerDetails(serviceDetails)
	version, err := vc.Run()
	if err != nil {
		return err
	}

	clientlog.Output("Pipelines Server Version: ", version)
	return nil
}
