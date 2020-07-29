package permissiontarget

import (
	"encoding/json"
	"fmt"
	"github.com/c-bata/go-prompt"
	"github.com/jfrog/jfrog-cli/artifactory/commands/utils"
	"github.com/jfrog/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"io/ioutil"
	"sort"
	"strings"
)

type PermissionTargetTemplateCommand struct {
	path string
}

const (
	// Strings for prompt questions
	SelectPermissionTargetSectionMsg = "Select the permission target section to configure" + utils.PressTabMsg
	LeaveEmptyForDefault             = "(press enter for default)"

	// Yes,No answers
	Yes = "yes"
	No  = "no"

	// Main permission target configuration JSON keys
	Name          = "name"
	Repo          = "repo"
	Build         = "build"
	ReleaseBundle = "releaseBundle"

	IncludePatternsDefault = "**"
	ExcludePatternsDefault = ""

	// Possible permissions
	read            = "read"
	write           = "write"
	annotate        = "annotate"
	delete          = "delete"
	manage          = "manage"
	managedXrayMeta = "managedXrayMeta"
	distribute      = "distribute"

	permissionSelectEnd = "-"
)

func NewPermissionTargetTemplateCommand() *PermissionTargetTemplateCommand {
	return &PermissionTargetTemplateCommand{}
}

func (pttc *PermissionTargetTemplateCommand) SetTemplatePath(path string) *PermissionTargetTemplateCommand {
	pttc.path = path
	return pttc
}

func (pttc *PermissionTargetTemplateCommand) RtDetails() (*config.ArtifactoryDetails, error) {
	// Since it's a local command, usage won't be reported.
	return nil, nil
}

func (pttc *PermissionTargetTemplateCommand) Run() (err error) {
	err = utils.ValidateTemplatePath(pttc.path)
	if err != nil {
		return
	}
	permissionTargetTemplateQuestionnaire := &utils.InteractiveQuestionnaire{
		MandatoryQuestionsKeys: []string{Name},
		QuestionsMap:           questionMap,
		OptionalKeysSuggests:   optionalSuggestsMap,
	}
	err = permissionTargetTemplateQuestionnaire.Perform()
	if err != nil {
		return err
	}
	resBytes, err := json.Marshal(permissionTargetTemplateQuestionnaire.AnswersMap)
	if err != nil {
		return errorutils.CheckError(err)
	}
	if err = ioutil.WriteFile(pttc.path, resBytes, 0644); err != nil {
		return errorutils.CheckError(err)
	}
	log.Info(fmt.Sprintf("Permission target configuration template successfully created at %s.", pttc.path))

	return nil
}

func (pttc *PermissionTargetTemplateCommand) CommandName() string {
	return "rt_permission_target_template"
}

var optionalSuggestsMap = []prompt.Suggest{
	{Text: utils.SaveAndExit},
	{Text: Repo},
	{Text: Build},
	{Text: ReleaseBundle},
}

// Each permission target section (repo/build/releaseBundle) can have the following keys:
//	* repos - Mandatory for repo and releaseBundle. Has a const default value for build.
//	* include/exclude-patterns - Optional, has a default value.
//	* actions - Optional,includes two maps (users and groups): user/group name -> permissions array.
func permissionSectionCallBack(iq *utils.InteractiveQuestionnaire, section string) (value string, err error) {
	if section == utils.SaveAndExit {
		return
	}
	var sectionAnswer PermissionSectionAnswer
	if section != Build {
		sectionAnswer.Repositories = utils.AskString(reposQuestionInfo.Msg, reposQuestionInfo.PromptPrefix, false)
	}
	sectionAnswer.IncludePatterns = utils.AskStringWithDefault(includePatternsQuestionInfo.Msg, includePatternsQuestionInfo.PromptPrefix, IncludePatternsDefault)
	sectionAnswer.ExcludePatterns = utils.AskString(excludePatternsQuestionInfo.Msg, excludePatternsQuestionInfo.PromptPrefix, true)
	configureActions := utils.AskFromList("", configureActionsQuestionInfo.PromptPrefix+"users?"+utils.PressTabMsg, false, configureActionsQuestionInfo.Options, Yes)
	if configureActions == Yes {
		sectionAnswer.ActionsUsers = make(map[string]string)
		readActionsMap("user", sectionAnswer.ActionsUsers)
	}
	configureActions = utils.AskFromList("", configureActionsQuestionInfo.PromptPrefix+"groups?", false, configureActionsQuestionInfo.Options, Yes)
	if configureActions == Yes {
		sectionAnswer.ActionsGroups = make(map[string]string)
		readActionsMap("group", sectionAnswer.ActionsGroups)
	}
	iq.AnswersMap[section] = sectionAnswer
	return
}

// We will read (user/group name, permissions) pairs until empty name is read.
func readActionsMap(actionsType string, actionsMap map[string]string) {
	//fmt.Println("Permissions value is a comma separated list, which can include the following values: read, write, annotate, delete, manage, managedXrayMeta, distribute")
	customKeyPrompt := "Insert " + actionsType + " name (press enter to finish) >"
	for {
		key := utils.AskString("", customKeyPrompt, true)
		if key == "" {
			return
		}
		value := strings.Join(readPermissionList(key), ",")
		actionsMap[key] = value
	}
}

func readPermissionList(permissionsOwner string) (permissions []string) {
	var permissionsMap = map[string]bool{
		read:            false,
		write:           false,
		annotate:        false,
		delete:          false,
		manage:          false,
		managedXrayMeta: false,
		distribute:      false,
	}
	for {
		answer := utils.AskFromList("", "Select permission value for "+permissionsOwner+" (press tab for options or enter to finish)", true, buildPermissionSuggestArray(permissionsMap), permissionSelectEnd)
		if answer == permissionSelectEnd {
			break
		}
		// if answer is a valid key we will mark it with false to remove it from the suggestion list
		if _, ok := permissionsMap[answer]; ok {
			permissionsMap[answer] = true
		} else {
			// answer is a var, we will add it to the final permissions slice result
			permissions = append(permissions, answer)
		}
	}
	for key, value := range permissionsMap {
		if value {
			permissions = append(permissions, key)
		}
	}
	return
}

func buildPermissionSuggestArray(permissionMap map[string]bool) (permissions []prompt.Suggest) {
	for key, value := range permissionMap {
		if !value {
			permissions = append(permissions, prompt.Suggest{Text: key})
		}
	}
	sort.Slice(permissions, func(i, j int) bool {
		return permissions[i].Text < permissions[j].Text
	})
	return
}

var questionMap = map[string]utils.QuestionInfo{
	Name: {
		Msg:          "",
		PromptPrefix: "Insert the permission target name >",
		AllowVars:    false,
		Writer:       utils.WriteStringAnswer,
		MapKey:       Name,
		Callback:     nil,
	},
	utils.OptionalKey: {
		Msg:          "",
		PromptPrefix: SelectPermissionTargetSectionMsg,
		AllowVars:    false,
		Writer:       nil,
		MapKey:       "",
		Callback:     permissionSectionCallBack,
	},
	Repo:          utils.FreeStringQuestionInfo,
	Build:         utils.FreeStringQuestionInfo,
	ReleaseBundle: utils.FreeStringQuestionInfo,
}

var reposQuestionInfo = utils.QuestionInfo{
	Msg:          "You can specify the name \"ANY\" to apply to all repositories, \"ANY REMOTE\" for all remote repositories or \"ANY LOCAL\" for all local repositories.\n" + utils.CommaSeparatedListMsg,
	PromptPrefix: "Insert the section's repositories value >",
}

var includePatternsQuestionInfo = utils.QuestionInfo{
	Msg:          utils.CommaSeparatedListMsg,
	PromptPrefix: "Insert value for include-patterns " + LeaveEmptyForDefault,
}

var excludePatternsQuestionInfo = utils.QuestionInfo{
	Msg:          utils.CommaSeparatedListMsg,
	PromptPrefix: "Insert value for exclude-patterns " + LeaveEmptyForDefault + " []:",
}

var configureActionsQuestionInfo = utils.QuestionInfo{
	PromptPrefix: "Configure actions for ",
	Options: []prompt.Suggest{
		{Text: Yes},
		{Text: No},
	},
}

type PermissionSectionAnswer struct {
	Repositories    string            `json:"repositories,omitempty"`
	IncludePatterns string            `json:"include-patterns,omitempty"`
	ExcludePatterns string            `json:"exclude-patterns,omitempty"`
	ActionsUsers    map[string]string `json:"actions-users,omitempty"`
	ActionsGroups   map[string]string `json:"actions-groups,omitempty"`
}
