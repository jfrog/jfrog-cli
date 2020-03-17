package replication

import (
	"encoding/json"

	"github.com/jfrog/jfrog-cli-go/artifactory/commands/utils"
	rtUtils "github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/utils/cliutils"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	clientutils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
)

type ReplicationCreateCommand struct {
	rtDetails    *config.ArtifactoryDetails
	templatePath string
	vars         string
}

func NewReplicationCreateCommand() *ReplicationCreateCommand {
	return &ReplicationCreateCommand{}
}

func (rcc *ReplicationCreateCommand) SetTemplatePath(path string) *ReplicationCreateCommand {
	rcc.templatePath = path
	return rcc
}

func (rcc *ReplicationCreateCommand) SetVars(vars string) *ReplicationCreateCommand {
	rcc.vars = vars
	return rcc
}

func (rcc *ReplicationCreateCommand) SetRtDetails(rtDetails *config.ArtifactoryDetails) *ReplicationCreateCommand {
	rcc.rtDetails = rtDetails
	return rcc
}

func (rcc *ReplicationCreateCommand) RtDetails() (*config.ArtifactoryDetails, error) {
	return rcc.rtDetails, nil
}

func (rcc *ReplicationCreateCommand) CommandName() string {
	return "rt_replication_create"
}

func (rcc *ReplicationCreateCommand) Run() (err error) {
	content, err := fileutils.ReadFile(rcc.templatePath)
	if errorutils.CheckError(err) != nil {
		return
	}
	// Replace vars string-by-string if needed
	if len(rcc.vars) > 0 {
		templateVars := cliutils.SpecVarsStringToMap(rcc.vars)
		content = cliutils.ReplaceVars(content, templateVars)
	}
	// Unmarshal template to a map
	var replicationConfigMap map[string]interface{}
	err = json.Unmarshal(content, &replicationConfigMap)
	if errorutils.CheckError(err) != nil {
		return
	}
	// All the values in the template are strings
	// Go over the the confMap and write the values with the correct type using the writersMap
	serverId := ""
	for key, value := range replicationConfigMap {
		if key == "serverId" {
			serverId = value.(string)
		} else {
			writersMap[key](&replicationConfigMap, key, value.(string))
		}
	}
	fillMissingDefaultValue(replicationConfigMap)
	// Write a JSON with the correct values
	content, err = json.Marshal(replicationConfigMap)
	if errorutils.CheckError(err) != nil {
		return
	}
	var params *clientutils.ReplicationParams
	err = json.Unmarshal(content, &params)
	if errorutils.CheckError(err) != nil {
		return
	}
	servicesManager, err := rtUtils.CreateServiceManager(rcc.rtDetails, false)
	if serverId != "" {
		updateArtifactoryInfo(params, serverId)
	}
	return servicesManager.CreateReplication(*params)
}

func fillMissingDefaultValue(replicationConfigMap map[string]interface{}) {
	if _, ok := replicationConfigMap["socketTimeoutMillis"]; !ok {
		writersMap["socketTimeoutMillis"](&replicationConfigMap, "socketTimeoutMillis", "15000")
	}
	if _, ok := replicationConfigMap["syncProperties"]; !ok {
		writersMap["syncProperties"](&replicationConfigMap, "syncProperties", "true")
	}
}

func updateArtifactoryInfo(param *clientutils.ReplicationParams, serverId string) error {
	singleConfig, err := config.GetArtifactorySpecificConfig(serverId)
	if err != nil {
		return err
	}
	param.URL, param.Password, param.Username = singleConfig.GetUrl(), singleConfig.GetPassword(), singleConfig.GetUser()
	return nil
}

var writersMap = map[string]utils.AnswerWriter{
	ServerId:               utils.WriteStringAnswer,
	RepoKey:                utils.WriteStringAnswer,
	CronExp:                utils.WriteStringAnswer,
	EnableEventReplication: utils.WriteBoolAnswer,
	Enabled:                utils.WriteBoolAnswer,
	SyncDeletes:            utils.WriteBoolAnswer,
	SyncProperties:         utils.WriteBoolAnswer,
	SyncStatistics:         utils.WriteBoolAnswer,
	PathPrefix:             utils.WriteStringAnswer,
	SocketTimeoutMillis:    utils.WriteIntAnswer,
}
