package buildinfo

import (
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"os"
	"strings"
)

type BuildCollectEnvCommand struct {
	buildConfiguration *utils.BuildConfiguration
}

func NewBuildCollectEnvCommand() *BuildCollectEnvCommand {
	return &BuildCollectEnvCommand{}
}

func (bcec *BuildCollectEnvCommand) SetBuildConfiguration(buildConfiguration *utils.BuildConfiguration) *BuildCollectEnvCommand {
	bcec.buildConfiguration = buildConfiguration
	return bcec
}

func (bcec *BuildCollectEnvCommand) Run() error {
	log.Info("Collecting environment variables...")
	err := utils.SaveBuildGeneralDetails(bcec.buildConfiguration.BuildName, bcec.buildConfiguration.BuildNumber)
	if err != nil {
		return err
	}
	populateFunc := func(partial *buildinfo.Partial) {
		partial.Env = getEnvVariables()
	}
	err = utils.SavePartialBuildInfo(bcec.buildConfiguration.BuildName, bcec.buildConfiguration.BuildNumber, populateFunc)
	if err != nil {
		return err
	}
	log.Info("Collected environment variables for", bcec.buildConfiguration.BuildName+"/"+bcec.buildConfiguration.BuildNumber+".")
	return nil
}

// Returns the default configured Artifactory server
func (bcec *BuildCollectEnvCommand) RtDetails() (*config.ArtifactoryDetails, error) {
	return config.GetDefaultArtifactoryConf()
}

func (bcec *BuildCollectEnvCommand) CommandName() string {
	return "rt_build_collect_env"
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
