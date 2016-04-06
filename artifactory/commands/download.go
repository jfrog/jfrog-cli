package commands

import (
	"encoding/json"
	"fmt"
	"github.com/jfrogdev/jfrog-cli-go/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	"strconv"
	"sync"
)

// Downloads the artifacts using the specified download pattern.
// Returns the AQL query used for the download.
func Download(downloadPattern string, flags *utils.Flags) string {
	if !flags.DryRun {
		ioutils.CreateTempDirPath()
		defer ioutils.RemoveTempDir()
	}

	if flags.ArtDetails.SshKeyPath != "" {
		utils.SshAuthentication(flags.ArtDetails)
	}
	if !flags.DryRun {
		utils.PingArtifactory(flags.ArtDetails)
	}

	aqlUrl := flags.ArtDetails.Url + "api/search/aql"
	data := utils.BuildAqlSearchQuery(downloadPattern, flags.Recursive, flags.Props)
	fmt.Println("Searching Artifactory using AQL query: " + data)

	if !flags.DryRun {
		utils.AddAuthHeaders(nil, flags.ArtDetails)
		httpClientsDetails := utils.GetArtifactoryHttpClientDetails(flags.ArtDetails)
		resp, json := ioutils.SendPost(aqlUrl, []byte(data), httpClientsDetails)
		fmt.Println("Artifactory response:", resp.Status)

		if resp.StatusCode == 200 {
			resultItems := parseAqlSearchResponse(json)
			downloadFiles(resultItems, flags, httpClientsDetails)
			fmt.Println("Downloaded " + strconv.Itoa(len(resultItems)) + " artifacts from Artifactory.")
		}
	}
	return data
}

func downloadFiles(resultItems []AqlSearchResultItem, flags *utils.Flags, httpClientsDetails ioutils.HttpClientDetails) {
	size := len(resultItems)
	var wg sync.WaitGroup
	for i := 0; i < flags.Threads; i++ {
		wg.Add(1)
		go func(threadId int) {
			logMsgPrefix := cliutils.GetLogMsgPrefix(threadId, flags.DryRun)
			for j := threadId; j < size; j += flags.Threads {
				downloadPath := buildDownloadUrl(flags.ArtDetails.Url, resultItems[j])
				fmt.Println(logMsgPrefix + "Downloading " + downloadPath)
				if !flags.DryRun {
					downloadFile(downloadPath, resultItems[j].Path, resultItems[j].Name, logMsgPrefix, flags, httpClientsDetails)
				}
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func downloadFile(downloadPath, localPath, localFileName, logMsgPrefix string, flags *utils.Flags, httpClientsDetails ioutils.HttpClientDetails) {
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

func buildDownloadUrl(baseUrl string, resultItem AqlSearchResultItem) string {
	if resultItem.Path == "." {
		return baseUrl + resultItem.Repo + "/" + resultItem.Name
	}
	return baseUrl + resultItem.Repo + "/" + resultItem.Path + "/" + resultItem.Name
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

func parseAqlSearchResponse(resp []byte) []AqlSearchResultItem {
	var result AqlSearchResult
	err := json.Unmarshal(resp, &result)
	cliutils.CheckError(err)
	return result.Results
}

type AqlSearchResult struct {
	Results []AqlSearchResultItem
}

type AqlSearchResultItem struct {
	Repo string
	Path string
	Name string
}
