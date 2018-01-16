package rtinstances

import (
	"encoding/json"
	"errors"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/missioncontrol/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/httputils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
	"net/http"
)

func AddInstance(instanceName string, flags *AddInstanceFlags) error {
	data := AddInstanceRequestContent{
		Name:        instanceName,
		Url:         flags.ArtifactoryInstanceDetails.Url,
		User:        flags.ArtifactoryInstanceDetails.User,
		Password:    flags.ArtifactoryInstanceDetails.Password,
		Description: flags.Description,
		Location:    flags.Location}
	requestContent, err := json.Marshal(data)
	if err != nil {
		return errorutils.CheckError(errors.New("Failed to execute request. " + cliutils.GetDocumentationMessage()))
	}
	missionControlUrl := flags.MissionControlDetails.Url + "api/v1/instances"
	httpClientDetails := utils.GetMissionControlHttpClientDetails(flags.MissionControlDetails)
	resp, body, err := httputils.SendPost(missionControlUrl, requestContent, httpClientDetails)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		return errorutils.CheckError(errors.New(resp.Status + ". " + utils.ReadMissionControlHttpMessage(body)))
	}

	log.Debug("Mission Control response: " + resp.Status)
	return nil
}

type AddInstanceFlags struct {
	MissionControlDetails      *config.MissionControlDetails
	Description                string
	Location                   string
	NodeId                     string
	ArtifactoryInstanceDetails *utils.ArtifactoryInstanceDetails
}

type AddInstanceRequestContent struct {
	Name        string `json:"name,omitempty"`
	Url         string `json:"url,omitempty"`
	User        string `json:"username,omitempty"`
	Password    string `json:"password,omitempty"`
	Location    string `json:"location,omitempty"`
	Description string `json:"description,omitempty"`
}
