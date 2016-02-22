package main

import (
	"github.com/codegangsta/cli"
	"github.com/jfrogdev/jfrog-cli-go/artifactory"
	"github.com/jfrogdev/jfrog-cli-go/bintray"
	"github.com/jfrogdev/jfrog-cli-go/cliutils"
	"os"
)

const cmdArtifactory = "arti"
const cmdBintray = "bt"

func main() {
	app := cli.NewApp()
	app.Name = "frog"
	app.Usage = "See https://github.com/jfrogdev/jfrog-cli-go for usage instructions."
	app.Version = "0.0.1"

	args := os.Args

	if showFrogCommands(args) {
		app.Commands = getCommands()
		app.Run(args)
	} else if args[1] == cmdArtifactory {
		app.Commands = artifactory.GetCommands()
		app.Run(args[1:])
	} else if args[1] == "bt" {
		app.Commands = bintray.GetCommands()
		app.Run(args[1:])
	} else {
		cliutils.Exit(cliutils.ExitCodeError, "Unknown command " + args[1] +
		    ". Expecting " + cmdArtifactory + " or " + cmdBintray + ".")
	}
}

func showFrogCommands(args []string) bool {
	if len(args) == 1 {
		return true
	}
	if args[1] != cmdArtifactory && args[1] != cmdBintray {
		return true
	}
	return false
}

func getCommands() []cli.Command {
	return []cli.Command{
		{
			Name:  cmdArtifactory,
			Usage: "Artifactory commands",
		},
		{
			Name:  cmdBintray,
			Usage: "Bintray commands",
		},
	}
}
