package utils

import (
	"fmt"

	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory"
	rtclientutils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/utils/io/content"
)

func ConfirmDelete(pathsToDeleteReader *content.ContentReader) (bool, error) {
	length, err := pathsToDeleteReader.Length()
	if err != nil || length < 1 {
		return false, err
	}
	for resultItem := new(rtclientutils.ResultItem); pathsToDeleteReader.NextRecord(resultItem) == nil; resultItem = new(rtclientutils.ResultItem) {
		fmt.Println("  " + resultItem.GetItemRelativePath())
	}
	if err := pathsToDeleteReader.GetError(); err != nil {
		return false, err
	}
	pathsToDeleteReader.Reset()
	return cliutils.AskYesNo("Are you sure you want to delete the above paths?", false), nil
}

func CreateDeleteServiceManager(artDetails *config.ArtifactoryDetails, threads int, dryRun bool) (*artifactory.ArtifactoryServicesManager, error) {
	return CreateServiceManagerWithThreads(artDetails, dryRun, threads)
}
