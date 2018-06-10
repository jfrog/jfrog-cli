package services

import (
	"errors"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/artifactory/auth"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/artifactory/services/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/httpclient"
	clientutils "github.com/jfrog/jfrog-cli-go/jfrog-client/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/errorutils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/log"
	"net/http"
)

type DeleteService struct {
	client     *httpclient.HttpClient
	ArtDetails auth.ArtifactoryDetails
	DryRun     bool
}

func NewDeleteService(client *httpclient.HttpClient) *DeleteService {
	return &DeleteService{client: client}
}

func (ds *DeleteService) GetArtifactoryDetails() auth.ArtifactoryDetails {
	return ds.ArtDetails
}

func (ds *DeleteService) SetArtifactoryDetails(rt auth.ArtifactoryDetails) {
	ds.ArtDetails = rt
}

func (ds *DeleteService) IsDryRun() bool {
	return ds.DryRun
}

func (ds *DeleteService) GetJfrogHttpClient() *httpclient.HttpClient {
	return ds.client
}

func (ds *DeleteService) GetPathsToDelete(deleteParams DeleteParams) (resultItems []utils.ResultItem, err error) {
	log.Info("Searching artifacts...")
	// Search paths using AQL.
	if deleteParams.GetSpecType() == utils.AQL {
		if resultItemsTemp, e := utils.AqlSearchBySpec(deleteParams.GetFile(), ds); e == nil {
			resultItems = append(resultItems, resultItemsTemp...)
		} else {
			err = e
			return
		}
	} else {

		deleteParams.SetIncludeDirs(true)
		tempResultItems, e := utils.AqlSearchDefaultReturnFields(deleteParams.GetFile(), ds)
		if e != nil {
			err = e
			return
		}
		paths := utils.ReduceDirResult(tempResultItems, utils.FilterTopChainResults)
		resultItems = append(resultItems, paths...)
	}
	utils.LogSearchResults(len(resultItems))
	return
}

func (ds *DeleteService) DeleteFiles(deleteItems []DeleteItem, conf utils.CommonConf) (int, error) {
	deletedCount := 0
	for _, v := range deleteItems {
		fileUrl, err := utils.BuildArtifactoryUrl(conf.GetArtifactoryDetails().GetUrl(), v.GetItemRelativePath(), make(map[string]string))
		if err != nil {
			return deletedCount, err
		}
		if conf.IsDryRun() {
			log.Info("[Dry run] Deleting:", v.GetItemRelativePath())
			continue
		}

		log.Info("Deleting:", v.GetItemRelativePath())
		httpClientsDetails := conf.GetArtifactoryDetails().CreateHttpClientDetails()
		resp, body, err := ds.client.SendDelete(fileUrl, nil, httpClientsDetails)
		if err != nil {
			log.Error(err)
			continue
		}
		if resp.StatusCode != http.StatusNoContent {
			log.Error(errorutils.CheckError(errors.New("Artifactory response: " + resp.Status + "\n" + clientutils.IndentJson(body))))
			continue
		}
		deletedCount++
		log.Debug("Artifactory response:", resp.Status)
	}
	return deletedCount, nil
}

type DeleteConfiguration struct {
	ArtDetails auth.ArtifactoryDetails
	DryRun     bool
}

func (conf *DeleteConfiguration) GetArtifactoryDetails() auth.ArtifactoryDetails {
	return conf.ArtDetails
}

func (conf *DeleteConfiguration) SetArtifactoryDetails(art auth.ArtifactoryDetails) {
	conf.ArtDetails = art
}

func (conf *DeleteConfiguration) IsDryRun() bool {
	return conf.DryRun
}

type DeleteParams interface {
	utils.FileGetter
	GetFile() *utils.ArtifactoryCommonParams
	SetIncludeDirs(includeDirs bool)
}

type DeleteParamsImpl struct {
	*utils.ArtifactoryCommonParams
}

func (ds *DeleteParamsImpl) GetFile() *utils.ArtifactoryCommonParams {
	return ds.ArtifactoryCommonParams
}

func (ds *DeleteParamsImpl) SetIncludeDirs(includeDirs bool) {
	ds.IncludeDirs = includeDirs
}

type DeleteItem interface {
	GetItemRelativePath() string
}
