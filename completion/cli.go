package completion

import (
	"github.com/codegangsta/cli"
	corecommon "github.com/jfrog/jfrog-cli-core/v2/docs/common"
	"github.com/jfrog/jfrog-cli/completion/shells/bash"
	"github.com/jfrog/jfrog-cli/completion/shells/zsh"
	bashdocs "github.com/jfrog/jfrog-cli/docs/completion/bash"
	zshdocs "github.com/jfrog/jfrog-cli/docs/completion/zsh"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
)

func GetCommands() []cli.Command {
	return cliutils.GetSortedCommands(cli.CommandsByName{
		{
			Name:         "bash",
			Description:  bashdocs.GetDescription(),
			HelpName:     corecommon.CreateUsage("completion bash", bashdocs.GetDescription(), bashdocs.Usage),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(*cli.Context) {
				bash.WriteBashCompletionScript()
			},
		},
		{
			Name:         "zsh",
			Description:  zshdocs.GetDescription(),
			HelpName:     corecommon.CreateUsage("completion zsh", zshdocs.GetDescription(), zshdocs.Usage),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(*cli.Context) {
				zsh.WriteZshCompletionScript()
			},
		},
	})
}
