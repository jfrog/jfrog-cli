package generic

import (
	"fmt"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/utils/cliutils"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	clientutils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

type GitLfsCommand struct {
	GenericCommand
	configuration *GitLfsCleanConfiguration
}

func (glc *GitLfsCommand) Configuration() *GitLfsCleanConfiguration {
	return glc.configuration
}

func (glc *GitLfsCommand) SetConfiguration(configuration *GitLfsCleanConfiguration) *GitLfsCommand {
	glc.configuration = configuration
	return glc
}

func NewGitLfsCommand() *GitLfsCommand {
	return &GitLfsCommand{GenericCommand: *NewGenericCommand()}
}

func (glc *GitLfsCommand) Run() error {
	rtDetails, err := glc.RtDetails()
	if errorutils.CheckError(err) != nil {
		return err
	}
	servicesManager, err := utils.CreateServiceManager(rtDetails, glc.DryRun())
	if err != nil {
		return err
	}

	gitLfsCleanParams := getGitLfsCleanParams(glc.configuration)

	filesToDelete, err := servicesManager.GetUnreferencedGitLfsFiles(gitLfsCleanParams)

	if err != nil || len(filesToDelete) < 1 {
		return err
	}

	if glc.configuration.Quiet {
		err = glc.deleteLfsFilesFromArtifactory(filesToDelete)
		return err
	}
	return glc.interactiveDeleteLfsFiles(filesToDelete)
}

func (glc *GitLfsCommand) CommandName() string {
	return "git_lfs_clean"
}

func (glc *GitLfsCommand) deleteLfsFilesFromArtifactory(deleteItems []clientutils.ResultItem) error {
	log.Info("Deleting", len(deleteItems), "files from", glc.configuration.Repo, "...")
	servicesManager, err := utils.CreateServiceManager(glc.rtDetails, glc.DryRun())
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
	Quiet   bool
	Refs    string
	Repo    string
	GitPath string
}

func getGitLfsCleanParams(configuration *GitLfsCleanConfiguration) (gitLfsCleanParams services.GitLfsCleanParams) {
	gitLfsCleanParams = services.NewGitLfsCleanParams()
	gitLfsCleanParams.GitPath = configuration.GitPath
	gitLfsCleanParams.Refs = configuration.Refs
	gitLfsCleanParams.Repo = configuration.Repo
	return
}

func (glc *GitLfsCommand) interactiveDeleteLfsFiles(filesToDelete []clientutils.ResultItem) error {
	for _, v := range filesToDelete {
		fmt.Println("  " + v.Name)
	}
	confirmed := cliutils.InteractiveConfirm("Are you sure you want to delete the above files?")
	if confirmed {
		err := glc.deleteLfsFilesFromArtifactory(filesToDelete)
		return err
	}
	return nil
}
