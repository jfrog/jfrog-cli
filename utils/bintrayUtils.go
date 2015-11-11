package utils

func DownloadBintrayFile(details *BintrayDetails, repo, path string) {
    url := details.DownloadServerUrl + details.Org + "/" + repo + "/" + path
    println("Downloading " + url)
    resp := DownloadFile(url, details.User, details.Key)
    println("Bintray response: " + resp.Status)
}