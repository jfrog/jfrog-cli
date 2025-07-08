package vscode

import (
	"fmt"

	coreVscode "github.com/jfrog/jfrog-cli-core/v2/general/ide/vscode"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/urfave/cli"
)

const (
	productJsonPath = "product-json-path"
)

func GetCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "set",
			Usage:     "Set configuration for VSCode",
			UsageText: "jf vscode set <subcommand>",
			Subcommands: []cli.Command{
				{
					Name:      "service-url",
					Usage:     "Set VSCode extensions service URL to use JFrog Artifactory",
					UsageText: "jf vscode set service-url <complete-service-url> [command options]",
					Flags:     getSetServiceUrlFlags(),
					Action:    setServiceUrlCmd,
					Description: `Set the complete VSCode extensions service URL to configure VSCode to download extensions from JFrog Artifactory.

The service URL should be in the format:
https://<artifactory-url>/artifactory/api/vscodeextensions/<repo-key>/_apis/public/gallery

Examples:
  jf vscode set service-url https://mycompany.jfrog.io/artifactory/api/vscodeextensions/vscode-extensions/_apis/public/gallery

This command will:
- Modify the VSCode product.json file to change the extensions gallery URL
- Create an automatic backup before making changes
- Require VSCode to be restarted to apply changes

Note: On macOS/Linux, you may need to run with sudo for system-installed VSCode.`,
				},
			},
		},
	}
}

func getSetServiceUrlFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  productJsonPath,
			Usage: "[Optional] Path to VSCode product.json file. If not provided, auto-detects VSCode installation.",
		},
	}
}

func setServiceUrlCmd(c *cli.Context) error {
	if c.NArg() == 0 {
		return fmt.Errorf("service URL is required\n\nUsage: jf vscode set service-url <service-url>\nExample: jf vscode set service-url https://mycompany.jfrog.io/artifactory/api/vscodeextensions/vscode-extensions/_apis/public/gallery")
	}

	serviceURL := c.Args().Get(0)
	if serviceURL == "" {
		return fmt.Errorf("service URL cannot be empty\n\nUsage: jf vscode set service-url <service-url>")
	}

	productPath := c.String(productJsonPath)

	// Create server details using the standard pattern
	serverDetails, err := cliutils.CreateArtifactoryDetailsByFlags(c)
	if err != nil {
		return err
	}

	// Use the core delegation layer
	vscodeCmd := coreVscode.NewVscodeCommand(serviceURL, productPath)
	vscodeCmd.SetServerDetails(serverDetails)

	return vscodeCmd.Run()
}
