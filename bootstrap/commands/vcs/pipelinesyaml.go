package commands

import (
	"fmt"
	"github.com/jfrog/jfrog-cli-core/utils/coreutils"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"gopkg.in/yaml.v2"
	"strconv"
	"strings"
)

const (
	jfrogCliFullImgName   = "releases-docker.jfrog.io/jfrog/jfrog-cli-full"
	jfrogCliFullImgTag    = "latest"
	m2pathCmd             = "MVN_PATH=`which mvn` && export M2_HOME=`readlink -f $MVN_PATH | xargs dirname | xargs dirname`"
	jfrogCliRtPrefix      = "jfrog rt"
	jfrogCliConfig        = "jfrog c add"
	jfrogCliBag           = "jfrog rt bag"
	jfrogCliBp            = "jfrog rt bp"
	buildNameEnvVar       = "JFROG_CLI_BUILD_NAME"
	buildNumberEnvVar     = "JFROG_CLI_BUILD_NUMBER"
	buildUrlEnvVar        = "JFROG_CLI_BUILD_URL"
	buildResultEnvVar     = "JFROG_BUILD_RESULTS"
	runNumberEnvVar       = "$run_number"
	stepUrlEnvVar         = "$step_url"
	updateCommitStatusCmd = "update_commit_status"

	passResult = "PASS"
	failResult = "FAIL"

	urlFlag    = "url"
	rtUrlFlag  = "artifactory-url"
	userFlag   = "user"
	apikeyFlag = "apikey"
)

func createPipelinesYaml(gitProvider, rtIntegration string, vcsData *VcsData) ([]byte, error) {
	log.Debug("Creating Pipelines Yaml...")
	pipelineName := createPipelineName(vcsData)
	gitResourceName := createGitResourceName(vcsData)

	converted, err := convertBuildCmd(vcsData.BuildCommand)
	if err != nil {
		return nil, err
	}

	pipelinesCommands := getPipelineCommands(rtIntegration, gitResourceName, converted, vcsData)
	gitResource := createGitResource(gitResourceName, gitProvider, getRepoFullName(vcsData), vcsData.GitBranch)
	pipeline := createPipeline(rtIntegration, pipelineName, gitResourceName, pipelinesCommands)
	pipelineYaml := PipelineYml{
		Resources: []Resource{gitResource},
		Pipelines: []Pipeline{pipeline},
	}
	pipelineBytes, err := yaml.Marshal(&pipelineYaml)
	if err != nil {
		return nil, errorutils.CheckError(err)
	}
	return pipelineBytes, nil
}

func getPipelineCommands(rtIntegration, gitResourceName, convertedBuildCmd string, vcsData *VcsData) []string {
	var commandsArray []string
	commandsArray = append(commandsArray, getCdToResourceCmd(gitResourceName))
	commandsArray = append(commandsArray, getExportsCommands(vcsData)...)
	commandsArray = append(commandsArray, getJfrogCliConfigCmd(rtIntegration))
	commandsArray = append(commandsArray, getTechConfigsCommands(rtIntegration, vcsData)...)
	commandsArray = append(commandsArray, convertedBuildCmd)
	commandsArray = append(commandsArray, jfrogCliBag)
	commandsArray = append(commandsArray, jfrogCliBp)
	return commandsArray
}

func createBuildCmdRegexp() string {
	// Beginning of line or with a preceding space.
	regexp := "(^|\\s)("
	// One of the supported executables names.
	regexp += strings.Join(techExecutablesNames, "|")
	// Has a succeeding space.
	regexp += ")(\\s)"
	return regexp
}

// Converts build tools commands to run via JFrog CLI.
func convertBuildCmd(buildCmd string) (string, error) {
	regexpStr := createBuildCmdRegexp()
	regexp, err := utils.GetRegExp(regexpStr)
	if err != nil {
		return "", err
	}
	// Replace exe (group 2) with "jfrog rt exe" while maintaining preceding (if any) and succeeding spaces.
	replacement := fmt.Sprintf("${1}%s ${2}${3}", jfrogCliRtPrefix)
	return regexp.ReplaceAllString(buildCmd, replacement), nil
}

func getCdToResourceCmd(gitResourceName string) string {
	return fmt.Sprintf("cd $res_%s_resourcePath", gitResourceName)
}

func getIntDetailForCmd(intName, detail string) string {
	return fmt.Sprintf("$int_%s_%s", intName, detail)
}

func getFlagSyntax(flagName string) string {
	return fmt.Sprintf("--%s", flagName)
}

func getJfrogCliConfigCmd(rtIntName string) string {
	return strings.Join([]string{
		jfrogCliConfig, rtIntName,
		getFlagSyntax(rtUrlFlag), getIntDetailForCmd(rtIntName, urlFlag),
		getFlagSyntax(userFlag), getIntDetailForCmd(rtIntName, userFlag),
		getFlagSyntax(apikeyFlag), getIntDetailForCmd(rtIntName, apikeyFlag),
		"--enc-password=false",
	}, " ")
}

func getTechConfigsCommands(serverId string, data *VcsData) []string {
	var configs []string
	if used, ok := data.DetectedTechnologies[Maven]; ok && used {
		configs = append(configs, m2pathCmd)
		configs = append(configs, getMavenConfigCmd(serverId, data.ArtifactoryVirtualRepos[Maven]))
	}
	if used, ok := data.DetectedTechnologies[Gradle]; ok && used {
		configs = append(configs, getBuildToolConfigCmd(cliutils.GradleConfig, serverId, data.ArtifactoryVirtualRepos[Gradle]))
	}
	if used, ok := data.DetectedTechnologies[Npm]; ok && used {
		configs = append(configs, getBuildToolConfigCmd(cliutils.NpmConfig, serverId, data.ArtifactoryVirtualRepos[Npm]))
	}
	return configs
}

func getMavenConfigCmd(serverId, repo string) string {
	return strings.Join([]string{
		jfrogCliRtPrefix, cliutils.MvnConfig,
		getFlagSyntax(cliutils.ServerIdResolve), serverId,
		getFlagSyntax(cliutils.RepoResolveReleases), repo,
		getFlagSyntax(cliutils.RepoResolveSnapshots), repo,
	}, " ")
}

func getBuildToolConfigCmd(configCmd, serverId, repo string) string {
	return strings.Join([]string{
		jfrogCliRtPrefix, configCmd,
		getFlagSyntax(cliutils.ServerIdResolve), serverId,
		getFlagSyntax(cliutils.RepoResolve), repo,
	}, " ")
}

func getExportsCommands(vcsData *VcsData) []string {
	return []string{
		getExportCmd(coreutils.CI, strconv.FormatBool(true)),
		getExportCmd(buildNameEnvVar, vcsData.BuildName),
		getExportCmd(buildNumberEnvVar, runNumberEnvVar),
		getExportCmd(buildUrlEnvVar, stepUrlEnvVar),
	}
}

func getExportCmd(key, value string) string {
	return fmt.Sprintf("export %s=%s", key, value)
}

func createGitResource(gitResourceName, gitProvider, gitRepoFullPath, branch string) Resource {
	return Resource{
		Name:         gitResourceName,
		ResourceType: GitRepo,
		ResourceConfiguration: ResourceConfiguration{
			Path:        gitRepoFullPath,
			GitProvider: gitProvider,
			BuildOn: BuildOn{
				PullRequestCreate: true,
			},
			Branches: IncludeExclude{Include: branch},
		},
	}
}

func createPipeline(rtIntegration, pipelineName, gitResourceName string, commands []string) Pipeline {
	return Pipeline{
		Name: pipelineName,
		Configuration: PipelineConfiguration{
			Runtime{
				RuntimeType: Image,
				Image: RuntimeImage{
					Custom: CustomImage{
						Name: jfrogCliFullImgName,
						Tag:  jfrogCliFullImgTag,
					},
				},
			},
		},
		Steps: []PipelineStep{
			{
				Name:     "Build",
				StepType: "Bash",
				Configuration: StepConfiguration{
					InputResources: []StepResource{
						{
							Name: gitResourceName,
						},
					},
					Integrations: []StepIntegration{
						{
							Name: rtIntegration,
						},
					},
				},
				Execution: StepExecution{
					OnExecute: commands,
					OnSuccess: getOnResultCmd(passResult, gitResourceName),
					OnFailure: getOnResultCmd(failResult, gitResourceName),
				},
			},
		},
	}
}

type PipelineYml struct {
	Resources []Resource `yaml:"resources,omitempty"`
	Pipelines []Pipeline `yaml:"pipelines,omitempty"`
}

type Resource struct {
	Name                  string `yaml:"name,omitempty"`
	ResourceType          `yaml:"type,omitempty"`
	ResourceConfiguration `yaml:"configuration,omitempty"`
}

type ResourceType string

const (
	GitRepo ResourceType = "GitRepo"
)

type ResourceConfiguration struct {
	Path        string `yaml:"path,omitempty"`
	GitProvider string `yaml:"gitProvider,omitempty"`
	BuildOn     `yaml:"buildOn,omitempty"`
	Branches    IncludeExclude `yaml:"branches,omitempty"`
}

type IncludeExclude struct {
	Include string `yaml:"include,omitempty"`
	Exclude string `yaml:"exclude,omitempty"`
}

type BuildOn struct {
	PullRequestCreate bool `yaml:"pullRequestCreate,omitempty"`
	Commit            bool `yaml:"commit,omitempty"`
}

type Pipeline struct {
	Name          string                `yaml:"name,omitempty"`
	Configuration PipelineConfiguration `yaml:"configuration,omitempty"`
	Steps         []PipelineStep        `yaml:"steps,omitempty"`
}

type PipelineConfiguration struct {
	Runtime `yaml:"runtime,omitempty"`
}

type RuntimeType string

const (
	Image RuntimeType = "image"
)

type Runtime struct {
	RuntimeType `yaml:"type,omitempty"`
	Image       RuntimeImage `yaml:"image,omitempty"`
}

type RuntimeImage struct {
	Auto   AutoImage   `yaml:"auto,omitempty"`
	Custom CustomImage `yaml:"custom,omitempty"`
}

type Language string

const (
	Java   Language = "java"
	NodeJs Language = "node"
)

type AutoImage struct {
	Language `yaml:"language,omitempty"`
	Versions []string `yaml:"versions,omitempty"`
}

type CustomImage struct {
	Name             string `yaml:"name,omitempty"`
	Tag              string `yaml:"tag,omitempty"`
	Options          string `yaml:"options,omitempty"`
	Registry         string `yaml:"registry,omitempty"`
	SourceRepository string `yaml:"sourceRepository,omitempty"`
	Region           string `yaml:"region,omitempty"`
}

type PipelineStep struct {
	Name          string            `yaml:"name,omitempty"`
	StepType      string            `yaml:"type,omitempty"`
	Configuration StepConfiguration `yaml:"configuration,omitempty"`
	Execution     StepExecution     `yaml:"execution,omitempty"`
}

type StepConfiguration struct {
	InputResources []StepResource    `yaml:"inputResources,omitempty"`
	Integrations   []StepIntegration `yaml:"integrations,omitempty"`
}

type StepResource struct {
	Name string `yaml:"name,omitempty"`
}

type StepIntegration struct {
	Name string `yaml:"name,omitempty"`
}

type StepExecution struct {
	OnStart    []string `yaml:"onStart,omitempty"`
	OnExecute  []string `yaml:"onExecute,omitempty"`
	OnComplete []string `yaml:"onComplete,omitempty"`
	OnSuccess  []string `yaml:"onSuccess,omitempty"`
	OnFailure  []string `yaml:"onFailure,omitempty"`
}

func getOnResultCmd(result, gitResourceName string) []string {
	return []string{getExportCmd(buildResultEnvVar, result), getUpdateCommitStatusCmd(gitResourceName)}
}

func getUpdateCommitStatusCmd(gitResourceName string) string {
	return updateCommitStatusCmd + " " + gitResourceName
}

func createGitResourceName(data *VcsData) string {
	return createPipelinesSuitableName(data, "gitResource")
}

func createPipelineName(data *VcsData) string {
	return createPipelinesSuitableName(data, "pipeline")
}
