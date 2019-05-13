package buildinfo

import (
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

type BuildCleanCommand struct {
	buildConfiguration *utils.BuildConfiguration
}

func (bcc *BuildCleanCommand) SetBuildConfiguration(buildConfiguration *utils.BuildConfiguration) *BuildCleanCommand {
	bcc.buildConfiguration = buildConfiguration
	return bcc
}

func (bcc *BuildCleanCommand) CommandName() string {
	return "rt_build_clean"
}

// Returns the default Artifactory server
func (bcc *BuildCleanCommand) RtDetails() *config.ArtifactoryDetails {
	details, err := config.GetDefaultArtifactoryConf()
	if err != nil {
		log.Debug(err)
		return nil
	}
	return details
}

func (bcc *BuildCleanCommand) Run() error {
	log.Info("Cleaning build info...")
	err := utils.RemoveBuildDir(bcc.buildConfiguration.BuildName, bcc.buildConfiguration.BuildNumber)
	if err != nil {
		return err
	}
	log.Info("Cleaned build info", bcc.buildConfiguration.BuildName+"/"+bcc.buildConfiguration.BuildNumber+".")
	return nil
}
