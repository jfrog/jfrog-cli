package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"os"
	"strings"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils/buildinfo"
)

func BuildCollectEnv(buildName, buildNumber string) (err error) {
	log.Info("Collecting environment variables...")
	if err = utils.SaveBuildGeneralDetails(buildName, buildNumber); err != nil {
		return
	}
	populateFunc := func(partial *buildinfo.Partial) {
		partial.Env = getEnvVariables()
	}
	err = utils.SavePartialBuildInfo(buildName, buildNumber, populateFunc)
	if err != nil {
		return
	}
	log.Info("Collected environment variables for", buildName+"/"+buildNumber+".")
	return
}

func getEnvVariables() buildinfo.Env {
	m := make(map[string]string)
	for _, e := range os.Environ() {
		pair := strings.Split(e, "=")
		if len(pair[0]) != 0 {
			m["buildInfo.env."+pair[0]] = pair[1]
		}
	}
	return m
}