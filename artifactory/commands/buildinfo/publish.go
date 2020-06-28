package buildinfo

import (
	"fmt"
	"github.com/jfrog/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"path/filepath"
	"sort"
	"strings"
)

const buildInfoPrefix = "buildInfo.env."

type BuildPublishCommand struct {
	buildConfiguration *utils.BuildConfiguration
	rtDetails          *config.ArtifactoryDetails
	config             *buildinfo.Configuration
}

func NewBuildPublishCommand() *BuildPublishCommand {
	return &BuildPublishCommand{}
}

func (bpc *BuildPublishCommand) SetConfig(config *buildinfo.Configuration) *BuildPublishCommand {
	bpc.config = config
	return bpc
}

func (bpc *BuildPublishCommand) SetRtDetails(rtDetails *config.ArtifactoryDetails) *BuildPublishCommand {
	bpc.rtDetails = rtDetails
	return bpc
}

func (bpc *BuildPublishCommand) SetBuildConfiguration(buildConfiguration *utils.BuildConfiguration) *BuildPublishCommand {
	bpc.buildConfiguration = buildConfiguration
	return bpc
}

func (bpc *BuildPublishCommand) CommandName() string {
	return "rt_build_publish"
}

func (bpc *BuildPublishCommand) RtDetails() (*config.ArtifactoryDetails, error) {
	return bpc.rtDetails, nil
}

func (bpc *BuildPublishCommand) Run() error {
	servicesManager, err := utils.CreateServiceManager(bpc.rtDetails, bpc.config.DryRun)
	if err != nil {
		return err
	}

	buildInfo, err := bpc.createBuildInfoFromPartials()
	if err != nil {
		return err
	}

	generatedBuildsInfo, err := utils.GetGeneratedBuildsInfo(bpc.buildConfiguration.BuildName, bpc.buildConfiguration.BuildNumber)
	if err != nil {
		return err
	}

	for _, v := range generatedBuildsInfo {
		buildInfo.Append(v)
	}

	if err = servicesManager.PublishBuildInfo(buildInfo); err != nil {
		return err
	}

	if !bpc.config.DryRun {
		return utils.RemoveBuildDir(bpc.buildConfiguration.BuildName, bpc.buildConfiguration.BuildNumber)
	}
	return nil
}

func (bpc *BuildPublishCommand) createBuildInfoFromPartials() (*buildinfo.BuildInfo, error) {
	buildName := bpc.buildConfiguration.BuildName
	buildNumber := bpc.buildConfiguration.BuildNumber
	partials, err := utils.ReadPartialBuildInfoFiles(buildName, buildNumber)
	if err != nil {
		return nil, err
	}
	sort.Sort(partials)

	buildInfo := buildinfo.New()
	buildInfo.SetAgentName(cliutils.ClientAgent)
	buildInfo.SetAgentVersion(cliutils.GetVersion())
	buildInfo.SetBuildAgentVersion(cliutils.GetVersion())
	buildInfo.SetArtifactoryPluginVersion(cliutils.GetUserAgent())
	buildInfo.Name = buildName
	buildInfo.Number = buildNumber
	buildGeneralDetails, err := utils.ReadBuildInfoGeneralDetails(buildName, buildNumber)
	if err != nil {
		return nil, err
	}
	buildInfo.Started = buildGeneralDetails.Timestamp.Format("2006-01-02T15:04:05.000-0700")
	modules, env, vcs, issues, err := extractBuildInfoData(partials, createIncludeFilter(bpc.config.EnvInclude), createExcludeFilter(bpc.config.EnvExclude))
	if err != nil {
		return nil, err
	}
	if len(env) != 0 {
		buildInfo.Properties = env
	}
	buildInfo.ArtifactoryPrincipal = bpc.rtDetails.User
	buildInfo.BuildUrl = bpc.config.BuildUrl
	if vcs != (buildinfo.Vcs{}) {
		buildInfo.Revision = vcs.Revision
		buildInfo.Url = vcs.Url
	}
	// Check for Tracker as it must be set
	if issues.Tracker != nil && issues.Tracker.Name != "" {
		buildInfo.Issues = &issues
	}
	for _, module := range modules {
		if module.Id == "" {
			module.Id = buildName
		}
		buildInfo.Modules = append(buildInfo.Modules, module)
	}
	return buildInfo, nil
}

func extractBuildInfoData(partials buildinfo.Partials, includeFilter, excludeFilter filterFunc) ([]buildinfo.Module, buildinfo.Env, buildinfo.Vcs, buildinfo.Issues, error) {
	var vcs buildinfo.Vcs
	var issues buildinfo.Issues
	env := make(map[string]string)
	partialModules := make(map[string]partialModule)
	issuesMap := make(map[string]*buildinfo.AffectedIssue)
	for _, partial := range partials {
		switch {
		case partial.Artifacts != nil:
			for _, artifact := range partial.Artifacts {
				addArtifactToPartialModule(artifact, partial.ModuleId, partialModules)
			}
		case partial.Dependencies != nil:
			for _, dependency := range partial.Dependencies {
				addDependencyToPartialModule(dependency, partial.ModuleId, partialModules)
			}
		case partial.Vcs != nil:
			vcs = *partial.Vcs
			if partial.Issues == nil {
				continue
			}
			// Collect issues.
			issues.Tracker = partial.Issues.Tracker
			issues.AggregateBuildIssues = partial.Issues.AggregateBuildIssues
			issues.AggregationBuildStatus = partial.Issues.AggregationBuildStatus
			// If affected issues exist, add them to issues map
			if partial.Issues.AffectedIssues != nil {
				for i, issue := range partial.Issues.AffectedIssues {
					issuesMap[issue.Key] = &partial.Issues.AffectedIssues[i]
				}
			}
		case partial.Env != nil:
			envAfterIncludeFilter, e := includeFilter(partial.Env)
			if errorutils.CheckError(e) != nil {
				return partialModulesToModules(partialModules), env, vcs, issues, e
			}
			envAfterExcludeFilter, e := excludeFilter(envAfterIncludeFilter)
			if errorutils.CheckError(e) != nil {
				return partialModulesToModules(partialModules), env, vcs, issues, e
			}
			for k, v := range envAfterExcludeFilter {
				env[k] = v
			}
		}
	}
	return partialModulesToModules(partialModules), env, vcs, issuesMapToArray(issues, issuesMap), nil
}

func partialModulesToModules(partialModules map[string]partialModule) []buildinfo.Module {
	var modules []buildinfo.Module
	for moduleId, singlePartialModule := range partialModules {
		moduleArtifacts := artifactsMapToList(singlePartialModule.artifacts)
		moduleDependencies := dependenciesMapToList(singlePartialModule.dependencies)
		modules = append(modules, *createModule(moduleId, moduleArtifacts, moduleDependencies))
	}
	return modules
}

func issuesMapToArray(issues buildinfo.Issues, issuesMap map[string]*buildinfo.AffectedIssue) buildinfo.Issues {
	for _, issue := range issuesMap {
		issues.AffectedIssues = append(issues.AffectedIssues, *issue)
	}
	return issues
}

func addDependencyToPartialModule(dependency buildinfo.Dependency, moduleId string, partialModules map[string]partialModule) {
	// init map if needed
	if partialModules[moduleId].dependencies == nil {
		partialModules[moduleId] =
			partialModule{artifacts: partialModules[moduleId].artifacts,
				dependencies: make(map[string]buildinfo.Dependency)}
	}
	key := fmt.Sprintf("%s-%s-%s-%s", dependency.Id, dependency.Sha1, dependency.Md5, dependency.Scopes)
	partialModules[moduleId].dependencies[key] = dependency
}

func addArtifactToPartialModule(artifact buildinfo.Artifact, moduleId string, partialModules map[string]partialModule) {
	// init map if needed
	if partialModules[moduleId].artifacts == nil {
		partialModules[moduleId] =
			partialModule{artifacts: make(map[string]buildinfo.Artifact),
				dependencies: partialModules[moduleId].dependencies}
	}
	key := fmt.Sprintf("%s-%s-%s", artifact.Name, artifact.Sha1, artifact.Md5)
	partialModules[moduleId].artifacts[key] = artifact
}

func artifactsMapToList(artifactsMap map[string]buildinfo.Artifact) []buildinfo.Artifact {
	var artifacts []buildinfo.Artifact
	for _, artifact := range artifactsMap {
		artifacts = append(artifacts, artifact)
	}
	return artifacts
}

func dependenciesMapToList(dependenciesMap map[string]buildinfo.Dependency) []buildinfo.Dependency {
	var dependencies []buildinfo.Dependency
	for _, dependency := range dependenciesMap {
		dependencies = append(dependencies, dependency)
	}
	return dependencies
}

func createModule(moduleId string, artifacts []buildinfo.Artifact, dependencies []buildinfo.Dependency) *buildinfo.Module {
	module := createDefaultModule(moduleId)
	if artifacts != nil && len(artifacts) > 0 {
		module.Artifacts = append(module.Artifacts, artifacts...)
	}
	if dependencies != nil && len(dependencies) > 0 {
		module.Dependencies = append(module.Dependencies, dependencies...)
	}
	return module
}

func createDefaultModule(moduleId string) *buildinfo.Module {
	return &buildinfo.Module{
		Id:           moduleId,
		Properties:   map[string][]string{},
		Artifacts:    []buildinfo.Artifact{},
		Dependencies: []buildinfo.Dependency{},
	}
}

type filterFunc func(map[string]string) (map[string]string, error)

func createIncludeFilter(pattern string) filterFunc {
	includePattern := strings.Split(pattern, ";")
	return func(tempMap map[string]string) (map[string]string, error) {
		result := make(map[string]string)
		for k, v := range tempMap {
			for _, filterPattern := range includePattern {
				matched, err := filepath.Match(strings.ToLower(filterPattern), strings.ToLower(strings.TrimPrefix(k, buildInfoPrefix)))
				if errorutils.CheckError(err) != nil {
					return nil, err
				}
				if matched {
					result[k] = v
					break
				}
			}
		}
		return result, nil
	}
}

func createExcludeFilter(pattern string) filterFunc {
	excludePattern := strings.Split(pattern, ";")
	return func(tempMap map[string]string) (map[string]string, error) {
		result := make(map[string]string)
		for k, v := range tempMap {
			include := true
			for _, filterPattern := range excludePattern {
				matched, err := filepath.Match(strings.ToLower(filterPattern), strings.ToLower(strings.TrimPrefix(k, buildInfoPrefix)))
				if errorutils.CheckError(err) != nil {
					return nil, err
				}
				if matched {
					include = false
					break
				}
			}
			if include {
				result[k] = v
			}
		}
		return result, nil
	}
}

type partialModule struct {
	artifacts    map[string]buildinfo.Artifact
	dependencies map[string]buildinfo.Dependency
}
