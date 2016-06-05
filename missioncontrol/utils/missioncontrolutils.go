package utils

 import (
	 "github.com/jfrogdev/jfrog-cli-go/utils/config"
	 "github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	 "encoding/json"
 )

func GetMissionControlHttpClientDetails(missionControlDetails *config.MissionControlDetails) ioutils.HttpClientDetails {
	return ioutils.HttpClientDetails{
		User:     missionControlDetails.User,
		Password: missionControlDetails.Password,
		Headers:  map[string]string{"Content-Type": "application/json"}}
}

func ReadMissionControlHttpMessage(resp []byte) string {
	var response map[string][]HttpResponse
	err := json.Unmarshal(resp, &response)
	if err != nil {
		return string(resp)
	}
	responseMessage := ""
	for i := range response["errors"] {
		item := response["errors"][i]
		if item.Message != "" {
			if responseMessage != "" {
				responseMessage += ", "
			}
			responseMessage += item.Message
			for i := 0; i < len(item.Details); i++ {
				responseMessage += " " + item.Details[i]
			}
		}
	}
	if responseMessage == ""{
		return string(resp)
	}
	return responseMessage
}

type HttpResponse struct {
	Message string
	Type string
	Details []string
}

type LicenseRequestContent struct {
	Name      string `json:"instanceName,omitempty"`
	NodeID 	  string `json:"nodeId,omitempty"`
	Deploy 	  bool   `json:"deploy,omitempty"`
}

type ArtifactoryInstanceDetails struct {
	Url      string `json:"url,omitempty"`
	User     string `json:"user,omitempty"`
	Password string `json:"password,omitempty"`
	Name     string	`json:"instanceName,omitempty"`
}

