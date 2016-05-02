package rtinstances


import (
	"github.com/jfrogdev/jfrog-cli-go/missioncontrol/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"encoding/json"
	"fmt"
)

func DetachLic(instanceName string, flags *DetachLicFlags) {
	bucketId := flags.BucketId
	postContent := utils.LicenseRequestContent{
		Name: 	  	 instanceName,
		NodeID:	     flags.NodeId}
	requestContent, err := json.Marshal(postContent)
	if err != nil {
		cliutils.Exit(cliutils.ExitCodeError, "Failed to marshal json. " + cliutils.GetDocumentationMessage())
	}
	missionControlUrl := flags.MissionControlDetails.Url + "api/v1/buckets/" + bucketId + "/licenses";
	httpClientDetails := utils.GetMissionControlHttpClientDetails(flags.MissionControlDetails)
	resp, body := ioutils.SendDelete(missionControlUrl, requestContent, httpClientDetails)
	if resp.StatusCode != 200 {
		cliutils.Exit(cliutils.ExitCodeError, resp.Status + ". " + utils.ReadMissionControlHttpMessage(body))
	}
	fmt.Println("Mission Control response: " + resp.Status)
}

type DetachLicFlags struct {
	MissionControlDetails *config.MissionControlDetails
	Interactive 	      bool
	NodeId 			      string
	BucketId 		      string
}
