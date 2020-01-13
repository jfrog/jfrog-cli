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

func LicenseAcquire(bucketId string, name string, mcDetails *config.MissionControlDetails) error {
	postContent := LicenseAcquireRequestContent{
		Name:         name,
		LicenseCount: 1,
	}
	requestContent, err := json.Marshal(postContent)
	if err != nil {
		return errorutils.CheckError(errors.New("Failed to marshal json: " + err.Error()))
	}
	missionControlUrl := mcDetails.Url + "api/v1/buckets/" + bucketId + "/acquire"
	httpClientDetails := utils.GetMissionControlHttpClientDetails(mcDetails)
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

	// Extract license from response
	var licenseKeys licenseKeys
	err = json.Unmarshal(body, &licenseKeys)
	if err != nil {
		return errorutils.CheckError(err)
	}
	if len(licenseKeys.LicenseKeys) < 1 {
		return errorutils.CheckError(errors.New("failed to acquire license key from Mission Control: received 0 license keys"))
	}
	// Print license to log
	log.Output(licenseKeys.LicenseKeys[0])
	return nil
}

type LicenseAcquireRequestContent struct {
	Name         string `json:"name,omitempty"`
	LicenseCount int    `json:"license_count,omitempty"`
}

type licenseKeys struct {
	LicenseKeys []string `json:"license_keys,omitempty"`
}
