package utils

import (
	"fmt"
)

func DownloadBintrayFile(bintrayDetails *BintrayDetails, versionDetails *VersionDetails, path string) {
    url := bintrayDetails.DownloadServerUrl + versionDetails.Subject + "/" + versionDetails.Repo + "/" + path
    fmt.Println("Downloading " + url)
    resp := DownloadFile(url, bintrayDetails.User, bintrayDetails.Key)
    fmt.Println("Bintray response: " + resp.Status)
}