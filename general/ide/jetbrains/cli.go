package jetbrains

import (
	"fmt"
	"net/url"
	"strings"

	coreJetbrains "github.com/jfrog/jfrog-cli-core/v2/general/ide/jetbrains"
	"github.com/urfave/cli"
)

func GetCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "config",
			Usage:     "Configure JetBrains IDEs to use JFrog Artifactory plugins repository.",
			UsageText: "jf jetbrains config <repository-url>",
			Action:    configCmd,
		},
	}
}

func configCmd(c *cli.Context) error {
	if c.NArg() != 1 {
		return fmt.Errorf("exactly one repository URL argument is required\n\nUsage: jf jetbrains config <repository-url>\nExample: jf jetbrains config 'http://productdemo.jfrog.io/artifactory/api/jetbrains/jetbrains-remote'")
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
	// Expected format: http://server/artifactory/api/jetbrains/repo-key
	pathParts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	if len(pathParts) < 4 || pathParts[0] != "artifactory" || pathParts[1] != "api" || pathParts[2] != "jetbrains" {
		return fmt.Errorf("invalid repository URL format. Expected: http://server/artifactory/api/jetbrains/repo-key")
	}

	repoKey := pathParts[3]
	artifactoryURL := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)

	return coreJetbrains.NewJetbrainsCommand(repoKey, artifactoryURL).Run()
}
