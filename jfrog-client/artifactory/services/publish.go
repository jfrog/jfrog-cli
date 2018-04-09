package services

import (
	"encoding/json"
	"errors"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/auth"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/httpclient"
	clientutils "github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/buildinfo"
)

type buildInfoPublishService struct {
	client     *httpclient.HttpClient
	ArtDetails auth.ArtifactoryDetails
	DryRun     bool
}

func NewBuildInfoPublishService(client *httpclient.HttpClient) *buildInfoPublishService {
	return &buildInfoPublishService{client: client}
}

func (bip *buildInfoPublishService) GetArtifactoryDetails() auth.ArtifactoryDetails {
	return bip.ArtDetails
}

func (bip *buildInfoPublishService) SetArtifactoryDetails(rt auth.ArtifactoryDetails) {
	bip.ArtDetails = rt
}

func (bip *buildInfoPublishService) IsDryRun() bool {
	return bip.DryRun
}

func (bip *buildInfoPublishService) PublishBuildInfo(build *buildinfo.BuildInfo) error {
	content, err := json.Marshal(build)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if bip.IsDryRun() {
		log.Output(clientutils.IndentJson(content))
		return nil
	}
	httpClientsDetails := bip.GetArtifactoryDetails().CreateHttpClientDetails()
	utils.SetContentType("application/vnd.org.jfrog.artifactory+json", &httpClientsDetails.Headers)
	log.Info("Deploying build info...")
	resp, body, err := bip.client.SendPut(bip.ArtDetails.GetUrl()+"api/build/", content, httpClientsDetails)
	if err != nil {
		return err
	}
	if resp.StatusCode != 204 {
		return errorutils.CheckError(errors.New("Artifactory response: " + resp.Status + "\n" + clientutils.IndentJson(body)))
	}

	log.Debug("Artifactory response:", resp.Status)
	log.Info("Build info successfully deployed. Browse it in Artifactory under " + bip.GetArtifactoryDetails().GetUrl() + "webapp/builds/" + build.Name + "/" + build.Number)
	return nil
}
