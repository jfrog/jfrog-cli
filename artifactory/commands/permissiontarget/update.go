package permissiontarget

import (
	"github.com/jfrog/jfrog-cli/utils/config"
)

type PermissionTargetUpdateCommand struct {
	PermissionTargetCommand
}

func NewPermissionTargetUpdateCommand() *PermissionTargetUpdateCommand {
	return &PermissionTargetUpdateCommand{}
}

func (ptuc *PermissionTargetUpdateCommand) SetTemplatePath(path string) *PermissionTargetUpdateCommand {
	ptuc.templatePath = path
	return ptuc
}

func (ptuc *PermissionTargetUpdateCommand) SetVars(vars string) *PermissionTargetUpdateCommand {
	ptuc.vars = vars
	return ptuc
}

func (ptuc *PermissionTargetUpdateCommand) SetRtDetails(rtDetails *config.ArtifactoryDetails) *PermissionTargetUpdateCommand {
	ptuc.rtDetails = rtDetails
	return ptuc
}

func (ptuc *PermissionTargetUpdateCommand) RtDetails() (*config.ArtifactoryDetails, error) {
	return ptuc.rtDetails, nil
}

func (ptuc *PermissionTargetUpdateCommand) CommandName() string {
	return "rt_permission_target_update"
}

func (ptuc *PermissionTargetUpdateCommand) Run() (err error) {
	return ptuc.PerformPermissionTargetCmd(true)
}
