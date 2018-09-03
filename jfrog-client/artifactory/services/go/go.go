package _go

import (
	"encoding/base64"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/artifactory/auth"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/artifactory/services/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/errors/httperrors"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/httpclient"
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
	url, err := utils.BuildArtifactoryUrl(gs.ArtDetails.GetUrl(), "api/go/"+params.GetTargetRepo(), make(map[string]string))
	clientDetails := gs.ArtDetails.CreateHttpClientDetails()

	utils.AddHeader("X-GO-MODULE-VERSION", params.GetVersion(), &clientDetails.Headers)
	utils.AddHeader("X-GO-MODULE-CONTENT", base64.StdEncoding.EncodeToString(params.GetModContent()), &clientDetails.Headers)
	addPropertiesHeaders(params.GetProps(), &clientDetails.Headers)
	resp, body, err := gs.client.UploadFile(params.GetZipPath(), url, clientDetails, 0)
	if err != nil {
		return err
	}
	return httperrors.CheckResponseStatus(resp, body, 201)
}

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

type GoParams interface {
	GetZipPath() string
	GetModContent() []byte
	GetProps() string
	GetVersion() string
	GetTargetRepo() string
}

type GoParamsImpl struct {
	ZipPath    string
	ModContent []byte
	Version    string
	Props      string
	TargetRepo string
}

func (gpi *GoParamsImpl) GetZipPath() string {
	return gpi.ZipPath
}

func (gpi *GoParamsImpl) GetModContent() []byte {
	return gpi.ModContent
}

func (gpi *GoParamsImpl) GetVersion() string {
	return gpi.Version
}

func (gpi *GoParamsImpl) GetProps() string {
	return gpi.Props
}

func (gpi *GoParamsImpl) GetTargetRepo() string {
	return gpi.TargetRepo
}
