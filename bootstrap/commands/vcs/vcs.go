package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/codegangsta/cli"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/jfrog/jfrog-cli-core/artifactory/commands"
	"github.com/jfrog/jfrog-cli-core/artifactory/commands/buildinfo"
	"github.com/jfrog/jfrog-cli-core/artifactory/commands/utils"
	rtutils "github.com/jfrog/jfrog-cli-core/artifactory/utils"
	corecommands "github.com/jfrog/jfrog-cli-core/common/commands"
	utilsconfig "github.com/jfrog/jfrog-cli-core/utils/config"
	"github.com/jfrog/jfrog-cli-core/utils/coreutils"
	artifactoryAuth "github.com/jfrog/jfrog-client-go/artifactory/auth"
	buildinfocmd "github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/jfrog/jfrog-client-go/auth"
	"github.com/jfrog/jfrog-client-go/config"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/jfrog/jfrog-client-go/xray"
	xrayutils "github.com/jfrog/jfrog-client-go/xray/services/utils"
)

const (
	ConfigServerId = "vcs-integration-platform"
	VcsConfigFile  = "jfrog-cli-vcs.conf"
	// Basic vcs questions keys
	JfrogUrl      = "jfrogUrl"
	JfrogUsername = "jfrogUsername"
	JfrogPassword = "jfrogPassword"
	VcsUrl        = "vcsUrl"
	VcsUsername   = "vcsUsername"
	VcsPassword   = "vcsPassword"
	VcsBranches   = "vcsBranches"
	LocalDir      = "localDir"

	// Technologies, repositories & build questions keys
	DefultFirstBuildNumber = "0"

	// Output questions keys
)

type VcsCommand struct {
	defaultData *VcsData
	data        *VcsData
}

type VcsData struct {
	ProjectName             string
	LocalDirPath            string
	VcsBranches             []string
	BuildCommand            string
	BuildName               string
	ArtifactoryVirtualRepos map[Technology]string
	// A collection of technologies that was found with a list of theirs indications
	DetectedTechnologies map[Technology]bool
	VcsCredentials       auth.ServiceDetails
}

func (vc *VcsCommand) SetData(data *VcsData) *VcsCommand {
	vc.data = data
	return vc
}
func (vc *VcsCommand) SetDefaultData(data *VcsData) *VcsCommand {
	vc.defaultData = data
	return vc
}

func (vc *VcsCommand) Run() error {
	vc.prepareConfigurationData()
	err := vc.VcsCmd()
	if err != nil {
		return err
	}
	return saveVcsConf(vc.data)

}

func (vc *VcsCommand) prepareConfigurationData() error {
	// If data is nil, initialize a new one
	if vc.data == nil {
		vc.data = new(VcsData)
	}

	// Get previous vcs data if exists
	vc.defaultData = readVcsConf()
	return nil
}

func readVcsConf() (conf *VcsData) {
	conf = &VcsData{}
	path, err := coreutils.GetJfrogHomeDir()
	if err != nil {
		return
	}
	configFile, err := fileutils.ReadFile(filepath.Join(path, VcsConfigFile))
	if err != nil {
		return
	}
	json.Unmarshal(configFile, conf)
	return
}

func saveVcsConf(conf *VcsData) error {
	path, err := coreutils.GetJfrogHomeDir()
	if err != nil {
		return err
	}
	bytesContent, err := json.Marshal(conf)
	if err != nil {
		return errorutils.CheckError(err)
	}
	var content bytes.Buffer
	err = json.Indent(&content, bytesContent, "", "  ")
	if err != nil {
		return errorutils.CheckError(err)
	}
	err = ioutil.WriteFile(path, []byte(content.String()), 0600)
	if err != nil {
		return errorutils.CheckError(err)
	}
	return nil
}

func VcsCmd(c *cli.Context) error {
	// Run JFrog config command
	err := runConfigCmd()
	if err != nil {
		return err
	}
	log.Info("Done config JFrog server - " + ConfigServerId)
	// Basic VCS questionnaire (URLs, Credentials, etc'...)
	err = runBasicAuthenticationQuestionnaire()
	if err != nil {
		return err
	}
	// Interactively create Artifactory repository based on the detected technologies and on going user input
	err = runBuildQuestionnaire(c)
	if err != nil || saveVcsConf(data) != nil {
		return err
	}
	// Publish empty build info.
	err = publishFirstBuild()
	if err != nil || saveVcsConf(data) != nil {
		return err
	}
	// Configure Xray to scan the new build.
	err = configureXray()
	if err != nil || saveVcsConf(data) != nil {
		return err
	}
	// Run jfrog-vcs-agent
	//buildConfig := convertVcsDataToBuildConfig(*data)

	// Output?

	return err
}

func runConfigCmd() (err error) {
	configCmd := corecommands.NewConfigCommand().SetInteractive(true).SetServerId(ConfigServerId).SetEncPassword(true)
	for {
		err = configCmd.Config()
		if err == nil {
			return nil
		}
		log.Error(err)
	}
}

func publishFirstBuild() (err error) {
	buildName := utils.AskStringWithDefault("", "> Enter name for your build.", "${projectName}-${branch}")
	if len(data.VcsBranches) > 1 && !strings.Contains(buildName, "${branch}") {
		return fmt.Errorf("Build name must be uniq, please include the ${branch} variable in vcs build name")
	}
	data.BuildName = buildName
	buildName = strings.Replace(buildName, "${projectName}", data.ProjectName, -1)
	for _, branch := range data.VcsBranches {

		buildName = strings.Replace(buildName, "${branch}", branch, -1)
		// Run BAG Command (in order to publish the first, empty, buildinfo)
		buildAddGitConfigurationCmd := buildinfo.NewBuildAddGitCommand().SetDotGitPath(data.LocalDirPath).SetServerId(ConfigServerId) //.SetConfigFilePath(c.String("config"))
		buildConfiguration := rtutils.BuildConfiguration{BuildName: buildName, BuildNumber: DefultFirstBuildNumber}
		buildAddGitConfigurationCmd = buildAddGitConfigurationCmd.SetBuildConfiguration(&buildConfiguration)
		err = commands.Exec(buildAddGitConfigurationCmd)
		if err != nil {
			return err
		}
		// Run BP Command.
		serviceDetails, err := utilsconfig.GetSpecificConfig(ConfigServerId, true, false)
		if err != nil {
			return err
		}
		buildInfoConfiguration := buildinfocmd.Configuration{DryRun: false}
		buildPublishCmd := buildinfo.NewBuildPublishCommand().SetServerDetails(serviceDetails).SetBuildConfiguration(&buildConfiguration).SetConfig(&buildInfoConfiguration)
		err = commands.Exec(buildPublishCmd)
		if err != nil {
			return err
		}
	}
	return
}

func configureXray() (err error) {
	serviceDetails, err := utilsconfig.GetSpecificConfig(ConfigServerId, true, false)
	if err != nil {
		return err
	}
	xrayDetails, err := serviceDetails.CreateXrayAuthConfig()
	serviceConfig, err := config.NewConfigBuilder().
		SetServiceDetails(xrayDetails).
		Build()
	if err != nil {
		return err
	}
	xrayManager, err := xray.New(&xrayDetails, serviceConfig)
	if err != nil {
		return err
	}
	// AddBuildsToIndexing.
	buildsToIndex := []string{data.BuildName}
	err = xrayManager.AddBuildsToIndexing(buildsToIndex)
	// Create new defult policy.
	policyParams := xrayutils.NewPolicyParams()
	policyParams.Name = "vcs-integration-security-policy"
	policyParams.Type = xrayutils.Security
	policyParams.Description = "Basic Security policy."
	policyParams.Rules = []xrayutils.PolicyRule{
		{
			Name:     "min-severity-rule",
			Criteria: *xrayutils.CreateSeverityPolicyCriteria(xrayutils.Critical),
			Priority: 1,
		},
	}
	err = xrayManager.CreatePolicy(policyParams)
	if err != nil {
		return err
	}
	// Create new defult watcher.
	watchParams := xrayutils.NewWatchParams()
	watchParams.Name = "vcs-integration-watch-all"
	watchParams.Description = "VCS Configured Build Watch"
	watchParams.Active = true

	// Need to be verified before merging
	watchParams.Builds.Type = xrayutils.WatchBuildAll
	watchParams.Policies = []xrayutils.AssignedPolicy{
		{
			Name: policyParams.Name,
			Type: "security",
		},
	}

	err = xrayManager.CreateWatch(watchParams)
	if err != nil {
		return err
	}
	return
}

func runBuildQuestionnaire(c *cli.Context) (err error) {
	data.ArtifactoryVirtualRepos = make(map[Technology]string)
	// First create repositories for each technology in Artifactory according to user input
	for tech, detected := range data.DetectedTechnologies {
		if detected && coreutils.AskYesNo(fmt.Sprintf("A %q technology has been detected, would you like %q repositories to be configured?", tech, tech), true) {
			err = interactivelyCreatRepos(tech)
			if err != nil {
				return
			}
		}
	}
	// Ask for working build command
	data.BuildCommand = utils.AskString("", "Please provide a single-line build command. You may use scripts or AND operator if necessary.", false, false)
	return nil
}

func interactivelyCreatRepos(technologyType Technology) (err error) {
	serviceDetails, err := utilsconfig.GetSpecificConfig(ConfigServerId, true, false)
	if err != nil {
		return err
	}
	// Get all relevant remotes to chose from
	remoteRepos, err := GetAllRepos(serviceDetails, Remote, string(technologyType))
	if err != nil {
		return err
	}
	repoNames := ConvertRepoDetailsToRepoNames(remoteRepos)
	// Add the option to create new remote repository
	repoNames = append(repoNames, NewRepository)

	// Ask if the user would like us to create a new remote or to chose from the exist repositories list
	remoteRepo := utils.AskFromList("", "Would you like to chose one of the following repositories or create a new one?", false, utils.ConvertToSuggests(repoNames), NewRepository)
	if remoteRepo == NewRepository {
		for {
			repoName := utils.AskStringWithDefault("", "Enter repository name >", GetRemoteDefaultName(technologyType))
			remoteUrl := utils.AskStringWithDefault("", "Enter repository url >", GetRemoteDefaultUrl(technologyType))
			err = CreateRemoteRepo(serviceDetails, technologyType, repoName, remoteUrl)
			if err != nil {
				log.Error(err)
			} else {
				remoteRepo = repoName
				break
			}
		}
	}
	// Create virtual repository
	for {
		virtualRepoName := utils.AskStringWithDefault(fmt.Sprintf("Creating %q virtual repository", technologyType), "Enter repository name >", GetVirtualDefaultName(technologyType))
		err = CreateVirtualRepo(serviceDetails, technologyType, virtualRepoName, remoteRepo)
		if err != nil {
			log.Error(err)
		} else {
			// Saves the new created repo name (key) in the results data structure.
			data.ArtifactoryVirtualRepos[technologyType] = virtualRepoName
			return
		}
	}

}

func runBasicAuthenticationQuestionnaire() (err error) {
	basicAuthenticationQuestionnaire := &utils.InteractiveQuestionnaire{
		MandatoryQuestionsKeys: []string{VcsUrl, VcsUsername, VcsPassword, VcsBranches, LocalDir},
		QuestionsMap:           basicAuthenticationQuestionMap,
	}
	err = basicAuthenticationQuestionnaire.Perform()
	if err != nil {
		return err
	}
	resBytes, err := json.Marshal(basicAuthenticationQuestionnaire.AnswersMap)
	if err != nil {
		return errorutils.CheckError(err)
	}
	if err = ioutil.WriteFile("./VCS-Authentication-config", resBytes, 0644); err != nil {
		return errorutils.CheckError(err)
	}
	log.Info(fmt.Sprintf("Basic VCS Authentication config template successfully created at curent directory"))
	return nil
}

func setVcsCredentials(iq *utils.InteractiveQuestionnaire, ans string) (value string, err error) {
	data.VcsCredentials = (artifactoryAuth.NewArtifactoryDetails())
	data.VcsCredentials.SetUrl(iq.AnswersMap[VcsUrl].(string))
	data.VcsCredentials.SetUser(iq.AnswersMap[VcsUsername].(string))
	data.VcsCredentials.SetPassword(iq.AnswersMap[VcsPassword].(string))
	return
}

func setAndPreformeClone(iq *utils.InteractiveQuestionnaire, ans string) (value string, err error) {
	data.VcsBranches = strings.Split(iq.AnswersMap[VcsBranches].(string), ",")
	data.LocalDirPath = iq.AnswersMap[LocalDir].(string)
	err = cloneProject()
	if err != nil {
		return
	}
	err = detectTechnologies()
	return
}

func cloneProject() (err error) {
	// Create the desired path if necessary
	err = os.MkdirAll(data.LocalDirPath, os.ModePerm)
	if err != nil {
		return err
	}
	cloneOption := &git.CloneOptions{
		URL:  data.VcsCredentials.GetUrl(),
		Auth: createCredentials(data.VcsCredentials),
		// Enable git submodules clone if there any.
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	}
	setProjectName()
	// Clone the given repository to the given directory from the given branch
	log.Info(fmt.Sprintf("git clone project %q from: %q to: %q", data.ProjectName, data.VcsCredentials.GetUrl(), data.LocalDirPath))
	_, err = git.PlainClone(data.LocalDirPath, false, cloneOption)
	log.Info(err)
	return
}

func setProjectName() {
	vcsUrl := data.VcsCredentials.GetUrl()
	// Trim trailing "/" if one exists
	vcsUrl = strings.TrimSuffix(vcsUrl, "/")
	data.VcsCredentials.SetUrl(vcsUrl)
	projectName := vcsUrl[strings.LastIndex(vcsUrl, "/")+1:]
	data.ProjectName = strings.TrimSuffix(projectName, ".git")
}

func detectTechnologies() (err error) {
	indicators := GetTechIndicators()
	log.Info(filepath.Join(data.LocalDirPath, data.ProjectName))
	filesList, err := fileutils.ListFilesRecursiveWalkIntoDirSymlink(data.LocalDirPath, false)
	if err != nil {
		return err
	}
	data.DetectedTechnologies = make(map[Technology]bool)
	for _, file := range filesList {
		for _, indicator := range indicators {
			if indicator.Indicates(file) {
				//log.Info(file)
				data.DetectedTechnologies[indicator.GetTechnology()] = true
				// Same file can't indicate on more than one technology.
				break
			}
		}
	}
	return
}

func createCredentials(serviceDetails auth.ServiceDetails) (auth transport.AuthMethod) {
	var password string
	if serviceDetails.GetApiKey() != "" {
		password = serviceDetails.GetApiKey()
	} else if serviceDetails.GetAccessToken() != "" {
		password = serviceDetails.GetAccessToken()
	} else {
		password = serviceDetails.GetPassword()
	}
	return &http.BasicAuth{Username: serviceDetails.GetUser(), Password: password}
}

var data = &VcsData{}

var basicAuthenticationQuestionMap = map[string]utils.QuestionInfo{
	VcsUrl: {
		Msg:          "",
		PromptPrefix: "Enter VCS URL >",
		AllowVars:    true,
		Writer:       utils.WriteStringAnswer,
		MapKey:       VcsUrl,
		Callback:     nil,
	},
	VcsUsername: {
		Msg:          "",
		PromptPrefix: "Enter VCS admin user name >",
		AllowVars:    true,
		Writer:       utils.WriteStringAnswer,
		MapKey:       VcsUsername,
		Callback:     nil,
	},
	VcsPassword: {
		Msg:          "",
		PromptPrefix: "Enter VCS password >",
		AllowVars:    true,
		Writer:       utils.WriteStringAnswer,
		MapKey:       VcsPassword,
		Callback:     setVcsCredentials,
	},
	VcsBranches: {
		Msg:          "",
		PromptPrefix: "Enter comma sperated list of git branches >",
		AllowVars:    true,
		Writer:       utils.WriteStringAnswer,
		MapKey:       VcsBranches,
		Callback:     setVcsCredentials,
	},
	LocalDir: {
		Options: []prompt.Suggest{
			{Text: "./workspace", Description: "a temp dir that will include a clone of the VCS project."},
		},
		Msg:          "",
		PromptPrefix: "Enter target directory for projet clone >",
		AllowVars:    true,
		Writer:       utils.WriteStringAnswer,
		MapKey:       LocalDir,
		Callback:     setAndPreformeClone,
	},
}
