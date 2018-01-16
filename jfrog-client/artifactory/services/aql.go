package services

import (
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/auth"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/httpclient"
)

type AqlService struct {
	client     *httpclient.HttpClient
	ArtDetails auth.ArtifactoryDetails
}

func NewAqlService(client *httpclient.HttpClient) *AqlService {
	return &AqlService{client: client}
}

func (s *AqlService) GetArtifactoryDetails() auth.ArtifactoryDetails {
	return s.ArtDetails
}

func (s *AqlService) SetArtifactoryDetails(rt auth.ArtifactoryDetails) {
	s.ArtDetails = rt
}

func (s *AqlService) IsDryRun() bool {
	return false
}

func (s *AqlService) GetJfrogHttpClient() *httpclient.HttpClient {
	return s.client
}

func (s *AqlService) ExecAql(aql string) ([]byte, error) {
	return s.exec(aql)
}

func (s *AqlService) exec(aql string) ([]byte, error) {
	return utils.ExecAql(aql, s)
}
