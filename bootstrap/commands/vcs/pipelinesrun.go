package commands

import (
	"errors"
	"github.com/jfrog/jfrog-cli-core/utils/config"
	"github.com/jfrog/jfrog-cli-core/utils/coreutils"
	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/auth"
	clientConfig "github.com/jfrog/jfrog-client-go/config"
	"github.com/jfrog/jfrog-client-go/pipelines"
	"github.com/jfrog/jfrog-client-go/pipelines/services"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"strings"
)

const (
	pipelinesYamlPath = "pipelines.yml"
)

func runPipelinesBootstrap(vcsData *VcsData, pipelinesToken string) ([]byte, error) {
	serviceDetails, err := config.GetSpecificConfig(ConfigServerId, true, false)
	if err != nil {
		return nil, err
	}

	psm, err := createPipelinesServiceManager(serviceDetails, pipelinesToken)
	if err != nil {
		return nil, err
	}

	vcsIntName, vcsIntId, err := createVcsIntegration(vcsData.GitProvider, psm, vcsData)
	if err != nil {
		return nil, err
	}

	rtIntName, err := createArtifactoryIntegration(psm, serviceDetails, vcsData)
	if err != nil {
		return nil, err
	}

	_, err = psm.AddPipelineSource(vcsIntId, getRepoFullName(vcsData), vcsData.GitBranch, pipelinesYamlPath)
	if err != nil {
		return nil, err
	}

	return createPipelinesYaml(vcsIntName, rtIntName, vcsData)
}

func createPipelinesServiceManager(details *config.ServerDetails, pipelinesToken string) (*pipelines.PipelinesServicesManager, error) {
	// Create new details with pipelines token.
	pipelinesDetails := *details
	pipelinesDetails.AccessToken = pipelinesToken

	certsPath, err := coreutils.GetJfrogCertsDir()
	if err != nil {
		return nil, err
	}
	pAuth, err := pipelinesDetails.CreatePipelinesAuthConfig()
	if err != nil {
		return nil, err
	}
	serviceConfig, err := clientConfig.NewConfigBuilder().
		SetServiceDetails(pAuth).
		SetCertificatesPath(certsPath).
		SetInsecureTls(pipelinesDetails.InsecureTls).
		SetDryRun(false).
		Build()
	if err != nil {
		return nil, err
	}
	return pipelines.New(&pAuth, serviceConfig)
}

func createVcsIntegration(gitProvider GitProvider, psm *pipelines.PipelinesServicesManager, vcsData *VcsData) (integrationName string, integrationId int, err error) {
	switch gitProvider {
	case Github:
		integrationName = createIntegrationName(services.GithubName, vcsData)
		integrationId, err = psm.CreateGithubIntegration(integrationName, vcsData.VcsCredentials.AccessToken)
	case Bitbucket:
		integrationName = createIntegrationName(services.BitbucketName, vcsData)
		integrationId, err = psm.CreateBitbucketIntegration(integrationName, vcsData.VcsCredentials.User, vcsData.VcsCredentials.AccessToken)
	case Gitlab:
		integrationName = createIntegrationName(services.GitlabName, vcsData)
		integrationId, err = psm.CreateGitlabIntegration(integrationName, vcsData.VcsCredentials.Url, vcsData.VcsCredentials.AccessToken)
	default:
		return "", -1, errorutils.CheckError(errors.New("vcs type is not supported at the moment"))
	}
	return
}

func createArtifactoryIntegration(psm *pipelines.PipelinesServicesManager, details *config.ServerDetails, vcsData *VcsData) (string, error) {
	integrationName := createIntegrationName("rt", vcsData)
	apiKey, err := getApiKey(details)
	if err != nil {
		return "", err
	}
	user := details.User
	if user == "" {
		user, err = auth.ExtractUsernameFromAccessToken(details.AccessToken)
		if err != nil {
			return "", err
		}
	}
	_, err = psm.CreateArtifactoryIntegration(integrationName, details.ArtifactoryUrl, user, apiKey)
	return integrationName, err
}

// Get API Key if exists, generate one if needed.
func getApiKey(details *config.ServerDetails) (string, error) {
	if details.ApiKey != "" {
		return details.ApiKey, nil
	}

	// Try getting API Key for the user if exists.
	asm, err := createRtServiceManager(details)
	if err != nil {
		return "", err
	}
	apiKey, err := asm.GetAPIKey()
	if err != nil || apiKey != "" {
		return apiKey, err
	}

	// Generate API Key for the user.
	return asm.CreateAPIKey()
}

func createIntegrationName(intType string, data *VcsData) string {
	return intType + "_" + createPipelinesSuitableName(data, "integration")
}

func createPipelinesSuitableName(data *VcsData, suffix string) string {
	name := strings.Join([]string{data.ProjectDomain, data.RepositoryName, suffix}, "_")
	// Pipelines does not allow "-" which might exist in repo names.
	return strings.Replace(name, "-", "_", -1)
}

func createRtServiceManager(artDetails *config.ServerDetails) (artifactory.ArtifactoryServicesManager, error) {
	certsPath, err := coreutils.GetJfrogCertsDir()
	if err != nil {
		return nil, err
	}
	artAuth, err := artDetails.CreateArtAuthConfig()
	if err != nil {
		return nil, err
	}
	serviceConfig, err := clientConfig.NewConfigBuilder().
		SetServiceDetails(artAuth).
		SetCertificatesPath(certsPath).
		SetInsecureTls(artDetails.InsecureTls).
		SetDryRun(false).
		Build()
	if err != nil {
		return nil, err
	}
	return artifactory.New(&artAuth, serviceConfig)
}

func getRepoFullName(data *VcsData) string {
	return data.ProjectDomain + "/" + data.RepositoryName
}
