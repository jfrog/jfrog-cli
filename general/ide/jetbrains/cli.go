package jetbrains

import (
	"fmt"

	coreJetbrains "github.com/jfrog/jfrog-cli-core/v2/general/ide/jetbrains"
	"github.com/urfave/cli"
)

const (
	repo           = "repo"
	artifactoryUrl = "artifactory-url"
)

func GetCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "config",
			Usage:     "Configure JetBrains IDEs to use JFrog Artifactory plugins repository.",
			UsageText: "jf jetbrains config --repo=<JETBRAINS_REPO_KEY> [command options]",
			Flags:     getConfigFlags(),
			Action:    configCmd,
		},
	}
}

func getConfigFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  repo,
			Usage: "[Mandatory] JetBrains repository key in Artifactory.",
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
		return fmt.Errorf("--repo flag is required\n\nUsage: jf jetbrains config --repo=<JETBRAINS_REPO_KEY>\nExample: jf jetbrains config --repo=jetbrains-repo")
	}

	artifactoryURL := c.String(artifactoryUrl)

	return coreJetbrains.NewJetbrainsCommand(repoKey, artifactoryURL).Run()
}
