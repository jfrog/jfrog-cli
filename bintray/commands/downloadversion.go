package commands

import (
	"encoding/json"
	"github.com/jfrogdev/jfrog-cli-go/bintray/utils"
	"github.com/jfrogdev/jfrog-cli-go/cliutils"
	"strings"
	"sync"
)

func DownloadVersion(versionDetails *utils.VersionDetails, flags *utils.DownloadFlags) {
	cliutils.CreateTempDirPath()
	defer cliutils.RemoveTempDir()

	if flags.BintrayDetails.User == "" {
		flags.BintrayDetails.User = versionDetails.Subject
	}
	path := flags.BintrayDetails.ApiUrl + "packages/" + versionDetails.Subject + "/" +
		versionDetails.Repo + "/" + versionDetails.Package + "/versions/" + versionDetails.Version + "/files"
	resp, body, _, _ := cliutils.SendGet(path, nil, true, flags.BintrayDetails.User, flags.BintrayDetails.Key)
	if resp.StatusCode != 200 {
		cliutils.Exit(cliutils.ExitCodeError, resp.Status+". "+utils.ReadBintrayMessage(body))
	}
	var results []VersionFilesResult
	err := json.Unmarshal(body, &results)
	cliutils.CheckError(err)

	downloadFiles(results, versionDetails, flags)
}

func downloadFiles(results []VersionFilesResult, versionDetails *utils.VersionDetails,
	flags *utils.DownloadFlags) {

	size := len(results)
	var wg sync.WaitGroup
	for i := 0; i < flags.Threads; i++ {
		wg.Add(1)
		go func(threadId int) {
			logMsgPrefix := cliutils.GetLogMsgPrefix(threadId, false)
			for j := threadId; j < size; j += flags.Threads {
				utils.DownloadBintrayFile(flags.BintrayDetails, versionDetails, results[j].Path,
					flags, logMsgPrefix)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func CreateVersionDetailsForDownloadVersion(versionStr string) *utils.VersionDetails {
	parts := strings.Split(versionStr, "/")
	if len(parts) != 4 {
		cliutils.Exit(cliutils.ExitCodeError, "Argument format should be subject/repository/package/version. Got "+versionStr)
	}
	return utils.CreateVersionDetails(versionStr)
}

type VersionFilesResult struct {
	Path string
}
