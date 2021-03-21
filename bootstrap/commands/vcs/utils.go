package commands

import (
	"encoding/json"
	"errors"
	"net/http"

	artUtils "github.com/jfrog/jfrog-cli-core/artifactory/utils"
	utilsconfig "github.com/jfrog/jfrog-cli-core/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	"github.com/jfrog/jfrog-client-go/auth"
	"github.com/jfrog/jfrog-client-go/http/jfroghttpclient"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
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

func GetVirtualRepo(serviceDetails *auth.ServiceDetails, repoKey string) (*services.VirtualRepositoryBaseParams, error) {

	httpClientsDetails := (*serviceDetails).CreateHttpClientDetails()
	client, err := jfroghttpclient.JfrogClientBuilder().SetInsecureTls(true).SetServiceDetails(serviceDetails).Build()
	if err != nil {
		return nil, err
	}
	api := "api/repositories/" + repoKey
	resp, body, _, err := client.SendGet((*serviceDetails).GetUrl()+api, true, &httpClientsDetails)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errorutils.CheckError(errors.New("Artifactory response: " + resp.Status + "\n" + clientutils.IndentJson(body)))
	}
	var repoDetailes services.VirtualRepositoryBaseParams
	err = json.Unmarshal(body, &repoDetailes)
	return &repoDetailes, err

}

func contains(arr []string, str string) bool {
	for _, element := range arr {
		if element == str {
			return true
		}
	}
	return false
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
