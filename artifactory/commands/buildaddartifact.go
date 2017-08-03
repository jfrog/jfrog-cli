package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/log"
)

func BuildAddArtifact(buildName, buildNumber, artifact string, flags *BuildAddArtifactFlags) (err error) {
	err = nil
	log.Info("Adding artifact '", artifact, "' to ", buildName, "#" + buildNumber)
	return
}

type BuildAddArtifactFlags struct {
	ArtDetails  *config.ArtifactoryDetails
}
