package pipelines

import (
	status "github.com/jfrog/jfrog-cli-core/v2/pipelines/commands"
	clientlog "github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/urfave/cli"
)

// getVersion version command handler
func getVersion(c *cli.Context) error {
	err := writePipelinesVersion(c)
	if err != nil {
		return err
	}
	return nil
}

// writePipelinesVersion writes pipelines server version to console
func writePipelinesVersion(c *cli.Context) error {
	serverID := c.String("server-id")
	serviceDetails, servErr := getServiceDetails(serverID)
	if servErr != nil {
		return servErr
	}
	vc := status.NewVersionCommand()
	vc.SetServerDetails(serviceDetails)
	version, runErr := vc.Run()
	if runErr != nil {
		return runErr
	}
	clientlog.Output("Pipelines Server Version: ", version)
	return nil
}
