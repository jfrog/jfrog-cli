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

func AddJpd(flags *AddJpdFlags) error {
	missionControlUrl := flags.MissionControlDetails.Url + "api/v1/jpds"
	httpClientDetails := utils.GetMissionControlHttpClientDetails(flags.MissionControlDetails)
	client, err := httpclient.ClientBuilder().Build()
	if err != nil {
		return err
	}
	resp, body, err := client.SendPost(missionControlUrl, flags.JpdSpec, httpClientDetails)
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

type AddJpdFlags struct {
	MissionControlDetails *config.MissionControlDetails
	JpdSpec               []byte
}

type AddJpdRequestContent struct {
	Name     string       `json:"name,omitempty"`
	Url      string       `json:"url,omitempty"`
	Location *JpdLocation `json:"location,omitempty"`
	Token    string       `json:"token,omitempty"`
	Tags     []string     `json:"tags,omitempty"`
}

type JpdLocation struct {
	CityName    string  `json:"city_name,omitempty"`
	CountryCode string  `json:"country_code,omitempty"`
	Latitude    float64 `json:"latitude,omitempty"`
	Longitude   float64 `json:"longitude,omitempty"`
}
