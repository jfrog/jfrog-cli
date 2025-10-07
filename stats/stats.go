package services

import (
	"fmt"
	"github.com/jfrog/jfrog-cli-artifactory/stats"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/urfave/cli"
)

func GetStats(c *cli.Context) error {
	format := c.String("format")
	accessToken := c.String("access-token")
	serverId := c.String("server-id")
	if c.NArg() != 1 {
		_ = cli.ShowSubcommandHelp(c)
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}
	productName := c.Args().First()
	if productName == "rt" || productName == "artifactory" {
		newStatsCommand := stats.NewStatsCommand().
			SetAccessToken(accessToken).
			SetServerId(serverId).
			SetFormat(format)
		return newStatsCommand.Run()
	} else {
		_ = cli.ShowSubcommandHelp(c)
		return fmt.Errorf("wrong product %s only artifactory or rt is supported", productName)
	}
}
