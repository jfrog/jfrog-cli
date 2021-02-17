package commands

import (
	artUtils "github.com/jfrog/jfrog-cli-core/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	"github.com/jfrog/jfrog-client-go/auth"
)

const (
	// Repo types
	Remote  = "remote"
	Virtual = "virtual"
	Local   = "local"

	NewRepository = "new reposetory"

	// Repos defaults
	MavenRemoteDefaultName   = "maven-remote"
	MavenRemoteDefaultUrl    = "https://jcenter.bintray.com"
	MavenVirtualDefaultName  = "maven-virtual"
	GradleRemoteDefaultName  = "gradle-remote"
	GradleRemoteDefaultUrl   = "https://jcenter.bintray.com"
	GradleVirtualDefaultName = "gradle-virtual"
	NpmRemoteDefaultName     = "npm-remote"
	NpmRemoteDefaultUrl      = "http://mirror.centos.org/centos/"
	NpmVirtualDefaultName    = "npm-virtual"
)

func GetAllRepos(serviceDetails auth.ServiceDetails, repoType, packageType string) (*[]services.RepositoryDetails, error) {
	servicesManager, err := artUtils.CreateServiceManager(convertServiceDetailesToRTDetails(serviceDetails), false)
	if err != nil {
		return nil, err
	}
	filterParams := services.RepositoriesFilterParams{RepoType: repoType, PackageType: packageType}
	return servicesManager.GetAllRepositoriesFiltered(filterParams)
}

func CreateRemoteRepo(serviceDetails auth.ServiceDetails, technologyType Technology, repoName, remoteUrl string) error {
	servicesManager, err := artUtils.CreateServiceManager(convertServiceDetailesToRTDetails(serviceDetails), false)
	if err != nil {
		return err
	}
	params := services.NewRemoteRepositoryBaseParams()
	params.PackageType = string(technologyType)
	params.Key = repoName
	params.Url = remoteUrl
	return servicesManager.CreateBasicRemoteRepository(params)
}

func CreateVirtualRepo(serviceDetails auth.ServiceDetails, technologyType Technology, repoName string, repositories ...string) error {
	servicesManager, err := artUtils.CreateServiceManager(convertServiceDetailesToRTDetails(serviceDetails), false)
	if err != nil {
		return err
	}
	params := services.NewVirtualRepositoryBaseParams()
	params.PackageType = string(technologyType)
	params.Key = repoName
	params.Repositories = repositories
	return servicesManager.CreateBasicVirtualRepository(params)
}

func convertServiceDetailesToRTDetails(serviceDetails auth.ServiceDetails) *config.ArtifactoryDetails {
	return &config.ArtifactoryDetails{
		Url:                  serviceDetails.GetUrl(),
		SshUrl:               "",
		DistributionUrl:      "",
		User:                 serviceDetails.GetUser(),
		Password:             serviceDetails.GetPassword(),
		SshKeyPath:           "",
		SshPassphrase:        "",
		AccessToken:          serviceDetails.GetAccessToken(),
		RefreshToken:         "",
		TokenRefreshInterval: 0,
		ClientCertPath:       "",
		ClientCertKeyPath:    "",
		ServerId:             "",
		IsDefault:            false,
		InsecureTls:          false,
		ApiKey:               serviceDetails.GetApiKey(),
	}
}

func ConvertRepoDetailsToRepoNames(reposDetails *[]services.RepositoryDetails) (reposNames []string) {
	for _, repo := range *reposDetails {
		reposNames = append(reposNames, repo.Key)
	}
	return
}

func GetRemoteDefaultName(technologyType Technology) string {
	switch technologyType {
	case Maven:
		return MavenRemoteDefaultName
	case Gradle:
		return GradleRemoteDefaultName
	case Npm:
		return NpmRemoteDefaultName
	default:
		return ""
	}
}

func GetRemoteDefaultUrl(technologyType Technology) string {
	switch technologyType {
	case Maven:
		return MavenRemoteDefaultUrl
	case Gradle:
		return GradleRemoteDefaultUrl
	case Npm:
		return NpmRemoteDefaultUrl
	default:
		return ""
	}
}
