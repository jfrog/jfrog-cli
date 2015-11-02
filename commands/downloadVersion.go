package commands

import (
    "encoding/json"
    "github.com/JFrogDev/bintray-cli-go/utils"
)

func DownloadVersion(flags *DownloadVersionFlags) {
    path := flags.BintrayDetails.ApiUrl + "packages/" + flags.BintrayDetails.Org + "/" +
        flags.Repo + "/" + flags.Package + "/versions/" + flags.Version + "/files"
    resp, body := utils.SendGet(path, nil, flags.BintrayDetails.User, flags.BintrayDetails.Key)
    if resp.StatusCode != 200 {
        utils.Exit(resp.Status + ". " + utils.ReadBintrayMessage(body))
    }
    var results []VersionFilesResult
    err := json.Unmarshal(body, &results)
    utils.CheckError(err)

    for _, result := range results {
        downloadFile(result.Path, flags)
    }
}

func downloadFile(path string, flags *DownloadVersionFlags) {
    url := flags.BintrayDetails.DownloadServerUrl + flags.BintrayDetails.Org + "/" + flags.Repo + "/" + path
    println("Downloading " + url)
    resp := utils.DownloadFile(url, flags.BintrayDetails.User, flags.BintrayDetails.Key)
    println("Bintray response: " + resp.Status)
}

type VersionFilesResult struct {
    Path string
}

type DownloadVersionFlags struct {
    BintrayDetails *utils.BintrayDetails
    Repo string
    Package string
    Version string
}