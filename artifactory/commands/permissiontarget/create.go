package permissiontarget

import (
	"encoding/json"
	"errors"
	"github.com/jfrog/jfrog-cli/artifactory/commands/utils"
	rtUtils "github.com/jfrog/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"strings"
)

const DefaultBuildRepositoriesValue = "artifactory-build-info"

type PermissionTargetCreateCommand struct {
	rtDetails    *config.ArtifactoryDetails
	templatePath string
	vars         string
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

func (ptcc *PermissionTargetCreateCommand) Vars() string {
	return ptcc.vars
}

func (ptcc *PermissionTargetCreateCommand) TemplatePath() string {
	return ptcc.templatePath
}

func (ptcc *PermissionTargetCreateCommand) RtDetails() (*config.ArtifactoryDetails, error) {
	return ptcc.rtDetails, nil
}

func (ptcc *PermissionTargetCreateCommand) CommandName() string {
	return "rt_permission_target_create"
}

func (ptcc *PermissionTargetCreateCommand) Run() (err error) {
	permissionTargetConfigMap, err := utils.ConvertTemplateToMap(ptcc)
	if err != nil {
		return err
	}
	// Go over the the confMap and write the values with the correct types
	for key, value := range permissionTargetConfigMap {
		isBuildSection := false
		switch key {
		case Name:
			if _, ok := value.(string); !ok {
				return errorutils.CheckError(errors.New("template syntax error: the value for the  key: \"Name\" is not a string type."))
			}
		case Build:
			isBuildSection = true
			fallthrough
		case Repo:
			fallthrough
		case ReleaseBundle:
			permissionSection, err := covertPermissionSection(value, isBuildSection)
			if err != nil {
				return err
			}
			permissionTargetConfigMap[key] = permissionSection
		default:
			return errorutils.CheckError(errors.New("template syntax error: unknown key: \"" + key + "\"."))
		}
	}
	// Convert the new JSON with the correct types to params struct
	content, err := json.Marshal(permissionTargetConfigMap)
	params := services.NewCreatePermissionTargetParams()
	err = json.Unmarshal(content, &params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	servicesManager, err := rtUtils.CreateServiceManager(ptcc.rtDetails, false)
	if err != nil {
		return err
	}
	return servicesManager.CreatePermissionTarget(params)
}

// Each section is map of string->interface{}. We need to convert each value to its correct type
func covertPermissionSection(value interface{}, isBuildSection bool) (*services.PermissionTargetSection, error) {
	content, err := json.Marshal(value)
	if errorutils.CheckError(err) != nil {
		return nil, err
	}
	var answer PermissionSectionAnswer
	err = json.Unmarshal(content, &answer)
	if errorutils.CheckError(err) != nil {
		return nil, err
	}
	var pts services.PermissionTargetSection
	if len(answer.IncludePatterns) > 0 {
		pts.IncludePatterns = strings.Split(answer.IncludePatterns, ",")
	}
	if len(answer.ExcludePatterns) > 0 {
		pts.ExcludePatterns = strings.Split(answer.ExcludePatterns, ",")
	}
	// 'build' permission target must contains repositories with a default value that cannot be changed.
	if isBuildSection {
		answer.Repositories = DefaultBuildRepositoriesValue
	}
	if len(answer.Repositories) > 0 {
		pts.Repositories = strings.Split(answer.Repositories, ",")
	}
	if answer.ActionsUsers != nil {
		convertActionMap(answer.ActionsUsers, &pts.Actions.Users)
	}
	if answer.ActionsGroups != nil {
		convertActionMap(answer.ActionsGroups, &pts.Actions.Groups)
	}
	return &pts, nil
}

// actionMap is map of string->string. We need to convert each value to []string
func convertActionMap(srcMap map[string]string, tgtMap *map[string][]string) {
	*tgtMap = make(map[string][]string)
	for key, permissionsStr := range srcMap {
		(*tgtMap)[key] = strings.Split(permissionsStr, ",")
	}

}
