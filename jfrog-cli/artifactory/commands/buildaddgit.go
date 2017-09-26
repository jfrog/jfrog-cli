package commands

import (
	"os"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils/buildinfo"
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

	populateFunc := func(partial *buildinfo.Partial) {
		partial.Vcs = &buildinfo.Vcs{
			Url:      gitManager.GetUrl() + ".git",
			Revision: gitManager.GetRevision(),
		}
	}
	err = utils.SavePartialBuildInfo(buildName, buildNumber, populateFunc)
	return
}
