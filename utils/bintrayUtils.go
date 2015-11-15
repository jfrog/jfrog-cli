package utils

func DownloadBintrayFile(bintrayDetails *BintrayDetails, versionDetails *VersionDetails, path string) {
    url := bintrayDetails.DownloadServerUrl + versionDetails.Subject + "/" + versionDetails.Repo + "/" + path
    println("Downloading " + url)
    resp := DownloadFile(url, bintrayDetails.User, bintrayDetails.Key)
    println("Bintray response: " + resp.Status)
}