package buildinfo

import (
	"fmt"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/config"
        "github.com/jfrog/jfrog-cli-go/jfrog-client/artifactory"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/artifactory/buildinfo"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/artifactory/services"
	rtclientutils "github.com/jfrog/jfrog-cli-go/jfrog-client/artifactory/services/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/errorutils"
	"path/filepath"
	"sort"
	"strings"
)

func Publish(buildName, buildNumber string, config *buildinfo.Configuration, artDetails *config.ArtifactoryDetails) error {
	servicesManager, err := utils.CreateServiceManager(artDetails, config.DryRun)
	if err != nil {
		return err
	}

	buildInfo, partials, err := createBuildInfoFromPartials(buildName, buildNumber, config, artDetails)
	if err != nil {
		return err
	}

	err = setBuildInfoPropertiesForArtifacts(servicesManager, buildInfo, partials)
	if err != nil {
		return err
	}

	generatedBuildsInfo, err := utils.GetGeneratedBuildsInfo(buildName, buildNumber)
	if err != nil {
		return err
	}

	for _, v := range generatedBuildsInfo {
		buildInfo.Append(v)
	}

	if err = servicesManager.PublishBuildInfo(buildInfo); err != nil {
		return err
	}

	if err = utils.RemoveBuildDir(buildName, buildNumber); err != nil {
		return err
	}
	return nil
}

func createBuildInfoFromPartials(buildName, buildNumber string, config *buildinfo.Configuration, artDetails *config.ArtifactoryDetails) (*buildinfo.BuildInfo, buildinfo.Partials, error) {
	partials, err := utils.ReadPartialBuildInfoFiles(buildName, buildNumber)
	if err != nil {
		return nil, nil, err
	}
	sort.Sort(partials)

	buildInfo := buildinfo.New()
	buildInfo.SetAgentName(cliutils.ClientAgent)
	buildInfo.SetAgentVersion(cliutils.GetVersion())
	buildInfo.SetBuildAgentVersion(cliutils.GetVersion())
	buildInfo.Name = buildName
	buildInfo.Number = buildNumber
	buildGeneralDetails, err := utils.ReadBuildInfoGeneralDetails(buildName, buildNumber)
	if err != nil {
		return nil, nil, err
	}
	buildInfo.Started = buildGeneralDetails.Timestamp.Format("2006-01-02T15:04:05.000-0700")
	modules, env, vcs, err := extractBuildInfoData(partials, createIncludeFilter(config.EnvInclude), createExcludeFilter(config.EnvExclude))
	if err != nil {
		return nil, nil, err
	}
	if len(env) != 0 {
		buildInfo.Properties = env
	}
	buildInfo.ArtifactoryPrincipal = artDetails.User
	buildInfo.BuildUrl = config.BuildUrl
	if vcs != (buildinfo.Vcs{}) {
		buildInfo.Revision = vcs.Revision
		buildInfo.Url = vcs.Url
	}
	for _, module := range modules {
		if module.Id == "" {
			module.Id = buildName
		}
		buildInfo.Modules = append(buildInfo.Modules, module)
	}
	return buildInfo, partials, nil
}

func extractBuildInfoData(partials buildinfo.Partials, includeFilter, excludeFilter filterFunc) ([]buildinfo.Module, buildinfo.Env, buildinfo.Vcs, error) {
	var vcs buildinfo.Vcs
	env := make(map[string]string)
	partialModules := make(map[string]partialModule)
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
		case partial.Env != nil:
			envAfterIncludeFilter, e := includeFilter(partial.Env)
			if errorutils.CheckError(e) != nil {
				return partialModulesToModules(partialModules), env, vcs, e
			}
			envAfterExcludeFilter, e := excludeFilter(envAfterIncludeFilter)
			if errorutils.CheckError(e) != nil {
				return partialModulesToModules(partialModules), env, vcs, e
			}
			for k, v := range envAfterExcludeFilter {
				env[k] = v
			}
		}
	}
	return partialModulesToModules(partialModules), env, vcs, nil
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

func addArtifactToPartialModule(internalArtifact buildinfo.InternalArtifact, moduleId string, partialModules map[string]partialModule) {
	artifact := internalArtifact.ToArtifact()

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
				matched, err := filepath.Match(strings.ToLower(filterPattern), strings.ToLower(k))
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
				matched, err := filepath.Match(strings.ToLower(filterPattern), strings.ToLower(k))
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

func setBuildInfoPropertiesForArtifacts(servicesManager *artifactory.ArtifactoryServicesManager, buildInfo *buildinfo.BuildInfo, partials buildinfo.Partials) error {
	var resultItems []rtclientutils.ResultItem
	for _, partial := range partials {
		switch {
		case partial.Artifacts != nil:
			for _, artifact := range partial.Artifacts {
				resultItems = append(resultItems, rtclientutils.ResultItem{Path: artifact.Path})
			}
		}
	}

	props := "build.name=" + buildInfo.Name + ";build.number=" + buildInfo.Number
	_, err := servicesManager.SetProps(&services.SetPropsParamsImpl{Items: resultItems, Props: props})
	return err
}

type partialModule struct {
	artifacts    map[string]buildinfo.Artifact
	dependencies map[string]buildinfo.Dependency
}
