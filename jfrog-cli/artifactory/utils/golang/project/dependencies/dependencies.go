package dependencies

import (
	"fmt"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils"
	golangutil "github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils/golang"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/jfrog/jfrog-client-go/artifactory/services/go"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"path/filepath"
)

func Load() ([]Package, error) {
	cachePath, err := GetCachePath()
	if err != nil {
		return nil, err
	}
	return loadDependencies(cachePath)
}

func GetCachePath() (string, error) {
	goPath, err := getGOPATH()
	if err != nil {
		return "", errorutils.CheckError(err)
	}
	return filepath.Join(goPath, "pkg", "mod", "cache", "download"), nil
}

// Represent go dependency package.
type Package struct {
	buildInfoDependencies []buildinfo.Dependency
	id                    string
	modContent            []byte
	zipPath               string
	version               string
}

func (dependencyPackage *Package) New(cachePath string, dep Package) GoPackage {
	dependencyPackage.modContent = dep.modContent
	dependencyPackage.zipPath = dep.zipPath
	dependencyPackage.version = dep.version
	dependencyPackage.id = dep.id
	dependencyPackage.buildInfoDependencies = dep.buildInfoDependencies
	return dependencyPackage
}

func (dependencyPackage *Package) GetId() string {
	return dependencyPackage.id
}

func (dependencyPackage *Package) GetModContent() []byte {
	return dependencyPackage.modContent
}

func (dependencyPackage *Package) SetModContent(modContent []byte) {
	dependencyPackage.modContent = modContent
}

func (dependencyPackage *Package) GetZipPath() string {
	return dependencyPackage.zipPath
}

// Init the dependency information if needed.
func (dependencyPackage *Package) Init() error {
	return nil
}

func (dependencyPackage *Package) PopulateModAndPublish(targetRepo string, cache *golangutil.DependenciesCache, details *config.ArtifactoryDetails) error {
	return dependencyPackage.prepareAndPublish(targetRepo, cache, details)
}

// Prepare for publishing and publish the dependency to Artifactory
func (dependencyPackage *Package) prepareAndPublish(targetRepo string, cache *golangutil.DependenciesCache, details *config.ArtifactoryDetails) error {
	serviceManager, err := utils.CreateServiceManager(details, false)
	if err != nil {
		cache.IncrementFailures()
		return err
	}
	successOutOfTotal := fmt.Sprintf("%d/%d", cache.GetSuccesses()+1, cache.GetTotal())
	err = dependencyPackage.Publish(successOutOfTotal, targetRepo, serviceManager)
	if err != nil {
		cache.IncrementFailures()
		return err
	}
	cache.IncrementSuccess()
	return nil
}

func (dependencyPackage *Package) Publish(summary string, targetRepo string, servicesManager *artifactory.ArtifactoryServicesManager) error {
	message := fmt.Sprintf("Publishing: %s to %s", dependencyPackage.id, targetRepo)
	if summary != "" {
		message += ":" + summary
	}
	log.Info(message)
	params := &_go.GoParamsImpl{}
	params.ZipPath = dependencyPackage.zipPath
	params.ModContent = dependencyPackage.modContent
	params.Version = dependencyPackage.version
	params.TargetRepo = targetRepo
	params.ModuleId = dependencyPackage.id

	return servicesManager.PublishGoProject(params)
}

func (dependencyPackage *Package) Dependencies() []buildinfo.Dependency {
	return dependencyPackage.buildInfoDependencies
}