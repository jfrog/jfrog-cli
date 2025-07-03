package newImpl

import (
	"errors"
	"fmt"
	buildinfo "github.com/jfrog/build-info-go/entities"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	buildUtils "github.com/jfrog/jfrog-cli-core/v2/common/build"
	"github.com/jfrog/jfrog-cli-core/v2/common/project"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	serviceutils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	specutils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/utils/io/content"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

type aqlResult struct {
	Results []*serviceutils.ResultItem `json:"results,omitempty"`
}

func GetFiles(pkgManager, fileUploaded string) ([]string, error) {
	switch pkgManager {
	case "gem":
		return []string{fileUploaded + ".gem", fileUploaded + ".gemspec.rz", fileUploaded}, nil
	default:
		return []string{}, errors.New("unsupported package manager")
	}
}

func GetBuildInfoForUploadedArtifacts(uploadedFile string, buildConfiguration *buildUtils.BuildConfiguration) error {
	repoConfig, err := extractRepositoryConfig()
	if err != nil {
		return err
	}
	serverDetails, err := repoConfig.ServerDetails()
	if err != nil {
		return err
	}
	err = saveBuildInfo(serverDetails, repoConfig.TargetRepo(), uploadedFile, buildConfiguration)
	return err
}

func saveBuildInfo(serverDetails *config.ServerDetails, searchRepo string, fileName string, buildConfiguration *buildUtils.BuildConfiguration) error {
	servicesManager, err := utils.CreateServiceManager(serverDetails, -1, 0, false)
	if err != nil {
		return err
	}

	var buildProps string
	var searchReader *content.ContentReader
	var buildName, buildNumber string
	buildName, err = buildConfiguration.GetBuildName()
	if err != nil {
		return err
	}
	buildNumber, err = buildConfiguration.GetBuildNumber()
	if err != nil {
		return err
	}
	buildProject := buildConfiguration.GetProject()

	buildProps, err = getBuildPropsForArtifact(buildName, buildNumber, buildProject)
	if err != nil {
		return err
	}

	searchQuery := CreateAqlQueryForSearch(searchRepo, fileName)
	searchParams := services.SearchParams{
		CommonParams: &specutils.CommonParams{
			Aql: specutils.Aql{
				ItemsFind: searchQuery,
			},
		},
	}
	searchReader, err = servicesManager.SearchFiles(searchParams)
	if err != nil {
		log.Error("Failed to get uploaded npm package: ", err.Error())
		return err
	}

	propsParams := services.PropsParams{
		Reader: searchReader,
		Props:  buildProps,
	}
	_, err = servicesManager.SetProps(propsParams)
	if err != nil {
		log.Warn("Unable to set build properties: ", err, "\nThis may cause build to not properly link with artifact, please add build name and build number properties on the tarball artifact manually")
	}
	buildInfoArtifacts, err := utils.ConvertArtifactsSearchDetailsToBuildInfoArtifacts(searchReader)
	err = createBuildInfo(buildName, buildNumber, buildProject, buildConfiguration.GetModule(), buildInfoArtifacts)
	if err != nil {
		return err
	}
	return nil
}

func getBuildPropsForArtifact(buildName, buildNumber, project string) (string, error) {
	err := buildUtils.SaveBuildGeneralDetails(buildName, buildNumber, project)
	if err != nil {
		return "", err
	}
	return buildUtils.CreateBuildProperties(buildName, buildNumber, project)
}

func createBuildInfo(buildName, buildNumber, project, moduleName string, artifacts []buildinfo.Artifact) error {
	buildInfoService := buildUtils.CreateBuildInfoService()
	build, err := buildInfoService.GetOrCreateBuildWithProject(buildName, buildNumber, project)
	if err != nil {
		return err
	}
	err = build.AddArtifacts(moduleName, "generic", artifacts...)
	return err
}

func extractRepositoryConfig() (*project.RepositoryConfig, error) {
	prefix := project.ProjectConfigDeployerPrefix
	configFilePath, exists, err := project.GetProjectConfFilePath(project.Gem)
	if !exists {
		return nil, fmt.Errorf("project configuration file not found for %s", project.Gem)
	}
	if err != nil {
		return nil, err
	}
	vConfig, err := project.ReadConfigFile(configFilePath, project.YAML)
	if err != nil {
		return nil, err
	}
	repoConfig, err := project.GetRepoConfigByPrefix(configFilePath, prefix, vConfig)
	if err != nil {
		return nil, err
	}
	return repoConfig, nil
}

func CreateAqlQueryForSearch(repo, file string) string {
	itemsPart :=
		`{` +
			`"repo": "%s",` +
			`"$or": [{` +
			`"$and":[{` +
			`"path": {"$match": "*"},` +
			`"name": {"$match": "%s"}` +
			`}]` +
			`}]` +
			`}`
	return fmt.Sprintf(itemsPart, repo, file)
}
