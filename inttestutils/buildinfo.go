package inttestutils

import (
	"github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/httpclient"
	"github.com/jfrog/jfrog-client-go/utils/io/httputils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"net/http"
	"path"
)

func DeleteBuild(artifactoryUrl, buildName string, artHttpDetails httputils.HttpClientDetails) {
	client, err := httpclient.ClientBuilder().Build()
	if err != nil {
		log.Error(err)
	}

	restApi := path.Join("api/build/", buildName)
	params := map[string]string{"deleteAll": "1"}
	requestFullUrl, err := utils.BuildArtifactoryUrl(artifactoryUrl, restApi, params)

	resp, body, err := client.SendDelete(requestFullUrl, nil, artHttpDetails)
	if err != nil {
		log.Error(err)
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		log.Error(resp.Status)
		log.Error(string(body))
	}
}
