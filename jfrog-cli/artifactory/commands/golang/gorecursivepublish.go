package golang

import (
	"github.com/jfrog/gocmd/executers"
	"github.com/jfrog/gocmd/utils/cmd"
	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"io/ioutil"
	"os"
	"strings"
)

func Execute(targetRepo, goModEditMessage string, serviceManager *artifactory.ArtifactoryServicesManager) error {
	modFileExists, err := fileutils.IsFileExists("go.mod", false)
	if err != nil {
		return err
	}
	gmi := goModInfo{}
	wd, err := os.Getwd()
	if err != nil {
		return gmi.revert(wd, err)
	}
	if !modFileExists {
		err = gmi.prepareModFile(wd, goModEditMessage)
		if err != nil {
			return err
		}
	} else {
		log.Debug("Using existing root mod file.")
		gmi.modFileContent, gmi.modFileStat, err = cmd.GetFileDetails("go.mod")
		if err != nil {
			return err
		}
	}

	err = executers.DownloadFromVcsWithPopulation(targetRepo, goModEditMessage, serviceManager)
	if err != nil {
		if !modFileExists {
			log.Debug("Graph failed, preparing to run go mod tidy on the root project since got the following error:", err.Error())
			err = gmi.prepareAndRunTidyOnFailedGraph(wd, targetRepo, goModEditMessage, serviceManager)
			if err != nil {
				return gmi.revert(wd, err)
			}
		} else {
			return gmi.revert(wd, err)
		}
	}

	err = os.Chdir(wd)
	if err != nil {
		return gmi.revert(wd, err)
	}
	return gmi.revert(wd, nil)
}

type goModInfo struct {
	modFileContent      []byte
	modFileStat         os.FileInfo
	shouldRevertModFile bool
}

func (gmi *goModInfo) revert(wd string, err error) error {
	if gmi.shouldRevertModFile {
		log.Debug("Reverting to original go.mod of the root project")
		revertErr := ioutil.WriteFile("go.mod", gmi.modFileContent, gmi.modFileStat.Mode())
		if revertErr != nil {
			if err != nil {
				log.Error(revertErr)
				return err
			} else {
				return revertErr
			}
		}
	}
	return nil
}

func (gmi *goModInfo) prepareModFile(wd, goModEditMessage string) error {
	err := cmd.RunGoModInit("")
	if err != nil {
		return err
	}
	regExp, err := executers.GetRegex()
	if err != nil {
		return err
	}
	notEmptyModRegex := regExp.GetNotEmptyModRegex()
	gmi.modFileContent, gmi.modFileStat, err = cmd.GetFileDetails("go.mod")
	if err != nil {
		return err
	}
	projectPackage := executers.Package{}
	projectPackage.SetModContent(gmi.modFileContent)
	packageWithDep := executers.PackageWithDeps{Dependency: &projectPackage}
	if !packageWithDep.PatternMatched(notEmptyModRegex) {
		log.Debug("Root mod is empty, preparing to run 'go mod tidy'")
		err = cmd.RunGoModTidy()
		if err != nil {
			return gmi.revert(wd, err)
		}
		gmi.shouldRevertModFile = true
	} else {
		log.Debug("Root project mod not empty.")
	}

	return nil
}

func (gmi *goModInfo) prepareAndRunTidyOnFailedGraph(wd, targetRepo, goModEditMessage string, serviceManager *artifactory.ArtifactoryServicesManager) error {
	// First revert the mod to an empty mod that includes only module name
	lines := strings.Split(string(gmi.modFileContent), "\n")
	emptyMod := strings.Join(lines[:3], "\n")
	gmi.modFileContent = []byte(emptyMod)
	gmi.shouldRevertModFile = true
	err := gmi.revert(wd, nil)
	if err != nil {
		log.Error(err)
	}
	// Run go mod tidy.
	err = cmd.RunGoModTidy()
	if err != nil {
		return err
	}
	// Perform collection again after tidy finished successfully.
	err = executers.DownloadFromVcsWithPopulation(targetRepo, goModEditMessage, serviceManager)
	if err != nil {
		return gmi.revert(wd, err)
	}
	return nil
}
