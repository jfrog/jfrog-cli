package commands

import (
	artUtils "github.com/jfrog/jfrog-cli-core/artifactory/utils"
	utilsconfig "github.com/jfrog/jfrog-cli-core/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
)

const (
	// Repo types
	Remote  = "remote"
	Virtual = "virtual"
	Local   = "local"

	NewRepository = "[Create new repository]"

	// Repos defaults
	MavenRemoteDefaultName   = "maven-remote"
	MavenRemoteDefaultUrl    = "https://jcenter.bintray.com"
	MavenVirtualDefaultName  = "maven-virtual"
	GradleRemoteDefaultName  = "gradle-remote"
	GradleRemoteDefaultUrl   = "https://jcenter.bintray.com"
	GradleVirtualDefaultName = "gradle-virtual"
	NpmRemoteDefaultName     = "npm-remote"
	NpmRemoteDefaultUrl      = "https://registry.npmjs.org"
	NpmVirtualDefaultName    = "npm-virtual"
)

func GetAllRepos(serviceDetails *utilsconfig.ServerDetails, repoType, packageType string) (*[]services.RepositoryDetails, error) {
	servicesManager, err := artUtils.CreateServiceManager(serviceDetails, false)
	if err != nil {
		return nil, err
	}
	filterParams := services.RepositoriesFilterParams{RepoType: repoType, PackageType: packageType}
	return servicesManager.GetAllRepositoriesFiltered(filterParams)
}

func CreateRemoteRepo(serviceDetails *utilsconfig.ServerDetails, technologyType Technology, repoName, remoteUrl string) error {
	servicesManager, err := artUtils.CreateServiceManager(serviceDetails, false)
	if err != nil {
		return err
	}
	params := services.NewRemoteRepositoryBaseParams()
	params.PackageType = string(technologyType)
	params.Key = repoName
	params.Url = remoteUrl
	return servicesManager.CreateBasicRemoteRepository(params)
}

func CreateVirtualRepo(serviceDetails *utilsconfig.ServerDetails, technologyType Technology, repoName string, repositories ...string) error {
	servicesManager, err := artUtils.CreateServiceManager(serviceDetails, false)
	if err != nil {
		return err
	}
	params := services.NewVirtualRepositoryBaseParams()
	params.PackageType = string(technologyType)
	params.Key = repoName
	params.Repositories = repositories
	return servicesManager.CreateBasicVirtualRepository(params)
}

// func convertVcsDataToBuildConfig(vcsData *VcsData) *agentutils.BuildConfig {
// 	serviceDetails, _ := utilsconfig.GetSpecificConfig(ConfigServerId, true, false)
// 	// rtDetails, _ := serviceDetails.CreateArtAuthConfig()
// 	// return &agentutils.BuildConfig{
// 	// 	ProjectName:  vcsData.ProjectName,
// 	// 	BuildCommand: vcsData.BuildCommand,
// 	// 	Vcs: &agentutils.Vcs{
// 	// 		Url:      vcsData.VcsCredentials.Url,
// 	// 		User:     vcsData.VcsCredentials.User,
// 	// 		Password: vcsData.VcsCredentials.Password,
// 	// 		Token:    vcsData.VcsCredentials.AccessToken,
// 	// 		Branches: []string{vcsData.VcsBranch},
// 	// 	},
// 	// 	Jfrog: &agentutils.JfrogDetails{
// 	// 		ArtUrl:   rtDetails.GetUrl(),
// 	// 		User:     rtDetails.GetUser(),
// 	// 		Password: rtDetails.GetPassword(),
// 	// 		//Repositories: vcsData.ArtifactoryVirtualRepos,
// 	// 		BuildName: vcsData.BuildName,
// 	// 	},
// 	// }
// }

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

func GetVirtualDefaultName(technologyType Technology) string {
	switch technologyType {
	case Maven:
		return MavenVirtualDefaultName
	case Gradle:
		return GradleVirtualDefaultName
	case Npm:
		return NpmVirtualDefaultName
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
