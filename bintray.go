package main

import (
    "com.jfrog/bintray/cli/client"
    "os"
    "github.com/codegangsta/cli"
    "encoding/json"
    "fmt"
    "com.jfrog/bintray/cli/command"
    "log"
    "reflect"
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
            Flags: append(globFlags, []cli.Flag{
                cli.StringFlag{
                    Name: "path",
                    Usage: "File or folder path to upload",
                },
                cli.StringFlag{
                    Name: "type",
                    Usage: "The repository type, used for automatic repo search/create. "+
                    "Can be one of: debian, rpm, docker, maven, generic",
                },
                cli.StringFlag{
                    Name: "repo",
                    Usage: "repository name",
                    Value: "generic",
                },
                cli.StringFlag{
                    Name: "package",
                    Usage: "package name",
                    Value: "default",
                },
                cli.StringFlag{
                    Name: "version",
                    Usage: "version name",
                    Value: "default",
                },
                cli.StringFlag{
                    Name: "publish",
                    Usage: "auto publish the uploaded files (0/[1])",
                    Value: "0",
                },
                cli.IntFlag{
                    Name: "parallel",
                    Usage: "The amount concurrenct uploads (default is 10)",
                    Value: 10,
                },
            }...),
            Action: func(c *cli.Context) {
                bt := newClient(c)
                results, err := command.Upload{}.Execute(bt)
                if err != nil {
                    log.Panicf("Upload failed %s\n", err)
                }
                /*ress := []*command.Result(results)
                for _, res := range ress {
                    fmt.Printf("RES: %s\n", res)
                }*/
                //cannot range over results (type interface {})
                s := reflect.ValueOf(results)
                for i := 0; i < s.Len(); i++ {
                    fmt.Printf("RES: %s\n", s.Index(i))
                }
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
        panic("Bintray username not set")
    }
    apiKey := c.String(flagApiKey)
    if apiKey == "" {
        panic("Bintray API key not set")
    }
    apiUrl := c.String(flagApiUrl)
    flags := makeFlagsMap(c)
    return client.New(username, apiKey, apiUrl, flags)
}

func makeFlagsMap(c *cli.Context) map[string]string {
    keyVals := make(map[string]string, len(c.FlagNames()))
    for _, name := range c.FlagNames() {
        keyVals[name] = c.String(name)
        log.Printf("flag: %s: %s\n", name, c.String(name))
    }
    subject := c.String(flagSubject)
    if (subject == "") {
        keyVals["subject"] = c.String(flagUsername)
    }
    return keyVals
}