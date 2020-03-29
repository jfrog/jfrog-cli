package completion

import (
	"github.com/codegangsta/cli"
	"github.com/jfrog/jfrog-cli/completion/shells/bash"
	"github.com/jfrog/jfrog-cli/completion/shells/zsh"
	"github.com/jfrog/jfrog-cli/docs/common"
	bash_docs "github.com/jfrog/jfrog-cli/docs/completion/bash"
	zsh_docs "github.com/jfrog/jfrog-cli/docs/completion/zsh"
)

func GetCommands() []cli.Command {
	return []cli.Command{
		{
			Name:         "bash",
			Usage:        bash_docs.Description,
			HelpName:     common.CreateUsage("completion bash", bash_docs.Description, bash_docs.Usage),
			BashComplete: common.CreateBashCompletionFunc(),
			Action: func(*cli.Context) {
				bash.WriteBashCompletionScript()
			},
		},
		{
			Name:         "zsh",
			Usage:        zsh_docs.Description,
			HelpName:     common.CreateUsage("completion zsh", zsh_docs.Description, zsh_docs.Usage),
			BashComplete: common.CreateBashCompletionFunc(),
			Action: func(*cli.Context) {
				zsh.WriteZshCompletionScript()
			},
		},
	}
}
