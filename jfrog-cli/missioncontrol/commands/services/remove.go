package services

import (
	"errors"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/missioncontrol/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/httputils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
)

func Remove(serviceName string, flags *RemoveFlags) error {
	missionControlUrl := flags.MissionControlDetails.Url + "api/v3/services/" + serviceName
	httpClientDetails := utils.GetMissionControlHttpClientDetails(flags.MissionControlDetails)
	resp, body, err := httputils.SendDelete(missionControlUrl, nil, httpClientDetails)
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
