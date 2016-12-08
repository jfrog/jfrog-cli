package commands

import (
	"encoding/json"
	"github.com/jfrogdev/jfrog-cli-go/bintray/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	"strings"
	"errors"
	"sync"
	"strconv"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/log"
)

func DownloadVersion(versionDetails *utils.VersionDetails, targetPath string, flags *utils.DownloadFlags) (totalDownloded, totalFailed int, err error) {
	ioutils.CreateTempDirPath()
	defer ioutils.RemoveTempDir()

	if flags.BintrayDetails.User == "" {
		flags.BintrayDetails.User = versionDetails.Subject
	}
	path := BuildDownloadVersionUrl(versionDetails, flags.BintrayDetails, flags.IncludeUnpublished)
	httpClientsDetails := utils.GetBintrayHttpClientDetails(flags.BintrayDetails)
	resp, body, _, _ := ioutils.SendGet(path, true, httpClientsDetails)
	if resp.StatusCode != 200 {
		err = cliutils.CheckError(errors.New(resp.Status + ". " + utils.ReadBintrayMessage(body)))
		return
	}
	var results []VersionFilesResult
	err = json.Unmarshal(body, &results)
	if cliutils.CheckError(err) != nil {
		return
	}

	totalDownloded, err = downloadFiles(results, versionDetails, targetPath, flags)
	log.Info("Downloaded", strconv.Itoa(totalDownloded), "artifacts.")
	totalFailed = len(results) - totalDownloded
	return
}

func BuildDownloadVersionUrl(versionDetails *utils.VersionDetails, bintrayDetails *config.BintrayDetails, includeUnpublished bool) string {
	urlPath := bintrayDetails.ApiUrl + "packages/" + versionDetails.Subject + "/" +
			versionDetails.Repo + "/" + versionDetails.Package + "/versions/" + versionDetails.Version + "/files"
	if includeUnpublished {
		urlPath += "?include_unpublished=1"
	}
	return urlPath
}

func downloadFiles(results []VersionFilesResult, versionDetails *utils.VersionDetails, targetPath string,
flags *utils.DownloadFlags) (totalDownloaded int, err error) {

	size := len(results)
	downloadedForThread := make([]int, flags.Threads)
	var wg sync.WaitGroup
	for i := 0; i < flags.Threads; i++ {
		wg.Add(1)
		go func(threadId int) {
			logMsgPrefix := cliutils.GetLogMsgPrefix(threadId, false)
			for j := threadId; j < size; j += flags.Threads {
				pathDetails := &utils.PathDetails{
					Subject: versionDetails.Subject,
					Repo:    versionDetails.Repo,
					Path:    results[j].Path}

				e := utils.DownloadBintrayFile(flags.BintrayDetails, pathDetails, targetPath,
					flags, logMsgPrefix)
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

func CreateVersionDetailsForDownloadVersion(versionStr string) (*utils.VersionDetails, error) {
	parts := strings.Split(versionStr, "/")
	if len(parts) != 4 {
		err := cliutils.CheckError(errors.New("Argument format should be subject/repository/package/version. Got " + versionStr))
		if err != nil {
			return nil, err
		}
	}
	return utils.CreateVersionDetails(versionStr)
}

type VersionFilesResult struct {
	Path string
}
