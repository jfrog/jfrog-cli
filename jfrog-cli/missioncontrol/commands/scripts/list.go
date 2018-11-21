package scripts

import (
	"fmt"
	"errors"
	"net/http"

	"github.com/jfrog/jfrog-cli-go/jfrog-cli/missioncontrol/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-client-go/httpclient"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

// ListScripts gets a list of MissionControl scripts
func ListScripts(c *config.MissionControlDetails) error {
	missionControlURL := c.Url + "api/v3/scripts"
	httpClientDetails := utils.GetMissionControlHttpClientDetails(c)
	client := httpclient.NewDefaultHttpClient()
	resp, body, _, err := client.SendGet(missionControlURL, true, httpClientDetails)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errorutils.CheckError(errors.New(resp.Status + ". " + utils.ReadMissionControlHttpMessage(body)))
	}

	log.Debug("Mission Control response: " + resp.Status)
	fmt.Println(string(body))
	return nil
}
