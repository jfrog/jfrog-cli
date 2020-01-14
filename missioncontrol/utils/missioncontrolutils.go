package utils

import (
	"encoding/json"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/utils/io/httputils"
)

func GetMissionControlHttpClientDetails(missionControlDetails *config.MissionControlDetails) httputils.HttpClientDetails {
	return httputils.HttpClientDetails{
		AccessToken: missionControlDetails.AccessToken,
		Headers:     map[string]string{"Content-Type": "application/json"}}
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
	if responseMessage == "" {
		return string(resp)
	}
	return responseMessage
}

type HttpResponse struct {
	Message string
	Type    string
	Details []string
}
