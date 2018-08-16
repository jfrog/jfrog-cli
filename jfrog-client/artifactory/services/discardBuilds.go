package services

import (
	"encoding/json"
	"errors"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/artifactory/auth"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/artifactory/services/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/httpclient"
	clientutils "github.com/jfrog/jfrog-cli-go/jfrog-client/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/errorutils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/log"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"
)

type DiscardBuildsService struct {
	client     *httpclient.HttpClient
	ArtDetails auth.ArtifactoryDetails
}

func NewDiscardBuildsService(client *httpclient.HttpClient) *DiscardBuildsService {
	return &DiscardBuildsService{client: client}
}

func (ds *DiscardBuildsService) DiscardBuilds(params DiscardBuildsParams) error {
	log.Info("Discarding builds...")

	discardUrl := ds.ArtDetails.GetUrl()
	restApi := path.Join("api/build/retention/", params.GetBuildName())
	requestFullUrl, err := utils.BuildArtifactoryUrl(discardUrl, restApi, make(map[string]string))
	if err != nil {
		return err
	}
	requestFullUrl += "?async=" + strconv.FormatBool(params.IsAsync())

	var excludeBuilds []string
	if params.GetExcludeBuilds() != "" {
		excludeBuilds = strings.Split(params.GetExcludeBuilds(), ",")
	}

	minimumBuildDateString, err := "", nil
	if params.GetMaxDays() != "" {
		minimumBuildDateString, err = calculateMinimumBuildDate(time.Now(), params.GetMaxDays())
		if err != nil {
			return err
		}
	}

	data := DiscardBuildsBody{
		ExcludeBuilds:    excludeBuilds,
		MinimumBuildDate: minimumBuildDateString,
		MaxBuilds:        params.GetMaxBuilds(),
		DeleteArtifacts:  params.IsDeleteArtifacts()}
	requestContent, err := json.Marshal(data)
	if err != nil {
		return errorutils.CheckError(err)
	}

	httpClientsDetails := ds.getArtifactoryDetails().CreateHttpClientDetails()
	utils.SetContentType("application/json", &httpClientsDetails.Headers)

	resp, body, err := ds.client.SendPost(requestFullUrl, requestContent, httpClientsDetails)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusNoContent {
		return errorutils.CheckError(errors.New("Artifactory response: " + resp.Status + "\n" + clientutils.IndentJson(body)))
	}

	if params.IsAsync() {
		log.Info("Builds are being discarded asynchronously.")
		return nil
	}

	log.Info("Builds discarded.")
	return nil
}

func calculateMinimumBuildDate(startingDate time.Time, maxDaysString string) (string, error) {
	maxDays, err := strconv.Atoi(maxDaysString)
	if err != nil {
		return "", errorutils.CheckError(err)
	}

	minimumBuildDate := startingDate.Add(-24 * time.Hour * (time.Duration(maxDays)))
	minimumBuildDateString := minimumBuildDate.Format("2006-01-02T15:04:05.000-0700")

	return minimumBuildDateString, nil
}

func (ds *DiscardBuildsService) getArtifactoryDetails() auth.ArtifactoryDetails {
	return ds.ArtDetails
}

type DiscardBuildsBody struct {
	MinimumBuildDate string   `json:"minimumBuildDate,omitempty"`
	MaxBuilds        string   `json:"count,omitempty"`
	ExcludeBuilds    []string `json:"buildNumbersNotToBeDiscarded,omitempty"`
	DeleteArtifacts  bool     `json:"deleteBuildArtifacts"`
}

type DiscardBuildsParams interface {
	GetBuildName() string
	GetMaxDays() string
	GetMaxBuilds() string
	GetExcludeBuilds() string
	IsDeleteArtifacts() bool
	IsAsync() bool
}

type DiscardBuildsParamsImpl struct {
	DeleteArtifacts bool
	BuildName       string
	MaxDays         string
	MaxBuilds       string
	ExcludeBuilds   string
	Async           bool
}

func (bd *DiscardBuildsParamsImpl) GetBuildName() string {
	return bd.BuildName
}

func (bd *DiscardBuildsParamsImpl) GetMaxDays() string {
	return bd.MaxDays
}

func (bd *DiscardBuildsParamsImpl) GetMaxBuilds() string {
	return bd.MaxBuilds
}

func (bd *DiscardBuildsParamsImpl) GetExcludeBuilds() string {
	return bd.ExcludeBuilds
}

func (bd *DiscardBuildsParamsImpl) IsDeleteArtifacts() bool {
	return bd.DeleteArtifacts
}

func (bd *DiscardBuildsParamsImpl) IsAsync() bool {
	return bd.Async
}
