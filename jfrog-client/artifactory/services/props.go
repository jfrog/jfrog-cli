package services

import (
	"github.com/jfrog/gofrog/parallel"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/artifactory/auth"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/artifactory/services/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/errors/httperrors"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/httpclient"
	clientutils "github.com/jfrog/jfrog-cli-go/jfrog-client/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/io/httputils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/log"
	"net/http"
	"net/url"
	"strings"
)

type PropsService struct {
	client     *httpclient.HttpClient
	ArtDetails auth.ArtifactoryDetails
	Threads    int
}

func NewPropsService(client *httpclient.HttpClient) *PropsService {
	return &PropsService{client: client}
}

func (ps *PropsService) GetArtifactoryDetails() auth.ArtifactoryDetails {
	return ps.ArtDetails
}

func (ps *PropsService) SetArtifactoryDetails(rt auth.ArtifactoryDetails) {
	ps.ArtDetails = rt
}

func (ps *PropsService) IsDryRun() bool {
	return false
}

func (ps *PropsService) GetThreads() int {
	return ps.Threads
}

func (ps *PropsService) SetProps(propsParams PropsParams) (int, error) {
	log.Info("Setting properties...")
	totalSuccess, err := ps.performRequest(propsParams, false)
	if err != nil {
		return totalSuccess, err
	}
	log.Info("Done setting properties.")
	return totalSuccess, nil
}

func (ps *PropsService) DeleteProps(propsParams PropsParams) (int, error) {
	log.Info("Deleting properties...")
	totalSuccess, err := ps.performRequest(propsParams, true)
	if err != nil {
		return totalSuccess, err
	}
	log.Info("Done deleting properties.")
	return totalSuccess, nil
}

type PropsParams interface {
	GetItems() []utils.ResultItem
	GetProps() string
}

type PropsParamsImpl struct {
	Items []utils.ResultItem
	Props string
}

func (sp *PropsParamsImpl) GetItems() []utils.ResultItem {
	return sp.Items
}

func (sp *PropsParamsImpl) GetProps() string {
	return sp.Props
}

func (ps *PropsService) performRequest(propsParams PropsParams, isDelete bool) (int, error) {
	updatePropertiesBaseUrl := ps.GetArtifactoryDetails().GetUrl() + "api/storage"
	var encodedParam string
	if !isDelete {
		props, err := utils.ParseProperties(propsParams.GetProps(), utils.JoinCommas)
		if err != nil {
			return 0, err
		}
		encodedParam = props.ToEncodedString()
	} else {
		propList := strings.Split(propsParams.GetProps(), ",")
		for _, prop := range propList {
			encodedParam += url.QueryEscape(prop) + ","
		}
		// Remove trailing comma
		if strings.HasSuffix(encodedParam, ",") {
			encodedParam = encodedParam[:len(encodedParam)-1]
		}

	}

	successCounters := make([]int, ps.GetThreads())
	producerConsumer := parallel.NewBounedRunner(ps.GetThreads(), true)
	errorsQueue := utils.NewErrorsQueue(1)

	go func() {
		for _, item := range propsParams.GetItems() {
			relativePath := item.GetItemRelativePath()
			setPropsTask := func(threadId int) error {
				logMsgPrefix := clientutils.GetLogMsgPrefix(threadId, ps.IsDryRun())
				httpClientsDetails := ps.GetArtifactoryDetails().CreateHttpClientDetails()
				setPropertiesUrl := updatePropertiesBaseUrl + "/" + relativePath + "?properties=" + encodedParam
				var resp *http.Response
				var body []byte
				var err error
				if isDelete {
					resp, body, err = ps.sendDeleteRequest(logMsgPrefix, relativePath, setPropertiesUrl, httpClientsDetails)
				} else {
					resp, body, err = ps.sendPutRequest(logMsgPrefix, relativePath, setPropertiesUrl, httpClientsDetails)
				}

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

	err := errorsQueue.GetError()
	if err != nil {
		return totalSuccess, err
	}
	return totalSuccess, nil
}

func (ps *PropsService) sendDeleteRequest(logMsgPrefix, relativePath, setPropertiesUrl string, httpClientsDetails httputils.HttpClientDetails) (resp *http.Response, body []byte, err error) {
	log.Info(logMsgPrefix+"Deleting properties on:", relativePath)
	log.Debug(logMsgPrefix+"Sending delete properties request:", setPropertiesUrl)
	resp, body, err = ps.client.SendDelete(setPropertiesUrl, nil, httpClientsDetails)
	return
}

func (ps *PropsService) sendPutRequest(logMsgPrefix, relativePath, setPropertiesUrl string, httpClientsDetails httputils.HttpClientDetails) (resp *http.Response, body []byte, err error) {
	log.Info(logMsgPrefix+"Setting properties on:", relativePath)
	log.Debug(logMsgPrefix+"Sending set properties request:", setPropertiesUrl)
	resp, body, err = ps.client.SendPut(setPropertiesUrl, nil, httpClientsDetails)
	return
}
