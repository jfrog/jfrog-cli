package services

import (
	"github.com/jfrog/gofrog/parallel"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/artifactory/auth"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/artifactory/services/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/errors/httperrors"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/httpclient"
	clientutils "github.com/jfrog/jfrog-cli-go/jfrog-client/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/log"
	"net/http"
)

type SetPropsService struct {
	client     *httpclient.HttpClient
	ArtDetails auth.ArtifactoryDetails
	Threads    int
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

func (sp *SetPropsService) GetThreads() int {
	return sp.Threads
}

func (sp *SetPropsService) SetProps(setPropsParams SetPropsParams) (int, error) {
	updatePropertiesBaseUrl := sp.GetArtifactoryDetails().GetUrl() + "api/storage"
	log.Info("Setting properties...")
	props, err := utils.ParseProperties(setPropsParams.GetProps(), utils.JoinCommas)
	if err != nil {
		return 0, err
	}
	successCounters := make([]int, sp.GetThreads())
	encodedParam := props.ToEncodedString()
	producerConsumer := parallel.NewBounedRunner(sp.GetThreads(), true)
	errorsQueue := utils.NewErrorsQueue(1)

	go func() {
		for _, item := range setPropsParams.GetItems() {
			relativePath := item.GetItemRelativePath()
			setPropsTask := func(threadId int) error {
				logMsgPrefix := clientutils.GetLogMsgPrefix(threadId, sp.IsDryRun())
				log.Info(logMsgPrefix+"Setting properties on:", relativePath)
				httpClientsDetails := sp.GetArtifactoryDetails().CreateHttpClientDetails()
				setPropertiesUrl := updatePropertiesBaseUrl + "/" + relativePath + "?properties=" + encodedParam
				log.Debug(logMsgPrefix+"Sending set properties request:", setPropertiesUrl)
				resp, body, err := sp.client.SendPut(setPropertiesUrl, nil, httpClientsDetails)
				if err != nil {
					return err
				}
				if err = httperrors.CheckResponseStatus(resp, body, http.StatusNoContent); err != nil {
					return err
				}
				successCounters[threadId]++
				return nil
			}

			producerConsumer.AddTaskWithError(setPropsTask, errorsQueue.AddError)
		}
		defer producerConsumer.Done()
	}()

	producerConsumer.Run()
	totalSuccess := 0
	for _, v := range successCounters {
		totalSuccess += v
	}

	err = errorsQueue.GetError()
	if err != nil {
		return totalSuccess, err
	}
	log.Info("Done setting properties.")
	return totalSuccess, nil
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
