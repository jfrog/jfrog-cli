package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils"
	rtclientutils "github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services/utils"
	"sort"
	"fmt"
	"strings"
	"encoding/json"
	"path/filepath"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/cliutils"
	"errors"
	clientuils "github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils/buildinfo"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"
)

func BuildPublish(buildName, buildNumber string, flags *buildinfo.Flags, artDetails *config.ArtifactoryDetails) error {
	artAuth := artDetails.CreateArtAuthConfig()
	flags.SetArtifactoryDetails(artAuth)

	buildInfo, err := createGenericBuildInfo(buildName, buildNumber, flags)
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
		fmt.Println(clientuils.IndentJson(marshaledBuildInfo))
		return nil
	}
	httpClientsDetails := flags.GetArtifactoryDetails().CreateArtifactoryHttpClientDetails()
	rtclientutils.SetContentType("application/vnd.org.jfrog.artifactory+json", &httpClientsDetails.Headers)
	cliutils.CliLogger.Info("Deploying build info...")
	resp, body, err := utils.PublishBuildInfo(flags.GetArtifactoryDetails().Url, marshaledBuildInfo, httpClientsDetails)
	if err != nil {
		return err
	}
	if resp.StatusCode != 204 {
		return errorutils.CheckError(errors.New("Artifactory response: " + resp.Status + "\n" + clientuils.IndentJson(body)))
	}

	cliutils.CliLogger.Debug("Artifactory response:", resp.Status)
	cliutils.CliLogger.Info("Build info successfully deployed. Browse it in Artifactory under " + flags.GetArtifactoryDetails().Url + "webapp/builds/" + buildName + "/" + buildNumber)
	if err = utils.RemoveBuildDir(buildName, buildNumber); err != nil {
		return err
	}
	return nil
}

func createGenericBuildInfo(buildName, buildNumber string, flags *buildinfo.Flags) (*buildinfo.BuildInfo, error) {
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
	artifactsSet, dependenciesSet, env, vcs, err := extractBuildInfoData(partials, createIncludeFilter(flags), createExcludeFilter(flags))
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
	if len(artifactsSet) > 0 || len(dependenciesSet) > 0 {
		module := createModule(buildName, artifactsSet, dependenciesSet)
		buildInfo.Modules = append(buildInfo.Modules, module)
	}
	return buildInfo, nil
}

func extractBuildInfoData(partials buildinfo.Partials, includeFilter, excludeFilter filterFunc) ([]buildinfo.Artifacts, []buildinfo.Dependencies, buildinfo.Env, buildinfo.Vcs, error) {
	var artifacts []buildinfo.Artifacts
	var dependencies []buildinfo.Dependencies
	var env buildinfo.Env
	var vcs buildinfo.Vcs
	env = make(map[string]string)
	for _, partial := range partials {
		switch {
		case partial.Artifacts != nil:
			for _, v := range partial.Artifacts {
				artifacts = append(artifacts, v)
			}
		case partial.Dependencies != nil:
			for _, v := range partial.Dependencies {
				dependencies = append(dependencies, v)
			}
		case partial.Vcs != nil:
			vcs = *partial.Vcs
		case partial.Env != nil:
			envAfterIncludeFilter, e := includeFilter(partial.Env)
			if errorutils.CheckError(e) != nil {
				return artifacts, dependencies, env, vcs, e
			}
			envAfterExcludeFilter, e := excludeFilter(envAfterIncludeFilter)
			if errorutils.CheckError(e) != nil {
				return artifacts, dependencies, env, vcs, e
			}
			for k, v := range envAfterExcludeFilter {
				env[k] = v
			}
		}
	}
	return artifacts, dependencies, env, vcs, nil
}

func createModule(buildName string, artifacts []buildinfo.Artifacts, dependencies []buildinfo.Dependencies) *buildinfo.Module {
	module := createDefaultModule(buildName)
	if artifacts != nil && len(artifacts) > 0 {
		module.Artifacts = append(module.Artifacts, artifacts...)
	}
	if dependencies != nil && len(dependencies) > 0 {
		module.Dependencies = append(module.Dependencies, dependencies...)
	}
	return module
}

func createDefaultModule(buildName string) *buildinfo.Module {
	return &buildinfo.Module{
		Id:           buildName,
		Properties:   map[string][]string{},
		Artifacts:    []buildinfo.Artifacts{},
		Dependencies: []buildinfo.Dependencies{},
	}
}

type filterFunc func(map[string]string) (map[string]string, error)

func createIncludeFilter(flags *buildinfo.Flags) filterFunc {
	includePatterns := strings.Split(flags.EnvInclude, ";")
	return func(tempMap map[string]string) (map[string]string, error) {
		result := make(map[string]string)
		for k, v := range tempMap {
			for _, filterPattern := range includePatterns {
				bool, err := filepath.Match(filterPattern, k)
				if errorutils.CheckError(err) != nil {
					return nil, err
				}
				if bool == true {
					result[k] = v
					break
				}
			}
		}
		return result, nil
	}
}

func createExcludeFilter(flags *buildinfo.Flags) filterFunc {
	excludePattern := strings.Split(flags.EnvExclude, ";")
	return func(tempMap map[string]string) (map[string]string, error) {
		result := make(map[string]string)
		for k, v := range tempMap {
			include := true
			for _, filterPattern := range excludePattern {
				bool, err := filepath.Match(filterPattern, k)
				if errorutils.CheckError(err) != nil {
					return nil, err
				}
				if bool == true {
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
