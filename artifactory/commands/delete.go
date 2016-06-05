package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/artifactory/utils"
	"fmt"
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

	deleteFiles(resultItems, flags)
}

func deleteFiles(resultItems []utils.AqlSearchResultItem, flags *utils.Flags) {
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
	}
}

// Simple directory path without wildcards.
func isDirectoryPath(path string) bool {
	if !strings.Contains(path, "*") && strings.HasSuffix(path, "/") {
		return true
	}
	return false
}
