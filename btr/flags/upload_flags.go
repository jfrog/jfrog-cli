package flags

import (
    "time"
    "github.com/codegangsta/cli"
)

type UploadFlags struct {
}

func (cmd UploadFlags) Get() []cli.Flag {
    return []cli.Flag{
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
    }
}

type Repository struct {
    Name         string     `json:"name"`
    Owner        string     `json:"owner"`
    Type         string     `json:"type"`
    Private      bool       `json:"private"`
    Premium      string     `json:"premium"`
    Description  string     `json:"desc"`
    Labels       []string   `json:"labels"`
    Created      time.Time  `json:"created"`
    PackageCount int        `json:"package_count`
}
