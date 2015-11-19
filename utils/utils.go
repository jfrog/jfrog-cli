package utils

import (
    "os"
    "bytes"
    "strings"
    "encoding/json"
)

func CheckError(err error) {
    if err != nil {
        panic(err)
    }
}

func Exit(msg string) {
    println(msg)
    os.Exit(1)
}

func AddTrailingSlashIfNeeded(url string) string {
    if url != "" && !strings.HasSuffix(url, "/") {
        url += "/"
    }
    return url
}

func GetFileNameFromUrl(url string) string {
    parts := strings.Split(url, "/")
    size := len(parts)
    if size == 0 {
        return url
    }
    return parts[size-1]
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
    if len(parts) != 4 {
        Exit("Expecting an argument in the form of subject/repository/package/version")
    }
    return &VersionDetails {
        Subject: parts[0],
        Repo: parts[1],
        Package: parts[2],
        Version: parts[3]}
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

func IndentJson(jsonStr []byte) string {
    var content bytes.Buffer
    err := json.Indent(&content, jsonStr, "", "  ")
    if err == nil {
        return content.String()
    }
    return string(jsonStr)
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