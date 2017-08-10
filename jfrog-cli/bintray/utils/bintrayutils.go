package utils

import (
	"errors"
	"encoding/json"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/httputils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/fileutils"
	"strings"
	"net/http"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
	"path"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
)

func DownloadBintrayFile(bintrayDetails *config.BintrayDetails, pathDetails *PathDetails, targetPath string,
	flags *DownloadFlags, logMsgPrefix string) (err error) {

	cleanPath := strings.Replace(pathDetails.Path, "(", "", -1)
	cleanPath = strings.Replace(cleanPath, ")", "", -1)
	downloadPath := path.Join(pathDetails.Subject, pathDetails.Repo, cleanPath)

	fileName, filePath := fileutils.GetFileAndDirFromPath(cleanPath)

	url := bintrayDetails.DownloadServerUrl + downloadPath
	if flags.IncludeUnpublished {
		url += "?include_unpublished=1"
	}
	log.Info(logMsgPrefix + "Downloading", downloadPath)

	httpClientsDetails := GetBintrayHttpClientDetails(bintrayDetails)
	var details *fileutils.FileDetails
	details, err = httputils.GetRemoteFileDetails(url, httpClientsDetails)
	if err != nil {
		err = errorutils.CheckError(errors.New("Bintray " + err.Error()))
		if err != nil {
			return
		}
	}

	regexpPattern := utils.PathToRegExp(pathDetails.Path)
	placeHolderTarget, e := utils.ReformatRegexp(regexpPattern, cleanPath, targetPath)
	if e != nil {
		err = e
		return
	}

	localPath, localFileName := fileutils.GetLocalPathAndFile(fileName, filePath, placeHolderTarget, flags.Flat)
	var shouldDownload bool
	shouldDownload, err = shouldDownloadFile(path.Join(localPath, localFileName), details)
	if err != nil {
		return
	}
	if !shouldDownload {
		log.Info(logMsgPrefix, "File already exists locally.")
		return
	}

	// Check if the file should be downloaded concurrently.
	if flags.SplitCount == 0 || flags.MinSplitSize < 0 || flags.MinSplitSize * 1000 > details.Size {
		// File should not be downloaded concurrently. Download it as one block.
		resp, err := httputils.DownloadFile(url, localPath, localFileName, httpClientsDetails)
		if err != nil {
			return err
		}
		log.Debug(logMsgPrefix, "Bintray response:", resp.Status)
	} else {
		// We should attempt to download the file concurrently, but only if it is provided through the DSN.
		// To check if the file is provided through the DSN, we first attempt to download the file
		// with 'follow redirect' disabled.

		var resp *http.Response
		var redirectUrl string
		resp, redirectUrl, err =
				httputils.DownloadFileNoRedirect(url, localPath, localFileName, httpClientsDetails)
		// There are two options now. Either the file has just been downloaded as one block, or
		// we got a redirect to DSN download URL. In case of the later, we should download the file
		// concurrently from the DSN URL.
		// 'err' is not nil in case 'redirectUrl' was returned.
		if redirectUrl != "" {
			concurrentDownloadFlags := httputils.ConcurrentDownloadFlags{
				DownloadPath: redirectUrl,
				FileName:     localFileName,
				LocalPath:    localPath,
				FileSize:     details.Size,
				SplitCount:   flags.SplitCount,
				Flat:         flags.Flat}

			httputils.DownloadFileConcurrently(concurrentDownloadFlags, "", httpClientsDetails)
		} else {
			if errorutils.CheckError(err) != nil {
				return
			}
			log.Info(logMsgPrefix, "Bintray response:", resp.Status)
		}
	}
	return
}

func shouldDownloadFile(localFilePath string, remoteFileDetails *fileutils.FileDetails) (bool, error) {
	exists, err := fileutils.IsFileExists(localFilePath)
	if err != nil {
		return false, err
	}
	if !exists {
		return true, nil
	}
	localFileDetails, err := fileutils.GetFileDetails(localFilePath)
	if err != nil {
		return false, err
	}
	return localFileDetails.Checksum.Sha1 != remoteFileDetails.Checksum.Sha1, nil
}

func ReadBintrayMessage(resp []byte) string {
	var response bintrayResponse
	err := json.Unmarshal(resp, &response)
	if err != nil {
		return string(resp)
	}
	return response.Message
}

func CreateVersionDetails(versionStr string) (*VersionDetails, error) {
	parts := strings.Split(versionStr, "/")
	size := len(parts)
	if size < 1 || size > 4 {
		err := errorutils.CheckError(errors.New("Unexpected format for argument: " + versionStr))
		if err != nil {
			return nil, err
		}
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
		Version: version}, nil
}

func CreatePackageDetails(packageStr string) (*VersionDetails, error) {
	parts := strings.Split(packageStr, "/")
	size := len(parts)
	if size != 3 {
		err := errorutils.CheckError(errors.New("Expecting an argument in the form of subject/repository/package"))
		if err != nil {
			return nil, err
		}
	}
	return &VersionDetails{
		Subject: parts[0],
		Repo:    parts[1],
		Package: parts[2]}, nil
}

func CreatePathDetails(str string) (*PathDetails, error) {
	parts := strings.Split(str, "/")
	size := len(parts)
	if size < 3 {
		err := errorutils.CheckError(errors.New("Expecting an argument in the form of subject/repository/file-path"))
		if err != nil {
			return nil, err
		}
	}
	path := strings.Join(parts[2:], "/")

	return &PathDetails{
		Subject: parts[0],
		Repo:    parts[1],
		Path:    path}, nil
}

func GetBintrayHttpClientDetails(bintrayDetails *config.BintrayDetails) httputils.HttpClientDetails {
	return httputils.HttpClientDetails{
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
	BintrayDetails     *config.BintrayDetails
	Threads            int
	MinSplitSize       int64
	SplitCount         int
	IncludeUnpublished bool
	Flat               bool
}
