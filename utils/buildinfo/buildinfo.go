package buildinfo

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jfrog/build-info-go/build"
	buildinfo "github.com/jfrog/build-info-go/entities"
	"github.com/jfrog/build-info-go/flexpack"
	"github.com/jfrog/gofrog/crypto"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	buildUtils "github.com/jfrog/jfrog-cli-core/v2/common/build"
	"github.com/jfrog/jfrog-cli-core/v2/common/project"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	specutils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/utils/io/content"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

const (
	// Configuration file paths
	mavenConfigPath = ".jfrog/projects/maven.yaml"

	// Environment variables
	autoPublishBuildInfoEnv = "JFROG_AUTO_PUBLISH_BUILD_INFO"
)

// createBuildInfoServiceWithAdapter creates a build info service with logger compatibility
// This wrapper fixes the logger interface incompatibility between jfrog-client-go and build-info-go
func createBuildInfoServiceWithAdapter() *build.BuildInfoService {
	// Now that we fixed the Log interface compatibility, we can use the original function
	return buildUtils.CreateBuildInfoService()
}

// GetBuildInfoForPackageManager determines the package manager and collects appropriate build info
func GetBuildInfoForPackageManager(pkgManager, workingDir string, buildConfiguration *buildUtils.BuildConfiguration) error {
	log.Debug("Collecting build info for package manager: " + pkgManager)

	switch pkgManager {
	case "poetry":
		return GetPoetryBuildInfo(workingDir, buildConfiguration)
	case "mvn", "maven":
		return GetMavenBuildInfo(workingDir, buildConfiguration)
	case "pip", "pipenv":
		// For now, fall back to generic build info collection
		// This can be extended with pip-specific native implementations
		return GetBuildInfoForUploadedArtifacts("", buildConfiguration)
	case "gem":
		// Use existing gem implementation
		return GetBuildInfoForUploadedArtifacts("", buildConfiguration)
	default:
		log.Warn("Package manager " + pkgManager + " not specifically supported, using generic build info collection")
		return GetBuildInfoForUploadedArtifacts("", buildConfiguration)
	}
}

// GetPoetryBuildInfo collects build info for Poetry projects
func GetPoetryBuildInfo(workingDir string, buildConfiguration *buildUtils.BuildConfiguration) error {
	log.Debug("Collecting Poetry build info from directory: " + workingDir)

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

	repoConfig, err := extractRepositoryConfigForProject(project.Poetry)
	if err != nil {
		log.Error("Failed to extract Poetry repository configuration: ", err)
		return fmt.Errorf("extractRepositoryConfigForProject failed: %w", err)
	}

	log.Debug("Poetry repo config - Repo: " + repoConfig.TargetRepo())

	// Get server details for build info collection
	serverDetails, err := repoConfig.ServerDetails()
	if err != nil {
		log.Error("Failed to retrieve server details: ", err)
		return fmt.Errorf("ServerDetails extraction failed: %w", err)
	}

	err = collectPoetryBuildInfo(workingDir, buildName, buildNumber, serverDetails, repoConfig.TargetRepo(), buildConfiguration)
	if err != nil {
		log.Warn("Enhanced Poetry collection failed, falling back to standard method: " + err.Error())
		err = saveBuildInfo(serverDetails, repoConfig.TargetRepo(), "", buildConfiguration)
		if err != nil {
			log.Error("Failed to save Poetry build info: ", err)
			return fmt.Errorf("saveBuildInfo failed: %w", err)
		}
	}

	// Check if auto-publish is enabled
	autoPublish := os.Getenv(autoPublishBuildInfoEnv)
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

// GetMavenBuildInfo collects build info for Maven projects
func GetMavenBuildInfo(workingDir string, buildConfiguration *buildUtils.BuildConfiguration) error {
	log.Debug("Collecting Maven build info from directory: " + workingDir)

	buildName, err := buildConfiguration.GetBuildName()
	if err != nil {
		return fmt.Errorf("failed to get build name: %w", err)
	}

	buildNumber, err := buildConfiguration.GetBuildNumber()
	if err != nil {
		return fmt.Errorf("failed to get build number: %w", err)
	}

	config := flexpack.MavenConfig{
		WorkingDirectory:        workingDir,
		IncludeTestDependencies: true,
	}

	mavenFlex, err := flexpack.NewMavenFlexPack(config)
	if err != nil {
		log.Warn("Failed to create Maven FlexPack instance: " + err.Error())
		return GetBuildInfoForUploadedArtifacts("", buildConfiguration)
	}

	buildInfo, err := mavenFlex.CollectBuildInfo(buildName, buildNumber)
	if err != nil {
		log.Warn("Failed to collect build info with Maven FlexPack: " + err.Error())
		return GetBuildInfoForUploadedArtifacts("", buildConfiguration)
	}

	log.Debug(fmt.Sprintf("Collected Maven build info with %d dependencies",
		len(buildInfo.Modules[0].Dependencies)))

	// Now collect artifacts that Maven deployed natively (true native approach)
	err = CollectDeployedMavenArtifacts(buildInfo, workingDir)
	if err != nil {
		log.Warn("Failed to collect deployed Maven artifacts: " + err.Error())
		// Continue anyway - dependencies are more important than artifacts
	}

	// Set build properties on deployed Maven artifacts (this tags them with build info)
	err = setMavenBuildProperties(buildInfo, buildConfiguration)
	if err != nil {
		log.Warn("Failed to set build properties on Maven artifacts: " + err.Error())
	}

	// Save complete build info (dependencies from FlexPack + artifacts from deployment)
	err = saveMavenBuildInfoForJfrogCli(buildInfo, buildConfiguration)
	if err != nil {
		log.Error("Failed to save Maven build info: " + err.Error())
		return fmt.Errorf("saveMavenBuildInfoForJfrogCli failed: %w", err)
	}

	// Check if auto-publish is enabled
	autoPublish := os.Getenv(autoPublishBuildInfoEnv)
	if autoPublish == "true" {

		// Get server details for publishing
		repoConfig, err := extractRepositoryConfig()
		if err != nil {
			log.Error("Failed to extract repository configuration for auto-publish: ", err)
			return fmt.Errorf("extractRepositoryConfig failed: %w", err)
		}

		serverDetails, err := repoConfig.ServerDetails()
		if err != nil {
			log.Error("Failed to retrieve server details for auto-publish: ", err)
			return fmt.Errorf("ServerDetails extraction failed: %w", err)
		}

		err = publishBuildInfo(serverDetails, buildName, buildNumber, buildConfiguration.GetProject())
		if err != nil {
			log.Error("Failed to auto-publish Maven build info: ", err)
			return fmt.Errorf("publishBuildInfo failed: %w", err)
		}
	}

	return nil
}

// saveMavenBuildInfoForJfrogCli saves Maven build info in a format compatible with jfrog-cli
func saveMavenBuildInfoForJfrogCli(buildInfo *buildinfo.BuildInfo, buildConfiguration *buildUtils.BuildConfiguration) error {
	log.Debug("Saving Maven build info for jfrog-cli compatibility")

	// Use build-info-go's build service to save build info
	buildInfoService := createBuildInfoServiceWithAdapter()

	// Get project key from build configuration
	projectKey := buildConfiguration.GetProject()

	// Get or create build instance
	buildInstance, err := buildInfoService.GetOrCreateBuildWithProject(
		buildInfo.Name,
		buildInfo.Number,
		projectKey,
	)
	if err != nil {
		return fmt.Errorf("failed to get or create build: %w", err)
	}

	// Convert entities.BuildInfo to the format expected by the build service
	// For now, save the build info directly using the SaveBuildInfo method
	err = buildInstance.SaveBuildInfo(buildInfo)
	if err != nil {
		return fmt.Errorf("failed to save build info: %w", err)
	}

	log.Debug("Successfully saved Maven build info for jfrog-cli")
	return nil
}

// GetBuildInfoForUploadedArtifacts handles build info for uploaded artifacts (generic fallback)
func GetBuildInfoForUploadedArtifacts(uploadedFile string, buildConfiguration *buildUtils.BuildConfiguration) error {
	repoConfig, err := extractRepositoryConfig()
	if err != nil {
		log.Error("Failed to extract repository configuration: ", err)
		return fmt.Errorf("extractRepositoryConfig failed: %w", err)
	}

	serverDetails, err := repoConfig.ServerDetails()
	if err != nil {
		log.Error("Failed to retrieve server details: ", err)
		return fmt.Errorf("ServerDetails extraction failed: %w", err)
	}

	err = saveBuildInfo(serverDetails, repoConfig.TargetRepo(), uploadedFile, buildConfiguration)
	if err != nil {
		log.Error("Failed to save build info: ", err)
		return fmt.Errorf("saveBuildInfo failed: %w", err)
	}

	return nil
}

// saveBuildInfo saves build info with the given configuration
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

// getBuildPropsForArtifact creates build properties for artifacts
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

// createBuildInfo creates build info with artifacts
func createBuildInfo(buildName, buildNumber, project, moduleName string, artifacts []buildinfo.Artifact) error {
	log.Debug("Creating build info service...")
	buildInfoService := createBuildInfoServiceWithAdapter()

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

	buildInfoService := createBuildInfoServiceWithAdapter()
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

// extractRepositoryConfig extracts generic repository configuration
func extractRepositoryConfig() (*project.RepositoryConfig, error) {
	return extractRepositoryConfigForProject(project.ProjectType(0)) // Use generic project type
}

// extractRepositoryConfigForProject extracts repository configuration for a specific project type
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

// CreateAqlQueryForSearch creates an AQL query for searching artifacts
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

// collectPoetryBuildInfo collects Poetry dependencies and artifacts for build info
func collectPoetryBuildInfo(workingDir, buildName, buildNumber string, serverDetails *config.ServerDetails, targetRepo string, buildConfiguration *buildUtils.BuildConfiguration) error {
	log.Debug("Initializing Poetry dependency collection...")

	// Create Poetry configuration
	config := flexpack.PoetryConfig{
		WorkingDirectory:       workingDir,
		IncludeDevDependencies: false, // Match standard behavior
	}

	// Create Poetry instance
	poetryCollector, err := flexpack.NewPoetryFlexPack(config)
	if err != nil {
		return fmt.Errorf("failed to create Poetry collector: %w", err)
	}

	// Collect build info with dependencies
	buildInfo, err := poetryCollector.CollectBuildInfo(buildName, buildNumber)
	if err != nil {
		return fmt.Errorf("failed to collect Poetry build info: %w", err)
	}

	log.Info(fmt.Sprintf("Collected build info with %d modules", len(buildInfo.Modules)))
	if len(buildInfo.Modules) > 0 {
		depCount := len(buildInfo.Modules[0].Dependencies)
		log.Info(fmt.Sprintf("Module '%s' has %d dependencies with checksums",
			buildInfo.Modules[0].Id, depCount))
	}

	// Collect artifacts and add them to the build info with proper repository paths
	err = addArtifactsToBuildInfo(buildInfo, serverDetails, targetRepo, workingDir)
	if err != nil {
		log.Warn("Failed to add artifacts to build info: " + err.Error())
		// Continue anyway - dependencies are more important than artifacts for now
	}

	// Then set build properties on uploaded artifacts (this tags them with build info)
	err = setPoetryBuildProperties(serverDetails, targetRepo, buildName, buildNumber, buildConfiguration.GetProject(), buildInfo)
	if err != nil {
		log.Warn("Failed to set build properties on artifacts: " + err.Error())
	}

	// Save complete build info (dependencies + artifacts) for jfrog-cli rt bp compatibility
	err = saveBuildInfoNative(buildInfo, buildConfiguration)
	if err != nil {
		return fmt.Errorf("failed to save build info: %w", err)
	}

	log.Info("Successfully saved build info for publishing")
	return nil
}

// wasDeployGoalExecuted checks if the deploy goal was part of the Maven command
func wasDeployGoalExecuted() bool {
	// Check os.Args for Maven goals
	for _, arg := range os.Args {
		if strings.Contains(arg, "deploy") {
			log.Debug("Found deploy goal in command arguments: " + arg)
			return true
		}
	}
	log.Debug("No deploy goal found in command arguments")
	return false
}

// CollectDeployedMavenArtifacts collects artifacts that Maven actually deployed to Artifactory
func CollectDeployedMavenArtifacts(buildInfo *buildinfo.BuildInfo, workingDir string) error {
	log.Debug("Checking for actually deployed Maven artifacts...")

	// Check if deploy goal was actually executed by examining command line arguments
	if !wasDeployGoalExecuted() {
		log.Debug("Deploy goal was not executed - skipping artifact collection for build info")
		return nil
	}

	// Look for artifacts in target directory (what Maven built)
	targetDir := filepath.Join(workingDir, "target")
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		log.Debug("Target directory not found, no artifacts to collect")
		return nil
	}

	// Read Maven project info
	pomPath := filepath.Join(workingDir, "pom.xml")
	pomContent, err := os.ReadFile(pomPath)
	if err != nil {
		return fmt.Errorf("failed to read pom.xml: %w", err)
	}

	pomStr := string(pomContent)
	artifactId := extractXMLValue(pomStr, "artifactId")
	groupId := extractXMLValue(pomStr, "groupId")
	version := extractXMLValue(pomStr, "version")

	if artifactId == "" || version == "" {
		return fmt.Errorf("could not extract artifactId or version from pom.xml")
	}

	// Get deployment repository from configuration
	deployRepo := getDeploymentRepoFromConfig(workingDir)
	if deployRepo == "" {
		log.Warn("No deployment repository found in Maven configuration (" + mavenConfigPath + ")")
		// If no deployment repository is configured, artifacts weren't deployed
		log.Info("No deployment repository configured - artifacts were not deployed to Artifactory")
		return nil
	}

	artifactPatterns := []string{
		fmt.Sprintf("%s-%s.jar", artifactId, version),
		fmt.Sprintf("%s-%s.war", artifactId, version),
		fmt.Sprintf("%s-%s.ear", artifactId, version),
		fmt.Sprintf("%s-%s-sources.jar", artifactId, version),
		fmt.Sprintf("%s-%s-javadoc.jar", artifactId, version),
		fmt.Sprintf("%s-%s-tests.jar", artifactId, version),
	}

	var artifacts []buildinfo.Artifact
	for _, pattern := range artifactPatterns {
		artifactPath := filepath.Join(targetDir, pattern)
		if _, err := os.Stat(artifactPath); err == nil {
			fileDetails, err := crypto.GetFileDetails(artifactPath, true)
			if err != nil {
				log.Warn(fmt.Sprintf("Failed to calculate checksums for %s: %v", pattern, err))
				continue
			}

			artifactType := "jar"
			switch {
			case strings.HasSuffix(pattern, ".war"):
				artifactType = "war"
			case strings.HasSuffix(pattern, ".ear"):
				artifactType = "ear"
			case strings.HasSuffix(pattern, "-sources.jar"):
				artifactType = "java-source-jar"
			case strings.HasSuffix(pattern, "-javadoc.jar"):
				artifactType = "javadoc"
			}

			// Construct the repository path
			repoPath := fmt.Sprintf("%s/%s/%s/%s",
				strings.ReplaceAll(groupId, ".", "/"),
				artifactId,
				version,
				pattern)

			artifact := buildinfo.Artifact{
				Name:                   pattern,
				Type:                   artifactType,
				Path:                   repoPath,
				OriginalDeploymentRepo: deployRepo,
				Checksum: buildinfo.Checksum{
					Sha1:   fileDetails.Checksum.Sha1,
					Sha256: fileDetails.Checksum.Sha256,
					Md5:    fileDetails.Checksum.Md5,
				},
			}

			artifacts = append(artifacts, artifact)
			log.Debug(fmt.Sprintf("Collected deployed artifact: %s with repository path: %s", pattern, repoPath))
		}
	}

	// Process POM file from project root
	pomPath = filepath.Join(workingDir, "pom.xml")
	if _, err := os.Stat(pomPath); err == nil {
		fileDetails, checksumErr := crypto.GetFileDetails(pomPath, true)
		if checksumErr != nil {
			log.Warn(fmt.Sprintf("Failed to calculate checksums for pom.xml: %v", checksumErr))
		} else {
			pomArtifact := buildinfo.Artifact{
				Name: fmt.Sprintf("%s-%s.pom", artifactId, version),
				Type: "pom",
				Path: fmt.Sprintf("%s/%s/%s/%s-%s.pom",
					strings.ReplaceAll(groupId, ".", "/"),
					artifactId,
					version,
					artifactId,
					version),
				OriginalDeploymentRepo: deployRepo,
				Checksum: buildinfo.Checksum{
					Sha1:   fileDetails.Checksum.Sha1,
					Sha256: fileDetails.Checksum.Sha256,
					Md5:    fileDetails.Checksum.Md5,
				},
			}

			artifacts = append(artifacts, pomArtifact)
			log.Debug(fmt.Sprintf("Collected POM artifact: %s", pomArtifact.Name))
		}
	}

	if len(buildInfo.Modules) > 0 && len(artifacts) > 0 {
		buildInfo.Modules[0].Artifacts = artifacts
		log.Debug(fmt.Sprintf("Added %d deployed artifacts to build info", len(artifacts)))
	}

	return nil
}

// extractXMLValue extracts a simple XML value using string parsing
func extractXMLValue(xmlContent, tagName string) string {
	startTag := fmt.Sprintf("<%s>", tagName)
	endTag := fmt.Sprintf("</%s>", tagName)

	startIdx := strings.Index(xmlContent, startTag)
	if startIdx == -1 {
		return ""
	}
	startIdx += len(startTag)

	endIdx := strings.Index(xmlContent[startIdx:], endTag)
	if endIdx == -1 {
		return ""
	}

	return strings.TrimSpace(xmlContent[startIdx : startIdx+endIdx])
}

// getDeploymentRepoFromConfig tries to get deployment repository from JFrog CLI config
func getDeploymentRepoFromConfig(workingDir string) string {
	configPath := filepath.Join(workingDir, mavenConfigPath)
	if content, err := os.ReadFile(configPath); err == nil {
		lines := strings.Split(string(content), "\n")
		inDeployer := false
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "deployer:" {
				inDeployer = true
				continue
			}
			if inDeployer && strings.HasPrefix(trimmed, "releaseRepo:") {
				repo := strings.TrimSpace(strings.TrimPrefix(trimmed, "releaseRepo:"))
				if repo != "" {
					return repo
				}
			}
		}
	}
	return ""
}

// saveBuildInfoNative saves build info for jfrog-cli rt bp compatibility (native path)
func saveBuildInfoNative(buildInfo *buildinfo.BuildInfo, buildConfiguration *buildUtils.BuildConfiguration) error {
	// Use the same approach as createBuildInfo but with the buildUtils service
	buildInfoService := createBuildInfoServiceWithAdapter()

	// Get project key from build configuration
	projectKey := buildConfiguration.GetProject()

	// Create or get build
	bld, err := buildInfoService.GetOrCreateBuildWithProject(buildInfo.Name, buildInfo.Number, projectKey)
	if err != nil {
		return fmt.Errorf("failed to create build: %w", err)
	}

	// Save the complete build info (artifacts + dependencies) using SaveBuildInfo
	log.Debug(fmt.Sprintf("Saving complete build info with %d modules", len(buildInfo.Modules)))

	// Use SaveBuildInfo to save both artifacts and dependencies together
	err = bld.SaveBuildInfo(buildInfo)
	if err != nil {
		return fmt.Errorf("failed to save complete build info: %w", err)
	}

	// Note: No need to call SaveBuildInfo here as AddArtifacts already saves the build
	// The native buildInfo object is used only for extracting artifacts and dependencies
	// The actual build persistence is handled by the build service methods

	log.Info("Build info with artifacts saved successfully")
	return nil
}

// addArtifactsToBuildInfo collects uploaded artifacts with proper repository paths and adds them to the build info
func addArtifactsToBuildInfo(buildInfo *buildinfo.BuildInfo, serverDetails *config.ServerDetails, targetRepo string, workingDir string) error {
	log.Debug("Collecting artifacts for build info...")

	if len(buildInfo.Modules) == 0 {
		return fmt.Errorf("no modules found in build info")
	}

	// Get the main module (should be the Poetry module)
	module := &buildInfo.Modules[0]

	// Collect artifacts with proper repository paths by searching Artifactory for recently uploaded files
	artifacts, err := collectArtifactsWithRepositoryPaths(serverDetails, targetRepo, workingDir)
	if err != nil {
		log.Warn("Failed to collect artifacts with repository paths: " + err.Error())
		// Don't fail the whole process for missing artifacts - continue with empty artifacts list
		artifacts = []buildinfo.Artifact{}
	}

	// Add artifacts to the module
	module.Artifacts = artifacts

	log.Info(fmt.Sprintf("Added %d artifacts to build info", len(artifacts)))
	return nil
}

// collectArtifactsWithRepositoryPaths collects artifacts with their actual repository paths
func collectArtifactsWithRepositoryPaths(serverDetails *config.ServerDetails, targetRepo, workingDir string) ([]buildinfo.Artifact, error) {
	log.Debug("Collecting artifacts with repository paths...")

	// First get the list of local artifacts to know what to search for
	localArtifacts, err := collectPoetryArtifacts(workingDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get local artifacts: %w", err)
	}

	if len(localArtifacts) == 0 {
		log.Debug("No local artifacts found")
		return []buildinfo.Artifact{}, nil
	}

	// Create services manager
	servicesManager, err := utils.CreateServiceManager(serverDetails, -1, 0, false)
	if err != nil {
		return nil, fmt.Errorf("failed to create services manager: %w", err)
	}

	var artifacts []buildinfo.Artifact
	for _, localArtifact := range localArtifacts {
		// Search for this specific artifact in Artifactory to get its actual path
		searchQuery := CreateAqlQueryForSearch(targetRepo, localArtifact.Name)
		searchParams := services.SearchParams{
			CommonParams: &specutils.CommonParams{
				Aql: specutils.Aql{
					ItemsFind: searchQuery,
				},
			},
		}

		searchReader, err := servicesManager.SearchFiles(searchParams)
		if err != nil {
			log.Warn(fmt.Sprintf("Failed to search for artifact %s: %v", localArtifact.Name, err))
			continue
		}

		// Get the most recent result (should be the one we just uploaded)
		for searchResult := new(specutils.ResultItem); searchReader.NextRecord(searchResult) == nil; searchResult = new(specutils.ResultItem) {
			// Create artifact with complete repository path including filename
			fullPath := searchResult.Path
			if !strings.HasSuffix(fullPath, "/") && !strings.HasSuffix(fullPath, searchResult.Name) {
				// Ensure path includes the filename: "test-project/1.0.1" + "/" + "filename.whl"
				fullPath = filepath.Join(searchResult.Path, searchResult.Name)
			}

			artifact := buildinfo.Artifact{
				Name:     searchResult.Name,
				Path:     fullPath, // Complete path: "test-project/1.0.1/test_project-1.0.1-py3-none-any.whl"
				Type:     getArtifactTypeFromName(searchResult.Name),
				Checksum: localArtifact.Checksum, // Use checksums from local file
			}

			artifacts = append(artifacts, artifact)
			break // Only take the first (most recent) result for each artifact
		}

		if err := searchReader.Close(); err != nil {
			log.Warn("Failed to close search reader:", err)
		}
	}

	log.Debug(fmt.Sprintf("Found %d artifacts with repository paths", len(artifacts)))
	return artifacts, nil
}

// collectPoetryArtifacts collects artifacts from the dist/ directory (legacy method)
func collectPoetryArtifacts(workingDir string) ([]buildinfo.Artifact, error) {
	distDir := filepath.Join(workingDir, "dist")

	// Check if dist directory exists
	if _, err := os.Stat(distDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("dist directory not found: %s", distDir)
	}

	// Read dist directory
	entries, err := os.ReadDir(distDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read dist directory: %w", err)
	}

	var artifacts []buildinfo.Artifact
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		filePath := filepath.Join(distDir, filename)

		// Only include Python package files
		if !strings.HasSuffix(filename, ".whl") && !strings.HasSuffix(filename, ".tar.gz") {
			continue
		}

		// Create artifact entry
		artifact := buildinfo.Artifact{
			Name: filename,
			Path: ".", // Relative path in repository
			Type: getArtifactTypeFromName(filename),
		}

		// Calculate checksums
		checksums, err := calculateFileChecksums(filePath)
		if err != nil {
			log.Warn(fmt.Sprintf("Failed to calculate checksums for %s: %v", filename, err))
		} else {
			artifact.Checksum = checksums
		}

		artifacts = append(artifacts, artifact)
	}

	return artifacts, nil
}

// calculateFileChecksums calculates SHA1, SHA256, and MD5 checksums for a file
func calculateFileChecksums(filePath string) (buildinfo.Checksum, error) {
	// Use crypto.GetFileDetails to calculate all checksums at once
	fileDetails, err := crypto.GetFileDetails(filePath, true)
	if err != nil {
		return buildinfo.Checksum{}, fmt.Errorf("failed to calculate checksums: %w", err)
	}

	return buildinfo.Checksum{
		Sha1:   fileDetails.Checksum.Sha1,
		Sha256: fileDetails.Checksum.Sha256,
		Md5:    fileDetails.Checksum.Md5,
	}, nil
}

// getArtifactTypeFromName determines artifact type based on filename
func getArtifactTypeFromName(filename string) string {
	if strings.HasSuffix(filename, ".whl") {
		return "wheel"
	} else if strings.HasSuffix(filename, ".tar.gz") {
		return "sdist"
	}
	return "unknown"
}

// setPoetryBuildProperties sets build properties on uploaded Poetry artifacts
// This ensures artifacts are tagged with build.name, build.number, and build.timestamp
// just like npm, maven, and gradle package managers do
func setPoetryBuildProperties(serverDetails *config.ServerDetails, targetRepo, buildName, buildNumber, project string, buildInfo *buildinfo.BuildInfo) error {
	log.Debug("Setting build properties on Poetry artifacts...")

	// Create services manager
	servicesManager, err := utils.CreateServiceManager(serverDetails, -1, 0, false)
	if err != nil {
		return fmt.Errorf("failed to create services manager: %w", err)
	}

	// Create build properties
	buildProps, err := getBuildPropsForArtifact(buildName, buildNumber, project)
	if err != nil {
		return fmt.Errorf("failed to create build properties: %w", err)
	}

	// Get artifacts from build info that were already collected
	if len(buildInfo.Modules) == 0 {
		log.Debug("No modules in build info, skipping property setting")
		return nil
	}

	module := &buildInfo.Modules[0] // Poetry should have one main module
	if len(module.Artifacts) == 0 {
		log.Debug("No artifacts in module, skipping property setting")
		return nil
	}

	// Set properties on each specific artifact by name and path
	for _, artifact := range module.Artifacts {
		fullPath := fmt.Sprintf("%s/%s/%s", targetRepo, artifact.Path, artifact.Name)
		log.Info(fmt.Sprintf("[Thread %d] Setting properties on: %s", 1, fullPath))

		// Create AQL query for this specific artifact
		searchQuery := CreateAqlQueryForSearch(targetRepo, artifact.Name)
		searchParams := services.SearchParams{
			CommonParams: &specutils.CommonParams{
				Aql: specutils.Aql{
					ItemsFind: searchQuery,
				},
			},
		}

		searchReader, err := servicesManager.SearchFiles(searchParams)
		if err != nil {
			log.Warn(fmt.Sprintf("Failed to find artifact %s: %v", artifact.Name, err))
			continue // Continue with other artifacts
		}

		// Set properties on this specific artifact
		propsParams := services.PropsParams{
			Reader: searchReader,
			Props:  buildProps,
		}

		_, err = servicesManager.SetProps(propsParams)
		if err != nil {
			log.Warn(fmt.Sprintf("Failed to set properties on artifact %s: %v", artifact.Name, err))
			continue // Continue with other artifacts
		}

		log.Debug(fmt.Sprintf("Successfully set properties on artifact: %s", artifact.Name))
	}

	log.Info(fmt.Sprintf("Successfully set build properties on %d Poetry artifacts", len(module.Artifacts)))
	return nil
}

// setMavenBuildProperties sets build properties on deployed Maven artifacts
// This ensures artifacts are tagged with build.name, build.number, and build.timestamp
// just like npm, maven, and gradle package managers do
func setMavenBuildProperties(buildInfo *buildinfo.BuildInfo, buildConfiguration *buildUtils.BuildConfiguration) error {
	log.Debug("Setting build properties on Maven artifacts...")

	// Get build configuration
	buildName, err := buildConfiguration.GetBuildName()
	if err != nil {
		return fmt.Errorf("failed to get build name: %w", err)
	}

	buildNumber, err := buildConfiguration.GetBuildNumber()
	if err != nil {
		return fmt.Errorf("failed to get build number: %w", err)
	}

	projectName := buildConfiguration.GetProject()

	// Extract repository configuration to get server details
	repoConfig, err := extractRepositoryConfigForProject(project.Maven)
	if err != nil {
		return fmt.Errorf("failed to extract repository configuration: %w", err)
	}

	// Get server details from repository configuration
	serverDetails, err := repoConfig.ServerDetails()
	if err != nil {
		return fmt.Errorf("failed to get server details: %w", err)
	}

	// Create services manager
	servicesManager, err := utils.CreateServiceManager(serverDetails, -1, 0, false)
	if err != nil {
		return fmt.Errorf("failed to create services manager: %w", err)
	}

	// Create build properties
	buildProps, err := getBuildPropsForArtifact(buildName, buildNumber, projectName)
	if err != nil {
		return fmt.Errorf("failed to create build properties: %w", err)
	}

	// Get artifacts from build info that were already collected
	if len(buildInfo.Modules) == 0 {
		log.Debug("No modules in build info, skipping property setting")
		return nil
	}

	module := &buildInfo.Modules[0] // Maven should have one main module
	if len(module.Artifacts) == 0 {
		log.Debug("No artifacts in module, skipping property setting")
		return nil
	}

	// Get the deployment repository from the module
	targetRepo := module.Repository
	if targetRepo == "" {
		log.Warn("No repository specified in module, trying to get from artifact deployment repo")
		if len(module.Artifacts) > 0 {
			targetRepo = module.Artifacts[0].OriginalDeploymentRepo
		}
		if targetRepo == "" {
			return fmt.Errorf("no target repository found for setting properties")
		}
	}

	// Set properties on each specific artifact by name and path
	for _, artifact := range module.Artifacts {
		fullPath := fmt.Sprintf("%s/%s", targetRepo, artifact.Path)
		log.Info(fmt.Sprintf("[Thread %d] Setting properties on Maven artifact: %s", 1, fullPath))

		// Create AQL query for this specific artifact
		searchQuery := CreateAqlQueryForSearch(targetRepo, artifact.Name)
		searchParams := services.SearchParams{
			CommonParams: &specutils.CommonParams{
				Aql: specutils.Aql{
					ItemsFind: searchQuery,
				},
			},
		}

		searchReader, err := servicesManager.SearchFiles(searchParams)
		if err != nil {
			log.Warn(fmt.Sprintf("Failed to find Maven artifact %s: %v", artifact.Name, err))
			continue // Continue with other artifacts
		}

		// Set properties on this specific artifact
		propsParams := services.PropsParams{
			Reader: searchReader,
			Props:  buildProps,
		}

		_, err = servicesManager.SetProps(propsParams)
		if err != nil {
			log.Warn(fmt.Sprintf("Failed to set properties on Maven artifact %s: %v", artifact.Name, err))
			continue // Continue with other artifacts
		}

		log.Debug(fmt.Sprintf("Successfully set properties on Maven artifact: %s", artifact.Name))
	}

	log.Info(fmt.Sprintf("Successfully set build properties on %d Maven artifacts", len(module.Artifacts)))
	return nil
}
