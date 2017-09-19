package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/log"
)

func SetProps(spec *utils.SpecFiles, flags utils.CommonFlags, props string) error {
	err := utils.PreCommandSetup(flags)
	if err != nil {
		return err
	}
	resultItems, err := utils.SearchBySpecFiles(spec, flags)
	if err != nil {
		return err
	}
	log.Info("Setting properties...")
	for _, item := range resultItems {
		log.Info("Setting properties on", item.GetFullUrl())
		utils.SetProps(item.GetFullUrl(), props, flags.GetArtifactoryDetails())
	}

	log.Info("Done setting properties.")
	return err
}