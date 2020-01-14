package commands

import (
	"errors"
	"github.com/jfrog/jfrog-cli-go/missioncontrol/utils"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/httpclient"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

func JpdDelete(jpdId string, mcDetails *config.MissionControlDetails) error {
	missionControlUrl := mcDetails.Url + "api/v1/jpds/" + jpdId
	httpClientDetails := utils.GetMissionControlHttpClientDetails(mcDetails)
	client, err := httpclient.ClientBuilder().Build()
	if err != nil {
		return err
	}
	resp, body, err := client.SendDelete(missionControlUrl, nil, httpClientDetails)
	if err != nil {
		return err
	}
	if resp.StatusCode != 204 {
		return errorutils.CheckError(errors.New(resp.Status + ". " + utils.ReadMissionControlHttpMessage(body)))
	}
	log.Debug("Mission Control response: " + resp.Status)
	return nil
}
