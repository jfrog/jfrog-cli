package commands

import (
    "encoding/json"
    "github.com/JFrogDev/bintray-cli-go/utils"
)

func DownloadVersion(versionDetails *utils.VersionDetails, flags *utils.BintrayDetails) {
    path := flags.ApiUrl + "packages/" + versionDetails.Subject + "/" +
        versionDetails.Repo + "/" + versionDetails.Package + "/versions/" + versionDetails.Version + "/files"
    resp, body := utils.SendGet(path, nil, flags.User, flags.Key)
    if resp.StatusCode != 200 {
        utils.Exit(resp.Status + ". " + utils.ReadBintrayMessage(body))
    }
    var results []VersionFilesResult
    err := json.Unmarshal(body, &results)
    utils.CheckError(err)

    for _, result := range results {
        utils.DownloadBintrayFile(flags, versionDetails, result.Path)
    }
}

type VersionFilesResult struct {
    Path string
}

type DownloadVersionFlags struct {
    BintrayDetails *utils.BintrayDetails
    Repo string
    Package string
}