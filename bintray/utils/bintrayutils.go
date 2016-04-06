package utils

import (
	"encoding/json"
	"fmt"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	"strings"
)

func BuildDownloadBintrayFileUrl(bintrayDetails *config.BintrayDetails,
    pathDetails *PathDetails) string {

	downloadPath := pathDetails.Subject + "/" + pathDetails.Repo + "/" +
	    pathDetails.Path
	return bintrayDetails.DownloadServerUrl + downloadPath
}

func DownloadBintrayFile(bintrayDetails *config.BintrayDetails, pathDetails *PathDetails,
	flags *DownloadFlags, logMsgPrefix string) {

	url := BuildDownloadBintrayFileUrl(bintrayDetails, pathDetails)
	fmt.Println(logMsgPrefix + "Downloading " + url)

	fileName, dir := ioutils.GetFileAndDirFromPath(pathDetails.Path)
	httpClientsDetails := GetBintrayHttpClientDetails(bintrayDetails)
	details := ioutils.GetRemoteFileDetails(url, httpClientsDetails)
	path := pathDetails.Path
	if flags.Flat {
	    path, _ = ioutils.GetFileAndDirFromPath(path)
	}
	if !shouldDownloadFile(path, details) {
	    fmt.Println(logMsgPrefix + "File already exists locally.")
		return
	}
	if flags.Flat {
		dir = ""
	}

    // Check if the file should be downloaded concurrently.
	if flags.SplitCount == 0 || flags.MinSplitSize < 0 || flags.MinSplitSize*1000 > details.Size {
	    // File should not be downloaded concurrently. Download it as one block.
		resp := ioutils.DownloadFile(url, dir, fileName, false, httpClientsDetails)
		fmt.Println(logMsgPrefix + "Bintray response: " + resp.Status)
	} else {
	    // We should attempt to download the file concurrently, but only if it is provided through the DSN.
	    // To check if the file is provided through the DSN, we first attempt to download the file
	    // with 'follow redirect' disabled.
	    resp, redirectUrl, err :=
	    ioutils.DownloadFileNoRedirect(url, dir, fileName, false, httpClientsDetails)
        // There are two options now. Either the file has just been downloaded as one block, or
        // we got a redirect to DSN download URL. In case of the later, we should download the file
        // concurrently from the DSN URL.
        // 'err' is not nil in case 'redirectUrl' was returned.
        if redirectUrl != "" {
            concurrentDownloadFlags := ioutils.ConcurrentDownloadFlags{
                DownloadPath: redirectUrl,
                FileName:     fileName,
                LocalPath:    dir,
                FileSize:     details.Size,
                SplitCount:   flags.SplitCount,
                Flat:         flags.Flat}

		ioutils.DownloadFileConcurrently(concurrentDownloadFlags, "", httpClientsDetails)
        } else {
            cliutils.CheckError(err)
            fmt.Println(logMsgPrefix + "Bintray response: " + resp.Status)
        }
	}
}

func shouldDownloadFile(localFilePath string, remoteFileDetails *ioutils.FileDetails) bool {
	if !ioutils.IsFileExists(localFilePath) {
		return true
	}
	localFileDetails := ioutils.GetFileDetails(localFilePath)
	if localFileDetails.Sha1 != remoteFileDetails.Sha1 {
		return true
	}
	return false
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
		cliutils.Exit(cliutils.ExitCodeError, "Unexpected format for argument: "+versionStr)
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
	return &VersionDetails{
		Subject: subject,
		Repo:    repo,
		Package: pkg,
		Version: version}
}

func CreatePackageDetails(packageStr string) *VersionDetails {
	parts := strings.Split(packageStr, "/")
	size := len(parts)
	if size != 3 {
		cliutils.Exit(cliutils.ExitCodeError, "Expecting an argument in the form of subject/repository/package")
	}
	return &VersionDetails{
		Subject: parts[0],
		Repo:    parts[1],
		Package: parts[2]}
}

func CreatePathDetails(str string) *PathDetails {
	parts := strings.Split(str, "/")
	size := len(parts)
	if size < 3 {
		cliutils.Exit(cliutils.ExitCodeError, "Expecting an argument in the form of subject/repository/file-path")
	}
	path := strings.Join(parts[2:], "/")

	return &PathDetails{
		Subject: parts[0],
		Repo:    parts[1],
		Path:    path}
}

func GetBintrayHttpClientDetails (bintrayDetails *config.BintrayDetails) ioutils.HttpClientDetails{
	return ioutils.HttpClientDetails{
		User:     bintrayDetails.User,
		Password: bintrayDetails.Key}
}

type bintrayResponse struct {
	Message string
}

type FileDetails struct {
	Sha1 string
	Size int64
}

type PathDetails struct {
	Subject string
	Repo    string
	Path    string
}

type VersionDetails struct {
	Subject string
	Repo    string
	Package string
	Version string
}

type DownloadFlags struct {
	BintrayDetails *config.BintrayDetails
	Threads        int
	MinSplitSize   int64
	SplitCount     int
	Flat           bool
}
