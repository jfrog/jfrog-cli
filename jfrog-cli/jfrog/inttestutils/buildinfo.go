package inttestutils

import (
	"encoding/json"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/buildinfo"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/httputils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
	"net/http"
	"testing"
)

func GetBuildInfo(artifactoryUrl, buildName, buildNumber string, t *testing.T, artHttpDetails httputils.HttpClientDetails) buildinfo.BuildInfo {
	_, body, _, err := httputils.SendGet(artifactoryUrl+"api/build/"+buildName+"/"+buildNumber, true, artHttpDetails)
	if err != nil {
		t.Error(err)
	}

	var buildInfoJson struct {
		BuildInfo buildinfo.BuildInfo `json:"buildInfo,omitempty"`
	}
	if err := json.Unmarshal(body, &buildInfoJson); err != nil {
		t.Error(err)
	}
	return buildInfoJson.BuildInfo
}

func DeleteBuild(artifactoryUrl, buildName string, artHttpDetails httputils.HttpClientDetails) {
	resp, body, err := httputils.SendDelete(artifactoryUrl+"api/build/"+buildName+"?deleteAll=1", nil, artHttpDetails)
	if err != nil {
		log.Error(err)
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		log.Error(resp.Status)
		log.Error(string(body))
	}
}
