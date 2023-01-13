package pipelines

import (
	status "github.com/jfrog/jfrog-cli-core/v2/pipelines/commands"
	clientlog "github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/urfave/cli"
)

// getVersion version command handler
func getVersion(c *cli.Context) error {
	return writePipelinesVersion(c)
}

// writePipelinesVersion writes pipelines server version to console
func writePipelinesVersion(c *cli.Context) error {
	serviceDetails, servErr := createPipelinesDetailsByFlags(c)
	if servErr != nil {
		return servErr
	}
	versionCommand := status.NewVersionCommand()
	versionCommand.SetServerDetails(serviceDetails)
	version, runErr := versionCommand.Run()
	if runErr != nil {
		return runErr
	}
	clientlog.Output("Pipelines Server Version: ", version)
	return nil
}
