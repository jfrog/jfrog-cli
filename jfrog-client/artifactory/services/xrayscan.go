package services

import (
	"encoding/json"
	"errors"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/auth"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/httpclient"
	clientutils "github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/httputils"
	"net/http"
	"time"
)

const SCAN_BUILD_API_URL = "api/xray/scanBuild"
const XRAY_SCAN_RETRY_CONSECUTIVE_RETRIES = 10           // Retrying to resume the scan 10 times after a stable connection
const XRAY_SCAN_CONNECTION_TIMEOUT = 90 * time.Second    // Expecting \r\n every 30 seconds
const XRAY_SCAN_SLEEP_BETWEEN_RETRIES = 15 * time.Second // 15 seconds sleep between retry
const XRAY_SCAN_STABLE_CONNECTION_WINDOW = 100 * time.Second
const XRAY_FATAL_FAIL_STATUS = -1

type XrayScanService struct {
	client     *httpclient.HttpClient
	ArtDetails auth.ArtifactoryDetails
}

func NewXrayScanService(client *httpclient.HttpClient) *XrayScanService {
	return &XrayScanService{client: client}
}

func (ps *XrayScanService) GetArtifactoryDetails() auth.ArtifactoryDetails {
	return ps.ArtDetails
}

func (ps *XrayScanService) SetArtifactoryDetails(rt auth.ArtifactoryDetails) {
	ps.ArtDetails = rt
}

func (ps *XrayScanService) ScanBuild(scanParams XrayScanParams) ([]byte, error) {
	url := ps.ArtDetails.GetUrl()
	requestFullUrl, err := utils.BuildArtifactoryUrl(url, SCAN_BUILD_API_URL, make(map[string]string))
	if err != nil {
		return []byte{}, err
	}
	data := XrayScanBody{
		BuildName:   scanParams.GetBuildName(),
		BuildNumber: scanParams.GetBuildNumber(),
		Context:     clientutils.ClientAgent,
	}

	requestContent, err := json.Marshal(data)
	if err != nil {
		return []byte{}, errorutils.CheckError(err)
	}

	httpClientsDetails := ps.ArtDetails.CreateHttpClientDetails()
	utils.SetContentType("application/json", &httpClientsDetails.Headers)

	connection := httputils.RetryableConnection{
		ReadTimeout:            XRAY_SCAN_CONNECTION_TIMEOUT,
		RetriesNum:             XRAY_SCAN_RETRY_CONSECUTIVE_RETRIES,
		StableConnectionWindow: XRAY_SCAN_STABLE_CONNECTION_WINDOW,
		SleepBetweenRetries:    XRAY_SCAN_SLEEP_BETWEEN_RETRIES,
		ConnectHandler: func() (*http.Response, error) {
			return execScanRequest(requestFullUrl, requestContent, httpClientsDetails)
		},
		ErrorHandler: func(content []byte) error {
			return checkForXrayResponseError(content, true)
		},
	}
	result, err := connection.Do()
	if err != nil {
		return []byte{}, err
	}

	return result, nil
}

func isFatalScanError(errResp *errorResponse) bool {
	if errResp == nil {
		return false
	}
	for _, v := range errResp.Errors {
		if v.Status == XRAY_FATAL_FAIL_STATUS {
			return true
		}
	}
	return false
}

func checkForXrayResponseError(content []byte, ignoreFatalError bool) error {
	respErrors := &errorResponse{}
	err := json.Unmarshal(content, respErrors)
	if errorutils.CheckError(err) != nil {
		return err
	}

	if respErrors.Errors == nil {
		return nil
	}

	if ignoreFatalError && isFatalScanError(respErrors) {
		// fatal error should be interpreted as no errors so no more retries will accrue
		return nil
	}
	return errorutils.CheckError(errors.New("Artifactory response: " + string(content)))
}

func execScanRequest(url string, content []byte, httpClientsDetails httputils.HttpClientDetails) (*http.Response, error) {
	resp, _, _, err := httputils.Send("POST", url, content, true, false, httpClientsDetails)
	if err != nil {
		return resp, err
	}

	if resp.StatusCode != http.StatusOK {
		err = errorutils.CheckError(errors.New("Artifactory Response: " + resp.Status))
	}
	return resp, err
}

type errorResponse struct {
	Errors []errorsStatusResponse `json:"errors,omitempty"`
}

type errorsStatusResponse struct {
	Status int `json:"status,omitempty"`
}

type XrayScanBody struct {
	BuildName   string `json:"buildName,omitempty"`
	BuildNumber string `json:"buildNumber,omitempty"`
	Context     string `json:"context,omitempty"`
}

type XrayScanParams interface {
	GetBuildName() string
	GetBuildNumber() string
}

type XrayScanParamsImpl struct {
	BuildName   string
	BuildNumber string
}

func (bp *XrayScanParamsImpl) GetBuildName() string {
	return bp.BuildName
}

func (bp *XrayScanParamsImpl) GetBuildNumber() string {
	return bp.BuildNumber
}
