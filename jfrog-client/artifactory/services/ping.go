package services

import (
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/auth"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/httpclient"
	"net/http"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
	clientutils "github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils"
	"errors"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services/utils"
)

type PingService struct {
	httpClient *httpclient.HttpClient
	ArtDetails auth.ArtifactoryDetails
}

func NewPingService(client *httpclient.HttpClient) *PingService {
	return &PingService{httpClient: client}
}

func (ps *PingService) GetArtifactoryDetails() auth.ArtifactoryDetails {
	return ps.ArtDetails
}

func (ps *PingService) SetArtifactoryDetails(rt auth.ArtifactoryDetails) {
	ps.ArtDetails = rt
}

func (ps *PingService) GetJfrogHttpClient() *httpclient.HttpClient {
	return ps.httpClient
}

func (ps *PingService) IsDryRun() bool {
	return false
}

func (ps *PingService) Ping() ([]byte, error) {
	url, err := utils.BuildArtifactoryUrl(ps.ArtDetails.GetUrl(),"api/system/ping", nil)
	if err != nil {
		return nil, err
	}
	resp, respBody, _, err := ps.httpClient.SendGet(url, true, ps.ArtDetails.CreateHttpClientDetails())
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errorutils.CheckError(errors.New("Artifactory response: " + resp.Status + "\n" + clientutils.IndentJson(respBody)))
	}
	log.Debug("Artifactory response: ", resp.Status)
	return respBody, nil
}
