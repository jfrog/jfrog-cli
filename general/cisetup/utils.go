package cisetup

import (
	artUtils "github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	utilsconfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	"github.com/jfrog/jfrog-client-go/config"
	"github.com/jfrog/jfrog-client-go/xray"
)

const (
	NewRepository = "[Create new repository]"

	// Build commands defaults
	mavenDefaultBuildCmd  = "mvn clean install"
	gradleDefaultBuildCmd = "gradle clean artifactoryPublish"
	npmDefaultBuildCmd    = "npm install"
)

var buildCmdByTech = map[coreutils.Technology]string{
	coreutils.Maven:  mavenDefaultBuildCmd,
	coreutils.Gradle: gradleDefaultBuildCmd,
	coreutils.Npm:    npmDefaultBuildCmd,
}

func CreateXrayServiceManager(serviceDetails *utilsconfig.ServerDetails) (*xray.XrayServicesManager, error) {
	xrayDetails, err := serviceDetails.CreateXrayAuthConfig()
	if err != nil {
		return nil, err
	}
	serviceConfig, err := config.NewConfigBuilder().
		SetServiceDetails(xrayDetails).
		Build()
	if err != nil {
		return nil, err
	}
	return xray.New(serviceConfig)
}

func GetAllRepos(serviceDetails *utilsconfig.ServerDetails, repoType, packageType string) (*[]services.RepositoryDetails, error) {
	servicesManager, err := artUtils.CreateServiceManager(serviceDetails, -1, 0, false)
	if err != nil {
		return nil, err
	}
	filterParams := services.RepositoriesFilterParams{RepoType: repoType, PackageType: packageType}
	return servicesManager.GetAllRepositoriesFiltered(filterParams)
}

func GetVirtualRepo(serviceDetails *utilsconfig.ServerDetails, repoKey string) (*services.VirtualRepositoryBaseParams, error) {
	servicesManager, err := artUtils.CreateServiceManager(serviceDetails, -1, 0, false)
	if err != nil {
		return nil, err
	}
	virtualRepoDetails := services.VirtualRepositoryBaseParams{}
	err = servicesManager.GetRepository(repoKey, &virtualRepoDetails)
	return &virtualRepoDetails, err
}

func CreateLocalRepo(serviceDetails *utilsconfig.ServerDetails, technologyType coreutils.Technology, repoName string) error {
	servicesManager, err := artUtils.CreateServiceManager(serviceDetails, -1, 0, false)
	if err != nil {
		return err
	}
	params := services.NewLocalRepositoryBaseParams()
	params.PackageType = string(technologyType)
	params.Key = repoName
	return servicesManager.CreateLocalRepositoryWithParams(params)
}

func CreateRemoteRepo(serviceDetails *utilsconfig.ServerDetails, technologyType coreutils.Technology, repoName, remoteUrl string) error {
	servicesManager, err := artUtils.CreateServiceManager(serviceDetails, -1, 0, false)
	if err != nil {
		return err
	}
	params := services.NewRemoteRepositoryBaseParams()
	params.PackageType = string(technologyType)
	params.Key = repoName
	params.Url = remoteUrl
	return servicesManager.CreateRemoteRepositoryWithParams(params)
}

func CreateVirtualRepo(serviceDetails *utilsconfig.ServerDetails, technologyType coreutils.Technology, repoName string, repositories ...string) error {
	servicesManager, err := artUtils.CreateServiceManager(serviceDetails, -1, 0, false)
	if err != nil {
		return err
	}
	params := services.NewVirtualRepositoryBaseParams()
	params.PackageType = string(technologyType)
	params.Key = repoName
	params.Repositories = repositories
	return servicesManager.CreateVirtualRepositoryWithParams(params)
}

func contains(arr []string, str string) bool {
	for _, element := range arr {
		if element == str {
			return true
		}
	}
	return false
}
