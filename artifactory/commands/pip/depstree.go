package pip

import (
	"encoding/json"
	"fmt"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils/dependenciestree"
	piputils "github.com/jfrog/jfrog-cli-go/artifactory/utils/pip"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils/pip/dependencies"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/pkg/errors"
	"strings"
)

type PipDepTreeCommand struct {
	*PipCommand
	DepsTreeRoot dependenciestree.Root
}

func NewPipDepTreeCommand() *PipDepTreeCommand {
	return &PipDepTreeCommand{PipCommand: &PipCommand{}}
}

func (pdtc *PipDepTreeCommand) Run() error {
	pythonExecutablePath, err := piputils.GetExecutablePath("python")
	if err != nil {
		return err
	}

	pipInstaller := &piputils.PipInstaller{Args: pdtc.args, RtDetails: pdtc.rtDetails, Repository: pdtc.repository, ShouldParseLogs: true}
	err = pipInstaller.Install()
	if err != nil {
		return err
	}

	// Extract dependencies from setup.py or requirements.txt.
	extractor, err := dependencies.CreateCompatibleExtractor(pythonExecutablePath, pdtc.args)
	if err != nil {
		return err
	}
	err = extractor.Extract()
	if err != nil {
		return err
	}

	allDependencies := extractor.AllDependencies()
	pdtc.removeUninstalledDependencies(allDependencies, pipInstaller.DependencyToFileMap)
	missingDeps, err := pdtc.pupulateDepsInfo(allDependencies, pipInstaller.DependencyToFileMap)
	if err != nil {
		return err
	}
	dependencies.UpdateDependenciesCache(allDependencies)

	// If missing dependencies information, fail the execution.
	if len(missingDeps) > 0 {
		return errorutils.CheckError(errors.New(fmt.Sprintf("Could not find information for the following packages: %s", strings.Join(missingDeps, ", "))))
	}

	// Build dependencies tree.
	rootDependencies := extractor.DirectDependencies()
	childrenMap := extractor.ChildrenMap()
	pdtc.DepsTreeRoot = dependenciestree.CreateDependencyTree(rootDependencies, allDependencies, childrenMap)

	// Output tree json.
	treeJson, err := json.Marshal(pdtc.DepsTreeRoot)
	if err != nil {
		return errorutils.CheckError(err)
	}
	log.Output(clientutils.IndentJson(treeJson))

	return nil
}

// Populate allDependencies map with dependencies information -> checksums and file-names.
func (pdtc *PipDepTreeCommand) pupulateDepsInfo(allDependencies map[string]*buildinfo.Dependency, depToFileMap map[string]string) ([]string, error) {
	dependenciesCache, err := dependencies.GetProjectDependenciesCache()
	if err != nil {
		return nil, err
	}
	servicesManager, err := utils.CreateServiceManager(pdtc.rtDetails, false)
	if err != nil {
		return nil, err
	}
	return dependencies.AddDepsInfoAndReturnMissingDeps(allDependencies, dependenciesCache, depToFileMap, servicesManager, pdtc.repository)
}

// Remove from allDependencies all the dependencies which weren't installed or collected during this execution,
// based on the information gathered in depToFileMap.
func (pdtc *PipDepTreeCommand) removeUninstalledDependencies(allDependencies map[string]*buildinfo.Dependency, depToFileMap map[string]string) {
	for dep := range allDependencies {
		if _, ok := depToFileMap[dep]; !ok {
			// Remove from allDependencies.
			delete(allDependencies, dep)
		}
	}
}

func (pdtc *PipDepTreeCommand) CommandName() string {
	return "rt_pip_deps_tree"
}

func (pdtc *PipDepTreeCommand) RtDetails() (*config.ArtifactoryDetails, error) {
	return pdtc.rtDetails, nil
}
