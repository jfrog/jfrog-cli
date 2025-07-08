package jetbrains

import (
	"fmt"

	coreJetbrains "github.com/jfrog/jfrog-cli-core/v2/general/ide/jetbrains"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/urfave/cli"
)

func GetCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "set",
			Usage:     "Set configuration for JetBrains IDEs",
			UsageText: "jf jetbrains set <subcommand>",
			Subcommands: []cli.Command{
				{
					Name:      "repository-url",
					Usage:     "Set JetBrains plugin repository URL to use JFrog Artifactory",
					UsageText: "jf jetbrains set repository-url <repository-url> [command options]",
					Action:    setRepositoryUrlCmd,
					Description: `Set the JetBrains plugin repository URL to configure JetBrains IDEs to download plugins from JFrog Artifactory.

The repository URL should be in the format:
https://<artifactory-url>/artifactory/api/jetbrainsplugins/<repo-key>

Examples:
  jf jetbrains set repository-url https://mycompany.jfrog.io/artifactory/api/jetbrainsplugins/jetbrains-plugins

This command will:
- Detect all installed JetBrains IDEs
- Update their plugin repository configuration
- Create automatic backups before making changes
- Require IDE restart to apply changes

Note: On macOS/Linux, you may need to run with sudo for system-installed IDEs.`,
				},
			},
		},
	}
}

func setRepositoryUrlCmd(c *cli.Context) error {
	if c.NArg() == 0 {
		return fmt.Errorf("repository URL is required\n\nUsage: jf jetbrains set repository-url <repository-url>\nExample: jf jetbrains set repository-url https://mycompany.jfrog.io/artifactory/api/jetbrainsplugins/jetbrains-plugins")
	}

	repositoryURL := c.Args().Get(0)
	if repositoryURL == "" {
		return fmt.Errorf("repository URL cannot be empty\n\nUsage: jf jetbrains set repository-url <repository-url>")
	}

	// Create server details using the standard pattern
	serverDetails, err := cliutils.CreateArtifactoryDetailsByFlags(c)
	if err != nil {
		return err
	}

	// Use the core delegation layer
	jetbrainsCmd := coreJetbrains.NewJetbrainsCommand(repositoryURL)
	jetbrainsCmd.SetServerDetails(serverDetails)

	return jetbrainsCmd.Run()
}
