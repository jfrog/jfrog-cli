package utils

import (
	"fmt"
	"github.com/jfrog/jfrog-cli-go/utils/cliutils"
	rtclientutils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
)

func ConfirmDelete(pathsToDelete []rtclientutils.ResultItem) bool {
	if len(pathsToDelete) < 1 {
		return false
	}
	for _, v := range pathsToDelete {
		fmt.Println("  " + v.GetItemRelativePath())
	}
	return cliutils.InteractiveConfirm("Are you sure you want to delete the above paths?")
}
