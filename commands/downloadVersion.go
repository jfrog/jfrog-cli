package commands

import (
    "github.com/JFrogDev/bintray-cli-go/utils"
)

func DownloadVersion(flags *DownloadVersionFlags) {
    path := ""
    println("path: " + path)
    resp, body := utils.SendGet(path, nil, flags.BintrayDetails.User, flags.BintrayDetails.Key)
    println(resp.Status)
    println(string(body))
}

type DownloadVersionFlags struct {
    BintrayDetails *utils.BintrayDetails
    Repo string
    Package string
    Version string
}