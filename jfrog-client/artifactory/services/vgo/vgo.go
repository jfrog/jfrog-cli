package vgo

import (
	"encoding/base64"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/auth"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/errors/httperrors"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/httpclient"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
	"os"
)

type VgoService struct {
	client     *httpclient.HttpClient
	ArtDetails auth.ArtifactoryDetails
}

func NewVgoService(client *httpclient.HttpClient) *VgoService {
	return &VgoService{client: client}
}

func (vs *VgoService) GetJfrogHttpClient() *httpclient.HttpClient {
	return vs.client
}

func (vs *VgoService) SetArtDetails(artDetails auth.ArtifactoryDetails) {
	vs.ArtDetails = artDetails
}

func (vs *VgoService) PublishPackage(params VgoParams) error {
	f, err := os.Open(params.GetZipPath())
	if err != nil {
		return errorutils.CheckError(err)
	}
	defer f.Close()

	url, err := utils.BuildArtifactoryUrl(vs.ArtDetails.GetUrl(), "api/go/"+params.GetTargetRepo(), make(map[string]string))
	clientDetails := vs.ArtDetails.CreateHttpClientDetails()

	utils.AddHeader("X-GO-MODULE-VERSION", params.GetVersion(), &clientDetails.Headers)
	utils.AddHeader("X-GO-MODULE-CONTENT", base64.StdEncoding.EncodeToString(params.GetModContent()), &clientDetails.Headers)
	addPropertiesHeaders(params.GetProps(), &clientDetails.Headers)

	resp, body, err := vs.client.UploadFile(f, url, clientDetails)
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

type VgoParams interface {
	GetZipPath() string
	GetModContent() []byte
	GetProps() string
	GetVersion() string
	GetTargetRepo() string
}

type VgoParamsImpl struct {
	ZipPath    string
	ModContent []byte
	Version    string
	Props      string
	TargetRepo string
}

func (vp *VgoParamsImpl) GetZipPath() string {
	return vp.ZipPath
}

func (vp *VgoParamsImpl) GetModContent() []byte {
	return vp.ModContent
}

func (vp *VgoParamsImpl) GetVersion() string {
	return vp.Version
}

func (vp *VgoParamsImpl) GetProps() string {
	return vp.Props
}

func (vp *VgoParamsImpl) GetTargetRepo() string {
	return vp.TargetRepo
}
