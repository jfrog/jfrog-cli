package commands

import (
	"errors"
	"github.com/jfrog/jfrog-cli-go/missioncontrol/utils"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/httpclient"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"net/http"
)

func JpdAdd(flags *JpdAddFlags) error {
	missionControlUrl := flags.MissionControlDetails.Url + "api/v1/jpds"
	httpClientDetails := utils.GetMissionControlHttpClientDetails(flags.MissionControlDetails)
	client, err := httpclient.ClientBuilder().Build()
	if err != nil {
		return err
	}
	resp, body, err := client.SendPost(missionControlUrl, flags.JpdConfig, httpClientDetails)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusCreated {
		return errorutils.CheckError(errors.New(resp.Status + ". " + utils.ReadMissionControlHttpMessage(body)))
	}

	log.Debug("Mission Control response: " + resp.Status)
	log.Output(clientutils.IndentJson(body))
	return nil
}

type JpdAddFlags struct {
	MissionControlDetails *config.MissionControlDetails
	JpdConfig             []byte
}
