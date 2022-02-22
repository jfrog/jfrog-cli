package buildtools

import (
	corecommon "github.com/jfrog/jfrog-cli-core/v2/docs/common"
	"github.com/jfrog/jfrog-cli/docs/buildtools/dockerpull"
	"github.com/jfrog/jfrog-cli/docs/buildtools/dockerpush"
	"github.com/jfrog/jfrog-cli/docs/buildtools/dockerscan"
	"github.com/jfrog/jfrog-cli/docs/common"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/urfave/cli"
)
// All the commands in this are hidden and have no logic. The purpose is to override the --help of the generic command.
// For example, 'jf docker scan --help' doesn't show the real help from the docker cli but gets redirects to 'dockerscanhelp' help output.
func GetBuildToolsHelpCommands() []cli.Command {
	return cliutils.GetSortedCommands(cli.CommandsByName{
		{
			Name:      "dockerpushhelp",
			Flags:     cliutils.GetCommandFlags(cliutils.DockerPush),
			Usage:     dockerpush.GetDescription(),
			HelpName:  corecommon.CreateUsage("docker push", dockerpush.GetDescription(), dockerpush.Usage),
			UsageText: dockerpush.GetArguments(),
			ArgsUsage: common.CreateEnvVars(),
			Hidden:    true,
		},
		{
			Name:      "dockerpullhelp",
			Flags:     cliutils.GetCommandFlags(cliutils.DockerPull),
			Usage:     dockerpull.GetDescription(),
			HelpName:  corecommon.CreateUsage("docker pull", dockerpull.GetDescription(), dockerpull.Usage),
			UsageText: dockerpull.GetArguments(),
			ArgsUsage: common.CreateEnvVars(),
			Hidden:    true,
		},
		{
			Name:      "dockerscanhelp",
			Flags:     cliutils.GetCommandFlags(cliutils.DockerScan),
			Usage:     dockerscan.GetDescription(),
			HelpName:  corecommon.CreateUsage("docker scan", dockerscan.GetDescription(), dockerscan.Usage),
			UsageText: dockerscan.GetArguments(),
			ArgsUsage: common.CreateEnvVars(),
			Hidden:    true,
		},
	})
}
