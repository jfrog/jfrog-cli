package pipelines

import (
	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	status "github.com/jfrog/jfrog-cli-core/v2/pipelines/commands"
	"github.com/urfave/cli"
)

// getVersion version command handler
func getVersion(c *cli.Context) error {
	serviceDetails, servErr := createPipelinesDetailsByFlags(c)
	if servErr != nil {
		return servErr
	}
	versionCommand := status.NewVersionCommand()
	versionCommand.SetServerDetails(serviceDetails)
	return commands.Exec(versionCommand)
}
