package utils

import (
	"fmt"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory"
	rtclientutils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
)

func ConfirmDelete(pathsToDelete []rtclientutils.ResultItem) bool {
	if len(pathsToDelete) < 1 {
		return false
	}
	for _, v := range pathsToDelete {
		fmt.Println("  " + v.GetItemRelativePath())
	}
	return cliutils.InteractiveConfirm("Are you sure you want to delete the above paths?", false)
}

func CreateDeleteServiceManager(artDetails *config.ArtifactoryDetails, threads int, dryRun bool) (*artifactory.ArtifactoryServicesManager, error) {
	return CreateServiceManagerWithThreads(artDetails, dryRun, threads)
}
