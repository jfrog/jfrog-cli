package buildinfo

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jfrog/build-info-go/build"
	buildinfo "github.com/jfrog/build-info-go/entities"
	"github.com/jfrog/build-info-go/flexpack"
	"github.com/jfrog/gofrog/crypto"
	"encoding/json"
	"io"

	artutils "github.com/jfrog/jfrog-cli-artifactory/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	buildUtils "github.com/jfrog/jfrog-cli-core/v2/common/build"
	"github.com/jfrog/jfrog-cli-core/v2/common/project"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	specutils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/utils/io/content"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/spf13/viper"
)

const (
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
	case "uv":
		// Pass "sync" as cmdName so dependency enrichment runs (same path as jf uv sync).
		return GetUvBuildInfo(workingDir, buildConfiguration, "", "sync", nil)
	case "poetry":
		return GetPoetryBuildInfo(workingDir, buildConfiguration, "") // Empty deployer repo - will use from pyproject.toml
	case "mvn", "maven":
		// Maven FlexPack is handled directly in jfrog-cli-artifactory Maven command
		return GetBuildInfoForUploadedArtifacts("", buildConfiguration)
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
func GetPoetryBuildInfo(workingDir string, buildConfiguration *buildUtils.BuildConfiguration, deployerRepo string) error {
	log.Debug("Collecting Poetry build info from directory: " + workingDir)
	log.Debug("Deployer repository: " + deployerRepo)

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

	// Use deployerRepo if provided, otherwise use resolver repo from config
	artifactRepo := deployerRepo
	if artifactRepo == "" {
		artifactRepo = repoConfig.TargetRepo()
	}
	log.Info(fmt.Sprintf("Using repository for artifacts: %s", artifactRepo))

	err = collectPoetryBuildInfo(workingDir, buildName, buildNumber, serverDetails, repoConfig.TargetRepo(), artifactRepo, buildConfiguration)
	if err != nil {
		log.Warn("Enhanced Poetry collection failed, falling back to standard method: " + err.Error())
		err = saveBuildInfo(serverDetails, artifactRepo, "", buildConfiguration)
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

// GetUvBuildInfo collects build info for UV projects using the FlexPack native implementation.
// cmdName is the uv subcommand that was run (sync, build, publish, etc.).
// serverDetails is the Artifactory server to use; if nil, the default configured server is used.
func GetUvBuildInfo(workingDir string, buildConfiguration *buildUtils.BuildConfiguration, deployerRepo, cmdName string, serverDetails *config.ServerDetails) error {
	log.Debug(fmt.Sprintf("Collecting UV build info for command '%s' in: %s", cmdName, workingDir))

	buildName, err := buildConfiguration.GetBuildName()
	if err != nil {
		return fmt.Errorf("GetBuildName failed: %w", err)
	}
	buildNumber, err := buildConfiguration.GetBuildNumber()
	if err != nil {
		return fmt.Errorf("GetBuildNumber failed: %w", err)
	}

	// Collect dependencies from pyproject.toml + uv.lock via FlexPack
	uvConfig := flexpack.UvConfig{
		WorkingDirectory:       workingDir,
		IncludeDevDependencies: false,
	}
	collector, err := flexpack.NewUvFlexPack(uvConfig)
	if err != nil {
		return fmt.Errorf("failed to create UV FlexPack collector: %w", err)
	}

	bi, err := collector.CollectBuildInfo(buildName, buildNumber)
	if err != nil {
		return fmt.Errorf("failed to collect UV build info: %w", err)
	}

	// Apply --module override: if the user passed --module, replace the auto-detected
	// module ID (name:version from pyproject.toml) with the user-supplied value.
	if customModule := buildConfiguration.GetModule(); customModule != "" && len(bi.Modules) > 0 {
		bi.Modules[0].Id = customModule
	}

	// For sync/install/lock: enrich dependency checksums from Artifactory.
	// uv.lock only has sha256. Artifactory needs sha1 + md5 to link a dependency to its
	// repo path in the build browser. We search each dependency by filename in the
	// configured index repo (from pyproject.toml [[tool.uv.index]]) to fetch sha1+md5.
	// This is the same mechanism pip/poetry use via UpdateDepsChecksumInfo.
	switch cmdName {
	case "sync", "install", "lock", "add", "remove", "run":
		if len(bi.Modules) > 0 && len(bi.Modules[0].Dependencies) > 0 {
			if repoKey := uvResolverRepoFromToml(workingDir); repoKey != "" {
				sd := serverDetails
				if sd == nil {
					sd, _ = config.GetDefaultServerConf()
				}
				if sd != nil {
					// Warn when the jf server config host doesn't match the index URL host.
					// In this case checksum enrichment (sha1/md5) will fail because the repo
					// lives on a different Artifactory instance than the configured server.
					// Fix: pass --server-id pointing to the instance that hosts the uv-virtual repo.
					if indexURL := uvIndexURLFromToml(workingDir); indexURL != "" && !uvServerHostMatches(indexURL, sd.ArtifactoryUrl) {
						log.Warn(fmt.Sprintf(
							"UV build-info: jf server config host (%s) differs from index URL host (%s) — "+
								"dependency checksum enrichment (sha1/md5) will be skipped. "+
								"Use --server-id to specify the Artifactory instance that hosts your uv packages.",
							uvHostOf(sd.ArtifactoryUrl), uvHostOf(indexURL)))
					}
					enrichUvDepsFromArtifactory(bi.Modules[0].Dependencies, repoKey, sd)
				}
			}
		}
	}

	// For build and publish commands, also collect artifacts from dist/
	switch cmdName {
	case "build":
		// Scan dist/ and attach artifacts with local checksums only
		if artifacts, scanErr := collectPythonDistArtifacts(workingDir); scanErr == nil && len(artifacts) > 0 {
			if len(bi.Modules) > 0 {
				bi.Modules[0].Artifacts = artifacts
				log.Info(fmt.Sprintf("Collected %d artifact(s) from dist/", len(artifacts)))
			}
		} else if scanErr != nil {
			log.Warn("Could not scan dist/ for artifacts: " + scanErr.Error())
		}
	case "publish":
		// Scan dist/, look up Artifactory repo paths, and set build properties on uploaded files.
		// deployerRepo may be a full URL (--publish-url https://.../api/pypi/repo) or a bare repo key.
		repoKey := extractRepoKeyFromURL(deployerRepo)
		sd := serverDetails
		if sd == nil {
			var sdErr error
			sd, sdErr = config.GetDefaultServerConf()
			if sdErr != nil {
				log.Warn("Could not load server config for artifact lookup: " + sdErr.Error())
				sd = nil
			}
		}
		if sd == nil {
			// Fall back to local artifacts only
			if artifacts, scanErr := collectPythonDistArtifacts(workingDir); scanErr == nil && len(bi.Modules) > 0 {
				bi.Modules[0].Artifacts = artifacts
			}
			break
		}
		if repoKey != "" {
			// Warn when the jf server config host doesn't match the publish URL host.
			// In this case artifact lookup and property setting will fail because the
			// artifacts live on a different Artifactory instance than the configured server.
			// Fix: pass --server-id pointing to the instance that hosts the uv-local repo.
			if deployerRepo != "" && !uvServerHostMatches(deployerRepo, sd.ArtifactoryUrl) {
				log.Warn(fmt.Sprintf(
					"UV build-info: jf server config host (%s) differs from publish URL host (%s) — "+
						"artifact lookup and build property setting will be skipped. "+
						"Use --server-id to specify the Artifactory instance that hosts your uv packages.",
					uvHostOf(sd.ArtifactoryUrl), uvHostOf(deployerRepo)))
			} else {
				if artErr := addArtifactsToBuildInfo(bi, sd, repoKey, workingDir); artErr != nil {
					log.Warn("Could not look up artifact repo paths, using local checksums: " + artErr.Error())
					if artifacts, scanErr := collectPythonDistArtifacts(workingDir); scanErr == nil && len(bi.Modules) > 0 {
						bi.Modules[0].Artifacts = artifacts
					}
				}
				// Set build.name / build.number / build.timestamp on the files in Artifactory
				// so the build browser can link artifacts to this build.
				if propErr := setPythonBuildProperties(sd, repoKey, buildName, buildNumber, buildConfiguration.GetProject(), bi); propErr != nil {
					log.Warn("Failed to set build properties on artifacts: " + propErr.Error())
				}
			}
		} else {
			if artifacts, scanErr := collectPythonDistArtifacts(workingDir); scanErr == nil && len(artifacts) > 0 {
				if len(bi.Modules) > 0 {
					bi.Modules[0].Artifacts = artifacts
					log.Info(fmt.Sprintf("Collected %d artifact(s) from dist/ (no deployer repo set)", len(artifacts)))
				}
			}
		}
	}

	if err = saveBuildInfoNative(bi, buildConfiguration); err != nil {
		return fmt.Errorf("failed to save UV build info: %w", err)
	}

	log.Info(fmt.Sprintf("UV build info collected. Use 'jf rt bp %s %s' to publish.", buildName, buildNumber))
	return nil
}

// uvResolverRepoFromToml reads the first [[tool.uv.index]] URL from pyproject.toml in workingDir
// and extracts the Artifactory repo key from it.
// e.g. "https://host/artifactory/api/pypi/agrasth-uv-local/simple" → "agrasth-uv-local"
// uvIndexURLFromToml returns the raw URL of the first [[tool.uv.index]] entry.
func uvIndexURLFromToml(workingDir string) string {
	pyprojectPath := filepath.Join(workingDir, "pyproject.toml")
	v := viper.New()
	v.SetConfigType("toml")
	v.SetConfigFile(pyprojectPath)
	if err := v.ReadInConfig(); err != nil {
		return ""
	}
	raw := v.Get("tool.uv.index")
	if raw == nil {
		return ""
	}
	if idxSlice, ok := raw.([]interface{}); ok && len(idxSlice) > 0 {
		if idxMap, ok := idxSlice[0].(map[string]interface{}); ok {
			if urlVal, ok := idxMap["url"].(string); ok {
				return urlVal
			}
		}
	}
	return ""
}

// uvHostOf returns the hostname from rawURL, or empty string on error.
func uvHostOf(rawURL string) string {
	if rawURL == "" {
		return ""
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return parsed.Hostname()
}

// uvServerHostMatches returns true when rawURL's hostname matches the Artifactory server URL hostname.
func uvServerHostMatches(rawURL, serverURL string) bool {
	h := uvHostOf(rawURL)
	return h != "" && h == uvHostOf(serverURL)
}

func uvResolverRepoFromToml(workingDir string) string {
	type UvIndex struct {
		URL string `mapstructure:"url"`
	}
	pyprojectPath := filepath.Join(workingDir, "pyproject.toml")
	v := viper.New()
	v.SetConfigType("toml")
	v.SetConfigFile(pyprojectPath)
	if err := v.ReadInConfig(); err != nil {
		return ""
	}
	raw := v.Get("tool.uv.index")
	if raw == nil {
		return ""
	}
	if idxSlice, ok := raw.([]interface{}); ok && len(idxSlice) > 0 {
		if idxMap, ok := idxSlice[0].(map[string]interface{}); ok {
			if urlVal, ok := idxMap["url"].(string); ok {
				return extractRepoKeyFromURL(urlVal)
			}
		}
	}
	return ""
}

// enrichUvDepsFromArtifactory searches each dependency by filename in the given Artifactory repo
// and updates the dependency's sha1 and md5 checksums. Artifactory uses sha1+md5 (not sha256)
// to link a dependency to its repo path in the build browser.
// This is the same mechanism used by pip/poetry via UpdateDepsChecksumInfo.
func enrichUvDepsFromArtifactory(deps []buildinfo.Dependency, repoKey string, serverDetails *config.ServerDetails) {
	servicesManager, err := utils.CreateServiceManager(serverDetails, -1, 0, false)
	if err != nil {
		log.Warn("Could not create services manager for dependency enrichment: " + err.Error())
		return
	}
	// If the repo is a remote repo, Artifactory stores cached files under "<repo>-cache"
	searchRepo, err := utils.GetRepoNameForDependenciesSearch(repoKey, servicesManager)
	if err != nil {
		log.Warn("Could not resolve repo for dependency search, using as-is: " + err.Error())
		searchRepo = repoKey
	}

	enriched := 0
	for i, dep := range deps {
		if dep.Id == "" {
			continue
		}
		// dep.Id is the wheel/sdist filename (e.g. "certifi-2026.2.25-py3-none-any.whl")
		aqlQuery := specutils.CreateAqlQueryForPypi(searchRepo, dep.Id)
		stream, err := servicesManager.Aql(aqlQuery)
		if err != nil {
			log.Debug(fmt.Sprintf("AQL search failed for %s: %v", dep.Id, err))
			continue
		}
		result, err := io.ReadAll(stream)
		stream.Close()
		if err != nil {
			continue
		}
		var aql struct {
			Results []struct {
				Actual_Sha1 string `json:"actual_sha1"`
				Actual_Md5  string `json:"actual_md5"`
				Sha256      string `json:"sha256"`
			} `json:"results"`
		}
		if err := json.Unmarshal(result, &aql); err != nil || len(aql.Results) == 0 {
			log.Debug(fmt.Sprintf("Dependency %s not found in repo %s", dep.Id, searchRepo))
			continue
		}
		r := aql.Results[0]
		deps[i].Checksum.Sha1 = r.Actual_Sha1
		deps[i].Checksum.Md5 = r.Actual_Md5
		if r.Sha256 != "" {
			deps[i].Checksum.Sha256 = r.Sha256
		}
		enriched++
		log.Debug(fmt.Sprintf("Enriched %s with sha1=%s md5=%s", dep.Id, r.Actual_Sha1, r.Actual_Md5))
	}
	if enriched > 0 {
		log.Info(fmt.Sprintf("Enriched %d/%d UV dependencies with Artifactory checksums (repo: %s)", enriched, len(deps), searchRepo))
	}
}

// extractRepoKeyFromURL returns just the repo key from a full Artifactory URL or a bare key.
// e.g. "https://host/artifactory/api/pypi/my-repo" → "my-repo"
//      "https://host/artifactory/api/pypi/my-repo/simple" → "my-repo"
//      "my-repo" → "my-repo"
func extractRepoKeyFromURL(repoOrURL string) string {
	if repoOrURL == "" {
		return ""
	}
	// If it doesn't look like a URL, treat it as a bare repo key
	if !strings.HasPrefix(repoOrURL, "http://") && !strings.HasPrefix(repoOrURL, "https://") {
		return repoOrURL
	}
	parsed, err := url.Parse(repoOrURL)
	if err != nil {
		return repoOrURL
	}
	// Strip trailing path segments like /simple
	segments := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	// Find the segment after "api/pypi" or "api/npm" etc., or just take the last meaningful segment
	for i, seg := range segments {
		if seg == "api" && i+2 < len(segments) {
			return segments[i+2]
		}
	}
	// Fallback: last non-empty segment
	for i := len(segments) - 1; i >= 0; i-- {
		if segments[i] != "" && segments[i] != "simple" {
			return segments[i]
		}
	}
	return repoOrURL
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

	// In native mode, completely ignore YAML and infer from pyproject.toml
	if artutils.ShouldRunNative("") {
		log.Info("Native mode enabled: inferring Poetry config from pyproject.toml")
		return inferPoetryConfigFromToml(projectType)
	}

	prefix := project.ProjectConfigDeployerPrefix

	// Handle invalid project types gracefully
	if projectType < 0 {
		return nil, fmt.Errorf("invalid project type")
	}

	configFilePath, exists, err := project.GetProjectConfFilePath(projectType)
	if !exists {
		log.Warn("Project configuration file not found, attempting to infer from pyproject.toml...")
		return inferPoetryConfigFromToml(projectType) // Fallback to TOML
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
func collectPoetryBuildInfo(workingDir, buildName, buildNumber string, serverDetails *config.ServerDetails, _ string, artifactRepo string, buildConfiguration *buildUtils.BuildConfiguration) error {
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
	err = addArtifactsToBuildInfo(buildInfo, serverDetails, artifactRepo, workingDir)
	if err != nil {
		log.Warn("Failed to add artifacts to build info: " + err.Error())
		// Continue anyway - dependencies are more important than artifacts for now
	}

	// Then set build properties on uploaded artifacts (this tags them with build info)
	err = setPythonBuildProperties(serverDetails, artifactRepo, buildName, buildNumber, buildConfiguration.GetProject(), buildInfo)
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
	localArtifacts, err := collectPythonDistArtifacts(workingDir)
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
			artifact := buildinfo.Artifact{
				Name:     searchResult.Name,
				Path:     searchResult.Path, // Directory only: "package-name/version"
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

// collectPythonDistArtifacts collects artifacts from the dist/ directory (legacy method)
func collectPythonDistArtifacts(workingDir string) ([]buildinfo.Artifact, error) {
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

// setPythonBuildProperties sets build properties on uploaded Python dist artifacts
// This ensures artifacts are tagged with build.name, build.number, and build.timestamp
// just like npm, maven, and gradle package managers do
func setPythonBuildProperties(serverDetails *config.ServerDetails, targetRepo, buildName, buildNumber, project string, buildInfo *buildinfo.BuildInfo) error {
	log.Debug("Setting build properties on Python dist artifacts...")

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
		log.Debug(fmt.Sprintf("Setting properties on: %s/%s/%s", targetRepo, artifact.Path, artifact.Name))

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

	log.Info(fmt.Sprintf("Successfully set build properties on %d artifacts", len(module.Artifacts)))
	return nil
}

// inferPoetryConfigFromToml reads pyproject.toml and infers repository configuration
func inferPoetryConfigFromToml(_ project.ProjectType) (*project.RepositoryConfig, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	tomlPath := filepath.Join(workingDir, "pyproject.toml")
	viper.SetConfigType("toml")
	viper.SetConfigFile(tomlPath)

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read pyproject.toml: %w", err)
	}

	sources := viper.Get("tool.poetry.source")
	if sources == nil {
		return nil, fmt.Errorf("no Poetry sources found in pyproject.toml")
	}

	// Get list of configured servers from jf config
	cmd := exec.Command("jf", "config", "show")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get jf config: %w", err)
	}

	sourcesList, ok := sources.([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid Poetry sources format in pyproject.toml")
	}

	for _, source := range sourcesList {
		sourceMap, ok := source.(map[string]interface{})
		if !ok {
			continue
		}

		sourceName, _ := sourceMap["name"].(string)
		sourceURL, _ := sourceMap["url"].(string)

		if sourceURL == "" {
			continue
		}

		// Parse the URL to extract host
		parsedURL, err := url.Parse(sourceURL)
		if err != nil {
			continue
		}

		// Check if this URL matches any configured server
		serverID := ""
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "Server ID:") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) > 1 {
					serverID = strings.TrimSpace(parts[1])
				}
			}
			if strings.Contains(line, "JFrog Platform URL:") && serverID != "" {
				// Extract URL after "JFrog Platform URL: " prefix
				parts := strings.SplitN(line, "JFrog Platform URL:", 2)
				if len(parts) < 2 {
					continue
				}
				configURL := strings.TrimSpace(parts[1])
				if strings.Contains(sourceURL, configURL) || strings.Contains(configURL, parsedURL.Host) {
					log.Info(fmt.Sprintf("Matched source '%s' to server '%s'", sourceName, serverID))

					// Extract repository name from URL
					repoName := extractRepoNameFromPypiURL(sourceURL)
					if repoName == "" {
						repoName = sourceName // Fallback to source name
					}

					log.Info(fmt.Sprintf("Inferred repository: %s", repoName))

					// Get server details from config
					serverDetails, err := config.GetSpecificConfig(serverID, false, true)
					if err != nil {
						return nil, fmt.Errorf("failed to get server details for %s: %w", serverID, err)
					}

					// Create and return repository config
					repoConfig := &project.RepositoryConfig{}
					repoConfig.SetServerDetails(serverDetails)
					repoConfig.SetTargetRepo(repoName)

					log.Info("Successfully inferred Poetry config from pyproject.toml (native mode)")
					return repoConfig, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("no matching JFrog server found for Poetry sources in pyproject.toml")
}

// extractRepoNameFromPypiURL extracts repository name from PyPI-format Artifactory URL
// Example: https://server.com/artifactory/api/pypi/poetry-local/simple -> poetry-local
func extractRepoNameFromPypiURL(urlStr string) string {
	// Remove trailing /simple or /simple/
	urlStr = strings.TrimSuffix(urlStr, "/")
	urlStr = strings.TrimSuffix(urlStr, "/simple")
	urlStr = strings.TrimSuffix(urlStr, "/")

	// Split by / and find "api/pypi/" pattern
	parts := strings.Split(urlStr, "/")
	for i := 0; i < len(parts)-1; i++ {
		if parts[i] == "api" && i+1 < len(parts) && parts[i+1] == "pypi" {
			// Repo name is after "pypi"
			if i+2 < len(parts) {
				return parts[i+2]
			}
		}
	}

	// Fallback: take the last meaningful part
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}

	return ""
}
