package rtinstances

import (
	"github.com/jfrogdev/jfrog-cli-go/missioncontrol/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"errors"
	"fmt"
)

func Remove(instanceName string, flags *RemoveFlags) error {
	missionControlUrl := flags.MissionControlDetails.Url + "api/v1/instances/" + instanceName;
	httpClientDetails := utils.GetMissionControlHttpClientDetails(flags.MissionControlDetails)
	resp, body, err := ioutils.SendDelete(missionControlUrl, nil, httpClientDetails)
	if err != nil {
	    return err
	}
	if resp.StatusCode != 204 {
		return cliutils.CheckError(errors.New(resp.Status + ". " + utils.ReadMissionControlHttpMessage(body)))
	}
	fmt.Println("Mission Control response: " + resp.Status)
	return nil
}

type RemoveFlags struct {
	MissionControlDetails *config.MissionControlDetails
	Interactive 	      bool
}
