package _go

import (
	"encoding/base64"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/artifactory/auth"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/artifactory/services/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/errors/httperrors"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/httpclient"
	"strings"
)

type GoService struct {
	client     *httpclient.HttpClient
	ArtDetails auth.ArtifactoryDetails
}

func NewGoService(client *httpclient.HttpClient) *GoService {
	return &GoService{client: client}
}

func (gs *GoService) GetJfrogHttpClient() *httpclient.HttpClient {
	return gs.client
}

func (gs *GoService) SetArtDetails(artDetails auth.ArtifactoryDetails) {
	gs.ArtDetails = artDetails
}

func (gs *GoService) PublishPackage(params GoParams) error {
	artifactsInfo := params.GetArtifactsInfo()
	url, err := utils.BuildArtifactoryUrl(gs.ArtDetails.GetUrl(), "api/go/"+artifactsInfo.GetTargetRepo(), make(map[string]string))
	clientDetails := gs.ArtDetails.CreateHttpClientDetails()

	utils.AddHeader("X-GO-MODULE-VERSION", artifactsInfo.GetVersion(), &clientDetails.Headers)
	utils.AddHeader("X-GO-MODULE-CONTENT", base64.StdEncoding.EncodeToString(artifactsInfo.GetModContent()), &clientDetails.Headers)

	if params.ShouldUseNewApi() {
		createUrlPath(artifactsInfo, &url)
	} else {
		addPropertiesHeaders(artifactsInfo.GetProps(), &clientDetails.Headers)
	}

	resp, body, err := gs.client.UploadFile(artifactsInfo.GetZipPath(), url, clientDetails, 0)
	if err != nil {
		return err
	}
	return httperrors.CheckResponseStatus(resp, body, 201)
}

// This is needed when using Artifactory older then 6.5.0
func addPropertiesHeaders(props string, headers *map[string]string) error {
	properties, err := utils.ParseProperties(props, utils.JoinCommas)
	if err != nil {
		return err
	}
	headersMap := properties.ToHeadersMap()
	for k, v := range headersMap {
		utils.AddHeader("X-ARTIFACTORY-PROPERTY-"+k, v, headers)
	}
	return nil
}

func createUrlPath(artifactsInfo ArtifactsInfo, url *string) error {
	*url = strings.Join([]string{*url, artifactsInfo.GetModuleId(), "@v", artifactsInfo.GetVersion() + ".zip"}, "/")
	properties, err := utils.ParseProperties(artifactsInfo.GetProps(), utils.JoinCommas)
	if err != nil {
		return err
	}
	*url = strings.Join([]string{*url, properties.ToEncodedString()}, ";")
	if strings.HasSuffix(*url, ";") {
		tempUrl := *url
		tempUrl = tempUrl[:len(tempUrl)-1]
		*url = tempUrl
	}
	return nil
}

type ArtifactsInfo interface {
	GetZipPath() string
	GetModContent() []byte
	GetProps() string
	GetVersion() string
	GetTargetRepo() string
	GetModuleId() string
}

type ArtifactsInfoImpl struct {
	ZipPath    string
	ModContent []byte
	Version    string
	Props      string
	TargetRepo string
	ModuleId   string
}

func (ai *ArtifactsInfoImpl) GetZipPath() string {
	return ai.ZipPath
}

func (ai *ArtifactsInfoImpl) GetModContent() []byte {
	return ai.ModContent
}

func (ai *ArtifactsInfoImpl) GetVersion() string {
	return ai.Version
}

func (ai *ArtifactsInfoImpl) GetProps() string {
	return ai.Props
}

func (ai *ArtifactsInfoImpl) GetTargetRepo() string {
	return ai.TargetRepo
}

func (ai *ArtifactsInfoImpl) GetModuleId() string {
	return ai.ModuleId
}

type GoParams interface {
	ShouldUseNewApi() bool
	GetArtifactsInfo() ArtifactsInfo
}

type GoParamsImpl struct {
	NewApi        bool
	ArtifactsInfo ArtifactsInfo
}

func (gpi *GoParamsImpl) ShouldUseNewApi() bool {
	return gpi.NewApi
}

func (gpi *GoParamsImpl) GetArtifactsInfo() ArtifactsInfo {
	return gpi.ArtifactsInfo
}
