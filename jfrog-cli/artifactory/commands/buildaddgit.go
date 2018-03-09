package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils/buildinfo"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
	"os"
	"strings"
)

func BuildAddGit(buildName, buildNumber, dotGitPath string) error {
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
	gitManager := utils.NewGitManager(dotGitPath)
	err = gitManager.ReadGitConfig()
	if err != nil {
		return err
	}

	gitUrl := gitManager.GetUrl()
	if !strings.HasSuffix(gitUrl,".git") {
		gitUrl += ".git"
	}

	populateFunc := func(partial *buildinfo.Partial) {
		partial.Vcs = &buildinfo.Vcs{
			Url:      gitManager.GetUrl(),
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
