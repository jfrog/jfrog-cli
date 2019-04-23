package generic

import (
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	clientutils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

func PrepareGitLfsClean(configuration *GitLfsCleanConfiguration) ([]clientutils.ResultItem, error) {
	servicesManager, err := utils.CreateServiceManager(configuration.ArtDetails, configuration.DryRun)
	if err != nil {
		return nil, err
	}

	gitLfsCleanParams := getGitLfsCleanParams(configuration)

	return servicesManager.GetUnreferencedGitLfsFiles(gitLfsCleanParams)
}

func DeleteLfsFilesFromArtifactory(deleteItems []clientutils.ResultItem, configuration *GitLfsCleanConfiguration) error {
	log.Info("Deleting", len(deleteItems), "files from", configuration.Repo, "...")
	servicesManager, err := utils.CreateServiceManager(configuration.ArtDetails, configuration.DryRun)
	if err != nil {
		return err
	}
	_, err = servicesManager.DeleteFiles(deleteItems)
	if err != nil {
		return errorutils.CheckError(err)
	}
	return nil
}

type GitLfsCleanConfiguration struct {
	ArtDetails *config.ArtifactoryDetails
	Quiet      bool
	DryRun     bool
	Refs       string
	Repo       string
	GitPath    string
}

func getGitLfsCleanParams(configuration *GitLfsCleanConfiguration) (gitLfsCleanParams services.GitLfsCleanParams) {
	gitLfsCleanParams = services.NewGitLfsCleanParams()
	gitLfsCleanParams.GitPath = configuration.GitPath
	gitLfsCleanParams.Refs = configuration.Refs
	gitLfsCleanParams.Repo = configuration.Repo
	return
}
