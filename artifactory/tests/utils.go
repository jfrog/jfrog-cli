package tests

import (
    "runtime"
    "github.com/jfrogdev/jfrog-cli-go/cliutils"
    "github.com/jfrogdev/jfrog-cli-go/artifactory/utils"
)

func GetFlags() *utils.Flags {
    flags := new(utils.Flags)
    flags.ArtDetails = new(cliutils.ArtifactoryDetails)
    flags.DryRun = true
    flags.EncPassword = true
    flags.Threads = 3

    return flags
}

func GetFileSeperator() string {
    if runtime.GOOS == "windows" {
        return "\\\\"
    }
    return "/"
}