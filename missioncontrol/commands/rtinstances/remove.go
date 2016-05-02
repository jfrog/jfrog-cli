package rtinstances

import (
	"github.com/jfrogdev/jfrog-cli-go/missioncontrol/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"fmt"
)

func Remove(instanceName string, flags *RemoveFlags) {
	missionControlUrl := flags.MissionControlDetails.Url + "api/v1/instances/" + instanceName;
	httpClientDetails := utils.GetMissionControlHttpClientDetails(flags.MissionControlDetails)
	resp, body := ioutils.SendDelete(missionControlUrl, nil, httpClientDetails)
	if resp.StatusCode != 204 {
		cliutils.Exit(cliutils.ExitCodeError, resp.Status + ". " + utils.ReadMissionControlHttpMessage(body))
	}
	fmt.Println("Mission Control response: " + resp.Status)
}

type RemoveFlags struct {
	MissionControlDetails *config.MissionControlDetails
	Interactive 	      bool
}
