package completion

import (
	"github.com/codegangsta/cli"
	"github.com/jfrog/jfrog-cli-go/completion/commands"
	"github.com/jfrog/jfrog-cli-go/docs/common"
	"github.com/jfrog/jfrog-cli-go/docs/completion/bash"
	"github.com/jfrog/jfrog-cli-go/utils/cliutils"
)

func GetCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "bash",
			Usage:     bash.Description,
			HelpName:  common.CreateUsage("completion bash", bash.Description, bash.Usage),
			ArgsUsage: common.CreateEnvVars(),
			Action:    completionBash,
		},
	}
}

func completionBash() {
	err := commands.CompletionBash()
	cliutils.ExitOnErr(err)
}
