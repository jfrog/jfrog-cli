package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/artifactory/utils"
	"fmt"
	"strings"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/logger"
)

func Delete(deletePattern string, flags *DeleteFlags) {
	utils.PreCommandSetup(flags)

	var resultItems []utils.AqlSearchResultItem
	if !utils.IsWildcardPattern(deletePattern) || isDirectoryPath(deletePattern) {
		simplePathItem := utils.AqlSearchResultItem{Path:deletePattern}
		resultItems = []utils.AqlSearchResultItem{simplePathItem}
	} else {
		resultItems = utils.AqlSearchDefaultReturnFields(deletePattern, flags)
	}

	deleteFiles(resultItems, flags)
}

func deleteFiles(resultItems []utils.AqlSearchResultItem, flags *DeleteFlags) {
	for _, v := range resultItems {
		fileUrl := utils.BuildArtifactoryUrl(flags.ArtDetails.Url, v.GetFullUrl(), make(map[string]string))
		if flags.DryRun {
			fmt.Println("[Dry run] Deleting: " + fileUrl)
			continue
		}

		logger.Logger.Info("Deleting: " + fileUrl)
		httpClientsDetails := utils.GetArtifactoryHttpClientDetails(flags.ArtDetails)
		resp, _ := ioutils.SendDelete(fileUrl, nil, httpClientsDetails)
		logger.Logger.Info("Artifactory response:", resp.Status)
	}
}

// Simple directory path without wildcards.
func isDirectoryPath(path string) bool {
	if !strings.Contains(path, "*") && strings.HasSuffix(path, "/") {
		return true
	}
	return false
}

type DeleteFlags struct {
	ArtDetails   *config.ArtifactoryDetails
	DryRun       bool
	Props        string
	Recursive    bool
}

func (flags *DeleteFlags) GetArtifactoryDetails() *config.ArtifactoryDetails {
	return flags.ArtDetails
}

func (flags *DeleteFlags) IsRecursive() bool {
	return flags.Recursive
}

func (flags *DeleteFlags) GetProps() string {
	return flags.Props
}

func (flags *DeleteFlags) IsDryRun() bool {
	return flags.DryRun
}