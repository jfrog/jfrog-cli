package main

import (
    "os"
    "github.com/codegangsta/cli"
    "github.com/jFrogdev/jfrog-cli-go/artifactory"
    "github.com/jFrogdev/jfrog-cli-go/bintray"
    "github.com/jFrogdev/jfrog-cli-go/cliutils"
)

func main() {
    app := cli.NewApp()
    app.Name = "frog"
    app.Usage = "See https://github.com/jFrogdev/jfrog-cli-go for usage instructions."
    app.Version = "0.0.1"

    args := os.Args

    if showFrogCommands(args) {
        app.Commands = getCommands()
        app.Run(args)
    } else
    if args[1] == "art" {
        app.Commands = artifactory.GetCommands()
        app.Run(args[1:])
    } else
    if args[1] == "bt" {
        app.Commands = bintray.GetCommands()
        app.Run(args[1:])
    } else {
        cliutils.Exit(cliutils.ExitCodeError, "Unknown command " + args[1] + ". Expecting art or bt.")
    }
}

func showFrogCommands(args []string) bool {
    if len(args) == 1 {
        return true
    }
    if args[1] != "art" && args[1] != "bt" {
        return true
    }
    return false
}

func getCommands() []cli.Command {
    return []cli.Command{
        {
            Name: "art",
            Usage: "Artifactory commands",
        },
        {
            Name: "bt",
            Usage: "Bintray commands",
        },
    }
}