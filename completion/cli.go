package completion

import (
	corecommon "github.com/jfrog/jfrog-cli-core/v2/docs/common"
	"github.com/jfrog/jfrog-cli/completion/shells/bash"
	"github.com/jfrog/jfrog-cli/completion/shells/fish"
	"github.com/jfrog/jfrog-cli/completion/shells/zsh"
	bash_docs "github.com/jfrog/jfrog-cli/docs/completion/bash"
	fish_docs "github.com/jfrog/jfrog-cli/docs/completion/fish"
	zsh_docs "github.com/jfrog/jfrog-cli/docs/completion/zsh"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/urfave/cli"
)

func GetCommands() []cli.Command {
	return cliutils.GetSortedCommands(cli.CommandsByName{
		{
			Name:         "bash",
			Flags:        cliutils.GetCommandFlags(cliutils.Completion),
			Usage:        bash_docs.GetDescription(),
			HelpName:     corecommon.CreateUsage("completion bash", bash_docs.GetDescription(), bash_docs.Usage),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) {
				bash.WriteBashCompletionScript(getInstallFlag(c))
			},
		},
		{
			Name:         "zsh",
			Flags:        cliutils.GetCommandFlags(cliutils.Completion),
			Usage:        zsh_docs.GetDescription(),
			HelpName:     corecommon.CreateUsage("completion zsh", zsh_docs.GetDescription(), zsh_docs.Usage),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) {
				zsh.WriteZshCompletionScript(getInstallFlag(c))
			},
		},
		{
			Name:         "fish",
			Flags:        cliutils.GetCommandFlags(cliutils.Completion),
			Usage:        fish_docs.GetDescription(),
			HelpName:     corecommon.CreateUsage("completion fish", fish_docs.GetDescription(), fish_docs.Usage),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) {
				fish.WriteFishCompletionScript(c, getInstallFlag(c))
			},
		},
	})
}

// Extract the --install flag from context
func getInstallFlag(c *cli.Context) bool {
	return c.Bool(cliutils.Install)
}
