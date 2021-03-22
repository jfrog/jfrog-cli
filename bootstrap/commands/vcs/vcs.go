package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"syscall"

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
	"github.com/jfrog/jfrog-cli-core/utils/ioutils"
	buildinfocmd "github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	"github.com/jfrog/jfrog-client-go/config"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/jfrog/jfrog-client-go/xray"
	xrayservices "github.com/jfrog/jfrog-client-go/xray/services"
	xrayutils "github.com/jfrog/jfrog-client-go/xray/services/utils"
	"golang.org/x/crypto/ssh/terminal"
)

const (
	ConfigServerId         = "vcs-integration-platform"
	VcsConfigFile          = "jfrog-cli-vcs.conf"
	DefultFirstBuildNumber = "0"
	DefultWorkSpace        = "./JFrogVcsWorkSpace"
)

type VcsCommand struct {
	defaultData *VcsData
	data        *VcsData
}

type VcsData struct {
	ProjectName             string
	LocalDirPath            string
	GitBranch               string
	BuildCommand            string
	BuildName               string
	ArtifactoryVirtualRepos map[Technology]string
	// A collection of technologies that was found with a list of theirs indications
	DetectedTechnologies map[Technology]bool
	VcsCredentials       VcsServerDetails
}
type VcsServerDetails struct {
	Url         string `json:"url,omitempty"`
	User        string `json:"user,omitempty"`
	Password    string `json:"-"`
	AccessToken string `json:"-"`
}

func (vc *VcsCommand) SetData(data *VcsData) *VcsCommand {
	vc.data = data
	return vc
}
func (vc *VcsCommand) SetDefaultData(data *VcsData) *VcsCommand {
	vc.defaultData = data
	return vc
}

func VcsCmd() error {
	vc := &VcsCommand{}
	vc.prepareConfigurationData()
	err := vc.Run()
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
	err = ioutil.WriteFile(filepath.Join(path, VcsConfigFile), []byte(content.String()), 0600)
	if err != nil {
		return errorutils.CheckError(err)
	}
	return nil
}

func (vc *VcsCommand) Run() error {
	// Run JFrog config command
	err := runConfigCmd()
	if err != nil {
		return err
	}

	// Basic VCS questionnaire (URLs, Credentials, etc'...)
	err = vc.getVcsCredentialsFromConsole()
	if err != nil || saveVcsConf(vc.data) != nil {
		return err
	}

	// Interactively create Artifactory repository based on the detected technologies and on going user input
	err = vc.runBuildQuestionnaire()
	if err != nil || saveVcsConf(vc.data) != nil {
		return err
	}
	// Publish empty build info.
	err = vc.publishFirstBuild()
	if err != nil || saveVcsConf(vc.data) != nil {
		return err
	}
	// Configure Xray to scan the new build.
	err = vc.configureXray()
	if err != nil || saveVcsConf(vc.data) != nil {
		return err
	}
	// Run jfrog-vcs-agent
	//buildConfig := convertVcsDataToBuildConfig(*data)

	// Output?

	return err
}

func runConfigCmd() (err error) {
	for {
		configCmd := corecommands.NewConfigCommand().SetInteractive(true).SetServerId(ConfigServerId).SetEncPassword(true)
		err = configCmd.Config()
		if err == nil {
			return nil
		}
		log.Error(err)
	}
}

func (vc *VcsCommand) publishFirstBuild() (err error) {
	println("Everytime the new pipeline builds the code, it generates a build entity (also known as build-info) and stores it in Artifactory.")
	ioutils.ScanFromConsole("Please choose a name of the build", &vc.data.BuildName, "${projectName}-${branch}")
	vc.data.BuildName = strings.Replace(vc.data.BuildName, "${projectName}", vc.data.ProjectName, -1)
	vc.data.BuildName = strings.Replace(vc.data.BuildName, "${branch}", vc.data.GitBranch, -1)
	// Run BAG Command (in order to publish the first, empty, buildinfo)
	buildAddGitConfigurationCmd := buildinfo.NewBuildAddGitCommand().SetDotGitPath(vc.data.LocalDirPath).SetServerId(ConfigServerId) //.SetConfigFilePath(c.String("config"))
	buildConfiguration := rtutils.BuildConfiguration{BuildName: vc.data.BuildName, BuildNumber: DefultFirstBuildNumber}
	buildAddGitConfigurationCmd = buildAddGitConfigurationCmd.SetBuildConfiguration(&buildConfiguration)
	log.Info("Generating an initial build-info...")
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
	return
}

func (vc *VcsCommand) configureXray() (err error) {
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
	buildsToIndex := []string{vc.data.BuildName}
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
		// In case the error is from type PolicyAlreadyExistsError, we should continue with the regular flow.
		if _, ok := err.(*xrayservices.PolicyAlreadyExistsError); !ok {
			return err
		} else {
			log.Debug(err.(*xrayservices.PolicyAlreadyExistsError).InnerError)
			err = nil
		}
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
		// In case the error is from type WatchAlreadyExistsError, we should continue with the regular flow.
		if _, ok := err.(*xrayservices.WatchAlreadyExistsError); !ok {
			return err
		} else {
			log.Debug(err.(*xrayservices.WatchAlreadyExistsError).InnerError)
			err = nil
		}
	}
	return
}

func (vc *VcsCommand) runBuildQuestionnaire() (err error) {

	vc.data.ArtifactoryVirtualRepos = make(map[Technology]string)
	// First create repositories for each technology in Artifactory according to user input
	for tech, detected := range vc.data.DetectedTechnologies {
		if detected && coreutils.AskYesNo(fmt.Sprintf(" It looks like the source code is built using %s. Would you like to resolve the %s dependencies from Artifactory?", tech, tech), true) {
			err = vc.interactivelyCreatRepos(tech)
			if err != nil {
				return
			}
		}
	}
	// Ask for working build command
	vc.data.BuildCommand = utils.AskString("", "Please provide a single-line build command. You may use scripts or the && operator:", false, false)
	return nil
}

func (vc *VcsCommand) interactivelyCreatRepos(technologyType Technology) (err error) {
	serviceDetails, err := utilsconfig.GetSpecificConfig(ConfigServerId, true, false)
	if err != nil {
		return err
	}
	// Get all relevant remotes to chose from
	remoteRepos, err := GetAllRepos(serviceDetails, Remote, string(technologyType))
	if err != nil {
		return err
	}

	// Ask if the user would like us to create a new remote or to chose from the exist repositories list
	remoteRepo, err := promptARepoSelection(remoteRepos, "Select remote repository")
	if err != nil {
		return nil
	}
	// The user choose to create a new remote repo
	if remoteRepo == NewRepository {
		for {
			var repoName, repoUrl string
			ioutils.ScanFromConsole("Repository Name", &repoName, GetRemoteDefaultName(technologyType))
			ioutils.ScanFromConsole("Repository URL", &repoUrl, GetRemoteDefaultUrl(technologyType))
			err = CreateRemoteRepo(serviceDetails, technologyType, repoName, repoUrl)
			if err != nil {
				log.Error(err)
			} else {
				remoteRepo = repoName
				// Create a new virtual repository as well
				ioutils.ScanFromConsole(fmt.Sprintf("Choose a name for a new virtual repository which will include %q remote repo", remoteRepo),
					&repoName, GetVirtualDefaultName(technologyType))
				err = CreateVirtualRepo(serviceDetails, technologyType, repoName, remoteRepo)
				if err != nil {
					log.Error(err)
				} else {
					// we created both remote and virtual repositories successfully
					vc.data.ArtifactoryVirtualRepos[technologyType] = repoName
					return
				}
			}
		}
	}
	// Else, the user choose an existing remote repo
	virtualRepos, err := GetAllRepos(serviceDetails, Virtual, string(technologyType))
	if err != nil {
		return err
	}
	// Ask if the user would like us to create a new virtual or to chose from the exist repositories list
	virtualRepo, err := promptARepoSelection(virtualRepos, fmt.Sprintf("Select a virtual repository, which includes %s or chose to create a new repo:", remoteRepo))
	if virtualRepo == NewRepository {
		// Create virtual repository
		for {
			var repoName string
			ioutils.ScanFromConsole("Repository Name", &repoName, GetVirtualDefaultName(technologyType))
			err = CreateVirtualRepo(serviceDetails, technologyType, repoName, remoteRepo)
			if err != nil {
				log.Error(err)
			} else {
				virtualRepo = repoName
				break
			}
		}
	} else {
		// Validate that the chosen virtual repo contains the chosen remote repo
		rtAuth, err := serviceDetails.CreateArtAuthConfig()
		if err != nil {
			return err
		}
		chosenVirtualRepo, err := GetVirtualRepo(&rtAuth, virtualRepo)
		if err != nil {
			return err
		}
		if !contains(chosenVirtualRepo.Repositories, remoteRepo) {
			log.Error(fmt.Sprintf("The chosen virtual repo %q does not contain the chosen remote repo %q", virtualRepo, remoteRepo))
			return vc.interactivelyCreatRepos(technologyType)
		}
	}
	// Saves the new created repo name (key) in the results data structure.
	vc.data.ArtifactoryVirtualRepos[technologyType] = virtualRepo
	return
}

func promptARepoSelection(repoDetailes *[]services.RepositoryDetails, promptMsg string) (selectedRepoName string, err error) {

	selectableItems := []ioutils.PromptItem{{Option: NewRepository, TargetValue: &selectedRepoName}}
	for _, repo := range *repoDetailes {
		selectableItems = append(selectableItems, ioutils.PromptItem{Option: repo.Key, TargetValue: &selectedRepoName, DefaultValue: repo.Url})
	}
	println(promptMsg)
	err = ioutils.SelectString(selectableItems, "", func(item ioutils.PromptItem) {
		*item.TargetValue = item.Option
	})
	return
}

func (vc *VcsCommand) prepareVcsData() (err error) {
	ioutils.ScanFromConsole("Git Branch", &vc.data.GitBranch, vc.defaultData.GitBranch)
	err = fileutils.CreateDirIfNotExist(DefultWorkSpace)
	if err != nil {
		return err
	}
	dirEmpty, err := fileutils.IsDirEmpty(DefultWorkSpace)
	if err != nil {
		return err
	}
	if !dirEmpty {
		ioutils.ScanFromConsole("WorkSpace Dir", &vc.data.LocalDirPath, "")
	} else {
		vc.data.LocalDirPath = DefultWorkSpace
	}
	err = vc.cloneProject()
	if err != nil {
		return
	}
	err = vc.detectTechnologies()
	return
}

func (vc *VcsCommand) cloneProject() (err error) {
	// Create the desired path if necessary
	err = os.MkdirAll(vc.data.LocalDirPath, os.ModePerm)
	if err != nil {
		return err
	}
	cloneOption := &git.CloneOptions{
		URL:  vc.data.VcsCredentials.Url,
		Auth: createCredentials(&vc.data.VcsCredentials),
		// Enable git submodules clone if there any.
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	}
	vc.setProjectName()
	// Clone the given repository to the given directory from the given branch
	log.Info(fmt.Sprintf("Cloning project %q from: %q into: %q", vc.data.ProjectName, vc.data.VcsCredentials.Url, vc.data.LocalDirPath))
	_, err = git.PlainClone(vc.data.LocalDirPath, false, cloneOption)
	if err != nil {
		return err
	}
	return
}

func (vc *VcsCommand) setProjectName() {
	vcsUrl := vc.data.VcsCredentials.Url
	// Trim trailing "/" if one exists
	vcsUrl = strings.TrimSuffix(vcsUrl, "/")
	vc.data.VcsCredentials.Url = vcsUrl
	projectName := vcsUrl[strings.LastIndex(vcsUrl, "/")+1:]
	vc.data.ProjectName = strings.TrimSuffix(projectName, ".git")
}

func (vc *VcsCommand) detectTechnologies() (err error) {
	indicators := GetTechIndicators()
	filesList, err := fileutils.ListFilesRecursiveWalkIntoDirSymlink(vc.data.LocalDirPath, false)
	if err != nil {
		return err
	}
	vc.data.DetectedTechnologies = make(map[Technology]bool)
	for _, file := range filesList {
		for _, indicator := range indicators {
			if indicator.Indicates(file) {
				vc.data.DetectedTechnologies[indicator.GetTechnology()] = true
				// Same file can't indicate on more than one technology.
				break
			}
		}
	}
	return
}

func createCredentials(serviceDetails *VcsServerDetails) (auth transport.AuthMethod) {
	var password, username string
	if serviceDetails.AccessToken != "" {
		password = serviceDetails.AccessToken
		// Authentication fails if the username string is empty. This can be anything except an empty string...
		username = "user"
	} else {
		password = serviceDetails.Password
		username = serviceDetails.User
	}
	return &http.BasicAuth{Username: username, Password: password}
}

func (vc *VcsCommand) getVcsCredentialsFromConsole() (err error) {
	for {
		ioutils.ScanFromConsole("Git project URL", &vc.data.VcsCredentials.Url, vc.defaultData.VcsCredentials.Url)
		print("Git Access token")
		byteToken, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			log.Error(err)
			continue
		}
		// New-line required after the access token input:
		fmt.Println()
		vc.data.VcsCredentials.AccessToken = string(byteToken)
		ioutils.ScanFromConsole("Git username", &vc.data.VcsCredentials.User, vc.defaultData.VcsCredentials.User)
		err = vc.prepareVcsData()
		if err != nil {
			log.Error(err)
		} else {
			return nil
		}
	}

}
