package utils

import (
    "os"
    "fmt"
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
    fmt.Println(msg)
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

func CreateBintrayPath(details *VersionDetails) string {
    if details.Version == "" {
        if details.Package == "" {
            return "repos/" + details.Subject + "/" + details.Repo
        }
        return "packages/" + details.Subject + "/" + details.Repo + "/" + details.Package
    } else {
        return "packages/" + details.Subject + "/" + details.Repo + "/" + details.Package +
            "/versions/" + details.Version
    }
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

func IndentJson(jsonStr []byte) string {
    var content bytes.Buffer
    err := json.Indent(&content, jsonStr, "", "  ")
    if err == nil {
        return content.String()
    }
    return string(jsonStr)
}

// Creates a string in the form of ["item-1","item-2","item-3"...] from an input
// in the form of item-1,item-1,item-1...
func BuildListString(listStr string) string {
    if listStr == "" {
        return ""
    }
    split := strings.Split(listStr, ",")
    size := len(split)
    str := "[\""
    for i := 0; i < size; i++ {
        str += split[i]
        if i+1 < size {
            str += "\",\""
        }
    }
    str += "\"]"
    return str
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