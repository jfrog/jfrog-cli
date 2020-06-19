package utils

import (
	"fmt"

	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory"
	rtclientutils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/utils/io/content"
)

func ConfirmDelete(pathsToDelete *content.ContentReader) (bool, error) {
	length, err := pathsToDelete.Length()
	if err != nil {
		return false, nil
	}
	if length < 1 {
		return false, nil
	}
	for resultItem := new(rtclientutils.ResultItem); pathsToDelete.NextRecord(resultItem) == nil; resultItem = new(rtclientutils.ResultItem) {
		fmt.Println("  " + resultItem.GetItemRelativePath())
	}
	if err := pathsToDelete.GetError(); err != nil {
		return false, err
	}
	return cliutils.AskYesNo("Are you sure you want to delete the above paths?", false)
}

func CreateDeleteServiceManager(artDetails *config.ArtifactoryDetails, threads int, dryRun bool) (*artifactory.ArtifactoryServicesManager, error) {
	return CreateServiceManagerWithThreads(artDetails, dryRun, threads)
}
