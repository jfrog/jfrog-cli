package commands

import (
	artUtils "github.com/jfrog/jfrog-cli-core/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/general/cisetup"
	utilsconfig "github.com/jfrog/jfrog-cli-core/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	"github.com/jfrog/jfrog-client-go/config"
	"github.com/jfrog/jfrog-client-go/xray"
)

const (
	// Repo types
	Remote  = "remote"
	Virtual = "virtual"
	Local   = "local"

	NewRepository = "[Create new repository]"

	// Repos defaults
	MavenRemoteDefaultName   = "maven-central-remote"
	MavenRemoteDefaultUrl    = "https://repo.maven.apache.org/maven2"
	MavenVirtualDefaultName  = "maven-virtual"
	GradleRemoteDefaultName  = "gradle-remote"
	GradleRemoteDefaultUrl   = "https://repo.maven.apache.org/maven2"
	GradleVirtualDefaultName = "gradle-virtual"
	NpmRemoteDefaultName     = "npm-remote"
	NpmRemoteDefaultUrl      = "https://registry.npmjs.org"
	NpmVirtualDefaultName    = "npm-virtual"
)

func CreateXrayServiceManager(serviceDetails *utilsconfig.ServerDetails) (*xray.XrayServicesManager, error) {
	xrayDetails, err := serviceDetails.CreateXrayAuthConfig()
	serviceConfig, err := config.NewConfigBuilder().
		SetServiceDetails(xrayDetails).
		Build()
	if err != nil {
		return nil, err
	}
	return xray.New(&xrayDetails, serviceConfig)
}

func GetAllRepos(serviceDetails *utilsconfig.ServerDetails, repoType, packageType string) (*[]services.RepositoryDetails, error) {
	servicesManager, err := artUtils.CreateServiceManager(serviceDetails, false)
	if err != nil {
		return nil, err
	}
	filterParams := services.RepositoriesFilterParams{RepoType: repoType, PackageType: packageType}
	return servicesManager.GetAllRepositoriesFiltered(filterParams)
}

func GetVirtualRepo(serviceDetails *utilsconfig.ServerDetails, repoKey string) (*services.VirtualRepositoryBaseParams, error) {
	servicesManager, err := artUtils.CreateServiceManager(serviceDetails, false)
	if err != nil {
		return nil, err
	}
	virtualRepoDetailes := services.VirtualRepositoryBaseParams{}
	err = servicesManager.GetRepository(repoKey, &virtualRepoDetailes)
	return &virtualRepoDetailes, err
}

func contains(arr []string, str string) bool {
	for _, element := range arr {
		if element == str {
			return true
		}
	}
	return false
}

func CreateRemoteRepo(serviceDetails *utilsconfig.ServerDetails, technologyType cisetup.Technology, repoName, remoteUrl string) error {
	servicesManager, err := artUtils.CreateServiceManager(serviceDetails, false)
	if err != nil {
		return err
	}
	params := services.NewRemoteRepositoryBaseParams()
	params.PackageType = string(technologyType)
	params.Key = repoName
	params.Url = remoteUrl
	return servicesManager.CreateRemoteRepositoryWithParams(params)
}

func CreateVirtualRepo(serviceDetails *utilsconfig.ServerDetails, technologyType cisetup.Technology, repoName string, repositories ...string) error {
	servicesManager, err := artUtils.CreateServiceManager(serviceDetails, false)
	if err != nil {
		return err
	}
	params := services.NewVirtualRepositoryBaseParams()
	params.PackageType = string(technologyType)
	params.Key = repoName
	params.Repositories = repositories
	return servicesManager.CreateVirtualRepositoryWithParams(params)
}

func GetRemoteDefaultName(technologyType cisetup.Technology) string {
	switch technologyType {
	case cisetup.Maven:
		return MavenRemoteDefaultName
	case cisetup.Gradle:
		return GradleRemoteDefaultName
	case cisetup.Npm:
		return NpmRemoteDefaultName
	default:
		return ""
	}
}

func GetVirtualDefaultName(technologyType cisetup.Technology) string {
	switch technologyType {
	case cisetup.Maven:
		return MavenVirtualDefaultName
	case cisetup.Gradle:
		return GradleVirtualDefaultName
	case cisetup.Npm:
		return NpmVirtualDefaultName
	default:
		return ""
	}
}

func GetRemoteDefaultUrl(technologyType cisetup.Technology) string {
	switch technologyType {
	case cisetup.Maven:
		return MavenRemoteDefaultUrl
	case cisetup.Gradle:
		return GradleRemoteDefaultUrl
	case cisetup.Npm:
		return NpmRemoteDefaultUrl
	default:
		return ""
	}
}
