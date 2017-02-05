package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/artifactory/utils"
	"os"
	"strings"
)

func BuildCollectEnv(buildName, buildNumber string) (err error) {
	if err = utils.SaveBuildGeneralDetails(buildName, buildNumber); err != nil {
		return
	}
	populateFunc := func(tempWrapper *utils.ArtifactBuildInfoWrapper) {
		tempWrapper.Env = getEnvVariables()
	}
	err = utils.SavePartialBuildInfo(buildName, buildNumber, populateFunc)
	return
}

func getEnvVariables() utils.BuildEnv {
	m := make(map[string]string)
	for _, e := range os.Environ() {
		pair := strings.Split(e, "=")
		if len(pair[0]) != 0 {
			m["buildInfo.env." + pair[0]] = pair[1]
		}
	}
	return m
}