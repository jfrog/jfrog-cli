package commands

import (
    "encoding/json"
    "github.com/JFrogDev/bintray-cli-go/utils"
)

func DownloadVersion(version string, flags *DownloadVersionFlags) {
    path := flags.BintrayDetails.ApiUrl + "packages/" + flags.BintrayDetails.Org + "/" +
        flags.Repo + "/" + flags.Package + "/versions/" + version + "/files"
    resp, body := utils.SendGet(path, nil, flags.BintrayDetails.User, flags.BintrayDetails.Key)
    if resp.StatusCode != 200 {
        utils.Exit(resp.Status + ". " + utils.ReadBintrayMessage(body))
    }
    var results []VersionFilesResult
    err := json.Unmarshal(body, &results)
    utils.CheckError(err)

    for _, result := range results {
        utils.DownloadBintrayFile(flags.BintrayDetails, flags.Repo, result.Path)
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