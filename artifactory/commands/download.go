package commands

import (
	"fmt"
	"github.com/jfrogdev/jfrog-cli-go/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	"strconv"
	"sync"
	"strings"
)

// Downloads the artifacts using the specified download pattern.
// Returns the AQL query used for the download.
func Download(downloadPattern string, flags *utils.Flags) {
	utils.PreCommandSetup(flags)
	if !flags.DryRun {
		ioutils.CreateTempDirPath()
		defer ioutils.RemoveTempDir()
	}
	if utils.IsWildcardPattern(downloadPattern) {
		resultItems := utils.AqlSearch(downloadPattern, flags)
		downloadFiles(resultItems, flags)
		fmt.Println("Downloaded " + strconv.Itoa(len(resultItems)) + " artifacts from Artifactory.")
	} else {
		props := "";
		if flags.Props != "" {
			props = ";" + flags.Props
		}
		downloadPath := utils.BuildArtifactoryUrl(flags.ArtDetails.Url, downloadPattern + props, make(map[string]string))
		logMsgPrefix := cliutils.GetLogMsgPrefix(0, flags.DryRun)
		if !flags.DryRun {
			localPath, localFileName := getDetailsFromDownloadPath(downloadPattern)
			downloadFile(downloadPath, localPath, localFileName, logMsgPrefix + ": ", flags)
		}
	}
}

func downloadFiles(resultItems []utils.AqlSearchResultItem, flags *utils.Flags) {
	size := len(resultItems)
	var wg sync.WaitGroup
	for i := 0; i < flags.Threads; i++ {
		wg.Add(1)
		go func(threadId int) {
			logMsgPrefix := cliutils.GetLogMsgPrefix(threadId, flags.DryRun)
			for j := threadId; j < size; j += flags.Threads {
				downloadPath := utils.BuildArtifactoryUrl(flags.ArtDetails.Url, resultItems[j].GetFullUrl(), make(map[string]string))
				if !flags.DryRun {
					downloadFile(downloadPath, resultItems[j].Path, resultItems[j].Name, logMsgPrefix, flags)
				} else {
					fmt.Println(logMsgPrefix + "Downloading " + downloadPath)
				}
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func getDetailsFromDownloadPath(downloadPattern string) (localPath, localFileName string) {
	firstSeparator := strings.Index(downloadPattern, "/")
	lastSeparator := strings.LastIndex(downloadPattern, "/")
	if firstSeparator != lastSeparator {
		localPath = downloadPattern[firstSeparator + 1:lastSeparator];
	} else {
		localPath = "."
	}
	localFileName = downloadPattern[lastSeparator + 1:];
	return
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
