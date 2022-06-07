package project

import (
	"os"
	"path/filepath"

	"github.com/jfrog/jfrog-cli-core/v2/common/commands"
	corecommon "github.com/jfrog/jfrog-cli-core/v2/docs/common"
	projectlogic "github.com/jfrog/jfrog-cli-core/v2/general/project"
	projectinit "github.com/jfrog/jfrog-cli/docs/general/project/init"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/urfave/cli"

	"github.com/jfrog/jfrog-cli/utils/cliutils"
)

func GetCommands() []cli.Command {
	return cliutils.GetSortedCommands(cli.CommandsByName{
		{
			Name:         "init",
			Usage:        projectinit.GetDescription(),
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
	if c.NArg() > 1 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}
	// The default project path is the current directory
	path, err := os.Getwd()
	if errorutils.CheckError(err) != nil {
		return err
	}
	if c.NArg() == 1 {
		path = c.Args().Get(0)
		path, err = filepath.Abs(path)
		if errorutils.CheckError(err) != nil {
			return err
		}
	}
	path = clientutils.AddTrailingSlashIfNeeded(path)
	initCmd := projectlogic.NewProjectInitCommand()
	initCmd.SetProjectPath(path).SetServerId(c.String("server-id"))
	return commands.Exec(initCmd)
}
