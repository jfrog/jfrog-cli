package utils

import (
	"fmt"
    "strings"
    "encoding/json"
    "github.com/jfrogdev/jfrog-cli-go/cliutils"
)

func DownloadBintrayFile(bintrayDetails *cliutils.BintrayDetails, versionDetails *VersionDetails, path string,
    logMsgPrefix string) {

    if logMsgPrefix != "" {
        logMsgPrefix += " "
    }
    downloadPath := versionDetails.Subject + "/" + versionDetails.Repo + "/" + path
    url := bintrayDetails.DownloadServerUrl + downloadPath
    fmt.Println(logMsgPrefix + "Downloading " + downloadPath)
    fileName := cliutils.GetFileNameFromUrl(url)
    resp := cliutils.DownloadFile(url, "", fileName, true, bintrayDetails.User, bintrayDetails.Key)
    fmt.Println(logMsgPrefix + "Bintray response: " + resp.Status)
}

func ReadBintrayMessage(resp []byte) string {
    var response bintrayResponse
    err := json.Unmarshal(resp, &response)
    if err != nil {
        return string(resp)
    }
    return response.Message
}

func CreateVersionDetails(versionStr string) *VersionDetails {
    parts := strings.Split(versionStr, "/")
    size := len(parts)
    if size < 1 || size > 4 {
        cliutils.Exit(cliutils.ExitCodeError, "Unexpected format for argument: " + versionStr)
    }
    var subject, repo, pkg, version string
    if size >= 2 {
        subject = parts[0]
        repo = parts[1]
    }
    if size >= 3 {
        pkg = parts[2]
    }
    if size == 4 {
        version = parts[3]
    }
    return &VersionDetails {
        Subject: subject,
        Repo: repo,
        Package: pkg,
        Version: version}
}

func CreatePackageDetails(packageStr string) *VersionDetails {
    parts := strings.Split(packageStr, "/")
    size := len(parts)
    if size != 3 {
        cliutils.Exit(cliutils.ExitCodeError, "Expecting an argument in the form of subject/repository/package")
    }
    return &VersionDetails {
        Subject: parts[0],
        Repo: parts[1],
        Package: parts[2]}
}

func CreateVersionDetailsAndPath(versionStr string) (versionDetails *VersionDetails, path string) {
    parts := strings.Split(versionStr, "/")
    size := len(parts)
    if size < 4 {
        cliutils.Exit(cliutils.ExitCodeError, "Expecting an argument in the form of subject/repository/package/version/path")
    }
    versionDetails = &VersionDetails {
        Subject: parts[0],
        Repo: parts[1],
        Package: parts[2],
        Version: parts[3]}

    if size > 3 {
        path = strings.Join(parts[4:],"/")
    }
    return
}

func CreatePathDetails(str string) *PathDetails {
    parts := strings.Split(str, "/")
    size := len(parts)
    if size < 3 {
        cliutils.Exit(cliutils.ExitCodeError, "Expecting an argument in the form of subject/repository/file-path")
    }
    path := strings.Join(parts[2:],"/")

    return &PathDetails {
        Subject: parts[0],
        Repo: parts[1],
        Path: path}
}

type bintrayResponse struct {
    Message string
}

type PathDetails struct {
    Subject string
    Repo string
    Path string
}

type VersionDetails struct {
    Subject string
    Repo string
    Package string
    Version string
}