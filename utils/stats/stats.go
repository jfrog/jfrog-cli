package services

import (
	coreStats "github.com/jfrog/jfrog-cli-core/utils/stats"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/urfave/cli"
)

func GetStats(c *cli.Context) error {
	productName := c.String("product")
	formatOutput := c.String("output")
	accessToken := c.String("access-token")

	if err := coreStats.GetStats(formatOutput, productName, accessToken); err != nil {
		log.Error("An error occurred:", err)
		return err
	}

	return nil
}
