package commands

import (
	artUtils "github.com/jfrog/jfrog-cli-core/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	"github.com/jfrog/jfrog-client-go/auth"
	agentutils "github.com/jfrog/jfrog-vcs-agent/utils"
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
	NpmRemoteDefaultUrl      = "https://registry.npmjs.org"
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
func convertVcsDataToBuildConfig(vcsData VcsData) *agentutils.BuildConfig {
	return &agentutils.BuildConfig{
		ProjectName:  vcsData.ProjectName,
		BuildCommand: vcsData.BuildCommand,
		Vcs: &agentutils.Vcs{
			Url:      vcsData.VcsCredentials.GetUrl(),
			User:     vcsData.VcsCredentials.GetUser(),
			Password: vcsData.VcsCredentials.GetPassword(),
			Token:    vcsData.VcsCredentials.GetAccessToken(),
			Branches: vcsData.VcsBranches,
		},
		Jfrog: &agentutils.JfrogDetails{
			ArtUrl:   vcsData.JfrogCredentials.GetUrl(), // need to seperate url to Artifactory, Xray and pipelines
			User:     vcsData.JfrogCredentials.GetUser(),
			Password: vcsData.JfrogCredentials.GetPassword(),
			//Repositories: vcsData.ArtifactoryVirtualRepos,
			BuildName: vcsData.BuildName,
		},
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
