package commands

import (
	"os"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils"
)

func BuildAddGit(buildName, buildNumber, dotGitPath string) (err error) {
	if err = utils.SaveBuildGeneralDetails(buildName, buildNumber); err != nil {
		return
	}
	if dotGitPath == "" {
		dotGitPath, err = os.Getwd()
		if err != nil {
			return
		}
	}
	gitManager := utils.NewGitManager(dotGitPath)
	err = gitManager.ReadGitConfig()
	if err != nil {
		return
	}

	populateFunc := func(tempWrapper *utils.ArtifactBuildInfoWrapper) {
		tempWrapper.Vcs = &utils.Vcs{
			VcsUrl: gitManager.GetUrl() + ".git",
			VcsRevision: gitManager.GetRevision(),
		}
	}
	err = utils.SavePartialBuildInfo(buildName, buildNumber, populateFunc)
	return
}
