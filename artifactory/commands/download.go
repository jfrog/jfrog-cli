package commands

import (
	"fmt"
	"github.com/jfrogdev/jfrog-cli-go/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	"strconv"
	"sync"
)

// Downloads the artifacts using the specified download pattern.
// Returns the AQL query used for the download.
func Download(downloadPattern string, flags *utils.Flags) int {
	utils.PreCommandSetup(flags)
	if !flags.DryRun {
		ioutils.CreateTempDirPath()
		defer ioutils.RemoveTempDir()
	}

	resultItems := utils.AqlSearch(downloadPattern, flags)
	downloadFiles(resultItems, flags)

	fmt.Println("Downloaded " + strconv.Itoa(len(resultItems)) + " artifacts from Artifactory.")
	return len(resultItems)
}

func downloadFiles(resultItems []utils.AqlSearchResultItem, flags *utils.Flags) {
	size := len(resultItems)
	var wg sync.WaitGroup
	for i := 0; i < flags.Threads; i++ {
		wg.Add(1)
		go func(threadId int) {
			logMsgPrefix := cliutils.GetLogMsgPrefix(threadId, flags.DryRun)
			for j := threadId; j < size; j += flags.Threads {
				downloadPath := flags.ArtDetails.Url + resultItems[j].GetFullUrl()
				fmt.Println(logMsgPrefix + "Downloading " + downloadPath)
				if !flags.DryRun {
					downloadFile(downloadPath, resultItems[j].Path, resultItems[j].Name, logMsgPrefix, flags)
				}
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func downloadFile(downloadPath, localPath, localFileName, logMsgPrefix string, flags *utils.Flags) {
	httpClientsDetails := utils.GetArtifactoryHttpClientDetails(flags.ArtDetails)
	details := ioutils.GetRemoteFileDetails(downloadPath, httpClientsDetails)
	localFilePath := localFileName
	if !flags.Flat {
        localFilePath = localPath + "/" + localFileName
	}

	if shouldDownloadFile(localFilePath, details) {
		if flags.SplitCount == 0 || flags.MinSplitSize < 0 || flags.MinSplitSize*1000 > details.Size || !details.AcceptRanges {
			resp := ioutils.DownloadFile(downloadPath, localPath, localFileName, flags.Flat, httpClientsDetails)
			fmt.Println(logMsgPrefix + "Artifactory response:", resp.Status)
		} else {
			concurrentDownloadFlags := ioutils.ConcurrentDownloadFlags{
				DownloadPath: downloadPath,
				FileName:     localFileName,
				LocalPath:    localPath,
				FileSize:     details.Size,
				SplitCount:   flags.SplitCount,
				Flat:         flags.Flat}

			ioutils.DownloadFileConcurrently(concurrentDownloadFlags, logMsgPrefix, httpClientsDetails)
		}
	} else {
		fmt.Println(logMsgPrefix + "File already exists locally.")
	}
}

func shouldDownloadFile(localFilePath string, artifactoryFileDetails *ioutils.FileDetails) bool {
	if !ioutils.IsFileExists(localFilePath) {
		return true
	}
	localFileDetails := ioutils.GetFileDetails(localFilePath)
	if localFileDetails.Md5 != artifactoryFileDetails.Md5 || localFileDetails.Sha1 != artifactoryFileDetails.Sha1 {
		return true
	}
	return false
}
