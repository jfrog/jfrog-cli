package commands

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/gookit/color"
	"github.com/jfrog/jfrog-cli-core/artifactory/commands/permissiontarget"
	"github.com/jfrog/jfrog-cli-core/artifactory/commands/usersmanagement"
	"github.com/jfrog/jfrog-cli-core/general/cisetup"
	"github.com/jfrog/jfrog-client-go/pipelines"
	pipelinesservices "github.com/jfrog/jfrog-client-go/pipelines/services"
	"github.com/jfrog/jfrog-client-go/utils"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/jfrog/jfrog-cli-core/artifactory/commands/buildinfo"
	rtutils "github.com/jfrog/jfrog-cli-core/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/common/commands"
	corecommoncommands "github.com/jfrog/jfrog-cli-core/common/commands"
	"github.com/jfrog/jfrog-cli-core/utils/config"
	utilsconfig "github.com/jfrog/jfrog-cli-core/utils/config"
	"github.com/jfrog/jfrog-cli-core/utils/coreutils"
	"github.com/jfrog/jfrog-cli-core/utils/ioutils"
	buildinfocmd "github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	clientConfig "github.com/jfrog/jfrog-client-go/config"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	xrayservices "github.com/jfrog/jfrog-client-go/xray/services"
	xrayutils "github.com/jfrog/jfrog-client-go/xray/services/utils"
	"golang.org/x/crypto/ssh/terminal"
)

const (
	VcsConfigFile            = "jfrog-cli-vcs.conf"
	DefaultFirstBuildNumber  = "0"
	DefaultWorkspace         = "./jfrog-vcs-workspace"
	pipelineUiPath           = "ui/pipelines/myPipelines/default/"
	permissionTargetName     = "jfrog-ide-developer-pt"
	permissionTargetTemplate = `{"build":{"include-patterns":"**","actions-groups":{"%s":"read"}},"name":"%s"}`
	pttFileName              = "ci-setup-ptt"
	ideGroupName             = "jfrog-ide-developer-group"
	ideUserName              = "ide-developer"
	ideUserPassPlaceholder   = "<INSERT-PASSWORD>"
	ideUserEmailPlaceholder  = "<INSERT-EMAIL>"
	createUserTemplate       = `jfrog rt user-create "%s" "%s" "%s" --users-groups="%s" --server-id="%s"`
)

type CiSetupCommand struct {
	defaultData *cisetup.CiSetupData
	data        *cisetup.CiSetupData
}

func RunCiSetupCmd() error {
	cc := &CiSetupCommand{}
	err := logBeginningInstructions()
	if err != nil {
		return err
	}
	err = cc.prepareConfigurationData()
	if err != nil {
		return err
	}
	err = cc.Run()
	if err != nil {
		return err
	}
	return saveVcsConf(cc.data)
}

func logBeginningInstructions() error {
	instructions := []string{
		"",
		colorTitle("About this command"),
		"This command sets up a basic CI pipeline which uses the JFrog Platform.",
		"It currently supports Maven, Gradle and npm, but additional package managers will be added in the future.",
		"The generated CI pipeline is based on JFrog Pipelines, but additional CI providers will be added in the future.",
		"The command takes care of configuring JFrog Artifactory, JFrog Xray and JFrog Pipelines for you.",
		"",
		colorTitle("Important"),
		" 1. If you don't have a JFrog Platform instance with admin credentials, head over to https://jfrog.com/start-free/ and get one for free.",
		" 2. When asked to provide credentials for your JFrog Platform and Git provider, please make sure the credentials have admin privileges.",
		" 3. You can exit the command by hitting 'control + C' at any time. The values you provided before exiting are saved (with the exception of passwords and tokens) and will be set as defaults the next time you run the command.",
		"", "",
	}
	return writeToScreen(strings.Join(instructions, "\n"))
}

func logActivatePipelinesInstructions() error {
	instructions := []string{
		"",
		colorTitle("Free"),
		"",
	}
	return writeToScreen(strings.Join(instructions, "\n"))
}

func colorTitle(title string) string {
	if terminal.IsTerminal(int(os.Stderr.Fd())) {
		return color.Green.Render(title)
	}
	return title
}

func (cc *CiSetupCommand) prepareConfigurationData() error {
	// If data is nil, initialize a new one
	if cc.data == nil {
		cc.data = new(cisetup.CiSetupData)
	}

	// Get previous vcs data if exists
	defaultData, err := readVcsConf()
	cc.defaultData = defaultData
	return err
}

func readVcsConf() (*cisetup.CiSetupData, error) {
	conf := &cisetup.CiSetupData{}
	path, err := coreutils.GetJfrogHomeDir()
	if err != nil {
		return nil, err
	}
	confPath := filepath.Join(path, VcsConfigFile)
	exists, err := fileutils.IsFileExists(confPath, false)
	if err != nil {
		return nil, err
	}
	if !exists {
		return conf, nil
	}
	configFile, err := fileutils.ReadFile(confPath)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(configFile, conf)
	return conf, errorutils.CheckError(err)
}

func saveVcsConf(conf *cisetup.CiSetupData) error {
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
	return errorutils.CheckError(err)
}

func (cc *CiSetupCommand) Run() error {
	// Run JFrog config command
	err := runConfigCmd()
	if err != nil {
		return err
	}
	// Basic VCS questionnaire (URLs, Credentials, etc'...)
	err = cc.gitPhase()
	err = saveIfNoError(err, cc.data)
	if err != nil {
		return err
	}
	// Ask the user which CI he tries to setup
	err = cc.desiredOutputPhase()
	err = saveIfNoError(err, cc.data)
	if err != nil {
		return err
	}
	// Interactively create Artifactory repository based on the detected technologies and on going user input
	err = cc.artifactoryConfigPhase()
	err = saveIfNoError(err, cc.data)
	if err != nil {
		return err
	}
	// Publish empty build info.
	err = cc.publishFirstBuild()
	err = saveIfNoError(err, cc.data)
	if err != nil {
		return err
	}
	// Configure Xray to scan the new build.
	err = cc.xrayConfigPhase()
	err = saveIfNoError(err, cc.data)
	if err != nil {
		return err
	}
	var ciFileName string
	switch cc.data.CiType {
	case cisetup.Pipelines:
		// Configure pipelines, create and stage pipelines.yml.
		ciFileName, err = cc.runPipelinesPhase()
		if err != nil {
			return err
		}
	case cisetup.Jenkins:
		// Create and stage Jenkinsfile.
		ciFileName, err = cc.runJenkinsPhase()
		if err != nil {
			return err
		}
	case cisetup.GithubActions:
		// Create and stage main.yml.
		ciFileName, err = cc.runGithubActionsPhase()
		if err != nil {
			return err
		}

	}
	// Create group and permission target if needed.
	err = runIdePhase()
	if err != nil {
		return err
	}
	return cc.logCompletionInstruction(ciFileName)
}

func saveIfNoError(errCheck error, conf *cisetup.CiSetupData) error {
	if errCheck != nil {
		return errCheck
	}
	return saveVcsConf(conf)
}

func runIdePhase() error {
	serverDetails, err := utilsconfig.GetSpecificConfig(cisetup.ConfigServerId, false, false)
	if err != nil {
		return err
	}
	err = createGroup(serverDetails)
	if err != nil {
		return err
	}
	return createPermissionTarget(serverDetails)
}

func createGroup(serverDetails *utilsconfig.ServerDetails) error {
	log.Info("Creating group...")
	groupCreateCmd := usersmanagement.NewGroupCreateCommand()
	groupCreateCmd.SetName(ideGroupName).SetServerDetails(serverDetails).SetReplaceIfExists(false)
	err := groupCreateCmd.Run()
	if err != nil {
		if _, ok := err.(*services.GroupAlreadyExistsError); !ok {
			return err
		}
		log.Debug("Group already exists, skipping...")
	}
	return nil
}

func createPermissionTarget(serverDetails *utilsconfig.ServerDetails) error {
	ptTemplate := fmt.Sprintf(permissionTargetTemplate, ideGroupName, permissionTargetName)
	tempDir, err := fileutils.CreateTempDir()
	if err != nil {
		return err
	}
	pttPath := filepath.Join(tempDir, pttFileName)
	err = ioutil.WriteFile(pttPath, []byte(ptTemplate), 0600)
	if err != nil {
		return err
	}
	permissionTargetCreateCmd := permissiontarget.NewPermissionTargetCreateCommand()
	permissionTargetCreateCmd.SetTemplatePath(pttPath).SetServerDetails(serverDetails).SetVars("")
	err = permissionTargetCreateCmd.Run()
	if err != nil {
		if _, ok := err.(*services.PermissionTargetAlreadyExistsError); !ok {
			return err
		}
		log.Debug("Permission target already exists, skipping...")
	}
	return nil
}

func writeToScreen(content string) error {
	_, err := fmt.Fprint(os.Stderr, content)
	return errorutils.CheckError(err)
}

func getPipelinesToken() (string, error) {
	var err error
	var byteToken []byte
	for len(byteToken) == 0 {
		err = writeToScreen("Please provide a JFrog Pipelines admin token (To generate the token, " +
			"log into the JFrog Platform UI --> Administration --> Identity and Access --> Access Tokens --> Generate Admin Token): ")
		if err != nil {
			return "", err
		}
		byteToken, err = terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return "", errorutils.CheckError(err)
		}
		// New-line required after the access token input:
		fmt.Println()
	}
	return string(byteToken), nil
}

func runConfigCmd() (err error) {
	for {
		configCmd := corecommoncommands.NewConfigCommand().SetInteractive(true).SetServerId(cisetup.ConfigServerId).SetEncPassword(true)
		err = configCmd.Config()
		if err != nil {
			log.Error(err)
			continue
		}
		// Validate JFrog credentials by execute get repo command
		serviceDetails, err := utilsconfig.GetSpecificConfig(cisetup.ConfigServerId, false, false)
		if err != nil {
			return err
		}
		_, err = GetAllRepos(serviceDetails, "", "")
		if err == nil {
			return nil
		}
		log.Error(err)
	}
}

func (cc *CiSetupCommand) runJenkinsPhase() (string, error) {
	return "", nil
}

func (cc *CiSetupCommand) runGithubActionsPhase() (string, error) {
	generator := cisetup.JFrogPipelinesYamlGenerator{
		SetupData: cc.data,
	}
	pipelinesYamlBytes, pipelineName, err := generator.Generate()
	if err != nil {
		return "", err
	}

	err = cc.saveCiConfigToFile(pipelinesYamlBytes, cisetup.PipelinesYamlName)
	if err != nil {
		return "", err
	}
	err = cc.stageCiConfigFile(cisetup.PipelinesYamlName)
	if err != nil {
		return "", err
	}
	return pipelineName, nil
}

func (cc *CiSetupCommand) runPipelinesPhase() (string, error) {
	var vcsIntName string
	var rtIntName string
	var err error
	configurator := cisetup.JFrogPipelinesConfigurator{
		SetupData: cc.data, PipelinesToken: "",
	}
	// Ask for token and config pipelines. Run again if authentication problem.
	for {
		// Ask for pipelines token.
		configurator.PipelinesToken, err = getPipelinesToken()
		if err != nil {
			return "", err
		}
		// Run Pipelines setup
		vcsIntName, rtIntName, err = configurator.Config()
		// If no error occurred, continue with flow. Elseif unauthorized error, ask for token again.
		if err == nil {
			break
		}
		if _, ok := err.(*pipelinesservices.IntegrationUnauthorizedError); !ok {
			return "", err
		}
		log.Debug(err.Error())
		log.Info("There seems to be an authorization problem with the pipelines token you entered. Please try again.")
	}
	generator := cisetup.JFrogPipelinesYamlGenerator{
		VcsIntName: vcsIntName,
		RtIntName:  rtIntName,
		SetupData:  cc.data,
	}
	pipelinesYamlBytes, pipelineName, err := generator.Generate()
	if err != nil {
		return "", err
	}

	err = cc.saveCiConfigToFile(pipelinesYamlBytes, cisetup.PipelinesYamlName)
	if err != nil {
		return "", err
	}
	err = cc.stageCiConfigFile(cisetup.PipelinesYamlName)
	if err != nil {
		return "", err
	}
	return pipelineName, nil
}

func (cc *CiSetupCommand) saveCiConfigToFile(yaml []byte, fileName string) error {
	path := filepath.Join(cc.data.LocalDirPath, fileName)
	log.Info(fmt.Sprintf("Generating %s at: %q ...", fileName, path))
	return ioutil.WriteFile(path, yaml, 0644)
}

func (cc *CiSetupCommand) logCompletionInstruction(pipelineName string) error {
	serviceDetails, err := utilsconfig.GetSpecificConfig(cisetup.ConfigServerId, false, false)
	if err != nil {
		return err
	}

	instructions := []string{
		"", colorTitle("Completing the setup"),
		"We configured the JFrog Platform and generated a pipelines.yml for you.",
		"To complete the setup, add the new pipelines.yml to your git repository by running the following commands:", "",
		"cd " + cc.data.LocalDirPath,
		"git commit -m \"Add pipelines.yml\"",
		"git push", "",
		"Although your pipeline is configured, it hasn't run yet.",
		"It will run and become visible in the following URL, after the next git commit:",
		getPipelineUiPath(serviceDetails.Url, pipelineName), "",
		colorTitle("Allowing developers to access this pipeline from their IDE"),
		"You have the option of viewing the new pipeline's runs from within IntelliJ IDEA. More IDEs will be supported in the future.",
		"To achieve this, follow these steps:",
		" 1. Make sure the latest version of JFrog Plugin is installed on IntelliJ IDEA.",
		" 2. Create a JFrog user for the IDE by running the following command:", "",
		"    " + fmt.Sprintf(createUserTemplate, ideUserName, ideUserPassPlaceholder, ideUserEmailPlaceholder, ideGroupName, cisetup.ConfigServerId), "",
		" 3. In IDEA, under 'JFrog Global Configuration', set the JFrog Platform URL and the user you created.",
		" 4. In IDEA, under 'JFrog CI Integration', set * as the 'Build name pattern'.",
		" 5. In IDEA, open the 'JFrog' panel at the bottom of the screen, choose the 'CI' tab to see the CI information.", "",
	}
	return writeToScreen(strings.Join(instructions, "\n"))
}

func getPipelineUiPath(pipelinesUrl, pipelineName string) string {
	return utils.AddTrailingSlashIfNeeded(pipelinesUrl) + pipelineUiPath + pipelineName
}

func (cc *CiSetupCommand) publishFirstBuild() (err error) {
	println("Everytime the new pipeline builds the code, it generates a build entity (also known as build-info) and stores it in Artifactory.")
	ioutils.ScanFromConsole("Please choose a name for the build", &cc.data.BuildName, "${vcs.repo.name}-${branch}")
	cc.data.BuildName = strings.Replace(cc.data.BuildName, "${vcs.repo.name}", cc.data.RepositoryName, -1)
	cc.data.BuildName = strings.Replace(cc.data.BuildName, "${branch}", cc.data.GitBranch, -1)
	// Run BAG Command (in order to publish the first, empty, build info)
	buildAddGitConfigurationCmd := buildinfo.NewBuildAddGitCommand().SetDotGitPath(cc.data.LocalDirPath).SetServerId(cisetup.ConfigServerId) //.SetConfigFilePath(c.String("config"))
	buildConfiguration := rtutils.BuildConfiguration{BuildName: cc.data.BuildName, BuildNumber: DefaultFirstBuildNumber}
	buildAddGitConfigurationCmd = buildAddGitConfigurationCmd.SetBuildConfiguration(&buildConfiguration)
	log.Info("Generating an initial build-info...")
	err = commands.Exec(buildAddGitConfigurationCmd)
	if err != nil {
		return err
	}
	// Run BP Command.
	serviceDetails, err := utilsconfig.GetSpecificConfig(cisetup.ConfigServerId, false, false)
	if err != nil {
		return err
	}
	buildInfoConfiguration := buildinfocmd.Configuration{DryRun: false}
	buildPublishCmd := buildinfo.NewBuildPublishCommand().SetServerDetails(serviceDetails).SetBuildConfiguration(&buildConfiguration).SetConfig(&buildInfoConfiguration)
	return commands.Exec(buildPublishCmd)
}

func (cc *CiSetupCommand) xrayConfigPhase() (err error) {
	serviceDetails, err := utilsconfig.GetSpecificConfig(cisetup.ConfigServerId, false, false)
	if err != nil {
		return err
	}
	xrayManager, err := CreateXrayServiceManager(serviceDetails)
	if err != nil {
		return err
	}
	// Index the build.
	buildsToIndex := []string{cc.data.BuildName}
	err = xrayManager.AddBuildsToIndexing(buildsToIndex)
	// Create new default policy.
	policyParams := xrayutils.NewPolicyParams()
	policyParams.Name = "ci-pipeline-security-policy"
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
	// Create new default watcher.
	watchParams := xrayutils.NewWatchParams()
	watchParams.Name = "ci-pipeline-watch-all"
	watchParams.Description = "CI Pipeline Build Watch"
	watchParams.Active = true
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

func (cc *CiSetupCommand) artifactoryConfigPhase() (err error) {
	err = cc.printDetectedTechs()
	if err != nil {
		return err
	}
	atLeastOneTechBuilt := false
	cc.data.BuiltTechnologies = make(map[cisetup.Technology]*cisetup.TechnologyInfo) // todo is pointer init
	// First create repositories for each technology in Artifactory according to user input
	for tech, detected := range cc.data.DetectedTechnologies {
		if detected && coreutils.AskYesNo(fmt.Sprintf("Would you like to use %s to build the code?", tech), true) {
			cc.data.BuiltTechnologies[tech] = &cisetup.TechnologyInfo{}
			err = cc.interactivelyCreateRepos(tech)
			if err != nil {
				return
			}
			cc.getBuildCmd(tech)
			atLeastOneTechBuilt = true
		} else {
			cc.data.DetectedTechnologies[tech] = false
		}
	}
	if !atLeastOneTechBuilt {
		return errorutils.CheckError(errors.New("at least one of the supported technologies is expected to be chosen for building"))
	}
	return nil
}

func (cc *CiSetupCommand) printDetectedTechs() error {
	var techs []string
	for tech, detected := range cc.data.DetectedTechnologies {
		if detected {
			techs = append(techs, string(tech))
		}
	}
	if len(techs) == 0 {
		return errorutils.CheckError(errors.New("no supported technology was found in the project"))
	}
	log.Info("The next step is to provide the commands to build your code. It looks like the code can be built with " + getExplicitTechsListByNumber(techs) + ".")
	return nil
}

// Get the explicit list of technologies, for ex: "maven, gradle and npm"
func getExplicitTechsListByNumber(techs []string) string {
	if len(techs) == 1 {
		return techs[0]
	}
	return strings.Join(techs[0:len(techs)-1], ", ") + " and " + techs[len(techs)-1]
}

func (cc *CiSetupCommand) getBuildCmd(tech cisetup.Technology) {
	defaultBuildCmd := buildCmdByTech[tech]
	if info, built := cc.defaultData.BuiltTechnologies[tech]; built {
		if info.BuildCmd != "" {
			defaultBuildCmd = info.BuildCmd
		}
	}
	// Ask for working build command.
	prompt := "Please provide a single-line " + string(tech) + " build command."
	ioutils.ScanFromConsole(prompt, &cc.data.BuiltTechnologies[tech].BuildCmd, defaultBuildCmd)
}

func (cc *CiSetupCommand) interactivelyCreateRepos(technologyType cisetup.Technology) (err error) {
	serviceDetails, err := utilsconfig.GetSpecificConfig(cisetup.ConfigServerId, false, false)
	if err != nil {
		return err
	}
	// Get all relevant remotes to choose from
	remoteRepos, err := GetAllRepos(serviceDetails, Remote, string(technologyType))
	if err != nil {
		return err
	}
	shouldPromptSelection := len(*remoteRepos) > 0
	var remoteRepo string
	if shouldPromptSelection {
		// Ask if the user would like us to create a new remote or to choose from the exist repositories list
		remoteRepo, err = promptARepoSelection(remoteRepos, "Select remote repository")
		if err != nil {
			return err
		}
	} else {
		remoteRepo = NewRepository
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
				for {
					// Create a new virtual repository as well
					ioutils.ScanFromConsole(fmt.Sprintf("Choose a name for a new virtual repository which will include %q remote repo", remoteRepo),
						&repoName, GetVirtualDefaultName(technologyType))
					err = CreateVirtualRepo(serviceDetails, technologyType, repoName, remoteRepo)
					if err != nil {
						log.Error(err)
					} else {
						// we created both remote and virtual repositories successfully
						cc.data.BuiltTechnologies[technologyType].VirtualRepo = repoName
						return
					}
				}
			}
		}
	}
	// Else, the user choose an existing remote repo
	virtualRepos, err := GetAllRepos(serviceDetails, Virtual, string(technologyType))
	if err != nil {
		return err
	}
	shouldPromptSelection = len(*virtualRepos) > 0
	var virtualRepo string
	if shouldPromptSelection {
		// Ask if the user would like us to create a new virtual or to choose from the exist repositories list
		virtualRepo, err = promptARepoSelection(virtualRepos, fmt.Sprintf("Select a virtual repository, which includes %s or choose to create a new repo:", remoteRepo))
		if err != nil {
			return err
		}
	} else {
		virtualRepo = NewRepository
	}
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
		chosenVirtualRepo, err := GetVirtualRepo(serviceDetails, virtualRepo)
		if err != nil {
			return err
		}
		if !contains(chosenVirtualRepo.Repositories, remoteRepo) {
			log.Error(fmt.Sprintf("The chosen virtual repo %q does not contain the chosen remote repo %q", virtualRepo, remoteRepo))
			return cc.interactivelyCreateRepos(technologyType)
		}
	}
	// Saves the new created repo name (key) in the results data structure.
	cc.data.BuiltTechnologies[technologyType].VirtualRepo = virtualRepo
	return
}

func promptARepoSelection(repoDetails *[]services.RepositoryDetails, promptMsg string) (selectedRepoName string, err error) {

	selectableItems := []ioutils.PromptItem{{Option: NewRepository, TargetValue: &selectedRepoName}}
	for _, repo := range *repoDetails {
		selectableItems = append(selectableItems, ioutils.PromptItem{Option: repo.Key, TargetValue: &selectedRepoName, DefaultValue: repo.Url})
	}
	println(promptMsg)
	err = ioutils.SelectString(selectableItems, "", func(item ioutils.PromptItem) {
		*item.TargetValue = item.Option
	})
	return
}

func promptGitProviderSelection() (selected string, err error) {
	gitProviders := []cisetup.GitProvider{
		cisetup.Github,
		cisetup.GithubEnterprise,
		cisetup.Bitbucket,
		cisetup.BitbucketServer,
		cisetup.Gitlab,
	}

	var selectableItems []ioutils.PromptItem
	for _, provider := range gitProviders {
		selectableItems = append(selectableItems, ioutils.PromptItem{Option: string(provider), TargetValue: &selected})
	}
	println("Choose your project Git provider:")
	err = ioutils.SelectString(selectableItems, "", func(item ioutils.PromptItem) {
		*item.TargetValue = item.Option
	})
	return
}

func promptCiTypeSelection() (selected string, err error) {
	ciTypes := []cisetup.CiType{
		cisetup.Pipelines,
		cisetup.Jenkins,
		cisetup.GithubActions,
	}

	var selectableItems []ioutils.PromptItem
	for _, ci := range ciTypes {
		selectableItems = append(selectableItems, ioutils.PromptItem{Option: string(ci), TargetValue: &selected})
	}
	println("Choose the desired CI type:")
	err = ioutils.SelectString(selectableItems, "", func(item ioutils.PromptItem) {
		*item.TargetValue = item.Option
	})
	return
}

func (cc *CiSetupCommand) prepareVcsData() (err error) {
	cc.data.LocalDirPath = DefaultWorkspace
	for {
		err = fileutils.CreateDirIfNotExist(cc.data.LocalDirPath)
		if err != nil {
			return err
		}
		dirEmpty, err := fileutils.IsDirEmpty(cc.data.LocalDirPath)
		if err != nil {
			return err
		}
		if dirEmpty {
			break
		} else {
			log.Error("The '" + cc.data.LocalDirPath + "' directory isn't empty.")
			ioutils.ScanFromConsole("Choose a name for a directory to be used as the command's workspace", &cc.data.LocalDirPath, "")
		}

	}
	err = cc.cloneProject()
	if err != nil {
		return
	}
	err = cc.detectTechnologies()
	return
}

func (cc *CiSetupCommand) cloneProject() (err error) {
	// Create the desired path if necessary
	err = os.MkdirAll(cc.data.LocalDirPath, os.ModePerm)
	if err != nil {
		return errorutils.CheckError(err)
	}
	cloneOption := &git.CloneOptions{
		URL:           cc.data.VcsCredentials.Url,
		Auth:          createCredentials(&cc.data.VcsCredentials),
		ReferenceName: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", cc.data.GitBranch)),
		// Enable git submodules clone if there any.
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	}
	err = cc.extractRepositoryName()
	if err != nil {
		return
	}
	// Clone the given repository to the given directory from the given branch
	log.Info(fmt.Sprintf("Cloning project %q from: %q into: %q", cc.data.RepositoryName, cc.data.VcsCredentials.Url, cc.data.LocalDirPath))
	_, err = git.PlainClone(cc.data.LocalDirPath, false, cloneOption)
	return errorutils.CheckError(err)
}

func (cc *CiSetupCommand) stageCiConfigFile(ciFileName string) error {
	log.Info(fmt.Sprintf("Staging %s for git commit...", ciFileName))
	repo, err := git.PlainOpen(cc.data.LocalDirPath)
	if err != nil {
		return errorutils.CheckError(err)
	}
	worktree, err := repo.Worktree()
	if err != nil {
		return errorutils.CheckError(err)
	}
	_, err = worktree.Add(ciFileName)
	return errorutils.CheckError(err)
}

func (cc *CiSetupCommand) extractRepositoryName() error {
	vcsUrl := cc.data.VcsCredentials.Url
	if vcsUrl == "" {
		return errorutils.CheckError(errors.New("vcs URL should not be empty"))
	}
	// Trim trailing "/" if one exists
	vcsUrl = strings.TrimSuffix(vcsUrl, "/")
	cc.data.VcsCredentials.Url = vcsUrl

	// Split vcs url.
	splitUrl := strings.Split(vcsUrl, "/")
	if len(splitUrl) < 3 {
		return errorutils.CheckError(errors.New("unexpected URL. URL is expected to contain the git provider URL, domain and repository names"))
	}
	cc.data.RepositoryName = strings.TrimSuffix(splitUrl[len(splitUrl)-1], ".git")
	cc.data.ProjectDomain = splitUrl[len(splitUrl)-2]
	cc.data.VcsBaseUrl = strings.Join(splitUrl[:len(splitUrl)-2], "/")
	return nil
}

func (cc *CiSetupCommand) detectTechnologies() (err error) {
	indicators := cisetup.GetTechIndicators()
	filesList, err := fileutils.ListFilesRecursiveWalkIntoDirSymlink(cc.data.LocalDirPath, false)
	if err != nil {
		return
	}
	cc.data.DetectedTechnologies = make(map[cisetup.Technology]bool)
	for _, file := range filesList {
		for _, indicator := range indicators {
			if indicator.Indicates(file) {
				cc.data.DetectedTechnologies[indicator.GetTechnology()] = true
				// Same file can't indicate more than one technology.
				break
			}
		}
	}
	return
}

func createCredentials(serviceDetails *cisetup.VcsServerDetails) (auth transport.AuthMethod) {
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

func (cc *CiSetupCommand) gitPhase() (err error) {
	for {
		gitProvider, err := promptGitProviderSelection()
		if err != nil {
			log.Error(err)
			continue
		}
		cc.data.GitProvider = cisetup.GitProvider(gitProvider)
		ioutils.ScanFromConsole("Git project URL", &cc.data.VcsCredentials.Url, cc.defaultData.VcsCredentials.Url)
		ioutils.ScanFromConsole("Git username", &cc.data.VcsCredentials.User, cc.defaultData.VcsCredentials.User)
		err = writeToScreen("Git access token (requires admin permissions): ")
		if err != nil {
			return err
		}
		byteToken, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			log.Error(err)
			continue
		}
		// New-line required after the access token input:
		fmt.Println()
		cc.data.VcsCredentials.AccessToken = string(byteToken)
		ioutils.ScanFromConsole("Git branch", &cc.data.GitBranch, cc.defaultData.GitBranch)
		err = cc.prepareVcsData()
		if err != nil {
			log.Error(err)
		} else {
			return nil
		}
	}
}

func (cc *CiSetupCommand) desiredOutputPhase() (err error) {
	var ciType string
	for {
		ciType, err = promptCiTypeSelection()
		if err != nil {
			log.Error(err)
			continue
		}
		if ciType == cisetup.Pipelines {
			// validate that pipelines is aviable
			serviceDetails, err := config.GetSpecificConfig(cisetup.ConfigServerId, false, false)
			if err != nil {
				log.Error(err)
				continue
			}
			pipelinesDetails := *serviceDetails
			pipelinesDetails.AccessToken = ""
			pipelinesDetails.User = ""
			pipelinesDetails.Password = ""
			pipelinesDetails.ApiKey = ""

			pAuth, err := pipelinesDetails.CreatePipelinesAuthConfig()
			if err != nil {
				log.Error(err)
				continue
			}
			serviceConfig, err := clientConfig.NewConfigBuilder().
				SetServiceDetails(pAuth).
				SetDryRun(false).
				Build()
			if err != nil {
				log.Error(err)
				continue
			}
			pipelinesMgr, err := pipelines.New(serviceConfig)
			if err != nil {
				log.Error(err)
				continue
			}
			_, err = pipelinesMgr.GetSystemInfo()
			if err == nil {
				cc.data.CiType = cisetup.CiType(ciType)
				return nil
			}
			if _, ok := err.(*pipelinesservices.PipelinesNotAvailableError); ok {
				logActivatePipelinesInstructions()
			}
			log.Error(err)
		} else { // The user doesn't choose Pipelines.
			cc.data.CiType = cisetup.CiType(ciType)
			return nil
		}
	}
}
