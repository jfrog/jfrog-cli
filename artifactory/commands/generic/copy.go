package generic

import (
	"errors"

	"github.com/jfrog/jfrog-cli/artifactory/spec"
	"github.com/jfrog/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

type CopyCommand struct {
	GenericCommand
}

func NewCopyCommand() *CopyCommand {
	return &CopyCommand{GenericCommand: *NewGenericCommand()}
}

func (cc *CopyCommand) CommandName() string {
	return "rt_copy"
}

// Copies the artifacts using the specified move pattern.
func (cc *CopyCommand) Run() error {
	// Create Service Manager:
	servicesManager, err := utils.CreateServiceManager(cc.rtDetails, cc.dryRun)
	if err != nil {
		return err
	}

	var errorOccurred = false
	// Copy Loop:
	for i := 0; i < len(cc.spec.Files); i++ {

		copyParams, err := getCopyParams(cc.spec.Get(i))
		if err != nil {
			errorOccurred = true
			log.Error(err)
			continue
		}

		partialSuccess, partialFailed, err := servicesManager.Copy(copyParams)
		success := cc.result.SuccessCount() + partialSuccess
		cc.result.SetSuccessCount(success)
		failed := cc.result.FailCount() + partialFailed
		cc.result.SetFailCount(failed)
		if err != nil {
			errorOccurred = true
			log.Error(err)
			continue
		}
	}
	if errorOccurred {
		return errors.New("Copy finished with errors, please review the logs.")
	}
	return err
}

func getCopyParams(f *spec.File) (copyParams services.MoveCopyParams, err error) {
	copyParams = services.NewMoveCopyParams()
	copyParams.ArtifactoryCommonParams = f.ToArtifactoryCommonParams()
	copyParams.Recursive, err = f.IsRecursive(true)
	if err != nil {
		return
	}

	copyParams.Flat, err = f.IsFlat(false)
	if err != nil {
		return
	}
	return
}
