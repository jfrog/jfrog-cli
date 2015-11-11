package commands

import (
    "github.com/JFrogDev/bintray-cli-go/utils"
)

func DownloadFile(path string, flags *DownloadFileFlags) {
    utils.DownloadBintrayFile(flags.BintrayDetails, flags.Repo, path)
}

type DownloadFileFlags struct {
    BintrayDetails *utils.BintrayDetails
    Repo string
}