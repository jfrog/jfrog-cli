package services

import (
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/generic"
	"github.com/urfave/cli"
)

func GetStats(c *cli.Context) error {
	productName := c.String("product")
	formatOutput := c.String("output")
	accessToken := c.String("access-token")
	serverId := c.String("server-id")
	newStatsCommand := generic.NewStatsCommand().SetAccessToken(accessToken).SetServerId(serverId).SetProductName(productName).SetFormatOutput(formatOutput).SetAccessToken(accessToken)
	err := newStatsCommand.Run()
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	return nil
}
