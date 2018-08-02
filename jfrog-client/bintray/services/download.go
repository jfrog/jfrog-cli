package services

import (
	"encoding/json"
	"errors"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/bintray/auth"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/bintray/services/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/bintray/services/versions"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/httpclient"
	clientutils "github.com/jfrog/jfrog-cli-go/jfrog-client/utils"
	logutil "github.com/jfrog/jfrog-cli-go/jfrog-client/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/errorutils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/io/fileutils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/log"
	"net/http"
	"path"
	"strconv"
	"strings"
	"sync"
	"path/filepath"
)

func NewDownloadService(client *httpclient.HttpClient) *DownloadService {
	ds := &DownloadService{client: client}
	return ds
}

func NewDownloadFileParams() *DownloadFileParams {
	return &DownloadFileParams{PathDetails: &utils.PathDetails{}}
}

func NewDownloadVersionParams() *DownloadVersionParams {
	return &DownloadVersionParams{Params: &versions.Params{}}
}

type DownloadService struct {
	client         *httpclient.HttpClient
	BintrayDetails auth.BintrayDetails
	Threads        int
	MinSplitSize   int64
	SplitCount     int
}

type DownloadFileParams struct {
	*utils.PathDetails
	TargetPath         string
	IncludeUnpublished bool
	Flat               bool
}

type DownloadVersionParams struct {
	*versions.Params
	TargetPath         string
	IncludeUnpublished bool
}

func (ds *DownloadService) DownloadFile(downloadParams *DownloadFileParams) (totalDownloded, totalFailed int, err error) {
	err = fileutils.CreateTempDirPath()
	if err != nil {
		return 0, 1, err
	}
	defer fileutils.RemoveTempDir()

	if ds.BintrayDetails.GetUser() == "" {
		ds.BintrayDetails.SetUser(downloadParams.Subject)
	}

	err = ds.downloadBintrayFile(downloadParams, "")
	if err != nil {
		return 0, 1, err
	}
	log.Info("Downloaded 1 artifact.")
	return 1, 0, nil
}

func (ds *DownloadService) DownloadVersion(downloadParams *DownloadVersionParams) (totalDownloded, totalFailed int, err error) {
	err = fileutils.CreateTempDirPath()
	if err != nil {
		return
	}
	defer fileutils.RemoveTempDir()

	versionPathUrl := buildDownloadVersionUrl(ds.BintrayDetails.GetApiUrl(), downloadParams)
	httpClientsDetails := ds.BintrayDetails.CreateHttpClientDetails()
	if httpClientsDetails.User == "" {
		httpClientsDetails.User = downloadParams.Subject
	}
	client := httpclient.NewDefaultHttpClient()
	resp, body, _, _ := client.SendGet(versionPathUrl, true, httpClientsDetails)
	if resp.StatusCode != http.StatusOK {
		err = errorutils.CheckError(errors.New(resp.Status + ". " + utils.ReadBintrayMessage(body)))
		return
	}
	var files []VersionFilesResult
	err = json.Unmarshal(body, &files)
	if errorutils.CheckError(err) != nil {
		return
	}

	totalDownloded, err = ds.downloadVersionFiles(files, downloadParams)
	log.Info("Downloaded", strconv.Itoa(totalDownloded), "artifacts.")
	totalFailed = len(files) - totalDownloded
	return
}

func buildDownloadVersionUrl(apiUrl string, downloadParams *DownloadVersionParams) string {
	urlPath := apiUrl + path.Join("packages/", downloadParams.Subject, downloadParams.Repo, downloadParams.Package, "versions", downloadParams.Version, "files")
	if downloadParams.IncludeUnpublished {
		urlPath += "?include_unpublished=1"
	}
	return urlPath
}

func (ds *DownloadService) downloadVersionFiles(files []VersionFilesResult, downloadParams *DownloadVersionParams) (totalDownloaded int, err error) {
	size := len(files)
	downloadedForThread := make([]int, ds.Threads)
	var wg sync.WaitGroup
	for i := 0; i < ds.Threads; i++ {
		wg.Add(1)
		go func(threadId int) {
			logMsgPrefix := logutil.GetLogMsgPrefix(threadId, false)
			for j := threadId; j < size; j += ds.Threads {
				pathDetails := &utils.PathDetails{
					Subject: downloadParams.Subject,
					Repo:    downloadParams.Repo,
					Path:    files[j].Path}

				downloadFileParams := &DownloadFileParams{PathDetails: pathDetails, TargetPath: downloadParams.TargetPath}
				e := ds.downloadBintrayFile(downloadFileParams, logMsgPrefix)
				if e != nil {
					err = e
					continue
				}
				downloadedForThread[threadId]++
			}
			wg.Done()
		}(i)
	}
	wg.Wait()

	for i := range downloadedForThread {
		totalDownloaded += downloadedForThread[i]
	}
	return
}

func CreateVersionDetailsForDownloadVersion(versionStr string) (*versions.Path, error) {
	parts := strings.Split(versionStr, "/")
	if len(parts) != 4 {
		err := errorutils.CheckError(errors.New("Argument format should be subject/repository/package/version. Got " + versionStr))
		if err != nil {
			return nil, err
		}
	}
	return versions.CreatePath(versionStr)
}

type VersionFilesResult struct {
	Path string
}

func (ds *DownloadService) downloadBintrayFile(downloadParams *DownloadFileParams, logMsgPrefix string) error {
	cleanPath := strings.Replace(downloadParams.Path, "(", "", -1)
	cleanPath = strings.Replace(cleanPath, ")", "", -1)
	downloadPath := path.Join(downloadParams.Subject, downloadParams.Repo, cleanPath)

	fileName, filePath := fileutils.GetFileAndDirFromPath(cleanPath)

	url := ds.BintrayDetails.GetDownloadServerUrl() + downloadPath
	if downloadParams.IncludeUnpublished {
		url += "?include_unpublished=1"
	}
	log.Info(logMsgPrefix+"Downloading", downloadPath)
	client := httpclient.NewDefaultHttpClient()

	httpClientsDetails := ds.BintrayDetails.CreateHttpClientDetails()
	details, err := client.GetRemoteFileDetails(url, httpClientsDetails)
	if err != nil {
		return errorutils.CheckError(errors.New("Bintray " + err.Error()))
	}

	regexpPattern := clientutils.PathToRegExp(downloadParams.Path)
	placeHolderTarget, err := clientutils.ReformatRegexp(regexpPattern, cleanPath, downloadParams.TargetPath)
	if err != nil {
		return err
	}

	localPath, localFileName := fileutils.GetLocalPathAndFile(fileName, filePath, placeHolderTarget, downloadParams.Flat)
	var shouldDownload bool
	shouldDownload, err = shouldDownloadFile(filepath.Join(localPath, localFileName), details)
	if err != nil {
		return err
	}
	if !shouldDownload {
		log.Info(logMsgPrefix, "File already exists locally.")
		return nil
	}

	// Check if the file should be downloaded concurrently.
	if ds.SplitCount == 0 || ds.MinSplitSize < 0 || ds.MinSplitSize*1000 > details.Size {
		// File should not be downloaded concurrently. Download it as one block.
		downloadDetails := &httpclient.DownloadFileDetails{
			FileName:      fileName,
			DownloadPath:  url,
			LocalPath:     localPath,
			LocalFileName: localFileName}

		resp, err := client.DownloadFile(downloadDetails, logMsgPrefix, httpClientsDetails, 0, false)
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
			client.DownloadFileNoRedirect(url, localPath, localFileName, httpClientsDetails)
		// There are two options now. Either the file has just been downloaded as one block, or
		// we got a redirect to DSN download URL. In case of the later, we should download the file
		// concurrently from the DSN URL.
		// 'err' is not nil in case 'redirectUrl' was returned.
		if redirectUrl != "" {
			err = nil
			concurrentDownloadFlags := httpclient.ConcurrentDownloadFlags{
				DownloadPath: redirectUrl,
				FileName:     localFileName,
				LocalPath:    localPath,
				FileSize:     details.Size,
				SplitCount:   ds.SplitCount,
				Flat:         downloadParams.Flat}
			err = client.DownloadFileConcurrently(concurrentDownloadFlags, "", httpClientsDetails)
			if err != nil {
				return err
			}
		} else {
			if errorutils.CheckError(err) != nil {
				return err
			}
			log.Debug(logMsgPrefix, "Bintray response:", resp.Status)
		}
	}
	return nil
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
