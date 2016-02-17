package commands

import (
    "github.com/jfrogdev/jfrog-cli-go/cliutils"
    "github.com/jfrogdev/jfrog-cli-go/bintray/utils"
)

func DownloadFile(versionDetails *utils.VersionDetails, path string, flags *utils.DownloadFlags) {
    cliutils.CreateTempDirPath()
    defer cliutils.RemoveTempDir()

    if flags.BintrayDetails.User == "" {
        flags.BintrayDetails.User = versionDetails.Subject
    }
    utils.DownloadBintrayFile(flags.BintrayDetails, versionDetails, path, flags, "")
}