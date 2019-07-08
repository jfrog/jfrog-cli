package completion

import (
	"github.com/codegangsta/cli"
	"github.com/jfrog/jfrog-cli-go/completion/shells"
	"github.com/jfrog/jfrog-cli-go/docs/common"
	"github.com/jfrog/jfrog-cli-go/docs/completion/bash"
	"github.com/jfrog/jfrog-cli-go/docs/completion/zsh"
)

func GetCommands() []cli.Command {
	return []cli.Command{
		{
			Name:         "bash",
			Usage:        bash.Description,
			HelpName:     common.CreateUsage("completion bash", bash.Description, bash.Usage),
			BashComplete: common.CreateBashCompletionFunc(),
			Action: func(*cli.Context) {
				shells.WriteBashCompletionScript()
			},
		},
		{
			Name:         "zsh",
			Usage:        zsh.Description,
			HelpName:     common.CreateUsage("completion zsh", zsh.Description, zsh.Usage),
			BashComplete: common.CreateBashCompletionFunc(),
			Action: func(*cli.Context) {
				shells.WriteZshCompletionScript()
			},
		},
	}
}
