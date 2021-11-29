package project

import (
	"os"

	"github.com/codegangsta/cli"
	corecommon "github.com/jfrog/jfrog-cli-core/v2/docs/common"
	projectlogic "github.com/jfrog/jfrog-cli-core/v2/general/project"
	projectinit "github.com/jfrog/jfrog-cli/docs/general/project/init"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"

	"github.com/jfrog/jfrog-cli/utils/cliutils"
)

func GetCommands() []cli.Command {
	return cliutils.GetSortedCommands(cli.CommandsByName{
		{
			Name:         "init",
			Description:  projectinit.GetDescription(),
			Flags:        cliutils.GetCommandFlags(cliutils.InitProject),
			HelpName:     corecommon.CreateUsage("project init", projectinit.GetDescription(), projectinit.Usage),
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action: func(c *cli.Context) error {
				return initProject(c)
			},
		},
	})
}

func initProject(c *cli.Context) error {
	if c.NArg() < 1 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	// The default project path is the current directory
	path, err := os.Getwd()
	if errorutils.CheckError(err) != nil {
		return err
	}
	if c.NArg() == 1 {
		path = c.Args().Get(0)
	}
	initCmd := projectlogic.NewProjectInitCommand(path)
	return initCmd.Run()
}
