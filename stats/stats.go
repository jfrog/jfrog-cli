package services

import (
	"github.com/jfrog/jfrog-cli-core/v2/general"
	"github.com/urfave/cli"
)

func GetStats(c *cli.Context) error {
	filter := c.String("filter")
	formatOutput := c.String("format-stats")
	accessToken := c.String("access-token")
	serverId := c.String("server-id")
	newStatsCommand := general.NewStatsCommand().SetAccessToken(accessToken).SetServerId(serverId).SetFilterName(filter).SetFormatOutput(formatOutput).SetAccessToken(accessToken)
	err := newStatsCommand.Run()
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	return nil
}
