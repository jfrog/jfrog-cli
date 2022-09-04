package inttestutils

import (
	"fmt"
	buildinfo "github.com/jfrog/build-info-go/entities"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"net/http"
	"path"
	"testing"

	coreutils "github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"

	"github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/http/httpclient"
	"github.com/jfrog/jfrog-client-go/utils/io/httputils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/stretchr/testify/assert"
)

func DeleteBuild(artifactoryUrl, buildName string, artHttpDetails httputils.HttpClientDetails) {
	client, err := httpclient.ClientBuilder().Build()
	if err != nil {
		log.Error(err)
		return
	}

	restApi := path.Join("api/build/", buildName)
	params := map[string]string{"deleteAll": "1"}
	requestFullUrl, err := utils.BuildArtifactoryUrl(artifactoryUrl, restApi, params)
	if err != nil {
		log.Error(err)
		return
	}

	resp, body, err := client.SendDelete(requestFullUrl, nil, artHttpDetails, "")
	if err != nil {
		log.Error(err)
		return
	}
	if err = errorutils.CheckResponseStatusWithBody(resp, body, http.StatusOK, http.StatusNotFound); err != nil {
		log.Error(err)
		return
	}

	log.Info("Build", buildName, "deleted successfully.")
}

func ValidateGeneratedBuildInfoModule(t *testing.T, buildName, buildNumber, projectKey string, moduleNames []string, moduleType buildinfo.ModuleType) {
	builds, err := coreutils.GetGeneratedBuildsInfo(buildName, buildNumber, projectKey)
	assert.NoError(t, err)
	assert.Len(t, builds, 1)
	for _, module := range builds[0].Modules {
		for _, moduleName := range moduleNames {
			if moduleName == module.Id {
				assert.Equal(t, moduleType, module.Type)
				return
			}
		}
		assert.Fail(t, fmt.Sprintf("Module '%s' with type of '%v' is expected to be one of %v", module.Id, module.Type, moduleNames))
	}
}
