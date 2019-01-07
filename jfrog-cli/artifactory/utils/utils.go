package utils

import (
	"errors"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/artifactory/auth"
	"github.com/jfrog/jfrog-client-go/artifactory/auth/cert"
	"github.com/jfrog/jfrog-client-go/httpclient"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
)

const repoDetailsUrl = "api/repositories/"

func GetJfrogSecurityDir() (string, error) {
	homeDir, err := config.GetJfrogHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, "security"), nil
}

func GetEncryptedPasswordFromArtifactory(artifactoryAuth auth.ArtifactoryDetails) (string, error) {
	u, err := url.Parse(artifactoryAuth.GetUrl())
	if err != nil {
		return "", err
	}
	u.Path = path.Join(u.Path, "api/security/encryptedPassword")
	httpClientsDetails := artifactoryAuth.CreateHttpClientDetails()
	securityDir, err := GetJfrogSecurityDir()
	if err != nil {
		return "", err
	}
	transport, err := cert.GetTransportWithLoadedCert(securityDir)
	client := httpclient.NewHttpClient(&http.Client{Transport: transport})
	resp, body, _, err := client.SendGet(u.String(), true, httpClientsDetails)
	if err != nil {
		return "", err
	}

	if resp.StatusCode == http.StatusOK {
		return string(body), nil
	}

	if resp.StatusCode == http.StatusConflict {
		message := "\nYour Artifactory server is not configured to encrypt passwords.\n" +
			"You may use \"art config --enc-password=false\""
		return "", errorutils.CheckError(errors.New(message))
	}

	return "", errorutils.CheckError(errors.New("Artifactory response: " + resp.Status))
}

func CreateServiceManager(artDetails *config.ArtifactoryDetails, isDryRun bool) (*artifactory.ArtifactoryServicesManager, error) {
	certPath, err := GetJfrogSecurityDir()
	if err != nil {
		return nil, err
	}
	artAuth, err := artDetails.CreateArtAuthConfig()
	if err != nil {
		return nil, err
	}
	serviceConfig, err := artifactory.NewConfigBuilder().
		SetArtDetails(artAuth).
		SetCertificatesPath(certPath).
		SetDryRun(isDryRun).
		SetLogger(log.Logger).
		Build()
	if err != nil {
		return nil, err
	}
	return artifactory.New(serviceConfig)
}

func isRepoExists(repository string, artDetails auth.ArtifactoryDetails) (bool, error) {
	artHttpDetails := artDetails.CreateHttpClientDetails()
	client := httpclient.NewDefaultHttpClient()
	resp, _, _, err := client.SendGet(artDetails.GetUrl()+repoDetailsUrl+repository, true, artHttpDetails)
	if err != nil {
		return false, errorutils.CheckError(err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		return true, nil
	}
	return false, nil
}

func CheckIfRepoExists(repository string, artDetails auth.ArtifactoryDetails) error {
	repoExists, err := isRepoExists(repository, artDetails)
	if err != nil {
		return err
	}

	if !repoExists {
		return errorutils.CheckError(errors.New("The repository '" + repository + "' does not exist."))
	}
	return nil
}
