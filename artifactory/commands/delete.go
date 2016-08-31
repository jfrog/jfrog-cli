package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/artifactory/utils"
	"fmt"
	"strings"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/logger"
)

func Delete(deletePattern string, flags *DeleteFlags) (err error) {
	utils.PreCommandSetup(flags)

	var resultItems []utils.AqlSearchResultItem
	if !utils.IsWildcardPattern(deletePattern) || isDirectoryPath(deletePattern) {
		simplePathItem := utils.AqlSearchResultItem{Path:deletePattern}
		resultItems = []utils.AqlSearchResultItem{simplePathItem}
	} else {
		resultItems, err = utils.AqlSearchDefaultReturnFields(deletePattern, flags.Recursive, flags.Props, flags)
		if err != nil {
			return
		}
	}

	err = deleteFiles(resultItems, flags)
	return
}

func deleteFiles(resultItems []utils.AqlSearchResultItem, flags *DeleteFlags) error {
	for _, v := range resultItems {
		fileUrl, err := utils.BuildArtifactoryUrl(flags.ArtDetails.Url, v.GetFullUrl(), make(map[string]string))
		if err != nil {
		    return err
		}
		if flags.DryRun {
			fmt.Println("[Dry run] Deleting: " + fileUrl)
			continue
		}

		logger.Logger.Info("Deleting: " + fileUrl)
		httpClientsDetails := utils.GetArtifactoryHttpClientDetails(flags.ArtDetails)
		resp, _, err := ioutils.SendDelete(fileUrl, nil, httpClientsDetails)
		if err != nil {
		    return err
		}
		logger.Logger.Info("Artifactory response:", resp.Status)
	}
	return nil
}

// Simple directory path without wildcards.
func isDirectoryPath(path string) bool {
	if !strings.Contains(path, "*") && strings.HasSuffix(path, "/") {
		return true
	}
	return false
}

type DeleteFlags struct {
	ArtDetails *config.ArtifactoryDetails
	DryRun     bool
	Props      string
	Recursive  bool
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