package main

import (
    "com.jfrog/bintray/cli/client"
    "os"
    "github.com/codegangsta/cli"
    "encoding/json"
    "fmt"
    "com.jfrog/bintray/cli/command"
    "log"
    "com.jfrog/bintray/cli/btr/flags"
)

const flagUsername = "username"
const flagApiKey = "apikey"
const flagApiUrl = "apiUrl"
const flagSubject = "subject"

var globFlags = []cli.Flag{
    cli.StringFlag{
        Name: flagUsername,
        EnvVar: "BINTRAY_USER",
        Usage: "Bintray username",
    },
    cli.StringFlag{
        Name: flagApiKey,
        EnvVar: "BINTRAY_KEY",
        Usage: "Bintray API key",
    },
    cli.StringFlag{
        Name: flagApiUrl,
        EnvVar: "BINTRAY_API_URL",
        Usage: "Bintray API url",
    },
    cli.StringFlag{
        Name: flagSubject,
        EnvVar: "BINTRAY_ORG",
        Usage: "Optional subject",
    },
}

func main() {
    app := cli.NewApp()
    app.Name = "btray"
    app.Usage = "task list on the command line"

    app.Commands = []cli.Command{
        {
            Name:    "upload",
            ShortName: "up",
            Usage:   "upload files",
            Flags: append(globFlags, flags.UploadFlags{}.Get() ...),
            Action: func(c *cli.Context) {
                bt := newClient(c)
                args := &command.UploadArgs{Parallel: uint32(c.Int("parallel")), FilePath: c.String("path"),
                    Subject: bt.Subject(), Repo: c.String("repo"), Pkg: c.String("package"),
                    Version: c.String("version"), Publish: c.Bool("publish")}
                upload := command.Upload{}
                upload.Execute(bt, args)
            },

        },
        {
            Name:    "search-repos",
            ShortName: "sr",
            Usage:   "find repositories",
            Flags: globFlags,
            Action: func(c *cli.Context) {
                bt := newClient(c)
                repos, _ := command.GetRepos{}.Execute(bt)
                buf, _ := json.MarshalIndent(repos, "", "  ")
                fmt.Printf("%s\n", buf)
            },

        },
    }
    app.Run(os.Args)

    /*(&cli.App{
        Flags: []cli.Flag{
            cli.StringFlag{Name: "count, c", EnvVar: "COMPAT_COUNT,APP_COUNT"},
        },
        Action: func(ctx *cli.Context) {
            if ctx.String("count") != "20" {
                t.Errorf("main name not set")
            }
            if ctx.String("c") != "20" {
                t.Errorf("short name not set")
            }
        },
    }).Run([]string{"run"})*/
}

func newClient(c *cli.Context) *client.Bintray {
    username := c.String(flagUsername)
    if username == "" {
        log.Panic("Bintray username not set")
    }
    apiKey := c.String(flagApiKey)
    if apiKey == "" {
        log.Panic("Bintray API key not set")
    }
    apiUrl := c.String(flagApiUrl)
    flags := makeFlagsMap(c)
    return client.New(username, apiKey, apiUrl, flags)
}

func makeFlagsMap(c *cli.Context) map[string]string {
    keyVals := make(map[string]string, len(c.FlagNames()))
    for _, name := range c.FlagNames() {
        keyVals[name] = c.String(name)
        //log.Printf("flag: %s: %s\n", name, c.String(name))
    }
    subject := c.String(flagSubject)
    if subject == "" {
        keyVals["subject"] = c.String(flagUsername)
    }
    return keyVals
}