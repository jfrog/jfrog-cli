package utils

import (
	"errors"
	"fmt"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/bintray/commands"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory"
	rthttpclient "github.com/jfrog/jfrog-client-go/artifactory/httpclient"
	"github.com/jfrog/jfrog-client-go/bintray"
	"github.com/jfrog/jfrog-client-go/bintray/auth"
	"github.com/jfrog/jfrog-client-go/bintray/services"
	"github.com/jfrog/jfrog-client-go/bintray/services/utils"
	"github.com/jfrog/jfrog-client-go/httpclient"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"net/http"
	"os"
	"path"
)

const (
	// This env var should be used for downloading the extractor jars through an Artifactory remote
	// repository, instead of downloading directly from jcenter. The remote repository should be
	// configured to proxy jcenter.
	// The env var should store a server ID configured by JFrog CLI.
	JCenterRemoteServerEnv = "JFROG_CLI_JCENTER_REMOTE_SERVER"
	// If the JCenterRemoteServerEnv env var is used, a maven remote repository named jcenter is assumed.
	// This env var can be used to use a different remote repository name.
	JCenterRemoteRepoEnv = "JFROG_CLI_JCENTER_REMOTE_REPO"
)

// Download the relevant build-info-extractor jar, if it does not already exist locally.
// By default, the jar is downloaded directly from jcenter.
// If the JCenterRemoteServerEnv environment variable is configured, the jar will be
// downloaded from a remote Artifactory repository which proxies jcenter.
//
// downloadPath: The Bintray or Artifactory download path.
// filename: The local file name.
// targetPath: The local download path (without the file name).
func DownloadExtractorIfNeeded(downloadPath, targetPath string) error {
	// If the file exists locally, we're done.
	exists, err := fileutils.IsFileExists(targetPath, false)
	if exists || err != nil {
		return err
	}

	artDetails, remotePath, err := GetJcenterRemoteDetails(downloadPath)
	if err != nil {
		return err
	}

	// Download through a remote repository in Artifactory, if configured to do so.
	if artDetails != nil {
		return downloadFileFromArtifactory(artDetails, remotePath, targetPath)
	}

	// If not configured to download through a remote repository in Artifactory,
	// download from jcenter.
	return downloadFileFromBintray(remotePath, targetPath)
}

func GetJcenterRemoteDetails(downloadPath string) (artDetails *config.ArtifactoryDetails, remotePath string, err error) {
	// Download through a remote repository in Artifactory, if configured to do so.
	serverId := os.Getenv(JCenterRemoteServerEnv)
	if serverId != "" {
		artDetails, err = config.GetArtifactoryConf(serverId)
		if err != nil {
			return
		}

		remotePath = path.Join(getJcenterRemoteRepoName(), downloadPath)
		return
	}

	// If not configured to download through a remote repository in Artifactory,
	// download from jcenter.
	remotePath = path.Join("bintray/jcenter", downloadPath)
	return
}

func getJcenterRemoteRepoName() string {
	jcenterRemoteRepo := os.Getenv(JCenterRemoteRepoEnv)
	if jcenterRemoteRepo == "" {
		jcenterRemoteRepo = "jcenter"
	}
	return jcenterRemoteRepo
}

func downloadFileFromArtifactory(artDetails *config.ArtifactoryDetails, downloadPath, targetPath string) error {
	downloadUrl := fmt.Sprintf("%s%s", artDetails.Url, downloadPath)
	log.Info("Downloading build-info-extractor from", downloadUrl)
	filename, localDir := fileutils.GetFileAndDirFromPath(targetPath)

	downloadFileDetails := &httpclient.DownloadFileDetails{
		FileName:      filename,
		DownloadPath:  downloadUrl,
		LocalPath:     localDir,
		LocalFileName: filename,
	}

	auth, err := artDetails.CreateArtAuthConfig()
	if err != nil {
		return err
	}
	securityDir, err := GetJfrogSecurityDir()
	if err != nil {
		return err
	}

	serviceConfig, err := artifactory.NewConfigBuilder().
		SetArtDetails(auth).
		SetCertificatesPath(securityDir).
		SetSkipCertsVerify(artDetails.SkipCertsVerify).
		SetDryRun(false).
		SetLogger(log.Logger).
		Build()
	if err != nil {
		return err
	}

	client, err := rthttpclient.ArtifactoryClientBuilder().
		SetCertificatesPath(serviceConfig.GetCertifactesPath()).
		SetSkipCertsVerify(artDetails.SkipCertsVerify).
		SetArtDetails(&auth).
		Build()
	if err != nil {
		return err
	}

	httpClientDetails := auth.CreateHttpClientDetails()
	resp, err := client.DownloadFile(downloadFileDetails, "", &httpClientDetails, 3, false)
	if err == nil && resp.StatusCode != http.StatusOK {
		err = errorutils.CheckError(errors.New(resp.Status + " received when attempting to download " + downloadUrl))
	}

	return err
}

func downloadFileFromBintray(downloadPath, targetPath string) error {
	bintrayConfig := auth.NewBintrayDetails()
	config := bintray.NewConfigBuilder().SetBintrayDetails(bintrayConfig).Build()

	pathDetails, err := utils.CreatePathDetails(downloadPath)
	if err != nil {
		return err
	}

	params := &services.DownloadFileParams{}
	params.PathDetails = pathDetails
	params.TargetPath = targetPath
	params.Flat = true

	_, _, err = commands.DownloadFile(config, params)
	return err
}
