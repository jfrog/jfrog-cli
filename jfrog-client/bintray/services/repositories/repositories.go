package repositories

import (
	"errors"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/bintray/auth"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/httpclient"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/errorutils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/io/httputils"
	"net/http"
	"path"
)

func NewService(client *httpclient.HttpClient) *RepositoryService {
	us := &RepositoryService{client: client}
	return us
}

type RepositoryService struct {
	client         *httpclient.HttpClient
	BintrayDetails auth.BintrayDetails
}

type Path struct {
	Subject string
	Repo    string
}

func (rs *RepositoryService) IsRepoExists(repositoryPath *Path) (bool, error) {
	url := rs.BintrayDetails.GetApiUrl() + path.Join("repos", repositoryPath.Subject, repositoryPath.Repo)
	httpClientsDetails := rs.BintrayDetails.CreateHttpClientDetails()

	resp, _, err := httputils.SendHead(url, httpClientsDetails)
	if err != nil {
		return false, err
	}
	if resp.StatusCode == http.StatusOK {
		return true, nil
	}
	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	return false, errorutils.CheckError(errors.New("Bintray response: " + resp.Status))
}
