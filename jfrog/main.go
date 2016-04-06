package main

import (
	"github.com/jfrogdev/jfrog-cli-go/Godeps/_workspace/src/github.com/codegangsta/cli"
	"github.com/jfrogdev/jfrog-cli-go/artifactory"
	"github.com/jfrogdev/jfrog-cli-go/bintray"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = "jfrog"
	app.Usage = "See https://github.com/jfrogdev/jfrog-cli-go for usage instructions."
	app.Version = cliutils.GetVersion()

	args := os.Args

	if showFrogCommands(args) {
		app.Commands = getCommands()
		app.Run(args)
	} else if args[1] == cliutils.CmdArtifactory {
		app.Commands = artifactory.GetCommands()
		app.Run(args[1:])
	} else if args[1] == cliutils.CmdBintray {
		app.Commands = bintray.GetCommands()
		app.Run(args[1:])
	} else {
		cliutils.Exit(cliutils.ExitCodeError, "Unknown command " + args[1] +
		    ". Expecting " + cliutils.CmdArtifactory + " or " + cliutils.CmdBintray + ".")
	}
}

func showFrogCommands(args []string) bool {
	if len(args) == 1 {
		return true
	}
	if args[1] != cliutils.CmdArtifactory && args[1] != cliutils.CmdBintray {
		return true
	}
	return false
}

func getCommands() []cli.Command {
	return []cli.Command{
		{
			Name:  cliutils.CmdArtifactory,
			Usage: "Artifactory commands",
		},
		{
			Name:  cliutils.CmdBintray,
			Usage: "Bintray commands",
		},
	}
}
