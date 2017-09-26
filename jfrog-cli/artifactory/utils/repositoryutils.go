package utils

import (
	"errors"
	"github.com/buger/jsonparser"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/httputils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services/utils/auth"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils"
)

type RepoType int

const (
	LOCAL   RepoType = iota
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

func GetRepositories(artDetails *auth.ArtifactoryDetails, repoType ...RepoType) ([]string, error) {
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

func execGetRepositories(artDetails *auth.ArtifactoryDetails, repoType RepoType) ([]string, error) {
	repos := []string{}
	artDetails.Url = cliutils.AddTrailingSlashIfNeeded(artDetails.Url)
	apiUrl := artDetails.Url + "api/repositories?type=" + repoType.String()

	httpClientsDetails := artDetails.CreateArtifactoryHttpClientDetails()
	resp, body, _, err := httputils.SendGet(apiUrl, true, httpClientsDetails)
	if err != nil {
		return repos, err
	}
	if resp.StatusCode != 200 {
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
