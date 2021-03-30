package completion

import (
	"github.com/codegangsta/cli"
	corecommon "github.com/jfrog/jfrog-cli-core/docs/common"
	"github.com/jfrog/jfrog-cli/completion/shells/bash"
	"github.com/jfrog/jfrog-cli/completion/shells/zsh"
	bash_docs "github.com/jfrog/jfrog-cli/docs/completion/bash"
	zsh_docs "github.com/jfrog/jfrog-cli/docs/completion/zsh"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
)

func GetCommands() []cli.Command {
	return cliutils.GetSortedCommands(cli.CommandsByName{
		{
			Name:         "bash",
			Description:  bash_docs.Description,
			HelpName:     corecommon.CreateUsage("completion bash", bash_docs.Description, bash_docs.Usage),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(*cli.Context) {
				bash.WriteBashCompletionScript()
			},
		},
		{
			Name:         "zsh",
			Description:  zsh_docs.Description,
			HelpName:     corecommon.CreateUsage("completion zsh", zsh_docs.Description, zsh_docs.Usage),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(*cli.Context) {
				zsh.WriteZshCompletionScript()
			},
		},
	})
}
