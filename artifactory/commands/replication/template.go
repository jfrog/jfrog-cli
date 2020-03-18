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
	// Strings for prompt questions
	SelectConfigKeyMsg = "Select the next configuration key" + utils.PressTabMsg

	// Template types
	TemplateType = "templateType"
	JobType      = "jobType"
	Create       = "create"
	Pull         = "pull"
	Push         = "push"

	// Common replication configuration JSON keys
	ServerId               = "serverId"
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
		Msg:          "",
		PromptPrefix: "Select the next property >",
		AllowVars:    false,
		Writer:       nil,
		MapKey:       "",
		Callback:     utils.OptionalKeyCallback,
	},
	TemplateType: {
		Options: []prompt.Suggest{
			{Text: Create, Description: "Template for creating a new replication"},
		},
		Msg:          "",
		PromptPrefix: "Select the template type >",
		AllowVars:    true,
		Writer:       nil,
		MapKey:       "",
		Callback:     nil,
	},
	JobType: {
		Options: []prompt.Suggest{
			{Text: Pull, Description: "Pull replication"},
			{Text: Push, Description: "Push replication"},
		},
		Msg:          "",
		PromptPrefix: "Select job type >",
		AllowVars:    true,
		Writer:       nil,
		MapKey:       "",
		Callback:     jobTypeCallback,
	},
	RepoKey: {
		Msg:          "",
		PromptPrefix: "Enter repo key >",
		AllowVars:    true,
		Writer:       utils.WriteStringAnswer,
		MapKey:       RepoKey,
		Callback:     nil,
	},
	ServerId: {
		Msg:          "",
		PromptPrefix: "Enter server id >",
		AllowVars:    true,
		Writer:       utils.WriteStringAnswer,
		MapKey:       ServerId,
		Callback:     nil,
	},
	CronExp: {
		Msg:          "",
		PromptPrefix: "Enter cron expression >",
		AllowVars:    true,
		Writer:       utils.WriteStringAnswer,
		MapKey:       CronExp,
		Callback:     nil,
	},
	EnableEventReplication: BoolToStringQuestionInfo,
	Enabled:                BoolToStringQuestionInfo,
	SyncDeletes:            BoolToStringQuestionInfo,
	SyncProperties:         BoolToStringQuestionInfo,
	SyncStatistics:         BoolToStringQuestionInfo,
	PathPrefix: {
		Msg:          "",
		PromptPrefix: "Enter path prefix >",
		AllowVars:    false,
		Writer:       utils.WriteStringAnswer,
		MapKey:       PathPrefix,
		Callback:     nil,
	},
	SocketTimeoutMillis: {
		Msg:          "",
		PromptPrefix: "Enter socket timeout millis >",
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
		iq.MandatoryQuestionsKeys = append(iq.MandatoryQuestionsKeys, ServerId)
		iq.OptionalKeysSuggests = getAllPossibleOptionalRepoConfKeys()
	}
	return "", nil
}

func getAllPossibleOptionalRepoConfKeys(values ...string) []prompt.Suggest {
	optionalKeys := []string{utils.SaveAndExit, Enabled, SyncDeletes, SyncProperties, SyncStatistics, PathPrefix, EnableEventReplication, SocketTimeoutMillis}
	if len(values) > 0 {
		optionalKeys = append(optionalKeys, values...)
	}
	return utils.GetSuggestsFromKeys(optionalKeys, suggestionMap)
}

// Specific writers for repo templates, since all the values in the templates should be written as string.
var BoolToStringQuestionInfo = utils.QuestionInfo{
	Options:   utils.GetBoolSuggests(),
	AllowVars: true,
	Writer:    utils.WriteStringAnswer,
}

var suggestionMap = map[string]prompt.Suggest{
	utils.SaveAndExit:      {Text: utils.SaveAndExit},
	ServerId:               {Text: ServerId},
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
