package permissiontarget

import (
	rtUtils "github.com/jfrog/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-cli/utils/config"
)

type PermissionTargetDeleteCommand struct {
	rtDetails            *config.ArtifactoryDetails
	permissionTargetName string
	quiet                bool
}

func NewPermissionTargetDeleteCommand() *PermissionTargetDeleteCommand {
	return &PermissionTargetDeleteCommand{}
}

func (ptdc *PermissionTargetDeleteCommand) SetPermissionTargetName(permissionTargetName string) *PermissionTargetDeleteCommand {
	ptdc.permissionTargetName = permissionTargetName
	return ptdc
}

func (ptdc *PermissionTargetDeleteCommand) SetQuiet(quiet bool) *PermissionTargetDeleteCommand {
	ptdc.quiet = quiet
	return ptdc
}

func (ptdc *PermissionTargetDeleteCommand) SetRtDetails(rtDetails *config.ArtifactoryDetails) *PermissionTargetDeleteCommand {
	ptdc.rtDetails = rtDetails
	return ptdc
}

func (ptdc *PermissionTargetDeleteCommand) RtDetails() (*config.ArtifactoryDetails, error) {
	return ptdc.rtDetails, nil
}

func (ptdc *PermissionTargetDeleteCommand) CommandName() string {
	return "rt_permission_target_delete"
}

func (ptdc *PermissionTargetDeleteCommand) Run() (err error) {
	if !ptdc.quiet && !cliutils.InteractiveConfirm("Are you sure you want to permanently delete the permission target "+ptdc.permissionTargetName+"?", true) {
		return nil
	}
	servicesManager, err := rtUtils.CreateServiceManager(ptdc.rtDetails, false)
	if err != nil {
		return err
	}
	return servicesManager.DeletePermissionTarget(ptdc.permissionTargetName)
}
