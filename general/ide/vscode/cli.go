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
	publisher       = "publisher"
	extensionName   = "extension-name"
	version         = "version"
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
		{
			Name:      "install",
			Usage:     "Install VSCode extension from JFrog Artifactory repository.",
			UsageText: "jf vscode install --publisher=<PUBLISHER> --extension-name=<EXTENSION_NAME> --repo=<VSCODE_REPO_KEY> [command options]",
			Flags:     getInstallFlags(),
			Action:    installCmd,
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

func getInstallFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  repo,
			Usage: "[Mandatory] VSCode repository key in Artifactory.",
		},
		cli.StringFlag{
			Name:  publisher,
			Usage: "[Mandatory] Extension publisher name.",
		},
		cli.StringFlag{
			Name:  extensionName,
			Usage: "[Mandatory] Extension name to install.",
		},
		cli.StringFlag{
			Name:  version,
			Usage: "[Optional] Specific extension version to install. If not provided, installs latest version.",
		},
		cli.StringFlag{
			Name:  artifactoryUrl,
			Usage: "[Optional] Artifactory server URL. If not provided, uses default server configuration.",
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

func installCmd(c *cli.Context) error {
	repoKey := c.String(repo)
	if repoKey == "" {
		return fmt.Errorf("--repo flag is required\n\nUsage: jf vscode install --publisher=<PUBLISHER> --extension-name=<EXTENSION_NAME> --repo=<VSCODE_REPO_KEY>")
	}

	publisherName := c.String(publisher)
	if publisherName == "" {
		return fmt.Errorf("--publisher flag is required\n\nUsage: jf vscode install --publisher=<PUBLISHER> --extension-name=<EXTENSION_NAME> --repo=<VSCODE_REPO_KEY>")
	}

	extName := c.String(extensionName)
	if extName == "" {
		return fmt.Errorf("--extension-name flag is required\n\nUsage: jf vscode install --publisher=<PUBLISHER> --extension-name=<EXTENSION_NAME> --repo=<VSCODE_REPO_KEY>")
	}

	extVersion := c.String(version)
	artifactoryURL := c.String(artifactoryUrl)

	return coreVscode.NewVscodeInstallCommand(repoKey, artifactoryURL, publisherName, extName, extVersion).Run()
}
