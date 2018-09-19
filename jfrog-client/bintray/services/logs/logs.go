package logs

import (
	"errors"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/bintray/auth"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/bintray/services/versions"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/httpclient"
	clientutils "github.com/jfrog/jfrog-cli-go/jfrog-client/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/errorutils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/log"
	"net/http"
	"path"
)

func NewService(client *httpclient.HttpClient) *LogsService {
	us := &LogsService{client: client}
	return us
}

type LogsService struct {
	client         *httpclient.HttpClient
	BintrayDetails auth.BintrayDetails
}

func (ls *LogsService) List(versionPath *versions.Path) error {
	if ls.BintrayDetails.GetUser() == "" {
		ls.BintrayDetails.SetUser(versionPath.Subject)
	}
	listUrl := ls.BintrayDetails.GetApiUrl() + path.Join("packages", versionPath.Subject, versionPath.Repo, versionPath.Package, "logs")
	httpClientsDetails := ls.BintrayDetails.CreateHttpClientDetails()
	log.Info("Getting logs...")
	client := httpclient.NewDefaultHttpClient()
	resp, body, _, _ := client.SendGet(listUrl, true, httpClientsDetails)

	if resp.StatusCode != http.StatusOK {
		return errorutils.CheckError(errors.New("Bintray response: " + resp.Status + "\n" + clientutils.IndentJson(body)))
	}

	log.Debug("Bintray response:", resp.Status)
	log.Output(clientutils.IndentJson(body))
	return nil
}

func (ls *LogsService) Download(versionPath *versions.Path, logName string) error {
	if ls.BintrayDetails.GetUser() == "" {
		ls.BintrayDetails.SetUser(versionPath.Subject)
	}
	downloadUrl := ls.BintrayDetails.GetApiUrl() + path.Join("packages", versionPath.Subject, versionPath.Repo, versionPath.Package, "logs")

	httpClientsDetails := ls.BintrayDetails.CreateHttpClientDetails()
	log.Info("Downloading logs...")
	client := httpclient.NewDefaultHttpClient()
	details := &httpclient.DownloadFileDetails{
		FileName:      logName,
		DownloadPath:  downloadUrl,
		LocalPath:     "",
		LocalFileName: logName}
	resp, err := client.DownloadFile(details, "", httpClientsDetails, 3, false)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errorutils.CheckError(errors.New("Bintray response: " + resp.Status))
	}
	log.Debug("Bintray response:", resp.Status)
	log.Info("Downloaded log.")
	return nil
}
