// Package packagealias provides the "jf package-alias" command implementation
// according to the Ghost Frog technical specification
package packagealias

import (
	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	corecommon "github.com/jfrog/jfrog-cli-core/v2/docs/common"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/urfave/cli"
)

const (
	packageAliasCategory = "Package Aliasing"
)

// GetCommands returns all package-alias sub-commands
func GetCommands() []cli.Command {
	return cliutils.GetSortedCommands(cli.CommandsByName{
		{
			Name:      "install",
			Usage:     "Install package manager aliases",
			HelpName:  corecommon.CreateUsage("package-alias install", "Install package manager aliases", []string{}),
			ArgsUsage: "",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "packages",
					Usage: "Comma-separated list of package managers to alias (default: all supported package managers)",
				},
			},
			Category:     packageAliasCategory,
			Action:       installCmd,
			BashComplete: corecommon.CreateBashCompletionFunc(),
		},
		{
			Name:         "uninstall",
			Usage:        "Uninstall package manager aliases",
			HelpName:     corecommon.CreateUsage("package-alias uninstall", "Uninstall package manager aliases", []string{}),
			ArgsUsage:    "",
			Category:     packageAliasCategory,
			Action:       uninstallCmd,
			BashComplete: corecommon.CreateBashCompletionFunc(),
		},
		{
			Name:         "status",
			Usage:        "Show package alias status",
			HelpName:     corecommon.CreateUsage("package-alias status", "Show package alias status", []string{}),
			ArgsUsage:    "",
			Category:     packageAliasCategory,
			Action:       statusCmd,
			BashComplete: corecommon.CreateBashCompletionFunc(),
		},
	})
}

func installCmd(c *cli.Context) error {
	installCmd := NewInstallCommand(c.String("packages"))
	return commands.Exec(installCmd)
}

func uninstallCmd(c *cli.Context) error {
	uninstallCmd := NewUninstallCommand()
	return commands.Exec(uninstallCmd)
}

func statusCmd(c *cli.Context) error {
	statusCmd := NewStatusCommand()
	return commands.Exec(statusCmd)
}
