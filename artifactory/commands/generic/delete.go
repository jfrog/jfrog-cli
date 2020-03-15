package generic

import (
	"github.com/jfrog/jfrog-cli-go/artifactory/spec"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	clientutils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
)

type DeleteCommand struct {
	deleteItems []clientutils.ResultItem
	GenericCommand
	quiet   bool
	threads int
}

func NewDeleteCommand() *DeleteCommand {
	return &DeleteCommand{GenericCommand: *NewGenericCommand()}
}

func (dc *DeleteCommand) Threads() int {
	return dc.threads
}

func (dc *DeleteCommand) SetThreads(threads int) *DeleteCommand {
	dc.threads = threads
	return dc
}

func (dc *DeleteCommand) Quiet() bool {
	return dc.quiet
}

func (dc *DeleteCommand) SetQuiet(quiet bool) *DeleteCommand {
	dc.quiet = quiet
	return dc
}

func (dc *DeleteCommand) DeleteItems() []clientutils.ResultItem {
	return dc.deleteItems
}

func (dc *DeleteCommand) SetDeleteItems(deleteItems []clientutils.ResultItem) *DeleteCommand {
	dc.deleteItems = deleteItems
	return dc
}

func (dc *DeleteCommand) CommandName() string {
	return "rt_delete"
}

func (dc *DeleteCommand) Run() error {
	err := dc.GetPathsToDelete()
	if err != nil {
		return err
	}
	if dc.quiet || utils.ConfirmDelete(dc.deleteItems) {
		success, failed, err := dc.DeleteFiles()
		result := dc.Result()
		result.SetFailCount(failed)
		result.SetSuccessCount(success)
		return err
	}
	return nil
}

func (dc *DeleteCommand) GetPathsToDelete() error {
	rtDetails, err := dc.RtDetails()
	if errorutils.CheckError(err) != nil {
		return err
	}
	servicesManager, err := utils.CreateServiceManager(rtDetails, dc.DryRun())
	if err != nil {
		return err
	}
	for i := 0; i < len(dc.Spec().Files); i++ {
		deleteParams, err := getDeleteParams(dc.Spec().Get(i))
		if err != nil {
			return err
		}

		currentResultItems, err := servicesManager.GetPathsToDelete(deleteParams)
		if err != nil {
			return err
		}
		dc.deleteItems = append(dc.deleteItems, currentResultItems...)
	}
	return nil
}

func (dc *DeleteCommand) DeleteFiles() (successCount, failedCount int, err error) {
	rtDetails, err := dc.RtDetails()
	if errorutils.CheckError(err) != nil {
		return 0, 0, err
	}
	servicesManager, err := utils.CreateDeleteServiceManager(rtDetails, dc.Threads(), dc.DryRun())
	if err != nil {
		return 0, 0, err
	}
	deletedCount, err := servicesManager.DeleteFiles(dc.deleteItems)
	return deletedCount, len(dc.deleteItems) - deletedCount, err
}

func getDeleteParams(f *spec.File) (deleteParams services.DeleteParams, err error) {
	deleteParams = services.NewDeleteParams()
	deleteParams.ArtifactoryCommonParams = f.ToArtifactoryCommonParams()
	deleteParams.Recursive, err = f.IsRecursive(true)
	if err != nil {
		return
	}
	return
}
