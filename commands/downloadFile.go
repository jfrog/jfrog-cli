package commands

import (
    "github.com/JFrogDev/bintray-cli-go/utils"
)

func DownloadFile(versionDetails *utils.VersionDetails, path string, bintrayDetails *utils.BintrayDetails) {
    utils.DownloadBintrayFile(bintrayDetails, versionDetails, path)
}

type DownloadFileFlags struct {
    BintrayDetails *utils.BintrayDetails
    Repo string
}