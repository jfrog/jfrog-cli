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
	if c.NArg() != 1 {
		_ = cli.ShowSubcommandHelp(c)
		return cli.NewExitError("nError: This command requires exactly one argument (product name).", 1)
	}
	productName := c.Args().First()
	newStatsCommand := general.NewStatsCommand().
		SetAccessToken(accessToken).SetServerId(serverId).
		SetFilterName(filter).SetFormatOutput(formatOutput).
		SetAccessToken(accessToken).SetProduct(productName)
	err := newStatsCommand.Run()
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	return nil
}
