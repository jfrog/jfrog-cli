package rtinstances

import (
	"github.com/jfrogdev/jfrog-cli-go/missioncontrol/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"encoding/json"
	"errors"
	"fmt"
)

func AddInstance(instanceName string, flags *AddInstanceFlags) error {
	data := AddInstanceRequestContent{
		Name: 	  	 instanceName,
		Url : 	  	 flags.ArtifactoryInstanceDetails.Url,
		User: 	  	 flags.ArtifactoryInstanceDetails.User,
		Password: 	 flags.ArtifactoryInstanceDetails.Password,
		Description: flags.Description,
		Location: 	 flags.Location}
	requestContent, err := json.Marshal(data)
	if err != nil {
		return cliutils.CheckError(errors.New("Failed to execute request. " + cliutils.GetDocumentationMessage()))
	}
	missionControlUrl := flags.MissionControlDetails.Url + "api/v1/instances";
	httpClientDetails := utils.GetMissionControlHttpClientDetails(flags.MissionControlDetails)
	resp, body, err := ioutils.SendPost(missionControlUrl, requestContent, httpClientDetails)
	if err != nil {
	    return err
	}
	if resp.StatusCode == 201 || resp.StatusCode == 204 {
		fmt.Println("Mission Control response: " + resp.Status)
	} else {
		return cliutils.CheckError(errors.New(resp.Status + ". " + utils.ReadMissionControlHttpMessage(body)))
	}
	return nil
}

type AddInstanceFlags struct {
	MissionControlDetails      *config.MissionControlDetails
	Description 			   string
	Location 				   string
	NodeId 					   string
	ArtifactoryInstanceDetails *utils.ArtifactoryInstanceDetails
}

type AddInstanceRequestContent struct {
	Name        string `json:"name,omitempty"`
	Url        	string `json:"url,omitempty"`
	User        string `json:"username,omitempty"`
	Password    string `json:"password,omitempty"`
	Location 	string `json:"location,omitempty"`
	Description string `json:"description,omitempty"`
}
