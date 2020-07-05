package permissiontarget

import (
	"github.com/jfrog/jfrog-cli/utils/config"
)

type PermissionTargetCreateCommand struct {
	PermissionTargetCommand
}

func NewPermissionTargetCreateCommand() *PermissionTargetCreateCommand {
	return &PermissionTargetCreateCommand{}
}

func (ptcc *PermissionTargetCreateCommand) SetTemplatePath(path string) *PermissionTargetCreateCommand {
	ptcc.templatePath = path
	return ptcc
}

func (ptcc *PermissionTargetCreateCommand) SetVars(vars string) *PermissionTargetCreateCommand {
	ptcc.vars = vars
	return ptcc
}

func (ptcc *PermissionTargetCreateCommand) SetRtDetails(rtDetails *config.ArtifactoryDetails) *PermissionTargetCreateCommand {
	ptcc.rtDetails = rtDetails
	return ptcc
}

func (ptcc *PermissionTargetCreateCommand) RtDetails() (*config.ArtifactoryDetails, error) {
	return ptcc.rtDetails, nil
}

func (ptcc *PermissionTargetCreateCommand) CommandName() string {
	return "rt_permission_target_create"
}

func (ptcc *PermissionTargetCreateCommand) Run() (err error) {
	return ptcc.PerformPermissionTargetCmd(false)
}
