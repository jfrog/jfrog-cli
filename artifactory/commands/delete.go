package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/artifactory/utils"
	"fmt"
	"strconv"
	"strings"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
)

func Delete(deletePattern string, flags *utils.Flags) {
	utils.PreCommandSetup(flags)

	var resultItems []utils.AqlSearchResultItem
	if !utils.IsWildcardPattern(deletePattern) || isDirectoryPath(deletePattern) {
		simplePathItem := utils.AqlSearchResultItem{Path:deletePattern}
		resultItems = []utils.AqlSearchResultItem{simplePathItem}
	} else {
		resultItems = utils.AqlSearch(deletePattern, flags)
	}

	deletedCount := deleteFiles(resultItems, flags)
	fmt.Println("Deleted " + strconv.Itoa(deletedCount) + " artifacts from Artifactory")
}

func deleteFiles(resultItems []utils.AqlSearchResultItem, flags *utils.Flags) int {
	deletedCount := 0
	for _, v := range resultItems {
		fileUrl := utils.BuildArtifactoryUrl(flags.ArtDetails.Url, v.GetFullUrl(), make(map[string]string))
		if flags.DryRun {
			fmt.Println("[Dry run] Deleting: " + fileUrl)
			continue
		}

		fmt.Println("Deleting: " + fileUrl)
		httpClientsDetails := utils.GetArtifactoryHttpClientDetails(flags.ArtDetails)
		resp, _ := ioutils.SendDelete(fileUrl, nil, httpClientsDetails)
		fmt.Println("Artifactory response:", resp.Status)

		deletedCount += cliutils.Bool2Int(resp.StatusCode == 204)
	}
	return deletedCount
}

// Simple directory path without wildcards.
func isDirectoryPath(path string) bool {
	if !strings.Contains(path, "*") && strings.HasSuffix(path, "/") {
		return true
	}
	return false
}
