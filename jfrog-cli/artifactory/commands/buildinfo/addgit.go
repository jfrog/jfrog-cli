package buildinfo

import (
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils/git"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/buildinfo"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
	"os"
)

func AddGit(buildName, buildNumber, dotGitPath string) error {
	log.Info("Collecting git revision and remote url...")
	err := utils.SaveBuildGeneralDetails(buildName, buildNumber)
	if err != nil {
		return err
	}
	if dotGitPath == "" {
		dotGitPath, err = os.Getwd()
		if err != nil {
			return err
		}
	}
	gitManager := git.NewManager(dotGitPath)
	err = gitManager.ReadConfig()
	if err != nil {
		return err
	}

	populateFunc := func(partial *buildinfo.Partial) {
		partial.Vcs = &buildinfo.Vcs{
			Url:      gitManager.GetUrl() + ".git",
			Revision: gitManager.GetRevision(),
		}
	}

	err = utils.SavePartialBuildInfo(buildName, buildNumber, populateFunc)
	if err != nil {
		return err
	}
	log.Info("Collected git revision and remote url for", buildName+"/"+buildNumber+".")
	return nil
}
