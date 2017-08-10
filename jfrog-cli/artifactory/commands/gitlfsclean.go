package commands

import (
	clientutils "github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/cliutils"
)

func PrepareGitLfsClean(flags *GitLfsCleanConfiguration) ([]clientutils.ResultItem, error) {
	servicesManager, err := utils.CreateDefaultServiceManager(flags.ArtDetails, flags.DryRun)
	if err != nil {
		return nil, err
	}
	return servicesManager.GetUnreferencedGitLfsFiles(flags)
}

func DeleteLfsFilesFromArtifactory(files []clientutils.ResultItem, flags *GitLfsCleanConfiguration) error {
	cliutils.CliLogger.Info("Deleting", len(files), "files from", flags.Repo, "...")
	servicesManager, err := utils.CreateDefaultServiceManager(flags.ArtDetails, flags.DryRun)
	if err != nil {
		return err
	}
	deleteItems := utils.ConvertResultItemArrayToDeleteItemArray(files)
	err = servicesManager.DeleteFiles(deleteItems)
	if err != nil {
		return errorutils.CheckError(err)
	}
	return nil
}

type GitLfsCleanConfiguration struct {
	*services.GitLfsCleanParamsImpl
	ArtDetails *config.ArtifactoryDetails
	Quiet bool
	DryRun bool
}
