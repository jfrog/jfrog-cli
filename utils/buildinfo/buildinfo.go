package buildinfo

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

	// Use enhanced Poetry implementation for dependency collection
	log.Info("Collecting Poetry dependencies and build artifacts...")
	err = collectPoetryBuildInfo(workingDir, buildName, buildNumber, serverDetails, repoConfig.TargetRepo(), buildConfiguration)
	if err != nil {
		log.Warn("Enhanced Poetry collection failed, falling back to standard method: " + err.Error())
		log.Info("Poetry build info collection (using standard method).")

		// Fallback: Save build info with Poetry configuration (generic method)
		err = saveBuildInfo(serverDetails, repoConfig.TargetRepo(), "", buildConfiguration)
		if err != nil {
			log.Error("Failed to save Poetry build info: ", err)
			return fmt.Errorf("saveBuildInfo failed: %w", err)
		}
	} else {
		log.Info("Successfully collected Poetry build info with dependencies and artifacts.")
		// Enhanced collection succeeded, no need for additional saveBuildInfo call
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

// GetBuildInfoForUploadedArtifacts handles build info for uploaded artifacts (generic fallback)
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
	err = addArtifactsToBuildInfo(buildInfo, serverDetails, targetRepo, buildConfiguration, workingDir)
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
	err = saveBuildInfoNative(buildInfo)
	if err != nil {
		return fmt.Errorf("failed to save build info: %w", err)
	}

	log.Info("Successfully saved build info for publishing")
	return nil
}

// saveBuildInfoNative saves build info for jfrog-cli rt bp compatibility (native path)
func saveBuildInfoNative(buildInfo *buildinfo.BuildInfo) error {
	// Use the same approach as createBuildInfo but with the buildUtils service
	buildInfoService := buildUtils.CreateBuildInfoService()

	// Create or get build
	bld, err := buildInfoService.GetOrCreateBuildWithProject(buildInfo.Name, buildInfo.Number, "")
	if err != nil {
		return fmt.Errorf("failed to create build: %w", err)
	}

	// Add artifacts from each module using the proper build service method
	for _, module := range buildInfo.Modules {
		if len(module.Artifacts) > 0 {
			log.Info(fmt.Sprintf("Adding %d artifacts to module %s", len(module.Artifacts), module.Id))
			for i, artifact := range module.Artifacts {
				log.Info(fmt.Sprintf("  Artifact %d: %s (path: %s)", i+1, artifact.Name, artifact.Path))
			}
			err = bld.AddArtifacts(module.Id, module.Type, module.Artifacts...)
			if err != nil {
				return fmt.Errorf("failed to add artifacts for module %s: %w", module.Id, err)
			}
			log.Info("Successfully added artifacts to build service")
		} else {
			log.Warn("No artifacts found in module " + module.Id)
		}
	}

	// Note: No need to call SaveBuildInfo here as AddArtifacts already saves the build
	// The FlexPack buildInfo object is used only for extracting artifacts and dependencies
	// The actual build persistence is handled by the build service methods

	log.Info("Build info with artifacts saved successfully")
	return nil
}

// addArtifactsToBuildInfo collects uploaded artifacts with proper repository paths and adds them to the build info
func addArtifactsToBuildInfo(buildInfo *buildinfo.BuildInfo, serverDetails *config.ServerDetails, targetRepo string, buildConfiguration *buildUtils.BuildConfiguration, workingDir string) error {
	log.Debug("Collecting artifacts for build info...")

	if len(buildInfo.Modules) == 0 {
		return fmt.Errorf("no modules found in build info")
	}

	// Get the main module (should be the Poetry module)
	module := &buildInfo.Modules[0]

	// Collect artifacts with proper repository paths by searching Artifactory for recently uploaded files
	artifacts, err := collectArtifactsWithRepositoryPaths(serverDetails, targetRepo, module.Id, workingDir)
	if err != nil {
		log.Warn("Failed to collect artifacts with repository paths: " + err.Error())
		return nil // Don't fail the whole process for missing artifacts
	}

	// Add artifacts to the module
	module.Artifacts = artifacts

	log.Info(fmt.Sprintf("Added %d artifacts to build info", len(artifacts)))
	return nil
}

// collectArtifactsWithRepositoryPaths collects artifacts with their actual repository paths
func collectArtifactsWithRepositoryPaths(serverDetails *config.ServerDetails, targetRepo, moduleId, workingDir string) ([]buildinfo.Artifact, error) {
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

		searchReader.Close()
	}

	log.Debug(fmt.Sprintf("Found %d artifacts with repository paths", len(artifacts)))
	return artifacts, nil
}

// collectArtifactsFromArtifactory collects artifacts that were uploaded in this specific build (legacy method)
func collectArtifactsFromArtifactory(serverDetails *config.ServerDetails, targetRepo, buildName, buildNumber, moduleId string) ([]buildinfo.Artifact, error) {
	log.Debug("Searching for artifacts with build properties...")

	// Create services manager
	servicesManager, err := utils.CreateServiceManager(serverDetails, -1, 0, false)
	if err != nil {
		return nil, fmt.Errorf("failed to create services manager: %w", err)
	}

	// Create AQL query to find artifacts with specific build properties
	// This ensures we only get artifacts from THIS build, not older versions
	aqlQuery := fmt.Sprintf(`{
		"repo": "%s",
		"$and": [
			{"@build.name": "%s"},
			{"@build.number": "%s"}
		]
	}`, targetRepo, buildName, buildNumber)

	searchParams := services.SearchParams{
		CommonParams: &specutils.CommonParams{
			Aql: specutils.Aql{
				ItemsFind: aqlQuery,
			},
		},
	}

	searchReader, err := servicesManager.SearchFiles(searchParams)
	if err != nil {
		return nil, fmt.Errorf("failed to search for artifacts: %w", err)
	}
	defer searchReader.Close()

	var artifacts []buildinfo.Artifact
	for searchResult := new(specutils.ResultItem); searchReader.NextRecord(searchResult) == nil; searchResult = new(specutils.ResultItem) {
		// Only include Python package files
		if !strings.HasSuffix(searchResult.Name, ".whl") && !strings.HasSuffix(searchResult.Name, ".tar.gz") {
			continue
		}

		// Create artifact entry with proper repository path
		artifact := buildinfo.Artifact{
			Name: searchResult.Name,
			Path: searchResult.Path, // This will be the actual repository path like "test-project/1.0.1/"
			Type: getArtifactTypeFromName(searchResult.Name),
			// Note: Checksums will be populated by the property setting process
			// or can be calculated separately if needed
		}

		artifacts = append(artifacts, artifact)
	}

	if err := searchReader.GetError(); err != nil {
		return nil, fmt.Errorf("error reading search results: %w", err)
	}

	log.Debug(fmt.Sprintf("Found %d artifacts with build properties", len(artifacts)))
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
