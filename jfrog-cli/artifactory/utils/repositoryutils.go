package utils

import (
	"errors"
	"github.com/buger/jsonparser"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/artifactory/auth"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/errorutils"
	"net/http"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/httpclient"
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

func GetRepositories(artDetails auth.ArtifactoryDetails, repoType ...RepoType) ([]string, error) {
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

func execGetRepositories(artDetails auth.ArtifactoryDetails, repoType RepoType) ([]string, error) {
	repos := []string{}
	artDetails.SetUrl(utils.AddTrailingSlashIfNeeded(artDetails.GetUrl()))
	apiUrl := artDetails.GetUrl() + "api/repositories?type=" + repoType.String()

	httpClientsDetails := artDetails.CreateHttpClientDetails()
	client := httpclient.NewDefaultHttpClient()
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
