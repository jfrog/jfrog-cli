package newImpl

import (
	"fmt"
	"os"

	buildinfo "github.com/jfrog/build-info-go/entities"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	buildUtils "github.com/jfrog/jfrog-cli-core/v2/common/build"
	"github.com/jfrog/jfrog-cli-core/v2/common/project"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
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
	case "poetry", "pip":
		return []string{fileUploaded + ".tar.gz", fileUploaded + ".whl"}, nil
	case "npm":
		return []string{fileUploaded + ".tgz", fileUploaded + ".tar.gz"}, nil
	case "nuget":
		return []string{fileUploaded + ".nupkg"}, nil
	case "maven", "gradle":
		return []string{fileUploaded + ".jar", fileUploaded + ".war", fileUploaded + ".pom"}, nil
	default:
		return []string{}, fmt.Errorf("unsupported package manager: %s", pkgManager)
	}
}

func GetBuildInfoForUploadedArtifacts(uploadedFile string, buildConfiguration *buildUtils.BuildConfiguration) error {
	log.Info("Extracting repository configuration for build info...")
	repoConfig, err := extractRepositoryConfig()
	if err != nil {
		log.Error("Failed to extract repository configuration: ", err)
		return fmt.Errorf("extractRepositoryConfig failed: %w", err)
	}

	log.Info("Retrieving server details...")
	serverDetails, err := repoConfig.ServerDetails()
	if err != nil {
		log.Error("Failed to retrieve server details: ", err)
		return fmt.Errorf("ServerDetails extraction failed: %w", err)
	}

	log.Info("Saving build info for uploaded file: " + uploadedFile)
	err = saveBuildInfo(serverDetails, repoConfig.TargetRepo(), uploadedFile, buildConfiguration)
	if err != nil {
		log.Error("Failed to save build info: ", err)
		return fmt.Errorf("saveBuildInfo failed: %w", err)
	}

	log.Info("Successfully saved build info for uploaded artifact.")
	return nil
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
		log.Error("Failed to get uploaded package: ", err.Error())
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
	if err != nil {
		log.Error("Failed to convert search results to build info artifacts: ", err)
		return fmt.Errorf("artifact conversion failed: %w", err)
	}

	err = createBuildInfo(buildName, buildNumber, buildProject, buildConfiguration.GetModule(), buildInfoArtifacts)
	if err != nil {
		log.Error("Failed to create build info: ", err)
		return fmt.Errorf("createBuildInfo failed: %w", err)
	}
	return nil
}

func getBuildPropsForArtifact(buildName, buildNumber, project string) (string, error) {
	log.Debug("Saving build general details...")
	err := buildUtils.SaveBuildGeneralDetails(buildName, buildNumber, project)
	if err != nil {
		log.Error("Failed to save build general details: ", err)
		return "", fmt.Errorf("SaveBuildGeneralDetails failed: %w", err)
	}

	log.Debug("Creating build properties...")
	props, err := buildUtils.CreateBuildProperties(buildName, buildNumber, project)
	if err != nil {
		log.Error("Failed to create build properties: ", err)
		return "", fmt.Errorf("CreateBuildProperties failed: %w", err)
	}

	return props, nil
}

func createBuildInfo(buildName, buildNumber, project, moduleName string, artifacts []buildinfo.Artifact) error {
	log.Debug("Creating build info service...")
	buildInfoService := buildUtils.CreateBuildInfoService()

	log.Debug("Getting or creating build: " + buildName + "-" + buildNumber)
	build, err := buildInfoService.GetOrCreateBuildWithProject(buildName, buildNumber, project)
	if err != nil {
		log.Error("Failed to get or create build: ", err)
		return fmt.Errorf("GetOrCreateBuildWithProject failed: %w", err)
	}

	log.Debug("Adding artifacts to module: " + moduleName)
	err = build.AddArtifacts(moduleName, "generic", artifacts...)
	if err != nil {
		log.Error("Failed to add artifacts to build: ", err)
		return fmt.Errorf("AddArtifacts failed: %w", err)
	}

	log.Info("Successfully created build info with artifacts.")
	return nil
}

// publishBuildInfo publishes the build info to Artifactory
func publishBuildInfo(serverDetails *config.ServerDetails, buildName, buildNumber, project string) error {
	log.Info("Publishing build info to Artifactory...")

	servicesManager, err := utils.CreateServiceManager(serverDetails, -1, 0, false)
	if err != nil {
		log.Error("Failed to create services manager: ", err)
		return fmt.Errorf("CreateServiceManager failed: %w", err)
	}

	buildInfoService := buildUtils.CreateBuildInfoService()
	build, err := buildInfoService.GetOrCreateBuildWithProject(buildName, buildNumber, project)
	if err != nil {
		log.Error("Failed to get or create build: ", err)
		return fmt.Errorf("GetOrCreateBuildWithProject failed: %w", err)
	}

	buildInfo, err := build.ToBuildInfo()
	if err != nil {
		log.Error("Failed to convert to build info: ", err)
		return fmt.Errorf("ToBuildInfo failed: %w", err)
	}

	// Set build agent information
	buildInfo.SetAgentName(coreutils.GetCliUserAgentName())
	buildInfo.SetAgentVersion(coreutils.GetCliUserAgentVersion())
	buildInfo.SetBuildAgentVersion(coreutils.GetClientAgentVersion())
	if serverDetails.User != "" {
		buildInfo.Principal = serverDetails.User
	}

	_, err = servicesManager.PublishBuildInfo(buildInfo, project)
	if err != nil {
		log.Error("Failed to publish build info: ", err)
		return fmt.Errorf("PublishBuildInfo failed: %w", err)
	}

	log.Info("Build info successfully published to Artifactory.")
	return nil
}

func extractRepositoryConfig() (*project.RepositoryConfig, error) {
	return extractRepositoryConfigForProject(project.ProjectType(0)) // Use generic project type
}

func extractRepositoryConfigForProject(projectType project.ProjectType) (*project.RepositoryConfig, error) {
	log.Debug("Extracting repository config for project type")
	prefix := project.ProjectConfigDeployerPrefix

	// Handle invalid project types gracefully
	if projectType < 0 {
		return nil, fmt.Errorf("invalid project type")
	}

	configFilePath, exists, err := project.GetProjectConfFilePath(projectType)
	if !exists {
		log.Warn("Project configuration file not found")
		return nil, fmt.Errorf("project configuration file not found")
	}
	if err != nil {
		log.Error("Failed to get project config file path: ", err)
		return nil, fmt.Errorf("GetProjectConfFilePath failed: %w", err)
	}

	log.Debug("Reading configuration file...")
	vConfig, err := project.ReadConfigFile(configFilePath, project.YAML)
	if err != nil {
		log.Error("Failed to read config file: ", err)
		return nil, fmt.Errorf("ReadConfigFile failed: %w", err)
	}

	log.Debug("Getting repository config by prefix...")
	repoConfig, err := project.GetRepoConfigByPrefix(configFilePath, prefix, vConfig)
	if err != nil {
		log.Error("Failed to get repo config by prefix: ", err)
		return nil, fmt.Errorf("GetRepoConfigByPrefix failed: %w", err)
	}

	log.Debug("Successfully extracted repository configuration.")
	return repoConfig, nil
}

// GetPoetryBuildInfo collects build info for Poetry projects
func GetPoetryBuildInfo(workingDir string, buildConfiguration *buildUtils.BuildConfiguration) error {
	log.Info("Collecting Poetry build info from directory: " + workingDir)

	buildName, err := buildConfiguration.GetBuildName()
	if err != nil {
		log.Error("Failed to get build name: ", err)
		return fmt.Errorf("GetBuildName failed: %w", err)
	}

	buildNumber, err := buildConfiguration.GetBuildNumber()
	if err != nil {
		log.Error("Failed to get build number: ", err)
		return fmt.Errorf("GetBuildNumber failed: %w", err)
	}

	log.Debug("Poetry build info collection for build: " + buildName + "-" + buildNumber)

	// Extract repository configuration specifically for Poetry
	log.Info("Extracting repository configuration for Poetry project...")
	repoConfig, err := extractRepositoryConfigForProject(project.Poetry)
	if err != nil {
		log.Error("Failed to extract Poetry repository configuration: ", err)
		return fmt.Errorf("extractRepositoryConfigForProject failed: %w", err)
	}

	log.Info("Retrieved Poetry repository configuration successfully")
	log.Debug("Poetry repo config - Repo: " + repoConfig.TargetRepo())

	// Get server details for build info collection
	serverDetails, err := repoConfig.ServerDetails()
	if err != nil {
		log.Error("Failed to retrieve server details: ", err)
		return fmt.Errorf("ServerDetails extraction failed: %w", err)
	}

	// TODO: Integrate with FlexPack Poetry implementation for dependency collection
	// For now, we'll collect build info with Poetry-specific configuration
	log.Info("Poetry build info collection (will be enhanced with FlexPack integration).")

	// Save build info with Poetry configuration
	err = saveBuildInfo(serverDetails, repoConfig.TargetRepo(), "", buildConfiguration)
	if err != nil {
		log.Error("Failed to save Poetry build info: ", err)
		return fmt.Errorf("saveBuildInfo failed: %w", err)
	}

	log.Info("Successfully collected Poetry build info.")

	// Check if auto-publish is enabled
	autoPublish := os.Getenv("JFROG_AUTO_PUBLISH_BUILD_INFO")
	if autoPublish == "true" {
		log.Info("Auto-publishing build info is enabled.")
		err = publishBuildInfo(serverDetails, buildName, buildNumber, buildConfiguration.GetProject())
		if err != nil {
			log.Error("Failed to auto-publish build info: ", err)
			return fmt.Errorf("publishBuildInfo failed: %w", err)
		}
	} else {
		log.Info("Build info saved locally. Use 'jf rt bp " + buildName + " " + buildNumber + "' to publish it to Artifactory.")
	}

	return nil
}

// GetBuildInfoForPackageManager determines the package manager and collects appropriate build info
func GetBuildInfoForPackageManager(pkgManager, workingDir string, buildConfiguration *buildUtils.BuildConfiguration) error {
	log.Info("Collecting build info for package manager: " + pkgManager)

	switch pkgManager {
	case "poetry":
		return GetPoetryBuildInfo(workingDir, buildConfiguration)
	case "pip", "pipenv":
		// For now, fall back to generic build info collection
		// This can be extended with pip-specific FlexPack implementations
		return GetBuildInfoForUploadedArtifacts("", buildConfiguration)
	case "gem":
		// Use existing gem implementation
		return GetBuildInfoForUploadedArtifacts("", buildConfiguration)
	default:
		log.Warn("Package manager " + pkgManager + " not specifically supported, using generic build info collection")
		return GetBuildInfoForUploadedArtifacts("", buildConfiguration)
	}
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
