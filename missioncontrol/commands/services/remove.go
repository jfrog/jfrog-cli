package services

import (
	"errors"
	"github.com/jfrog/jfrog-cli-go/missioncontrol/utils"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/jfrog/jfrog-client-go/httpclient"
)

func Remove(serviceName string, flags *RemoveFlags) error {
	missionControlUrl := flags.MissionControlDetails.Url + "api/v3/services/" + serviceName
	httpClientDetails := utils.GetMissionControlHttpClientDetails(flags.MissionControlDetails)
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

type RemoveFlags struct {
	MissionControlDetails *config.MissionControlDetails
	Interactive           bool
}
