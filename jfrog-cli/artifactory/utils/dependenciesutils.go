package utils

import (
	"errors"
	"fmt"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/bintray/commands"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory"
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
)

// Download the relevant build-info-extractor jar, if it does not already exist locally.
// By default, the jar is downloaded directly from jcenter.
// If the JFROG_CLI_JCENTER_REMOTE_SERVER environment variable is configured, the jar will be
// downloaded from a remote Artifactory repository which proxies jcenter.
//
// downloadPath: The Bintray or Artifactory download path.
// filename: The local file name.
// targetPath: The local download path (without the file name).
func DownloadExreactorIfNeeded(downloadPath, targetPath string) error {
	// If the file exists locally, we're done.
	exists, err := fileutils.IsFileExists(targetPath, false)
	if exists || err != nil {
		return err
	}

	// Download through a remote repository in Artifactory, if configured to do so.
	jcenterRemoteServerId := os.Getenv("JFROG_CLI_JCENTER_REMOTE_SERVER")
	if jcenterRemoteServerId != "" {
		artDetails, err := config.GetArtifactoryConf(jcenterRemoteServerId)
		if err != nil {
			return err
		}

		downloadPath := fmt.Sprintf("%s/%s", getJcenterRemoteRepoName(), downloadPath)
		return downloadFileFromArtifactory(artDetails, downloadPath, targetPath)
	}

	// If not configured to download through a remote repository in Artifactory,
	// download from jcenter.
	return downloadFileFromBintray("bintray/jcenter/"+downloadPath, targetPath)
}

func getJcenterRemoteRepoName() string {
	jcenterRemoteRepo := os.Getenv("JFROG_CLI_JCENTER_REMOTE_REPO")
	if jcenterRemoteRepo == "" {
		jcenterRemoteRepo = "jcenter"
	}
	return jcenterRemoteRepo
}

func downloadFileFromArtifactory(artDetails *config.ArtifactoryDetails, downloadPath, targetPath string) error {
	downloadUrl := fmt.Sprintf("%s%s", artDetails.Url, downloadPath)
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
		SetDryRun(false).
		SetLogger(log.Logger).
		Build()
	if err != nil {
		return err
	}

	client, err := artifactory.CreateArtifactoryHttpClient(serviceConfig)
	if err != nil {
		return err
	}

	resp, err := client.DownloadFile(downloadFileDetails, "", auth.CreateHttpClientDetails(), 3, false)
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
