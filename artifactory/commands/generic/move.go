package generic

import (
	"errors"

	"github.com/jfrog/jfrog-cli/artifactory/spec"
	"github.com/jfrog/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

type MoveCommand struct {
	GenericCommand
}

func NewMoveCommand() *MoveCommand {
	return &MoveCommand{GenericCommand: *NewGenericCommand()}
}

// Moves the artifacts using the specified move pattern.
func (mc *MoveCommand) Run() error {
	// Create Service Manager:
	servicesManager, err := utils.CreateServiceManager(mc.rtDetails, mc.DryRun())
	if err != nil {
		return err
	}

	var errorOccurred = false
	// Move Loop:
	for i := 0; i < len(mc.Spec().Files); i++ {

		moveParams, err := getMoveParams(mc.Spec().Get(i))
		if err != nil {
			errorOccurred = true
			log.Error(err)
			continue
		}

		partialSuccess, partialFailed, err := servicesManager.Move(moveParams)
		success := mc.result.SuccessCount() + partialSuccess
		mc.result.SetSuccessCount(success)
		failed := mc.result.FailCount() + partialFailed
		mc.result.SetFailCount(failed)
		if err != nil {
			errorOccurred = true
			log.Error(err)
			continue
		}
	}
	if errorOccurred {
		return errors.New("Move finished with errors, please review the logs.")
	}
	return err

}

func (mc *MoveCommand) CommandName() string {
	return "rt_move"
}

func getMoveParams(f *spec.File) (moveParams services.MoveCopyParams, err error) {
	moveParams = services.NewMoveCopyParams()
	moveParams.ArtifactoryCommonParams = f.ToArtifactoryCommonParams()
	moveParams.Recursive, err = f.IsRecursive(true)
	if err != nil {
		return
	}

	moveParams.Flat, err = f.IsFlat(false)
	if err != nil {
		return
	}
	return
}
