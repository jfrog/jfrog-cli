package replication

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/c-bata/go-prompt"
	"github.com/jfrog/jfrog-cli-go/artifactory/commands/utils"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

const (
	// Template types
	TemplateType = "templateType"
	JobType      = "jobType"
	Create       = "create"
	Pull         = "pull"
	Push         = "push"

	// Common replication configuration JSON keys
	Username               = "username"
	Password               = "password"
	URL                    = "url"
	CronExp                = "cronExp"
	RepoKey                = "repoKey"
	EnableEventReplication = "enableEventReplication"
	Enabled                = "enabled"
	SyncDeletes            = "syncDeletes"
	SyncProperties         = "syncProperties"
	SyncStatistics         = "syncStatistics"
	PathPrefix             = "pathPrefix"
	SocketTimeoutMillis    = "socketTimeoutMillis"
)

type ReplicationTemplateCommand struct {
	path string
}

func NewReplicationTemplateCommand() *ReplicationTemplateCommand {
	return &ReplicationTemplateCommand{}
}

func (rtc *ReplicationTemplateCommand) SetTemplatePath(path string) *ReplicationTemplateCommand {
	rtc.path = path
	return rtc
}

func (rtc *ReplicationTemplateCommand) CommandName() string {
	return "rt_replication_template"
}

func (rtc *ReplicationTemplateCommand) RtDetails() (*config.ArtifactoryDetails, error) {
	// Since it's a local command, usage won't be reported.
	return nil, nil
}

func (rtc *ReplicationTemplateCommand) Run() (err error) {
	replicationTemplateQuestionnaire := &utils.InteractiveQuestionnaire{
		MandatoryQuestionsKeys: []string{JobType, RepoKey},
		QuestionsMap:           questionMap,
	}
	err = replicationTemplateQuestionnaire.Perform()
	if err != nil {
		return err
	}
	resBytes, err := json.Marshal(replicationTemplateQuestionnaire.AnswersMap)
	if err != nil {
		return errorutils.CheckError(err)
	}
	if err = ioutil.WriteFile(rtc.path, resBytes, 0644); err != nil {
		return errorutils.CheckError(err)
	}
	log.Info(fmt.Sprintf("Replication creation config template successfully created at %s.", rtc.path))
	return nil
}

var questionMap = map[string]utils.QuestionInfo{
	utils.OptionalKey: {
		Msg:          "Select the next property",
		PromptPrefix: ">",
		AllowVars:    false,
		Writer:       nil,
		MapKey:       "",
		Callback:     optionalKeyCallback,
	},
	TemplateType: {
		Options: []prompt.Suggest{
			{Text: Create, Description: "Template for creating a new replication job"},
		},
		Msg:          "Select the template type",
		PromptPrefix: ">",
		AllowVars:    true,
		Writer:       nil,
		MapKey:       "",
		Callback:     nil,
	},
	JobType: {
		Options: []prompt.Suggest{
			{Text: Pull, Description: "Pull replication job"},
			{Text: Push, Description: "Push replication job"},
		},
		Msg:          "Select job type",
		PromptPrefix: ">",
		AllowVars:    true,
		Writer:       utils.WriteStringAnswer,
		MapKey:       JobType,
		Callback:     jobTypeCallback,
	},
	RepoKey: {
		Msg:          "Enter repo key",
		PromptPrefix: ">",
		AllowVars:    true,
		Writer:       utils.WriteStringAnswer,
		MapKey:       RepoKey,
		Callback:     nil,
	},
	Username: {
		Msg:          "Enter user name",
		PromptPrefix: ">",
		AllowVars:    true,
		Writer:       utils.WriteStringAnswer,
		MapKey:       Username,
		Callback:     nil,
	},
	Password: {
		Msg:          "Enter password",
		PromptPrefix: ">",
		AllowVars:    true,
		Writer:       utils.WriteStringAnswer,
		MapKey:       Password,
		Callback:     nil,
	},
	URL: {
		Msg:          "Enter URL",
		PromptPrefix: ">",
		AllowVars:    true,
		Writer:       utils.WriteStringAnswer,
		MapKey:       URL,
		Callback:     nil,
	},
	CronExp: {
		Msg:          "Enter cron expression",
		PromptPrefix: ">",
		AllowVars:    true,
		Writer:       utils.WriteStringAnswer,
		MapKey:       CronExp,
		Callback:     nil,
	},
	EnableEventReplication: utils.BoolQuestionInfo,
	Enabled:                utils.BoolQuestionInfo,
	SyncDeletes:            utils.BoolQuestionInfo,
	SyncProperties:         utils.BoolQuestionInfo,
	SyncStatistics:         utils.BoolQuestionInfo,
	PathPrefix: {
		Msg:          "Enter path prefix",
		PromptPrefix: ">",
		AllowVars:    false,
		Writer:       utils.WriteStringAnswer,
		MapKey:       PathPrefix,
		Callback:     nil,
	},
	SocketTimeoutMillis: {
		Msg:          "Enter socket timeout millis",
		PromptPrefix: ">",
		AllowVars:    false,
		Writer:       utils.WriteStringAnswer,
		MapKey:       SocketTimeoutMillis,
		Callback:     nil,
	},
}

func jobTypeCallback(iq *utils.InteractiveQuestionnaire, jobType string) (string, error) {
	iq.MandatoryQuestionsKeys = append(iq.MandatoryQuestionsKeys, CronExp)
	if jobType == Pull {
		iq.OptionalKeysSuggests = getAllPossibleOptionalRepoConfKeys()
	} else {
		iq.MandatoryQuestionsKeys = append(iq.MandatoryQuestionsKeys, Username, Password, URL)
		iq.OptionalKeysSuggests = getAllPossibleOptionalRepoConfKeys()
	}
	return "", nil
}

func getAllPossibleOptionalRepoConfKeys(values ...string) []prompt.Suggest {
	optionalKeys := []string{utils.WriteAndExist, Enabled, SyncDeletes, SyncProperties, SyncStatistics, PathPrefix, EnableEventReplication, SocketTimeoutMillis}
	if len(values) > 0 {
		optionalKeys = append(optionalKeys, values...)
	}
	return utils.GetSuggestsFromKeys(optionalKeys, suggestionMap)
}

//duplicate code
func optionalKeyCallback(iq *utils.InteractiveQuestionnaire, key string) (value string, err error) {
	if key != utils.WriteAndExist {
		valueQuestion := iq.QuestionsMap[key]
		valueQuestion.MapKey = key
		valueQuestion.PromptPrefix = "Insert the value for " + key
		if valueQuestion.Options != nil {
			valueQuestion.PromptPrefix += utils.PressTabMsg
		}
		valueQuestion.PromptPrefix += " >"
		value, err = iq.AskQuestion(valueQuestion)
	}
	return value, err
}

var suggestionMap = map[string]prompt.Suggest{
	utils.WriteAndExist:    {Text: utils.WriteAndExist},
	Username:               {Text: Username},
	Password:               {Text: Password},
	URL:                    {Text: URL},
	RepoKey:                {Text: RepoKey},
	CronExp:                {Text: CronExp},
	EnableEventReplication: {Text: EnableEventReplication},
	Enabled:                {Text: Enabled},
	SyncDeletes:            {Text: SyncDeletes},
	SyncProperties:         {Text: SyncProperties},
	SyncStatistics:         {Text: SyncStatistics},
	PathPrefix:             {Text: PathPrefix},
	SocketTimeoutMillis:    {Text: SocketTimeoutMillis},
}
