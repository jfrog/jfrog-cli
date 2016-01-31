package commands

import (
    "sync"
    "strings"
    "encoding/json"
    "github.com/JFrogDev/jfrog-cli-go/cliutils"
    "github.com/JFrogDev/jfrog-cli-go/bintray/utils"
)

func DownloadVersion(versionDetails *utils.VersionDetails, bintrayDetails *utils.BintrayDetails,
    threads int) {

    if bintrayDetails.User == "" {
        bintrayDetails.User = versionDetails.Subject
    }
    path := bintrayDetails.ApiUrl + "packages/" + versionDetails.Subject + "/" +
        versionDetails.Repo + "/" + versionDetails.Package + "/versions/" + versionDetails.Version + "/files"
    resp, body := cliutils.SendGet(path, nil, bintrayDetails.User, bintrayDetails.Key)
    if resp.StatusCode != 200 {
        cliutils.Exit(cliutils.ExitCodeError, resp.Status + ". " + utils.ReadBintrayMessage(body))
    }
    var results []VersionFilesResult
    err := json.Unmarshal(body, &results)
    cliutils.CheckError(err)

    downloadFiles(results, versionDetails, bintrayDetails, threads)
}

func downloadFiles(results []VersionFilesResult, versionDetails *utils.VersionDetails,
    bintrayDetails *utils.BintrayDetails, threads int) {

    size := len(results)
    var wg sync.WaitGroup
    for i := 0; i < threads; i++ {
        wg.Add(1)
        go func(threadId int) {
            logMsgPrefix := cliutils.GetLogMsgPrefix(threadId, false)
            for j := threadId; j < size; j += threads {
                utils.DownloadBintrayFile(bintrayDetails, versionDetails, results[j].Path, logMsgPrefix)
            }
            wg.Done()
        }(i)
    }
    wg.Wait()
}

func CreateVersionDetailsForDownloadVersion(versionStr string) *utils.VersionDetails {
    parts := strings.Split(versionStr, "/")
    if len(parts) != 4 {
        cliutils.Exit(cliutils.ExitCodeError, "Argument format should be subject/repository/package/version. Got " + versionStr)
    }
    return utils.CreateVersionDetails(versionStr)
}

type VersionFilesResult struct {
    Path string
}