package services

import (
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/auth"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/errors/httperrors"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/httpclient"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
	"net/http"
)

type SetPropsService struct {
	client     *httpclient.HttpClient
	ArtDetails auth.ArtifactoryDetails
}

func NewSetPropsService(client *httpclient.HttpClient) *SetPropsService {
	return &SetPropsService{client: client}
}

func (sp *SetPropsService) GetArtifactoryDetails() auth.ArtifactoryDetails {
	return sp.ArtDetails
}

func (sp *SetPropsService) SetArtifactoryDetails(rt auth.ArtifactoryDetails) {
	sp.ArtDetails = rt
}

func (sp *SetPropsService) IsDryRun() bool {
	return false
}

func (sp *SetPropsService) SetProps(setPropsParams SetPropsParams) (int, error) {
	updatePropertiesBaseUrl := sp.GetArtifactoryDetails().GetUrl() + "api/storage"
	log.Info("Setting properties...")
	props, err := utils.ParseProperties(setPropsParams.GetProps(), utils.JoinCommas)
	if err != nil {
		return 0, err
	}
	successCount := 0
	encodedParam := props.ToEncodedString()
	for _, item := range setPropsParams.GetItems() {
		log.Info("Setting properties on:", item.GetItemRelativePath())
		httpClientsDetails := sp.GetArtifactoryDetails().CreateHttpClientDetails()
		setPropertiesUrl := updatePropertiesBaseUrl + "/" + item.GetItemRelativePath() + "?properties=" + encodedParam
		log.Debug("Sending set properties request:", setPropertiesUrl)
		resp, body, err := sp.client.SendPut(setPropertiesUrl, nil, httpClientsDetails)
		if err != nil {
			log.Error(err)
			continue
		}
		if err = httperrors.CheckResponseStatus(resp, body, http.StatusNoContent); err != nil {
			log.Error(err)
			continue
		}
		successCount++
	}

	log.Info("Done setting properties.")
	return successCount, err
}

type SetPropsParams interface {
	GetItems() []utils.ResultItem
	GetProps() string
}

type SetPropsParamsImpl struct {
	Items []utils.ResultItem
	Props string
}

func (sp *SetPropsParamsImpl) GetItems() []utils.ResultItem {
	return sp.Items
}

func (sp *SetPropsParamsImpl) GetProps() string {
	return sp.Props
}
