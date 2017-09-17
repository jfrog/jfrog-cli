package utils

import (
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"
	"net/http"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services/utils/auth"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services/utils/auth/cert"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/types/httpclient"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services"
	clientutils "github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/cliutils"
)

func GetJfrogSecurityDir() (string, error) {
	confPath, err := config.GetJfrogHomeDir()
	if err != nil {
		return "", err
	}
	return confPath + "security/", nil
}

func GetEncryptedPasswordFromArtifactory(artifactoryAuth *auth.ArtifactoryDetails) (*http.Response, string, error) {
	apiUrl := artifactoryAuth.Url + "api/security/encryptedPassword"
	httpClientsDetails := artifactoryAuth.CreateArtifactoryHttpClientDetails()
	securityDir, err := GetJfrogSecurityDir()
	if err != nil {
		return nil, "", err
	}
	transport, err := cert.GetTransportWithLoadedCert(securityDir)
	client := httpclient.NewHttpClient(&http.Client{Transport: transport})
	resp, body, _, err := client.SendGet(apiUrl, true, httpClientsDetails)
	return resp, string(body), err
}

func CreateServiceManager(artDetails *config.ArtifactoryDetails, isDryRun bool) (*artifactory.ArtifactoryServicesManager, error) {
	certPath, err := GetJfrogSecurityDir()
	if err != nil {
		return nil, err
	}
	authConfig := artDetails.CreateArtAuthConfig()
	serviceConfig, err := (&artifactory.ArtifactoryServicesConfigBuilder{}).
		SetArtDetails(authConfig).
		SetCertificatesPath(certPath).
		SetDryRun(isDryRun).
		SetLogger(cliutils.CliLogger).
		Build()
	if err != nil {
		return nil, err
	}
	return artifactory.NewArtifactoryService(serviceConfig)
}

func ConvertResultItemArrayToDeleteItemArray(resultItems []clientutils.ResultItem) ([]services.DeleteItem) {
	var deleteItems []services.DeleteItem = make([]services.DeleteItem, len(resultItems))
	for i, item := range resultItems {
		deleteItems[i] = item
	}
	return deleteItems
}