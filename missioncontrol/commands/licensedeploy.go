package commands

import (
	"encoding/json"
	"errors"
	"github.com/jfrog/jfrog-cli-go/missioncontrol/utils"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/httpclient"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"net/http"
)

func LicenseDeploy(bucketId, jpdId string, flags *LicenseDeployFlags) error {
	postContent := LicenseDeployRequestContent{
		JpdId:        jpdId,
		LicenseCount: flags.LicenseCount,
	}
	requestContent, err := json.Marshal(postContent)
	if err != nil {
		return errorutils.CheckError(errors.New("Failed to marshal json: " + err.Error()))
	}
	missionControlUrl := flags.MissionControlDetails.Url + "api/v1/buckets/" + bucketId + "/deploy"
	httpClientDetails := utils.GetMissionControlHttpClientDetails(flags.MissionControlDetails)
	client, err := httpclient.ClientBuilder().Build()
	if err != nil {
		return err
	}
	resp, body, err := client.SendPost(missionControlUrl, requestContent, httpClientDetails)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errorutils.CheckError(errors.New(resp.Status + ". " + utils.ReadMissionControlHttpMessage(body)))
	}
	log.Debug("Mission Control response: " + resp.Status)
	log.Output(clientutils.IndentJson(body))
	return nil
}

const (
	DefaultLicenseCount = 1
)

type LicenseDeployRequestContent struct {
	JpdId        string `json:"jpd_id,omitempty"`
	LicenseCount int    `json:"license_count,omitempty"`
}

type LicenseDeployFlags struct {
	MissionControlDetails *config.MissionControlDetails
	LicenseCount          int
}
