package vscode

import (
	"fmt"

	coreVscode "github.com/jfrog/jfrog-cli-core/v2/general/ide/vscode"
	"github.com/urfave/cli"
)

const (
	repo            = "repo"
	artifactoryUrl  = "artifactory-url"
	productJsonPath = "product-json-path"
)

func GetCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "config",
			Usage:     "Configure VSCode to use JFrog Artifactory extensions repository.",
			UsageText: "jf vscode config --repo=<VSCODE_REPO_KEY> [command options]",
			Flags:     getConfigFlags(),
			Action:    configCmd,
		},
	}
}

func getConfigFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  repo,
			Usage: "[Mandatory] VSCode repository key in Artifactory.",
		},
		cli.StringFlag{
			Name:  artifactoryUrl,
			Usage: "[Optional] Artifactory server URL. If not provided, uses default server configuration.",
		},
		cli.StringFlag{
			Name:  productJsonPath,
			Usage: "[Optional] Path to VSCode product.json file. If not provided, auto-detects VSCode installation.",
		},
	}
}

func configCmd(c *cli.Context) error {
	repoKey := c.String(repo)
	if repoKey == "" {
		return fmt.Errorf("--repo flag is required\n\nUsage: jf vscode config --repo=<VSCODE_REPO_KEY>\nExample: jf vscode config --repo=vscode-repo")
	}

	artifactoryURL := c.String(artifactoryUrl)
	productPath := c.String(productJsonPath)

	return coreVscode.NewVscodeCommand(repoKey, artifactoryURL, productPath).Run()
}
