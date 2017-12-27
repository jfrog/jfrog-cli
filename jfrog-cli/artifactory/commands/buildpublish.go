package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils"
	rtclientutils "github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services/utils"
	"sort"
	"strings"
	"encoding/json"
	"path/filepath"
	"errors"
	clientuils "github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils/buildinfo"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
	"fmt"
)

func BuildPublish(buildName, buildNumber string, flags *buildinfo.Flags, artDetails *config.ArtifactoryDetails) error {
	artAuth, err := artDetails.CreateArtAuthConfig()
	if err != nil {
		return err
	}
	flags.SetArtifactoryDetails(artAuth)

	buildInfo, err := createBuildInfoFromPartials(buildName, buildNumber, flags)
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

	return sendBuildInfo(buildName, buildNumber, buildInfo, flags)
}

func sendBuildInfo(buildName, buildNumber string, buildInfo *buildinfo.BuildInfo, flags *buildinfo.Flags) error {
	marshaledBuildInfo, err := json.Marshal(buildInfo)
	if errorutils.CheckError(err) != nil {
		return err
	}
	if flags.IsDryRun() {
		log.Output(clientuils.IndentJson(marshaledBuildInfo))
		return nil
	}
	httpClientsDetails := flags.GetArtifactoryDetails().CreateArtifactoryHttpClientDetails()
	rtclientutils.SetContentType("application/vnd.org.jfrog.artifactory+json", &httpClientsDetails.Headers)
	log.Info("Deploying build info...")
	resp, body, err := utils.PublishBuildInfo(flags.GetArtifactoryDetails().Url, marshaledBuildInfo, httpClientsDetails)
	if err != nil {
		return err
	}
	if resp.StatusCode != 204 {
		return errorutils.CheckError(errors.New("Artifactory response: " + resp.Status + "\n" + clientuils.IndentJson(body)))
	}

	log.Debug("Artifactory response:", resp.Status)
	log.Info("Build info successfully deployed. Browse it in Artifactory under " + flags.GetArtifactoryDetails().Url + "webapp/builds/" + buildName + "/" + buildNumber)
	if err = utils.RemoveBuildDir(buildName, buildNumber); err != nil {
		return err
	}
	return nil
}

func createBuildInfoFromPartials(buildName, buildNumber string, flags *buildinfo.Flags) (*buildinfo.BuildInfo, error) {
	partials, err := utils.ReadPartialBuildInfoFiles(buildName, buildNumber)
	if err != nil {
		return nil, err
	}
	sort.Sort(partials)

	buildInfo := buildinfo.New()
	buildInfo.Name = buildName
	buildInfo.Number = buildNumber
	buildGeneralDetails, err := utils.ReadBuildInfoGeneralDetails(buildName, buildNumber)
	if err != nil {
		return nil, err
	}
	buildInfo.Started = buildGeneralDetails.Timestamp.Format("2006-01-02T15:04:05.000-0700")
	modules, env, vcs, err := extractBuildInfoData(partials, createIncludeFilter(flags.EnvInclude), createExcludeFilter(flags.EnvExclude))
	if err != nil {
		return nil, err
	}
	if len(env) != 0 {
		buildInfo.Properties = env
	}
	buildInfo.ArtifactoryPrincipal = flags.GetArtifactoryDetails().GetUser()
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
	return buildInfo, nil
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

func addDependencyToPartialModule(dependency buildinfo.Dependencies, moduleId string, partialModules map[string]partialModule) {
	// init map if needed
	if partialModules[moduleId].dependencies == nil {
		partialModules[moduleId] =
			partialModule{artifacts: partialModules[moduleId].artifacts,
				dependencies: make(map[string]buildinfo.Dependencies)}
	}
	key := fmt.Sprintf("%s-%s-%s-%s", dependency.Id, dependency.Sha1, dependency.Md5, dependency.Scopes)
	partialModules[moduleId].dependencies[key] = dependency
}

func addArtifactToPartialModule(artifact buildinfo.Artifacts, moduleId string, partialModules map[string]partialModule) {
	// init map if needed
	if partialModules[moduleId].artifacts == nil {
		partialModules[moduleId] =
			partialModule{artifacts: make(map[string]buildinfo.Artifacts),
				dependencies: partialModules[moduleId].dependencies}
	}
	key := fmt.Sprintf("%s-%s-%s", artifact.Name, artifact.Sha1, artifact.Md5)
	partialModules[moduleId].artifacts[key] = artifact
}

func artifactsMapToList(artifactsMap map[string]buildinfo.Artifacts) []buildinfo.Artifacts {
	var artifacts []buildinfo.Artifacts
	for _, artifact := range artifactsMap {
		artifacts = append(artifacts, artifact)
	}
	return artifacts
}

func dependenciesMapToList(dependenciesMap map[string]buildinfo.Dependencies) []buildinfo.Dependencies {
	var dependencies []buildinfo.Dependencies
	for _, dependency := range dependenciesMap {
		dependencies = append(dependencies, dependency)
	}
	return dependencies
}

func createModule(moduleId string, artifacts []buildinfo.Artifacts, dependencies []buildinfo.Dependencies) *buildinfo.Module {
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
		Artifacts:    []buildinfo.Artifacts{},
		Dependencies: []buildinfo.Dependencies{},
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

type partialModule struct {
	artifacts    map[string]buildinfo.Artifacts
	dependencies map[string]buildinfo.Dependencies
}