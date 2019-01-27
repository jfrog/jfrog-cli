package golang

import (
	"github.com/jfrog/gocmd/dependencies"
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
	gci := goConfigInfo{}
	wd, err := os.Getwd()
	if err != nil {
		return gci.revert(wd, err)
	}
	if !modFileExists {
		err = gci.prepareModFile(wd, goModEditMessage)
		if err != nil {
			return err
		}
	} else {
		log.Debug("Using existing root mod file.")
		gci.modFileContent, gci.modFileStat, err = cmd.GetFileDetails("go.mod")
		if err != nil {
			return err
		}
		gci.shouldRevertSumFile, err = fileutils.IsFileExists("go.sum", false)
		if err != nil {
			return err
		}
		if gci.shouldRevertSumFile {
			gci.sumFileContent, gci.sumFileStat, err = cmd.GetFileDetails("go.sum")
			if err != nil {
				return err
			}
		}
	}

	err = executers.DownloadFromVcsWithPopulation(targetRepo, goModEditMessage, serviceManager)
	if err != nil {
		if !modFileExists {
			log.Debug("Graph failed, preparing to run go mod tidy on the root project since got the following error:", err.Error())
			err = gci.prepareAndRunTidyOnFailedGraph(wd, targetRepo, goModEditMessage, serviceManager)
			if err != nil {
				return gci.revert(wd, err)
			}
		} else {
			return gci.revert(wd, err)
		}
	}

	err = os.Chdir(wd)
	if err != nil {
		return gci.revert(wd, err)
	}
	if gci.shouldRevertModFile || gci.shouldRevertSumFile {
		return gci.revert(wd, nil)
	}
	return nil
}

type goConfigInfo struct {
	modFileContent      []byte
	modFileStat         os.FileInfo
	shouldRevertModFile bool
	shouldRevertSumFile bool
	sumFileContent      []byte
	sumFileStat         os.FileInfo
}

func (gci *goConfigInfo) revert(wd string, err error) error {
	if gci.shouldRevertModFile {
		log.Debug("Reverting to original go.mod of the root project")
		revertErr := ioutil.WriteFile("go.mod", gci.modFileContent, gci.modFileStat.Mode())
		if revertErr != nil {
			if err != nil {
				log.Error(revertErr)
				return err
			} else {
				return revertErr
			}
		}
	}

	if gci.shouldRevertSumFile {
		log.Debug("Reverting to original go.sum of the root project")
		revertErr := ioutil.WriteFile("go.sum", gci.sumFileContent, gci.sumFileStat.Mode())
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

func (gci *goConfigInfo) prepareModFile(wd, goModEditMessage string) error {
	err := cmd.RunGoModInit("", goModEditMessage)
	if err != nil {
		return err
	}
	regExp, err := dependencies.GetRegex()
	if err != nil {
		return err
	}
	notEmptyModRegex := regExp.GetNotEmptyModRegex()
	gci.modFileContent, gci.modFileStat, err = cmd.GetFileDetails("go.mod")
	if err != nil {
		return err
	}
	projectPackage := dependencies.Package{}
	projectPackage.SetModContent(gci.modFileContent)
	packageWithDep := dependencies.PackageWithDeps{Dependency: &projectPackage}
	if !packageWithDep.PatternMatched(notEmptyModRegex) {
		log.Debug("Root mod is empty, preparing to run 'go mod tidy'")
		err = cmd.RunGoModTidy()
		if err != nil {
			return gci.revert(wd, err)
		}
		gci.shouldRevertModFile = true
	} else {
		log.Debug("Root project mod not empty.")
	}

	return nil
}

func (gci *goConfigInfo) prepareAndRunTidyOnFailedGraph(wd, targetRepo, goModEditMessage string, serviceManager *artifactory.ArtifactoryServicesManager) error {
	// First revert the mod to an empty mod that includes only module name
	lines := strings.Split(string(gci.modFileContent), "\n")
	emptyMod := strings.Join(lines[:3], "\n")
	gci.modFileContent = []byte(emptyMod)
	gci.shouldRevertModFile = true
	err := gci.revert(wd, nil)
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
		return gci.revert(wd, err)
	}
	return nil
}
