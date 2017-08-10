package services

import (
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/errors/httperrors"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/types/httpclient"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services/utils/auth"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services/utils"
)

type SetPropsService struct {
	client     *httpclient.HttpClient
	ArtDetails *auth.ArtifactoryDetails
}

func NewSetPropsService(client *httpclient.HttpClient) *SetPropsService {
	return &SetPropsService{client: client}
}

func (sp *SetPropsService) GetArtifactoryDetails() *auth.ArtifactoryDetails {
	return sp.ArtDetails
}

func (sp *SetPropsService) SetArtifactoryDetails(rt *auth.ArtifactoryDetails) {
	sp.ArtDetails = rt
}

func (sp *SetPropsService) IsDryRun() bool {
	return false
}

func (sp *SetPropsService) SetProps(setPropsParams SetPropsParams) error {
	updatePropertiesBaseUrl := sp.GetArtifactoryDetails().Url + "api/storage"
	log.Info("Setting properties...")
	encodedParam, err := utils.EncodeParams(setPropsParams.GetProps())
	if err != nil {
		return err
	}
	for _, item := range setPropsParams.GetItems() {
		log.Info("Setting properties on:", item.GetItemRelativePath())
		httpClientsDetails := sp.GetArtifactoryDetails().CreateArtifactoryHttpClientDetails()
		setPropertiesUrl := updatePropertiesBaseUrl + "/" + item.GetItemRelativePath() + "?properties=" + encodedParam
		log.Debug("Sending set properties request:", setPropertiesUrl)
		resp, body, err := sp.client.SendPut(setPropertiesUrl, nil, httpClientsDetails)
		if err != nil {
			return err
		}
		if err = httperrors.CheckResponseStatus(resp, body, 204); err != nil {
			return err
		}
	}

	log.Info("Done setting properties.")
	return err
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
