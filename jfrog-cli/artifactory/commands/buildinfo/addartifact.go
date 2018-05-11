package buildinfo

import (
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/buildinfo"
)

func AddArtifact(flags *BuildAddArtifactConfiguration) (err error) {
	log.Info("Adding artifact '" + flags.Artifact + "' to build info " + flags.BuildName + " #" + flags.BuildNumber + "...")

	err = utils.SaveBuildGeneralDetails(flags.BuildName, flags.BuildNumber); if err != nil {
		return
	}

	serviceManager, err := utils.CreateServiceManager(flags.ArtDetails, false); if err != nil {
		return err
	}

	fileInfo, err := serviceManager.GetFileInfo(flags.Artifact); if err != nil {
		return err
	}

	buildArtifacts := []buildinfo.InternalArtifact{fileInfo.ToBuildArtifact()}
	populateFunc := func(partial *buildinfo.Partial) {
		partial.Artifacts = buildArtifacts
	}
	err = utils.SavePartialBuildInfo(flags.BuildName, flags.BuildNumber, populateFunc); if err != nil {
		return
	}

	log.Info("Successfully added artifact to build info")
	return
}

type BuildAddArtifactConfiguration struct {
	ArtDetails  *config.ArtifactoryDetails
	Artifact    string
	BuildName   string
	BuildNumber string
}
