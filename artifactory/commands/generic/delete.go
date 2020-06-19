package generic

import (
	"github.com/jfrog/jfrog-cli/artifactory/spec"
	"github.com/jfrog/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/content"
)

type DeleteCommand struct {
	deleteItems *content.ContentReader
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

func (dc *DeleteCommand) DeleteItems() *content.ContentReader {
	return dc.deleteItems
}

func (dc *DeleteCommand) SetDeleteItems(deleteItems *content.ContentReader) *DeleteCommand {
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
	allowDelete := true
	if !dc.quiet {
		allowDelete, err = utils.ConfirmDelete(dc.deleteItems)
		if err != nil {
			return err
		}
	}
	if allowDelete {
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
	temp := []*content.ContentReader{}
	for i := 0; i < len(dc.Spec().Files); i++ {
		deleteParams, err := getDeleteParams(dc.Spec().Get(i))
		if err != nil {
			return err
		}

		currentResultsReader, err := servicesManager.GetPathsToDelete(deleteParams)
		if err != nil {
			return err
		}
		length, err := currentResultsReader.Length()
		if err != nil {
			return err
		}
		if length > 0 {
			temp = append(temp, currentResultsReader)
		}
	}
	dc.deleteItems, err = content.MergeReaders(temp, "results")
	if err != nil {
		return err
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
	length, err := dc.deleteItems.Length()
	if err != nil {
		return 0, 0, err
	}
	return deletedCount, length - deletedCount, err
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
