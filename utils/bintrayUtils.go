package utils

import (
	"fmt"
    "strings"
    "encoding/json"
)

func DownloadBintrayFile(bintrayDetails *BintrayDetails, versionDetails *VersionDetails, path string) {
    url := bintrayDetails.DownloadServerUrl + versionDetails.Subject + "/" + versionDetails.Repo + "/" + path
    fmt.Println("Downloading " + url)
    resp := DownloadFile(url, bintrayDetails.User, bintrayDetails.Key)
    fmt.Println("Bintray response: " + resp.Status)
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
        Exit("Unexpected format for argument: " + versionStr)
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
        Exit("Expecting an argument in the form of subject/repository/package")
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
        Exit("Expecting an argument in the form of subject/repository/package/version/path")
    }
    versionDetails = &VersionDetails {
        Subject: parts[0],
        Repo: parts[1],
        Package: parts[2],
        Version: parts[3]}

    for i := 4; i < size; i++ {
        path += parts[i]
        if i+1 < size {
            path += "/"
        }
    }
    return
}

type bintrayResponse struct {
    Message string
}

type VersionDetails struct {
    Subject string
    Repo string
    Package string
    Version string
}

type BintrayDetails struct {
    ApiUrl string
    DownloadServerUrl string
    User string
    Key string
}