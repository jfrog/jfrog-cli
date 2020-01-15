package commands

import (
	"encoding/json"
	"errors"
	"github.com/jfrog/jfrog-cli-go/missioncontrol/utils"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/httpclient"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"net/http"
)

func LicenseRelease(bucketId, jpdId string, mcDetails *config.MissionControlDetails) error {
	postContent := LicenseReleaseRequestContent{
		Name: jpdId}
	requestContent, err := json.Marshal(postContent)
	if err != nil {
		return errorutils.CheckError(errors.New("Failed to marshal json: " + err.Error()))
	}
	missionControlUrl := mcDetails.Url + "api/v1/buckets/" + bucketId + "/release"
	httpClientDetails := utils.GetMissionControlHttpClientDetails(mcDetails)
	client, err := httpclient.ClientBuilder().Build()
	if err != nil {
		return err
	}
	resp, body, err := client.SendPost(missionControlUrl, requestContent, httpClientDetails)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return errorutils.CheckError(errors.New(resp.Status + ". " + utils.ReadMissionControlHttpMessage(body)))
	}
	log.Debug("Mission Control response: " + resp.Status)
	return nil
}

type LicenseReleaseRequestContent struct {
	Name string `json:"name,omitempty"`
}
