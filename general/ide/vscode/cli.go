package vscode

import (
	"fmt"
	"net/url"
	"strings"

	coreVscode "github.com/jfrog/jfrog-cli-core/v2/general/ide/vscode"
	"github.com/urfave/cli"
)

const (
	productJsonPath = "product-json-path"
)

func GetCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "config",
			Usage:     "Configure VSCode to use JFrog Artifactory extensions repository.",
			UsageText: "jf vscode config <repository-url> [command options]",
			Flags:     getConfigFlags(),
			Action:    configCmd,
		},
	}
}

func getConfigFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  productJsonPath,
			Usage: "[Optional] Path to VSCode product.json file. If not provided, auto-detects VSCode installation.",
		},
	}
}

func configCmd(c *cli.Context) error {
	if c.NArg() != 1 {
		return fmt.Errorf("exactly one repository URL argument is required\n\nUsage: jf vscode config <repository-url>\nExample: jf vscode config 'http://productdemo.jfrog.io/artifactory/api/vscode/extensions-remote'")
	}

	repoURL := c.Args().Get(0)
	if repoURL == "" {
		return fmt.Errorf("repository URL is required")
	}

	// Parse and validate the URL
	parsedURL, err := url.Parse(repoURL)
	if err != nil {
		return fmt.Errorf("invalid repository URL: %w", err)
	}

	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return fmt.Errorf("repository URL must include scheme and host (e.g., http://example.com/...)")
	}

	// Extract components from the URL
	// Expected format: http://server/artifactory/api/vscode/repo-key
	pathParts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	if len(pathParts) < 4 || pathParts[0] != "artifactory" || pathParts[1] != "api" || pathParts[2] != "vscode" {
		return fmt.Errorf("invalid repository URL format. Expected: http://server/artifactory/api/vscode/repo-key")
	}

	repoKey := pathParts[3]
	artifactoryURL := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)

	productPath := c.String(productJsonPath)

	return coreVscode.NewVscodeCommand(repoKey, artifactoryURL, productPath).Run()
}
