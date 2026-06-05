package mcp

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	commonCliUtils "github.com/jfrog/jfrog-cli-core/v2/common/cliutils"
	corecommon "github.com/jfrog/jfrog-cli-core/v2/docs/common"
	coreconfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli/docs/common"
	installDocs "github.com/jfrog/jfrog-cli/docs/mcp/install"
	showDocs "github.com/jfrog/jfrog-cli/docs/mcp/show"
	uninstallDocs "github.com/jfrog/jfrog-cli/docs/mcp/uninstall"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-cli/utils/usage"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/urfave/cli"
)

func GetCommands() []cli.Command {
	return cliutils.GetSortedCommands(cli.CommandsByName{
		{
			Name:         "show",
			Flags:        cliutils.GetCommandFlags(cliutils.McpShow),
			Usage:        showDocs.GetDescription(),
			HelpName:     corecommon.CreateUsage("mcp show", showDocs.GetDescription(), showDocs.Usage),
			UsageText:    showDocs.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action:       showCmd,
		},
		{
			Name:         "install",
			Flags:        cliutils.GetCommandFlags(cliutils.McpInstall),
			Usage:        installDocs.GetDescription(),
			HelpName:     corecommon.CreateUsage("mcp install", installDocs.GetDescription(), installDocs.Usage),
			UsageText:    installDocs.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc("cursor", "claude"),
			Action:       installCmd,
		},
		{
			Name:         "uninstall",
			Flags:        cliutils.GetCommandFlags(cliutils.McpUninstall),
			Usage:        uninstallDocs.GetDescription(),
			HelpName:     corecommon.CreateUsage("mcp uninstall", uninstallDocs.GetDescription(), uninstallDocs.Usage),
			UsageText:    uninstallDocs.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			BashComplete: corecommon.CreateBashCompletionFunc("cursor", "claude"),
			Action:       uninstallCmd,
		},
	})
}

func collectFlags(c *cli.Context) []string {
	var flagsUsed []string
	for _, f := range c.Command.Flags {
		name := f.GetName()
		if c.IsSet(name) {
			flagsUsed = append(flagsUsed, name)
		}
	}
	return flagsUsed
}

func showCmd(c *cli.Context) error {
	if c.NArg() != 0 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}
	serverDetails, err := cliutils.CreateServerDetailsWithConfigOffer(c, true, commonCliUtils.Platform)
	if err != nil {
		return err
	}
	const cmdName = "jf mcp show"
	wait := usage.StartReport(cmdName, collectFlags(c), serverDetails)
	defer usage.WaitForReport(cmdName, wait, usage.DefaultReportTimeout)

	return runShow(c, serverDetails, os.Stdout)
}

func runShow(c *cli.Context, serverDetails *coreconfig.ServerDetails, out io.Writer) error {
	mcpURL, err := ResolveMcpURL(c.String("mcp-url"), serverDetails)
	if err != nil {
		return err
	}
	info := EndpointInfo{
		ServerId:    serverDetails.ServerId,
		PlatformUrl: serverDetails.GetUrl(),
		McpUrl:      mcpURL,
		Transport:   "http",
	}
	if strings.EqualFold(c.String("format"), "json") {
		data, err := json.MarshalIndent(info, "", "  ")
		if err != nil {
			return errorutils.CheckError(err)
		}
		return fprintf(out, "%s\n", string(data))
	}
	msg := fmt.Sprintf("MCP endpoint:  %s\n", info.McpUrl)
	if info.ServerId != "" {
		msg += fmt.Sprintf("Server ID:     %s\n", info.ServerId)
	}
	msg += "Auth:          OAuth (completed in the agent / client platform)\n"
	return fprintf(out, "%s", msg)
}

func installCmd(c *cli.Context) error {
	if c.NArg() != 0 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}
	serverDetails, err := cliutils.CreateServerDetailsWithConfigOffer(c, true, commonCliUtils.Platform)
	if err != nil {
		return err
	}
	const cmdName = "jf mcp install"
	wait := usage.StartReport(cmdName, collectFlags(c), serverDetails)
	defer usage.WaitForReport(cmdName, wait, usage.DefaultReportTimeout)

	mcpURL, err := ResolveMcpURL(c.String("mcp-url"), serverDetails)
	if err != nil {
		return err
	}
	return Install(InstallParams{
		Agent:         c.String("agent"),
		ServerName:    serverNameOrDefault(c),
		McpURL:        mcpURL,
		ProjectDir:    c.String("project-dir"),
		Global:        c.Bool("global"),
		DryRun:        c.Bool("dry-run"),
		ServerDetails: serverDetails,
		SkipCheck:     c.Bool("skip-check"),
	}, os.Stdout)
}

func uninstallCmd(c *cli.Context) error {
	if c.NArg() != 0 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}
	serverDetails, err := cliutils.CreateServerDetailsWithConfigOffer(c, true, commonCliUtils.Platform)
	if err != nil {
		return err
	}
	const cmdName = "jf mcp uninstall"
	wait := usage.StartReport(cmdName, collectFlags(c), serverDetails)
	defer usage.WaitForReport(cmdName, wait, usage.DefaultReportTimeout)

	return Uninstall(UninstallParams{
		Agent:      c.String("agent"),
		ServerName: serverNameOrDefault(c),
		ProjectDir: c.String("project-dir"),
		Global:     c.Bool("global"),
		DryRun:     c.Bool("dry-run"),
	}, os.Stdout)
}

func serverNameOrDefault(c *cli.Context) string {
	if name := strings.TrimSpace(c.String("name")); name != "" {
		return name
	}
	return DefaultServerName
}
