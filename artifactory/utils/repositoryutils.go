package utils

import (
	"errors"
	"github.com/buger/jsonparser"
	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/auth"
	"github.com/jfrog/jfrog-client-go/httpclient"
	"github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"net/http"
)

type RepoType int

const (
	LOCAL RepoType = iota
	REMOTE
	VIRTUAL
)

var RepoTypes = []string{
	"local",
	"remote",
	"virtual",
}

func (repoType RepoType) String() string {
	return RepoTypes[repoType]
}

func GetRepositories(artDetails auth.ServiceDetails, repoType ...RepoType) ([]string, error) {
	repos := []string{}
	for _, v := range repoType {
		r, err := execGetRepositories(artDetails, v)
		if err != nil {
			return repos, err
		}
		if len(r) > 0 {
			repos = append(repos, r...)
		}
	}

	return repos, nil
}

func execGetRepositories(artDetails auth.ServiceDetails, repoType RepoType) ([]string, error) {
	repos := []string{}
	artDetails.SetUrl(utils.AddTrailingSlashIfNeeded(artDetails.GetUrl()))
	apiUrl := artDetails.GetUrl() + "api/repositories?type=" + repoType.String()

	httpClientsDetails := artDetails.CreateHttpClientDetails()
	client, err := httpclient.ClientBuilder().Build()
	if err != nil {
		return repos, err
	}
	resp, body, _, err := client.SendGet(apiUrl, true, httpClientsDetails)
	if err != nil {
		return repos, err
	}
	if resp.StatusCode != http.StatusOK {
		return repos, errorutils.CheckError(errors.New("Artifactory response: " + resp.Status + "\n" + utils.IndentJson(body)))
	}

	jsonparser.ArrayEach(body, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		val, _, _, err := jsonparser.Get(value, "key")
		if err == nil {
			repos = append(repos, string(val))
		}
	})
	return repos, nil
}

// Since we can't search dependencies in a remote repository, we will turn the search to the repository's cache.
// Local/Virtual repository name will be returned as is.
func GetRepoNameForDependenciesSearch(repoName string, serviceManager *artifactory.ArtifactoryServicesManager) (string, error) {
	repoDetails, err := serviceManager.GetRepository(repoName)
	if err != nil {
		return "", err
	}
	if repoDetails.Rclass == "remote" {
		repoName += "-cache"
	}
	return repoName, err
}
