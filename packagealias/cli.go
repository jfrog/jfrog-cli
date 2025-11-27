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
			Name:         "install",
			Usage:        "Install package manager aliases",
			HelpName:     corecommon.CreateUsage("package-alias install", "Install package manager aliases", []string{}),
			ArgsUsage:    "",
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
			Name:         "enable",
			Usage:        "Enable package manager aliases",
			HelpName:     corecommon.CreateUsage("package-alias enable", "Enable package manager aliases", []string{}),
			ArgsUsage:    "",
			Category:     packageAliasCategory,
			Action:       enableCmd,
			BashComplete: corecommon.CreateBashCompletionFunc(),
		},
		{
			Name:         "disable",
			Usage:        "Disable package manager aliases",
			HelpName:     corecommon.CreateUsage("package-alias disable", "Disable package manager aliases", []string{}),
			ArgsUsage:    "",
			Category:     packageAliasCategory,
			Action:       disableCmd,
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
		{
			Name:         "exclude",
			Usage:        "Exclude a tool from Ghost Frog interception (run natively)",
			HelpName:     corecommon.CreateUsage("package-alias exclude <tool>", "Exclude a tool from Ghost Frog interception", []string{"tool"}),
			ArgsUsage:    "<tool>",
			Category:     packageAliasCategory,
			Action:       excludeCmd,
			BashComplete: corecommon.CreateBashCompletionFunc(SupportedTools...),
		},
		{
			Name:         "include",
			Usage:        "Include a tool in Ghost Frog interception (run via JFrog CLI)",
			HelpName:     corecommon.CreateUsage("package-alias include <tool>", "Include a tool in Ghost Frog interception", []string{"tool"}),
			ArgsUsage:    "<tool>",
			Category:     packageAliasCategory,
			Action:       includeCmd,
			BashComplete: corecommon.CreateBashCompletionFunc(SupportedTools...),
		},
	})
}

func installCmd(c *cli.Context) error {
	installCmd := NewInstallCommand()
	return commands.Exec(installCmd)
}

func uninstallCmd(c *cli.Context) error {
	uninstallCmd := NewUninstallCommand()
	return commands.Exec(uninstallCmd)
}

func enableCmd(c *cli.Context) error {
	enableCmd := NewEnableCommand()
	return commands.Exec(enableCmd)
}

func disableCmd(c *cli.Context) error {
	disableCmd := NewDisableCommand()
	return commands.Exec(disableCmd)
}

func statusCmd(c *cli.Context) error {
	statusCmd := NewStatusCommand()
	return commands.Exec(statusCmd)
}

func excludeCmd(c *cli.Context) error {
	if c.NArg() < 1 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}
	tool := c.Args().Get(0)
	excludeCmd := NewExcludeCommand(tool)
	return commands.Exec(excludeCmd)
}

func includeCmd(c *cli.Context) error {
	if c.NArg() < 1 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}
	tool := c.Args().Get(0)
	includeCmd := NewIncludeCommand(tool)
	return commands.Exec(includeCmd)
}
