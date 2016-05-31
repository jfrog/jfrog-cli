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
			details := getFileRemoteDetails(downloadPath, flags)
			if shouldDownloadFile(getFileLocalPath(localPath, localFileName, flags), details.Md5, details.Sha1) {
				downloadFileDetails := createDownloadFileDetails(downloadPath, localPath, localFileName, details.AcceptRanges, details.Size, flags)
				downloadFile(downloadFileDetails, logMsgPrefix + ": ", flags)
			} else {
				fmt.Println(logMsgPrefix + "File already exists locally.")
			}
		} else {
			fmt.Println(logMsgPrefix + "Downloading " + downloadPath)
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
				if flags.DryRun {
					fmt.Println(logMsgPrefix + "Downloading " + downloadPath)
					continue
				}
				if shouldDownloadFile(getFileLocalPath(resultItems[j].Path, resultItems[j].Name, flags), resultItems[j].Actual_Md5, resultItems[j].Actual_Sha1) {
					downloadFileDetails := createDownloadFileDetails(downloadPath, resultItems[j].Path, resultItems[j].Name, cliutils.NotDefined, resultItems[j].Size, flags)
					downloadFile(downloadFileDetails, logMsgPrefix, flags)
				} else {
					fmt.Println(logMsgPrefix + "File already exists locally.")
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

func createDownloadFileDetails(downloadPath, localPath, localFileName string, acceptRanges cliutils.BoolEnum, size int64, flags *utils.Flags) (details *DownloadFileDetails) {
	details = &DownloadFileDetails{
		DownloadPath: downloadPath,
		LocalPath: localPath,
		LocalFileName: localFileName,
		AcceptRanges: acceptRanges,
		Size: size}
	return
}

func getFileRemoteDetails(downloadPath string, flags *utils.Flags) (details *ioutils.FileDetails) {
	httpClientsDetails := utils.GetArtifactoryHttpClientDetails(flags.ArtDetails)
	details = ioutils.GetRemoteFileDetails(downloadPath, httpClientsDetails)
	return
}

func getFileLocalPath(localPath, localFileName string, flags *utils.Flags) (localFilePath string){
	localFilePath = localFileName
	if !flags.Flat {
		localFilePath = localPath + "/" + localFileName
	}
	return
}

func downloadFile(downloadFileDetails *DownloadFileDetails, logMsgPrefix string, flags *utils.Flags) {
	httpClientsDetails := utils.GetArtifactoryHttpClientDetails(flags.ArtDetails)
	if flags.SplitCount == 0 || flags.MinSplitSize < 0 || flags.MinSplitSize*1000 > downloadFileDetails.Size || !isFileAcceptRange(downloadFileDetails, flags) {
		resp := ioutils.DownloadFile(downloadFileDetails.DownloadPath, downloadFileDetails.LocalPath, downloadFileDetails.LocalFileName, flags.Flat, httpClientsDetails)
		fmt.Println(logMsgPrefix + "Artifactory response:", resp.Status)
	} else {
		concurrentDownloadFlags := ioutils.ConcurrentDownloadFlags{
			DownloadPath: downloadFileDetails.DownloadPath,
			FileName:     downloadFileDetails.LocalFileName,
			LocalPath:    downloadFileDetails.LocalPath,
			FileSize:     downloadFileDetails.Size,
			SplitCount:   flags.SplitCount,
			Flat:         flags.Flat}

		ioutils.DownloadFileConcurrently(concurrentDownloadFlags, logMsgPrefix, httpClientsDetails)
	}
}

func isFileAcceptRange(downloadFileDetails *DownloadFileDetails, flags *utils.Flags) bool {
	if downloadFileDetails.AcceptRanges == cliutils.NotDefined {
		details := getFileRemoteDetails(downloadFileDetails.DownloadPath, flags)
		return details.AcceptRanges == cliutils.True
	}
	return downloadFileDetails.AcceptRanges == cliutils.True
}

func shouldDownloadFile(localFilePath, md5, sha1 string) bool {
	if !ioutils.IsFileExists(localFilePath) {
		return true
	}
	localFileDetails := ioutils.GetFileDetails(localFilePath)
	if localFileDetails.Md5 != md5 || localFileDetails.Sha1 != sha1 {
		return true
	}
	return false
}

type DownloadFileDetails struct {
	DownloadPath  			 string 		     `json:"DownloadPath,omitempty"`
	LocalPath     			 string 		     `json:"LocalPath,omitempty"`
	LocalFileName 			 string 		     `json:"LocalFileName,omitempty"`
	AcceptRanges  			 cliutils.BoolEnum   `json:"AcceptRanges,omitempty"`
	Size  			 		 int64  		     `json:"Size,omitempty"`
}

