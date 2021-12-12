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
			Description:  bash_docs.GetDescription(),
			HelpName:     corecommon.CreateUsage("completion bash", bash_docs.GetDescription(), bash_docs.Usage),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(*cli.Context) {
				bash.WriteBashCompletionScript()
			},
		},
		{
			Name:         "zsh",
			Description:  zsh_docs.GetDescription(),
			HelpName:     corecommon.CreateUsage("completion zsh", zsh_docs.GetDescription(), zsh_docs.Usage),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(*cli.Context) {
				zsh.WriteZshCompletionScript()
			},
		},
		{
			Name:         "fish",
			Description:  fish_docs.GetDescription(),
			HelpName:     corecommon.CreateUsage("completion fish", fish_docs.GetDescription(), fish_docs.Usage),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) {
				fish.WriteFishCompletionScript(c)
			},
		},
	})
}
