package utils

import (
    "os"
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

type bintrayResponse struct {
    Message string
}

type BintrayDetails struct {
    ApiUrl string
    DownloadServerUrl string
    Org string
    User string
    Key string
}