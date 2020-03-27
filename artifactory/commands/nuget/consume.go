package nuget

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"strings"

	gofrogcmd "github.com/jfrog/gofrog/io"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils/nuget"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils/nuget/solution"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/auth"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

type NugetCommandArgs struct {
	args               string
	flags              string
	repoName           string
	solutionPath       string
	buildConfiguration *utils.BuildConfiguration
	rtDetails          *config.ArtifactoryDetails
}

type NugetCommand struct {
	configFilePath string
	*NugetCommandArgs
}

func NewNugetCommand() *NugetCommand {
	return &NugetCommand{"", &NugetCommandArgs{}}
}

func (nc *NugetCommand) SetConfigFilePath(configFilePath string) *NugetCommand {
	nc.configFilePath = configFilePath
	return nc
}

func (nca *NugetCommandArgs) SetRtDetails(rtDetails *config.ArtifactoryDetails) *NugetCommandArgs {
	nca.rtDetails = rtDetails
	return nca
}

func (nca *NugetCommandArgs) SetBuildConfiguration(buildConfiguration *utils.BuildConfiguration) *NugetCommandArgs {
	nca.buildConfiguration = buildConfiguration
	return nca
}

func (nca *NugetCommandArgs) SetSolutionPath(solutionPath string) *NugetCommandArgs {
	nca.solutionPath = solutionPath
	return nca
}

func (nca *NugetCommandArgs) SetRepoName(repoName string) *NugetCommandArgs {
	nca.repoName = repoName
	return nca
}

func (nca *NugetCommandArgs) SetFlags(flags string) *NugetCommandArgs {
	nca.flags = flags
	return nca
}

func (nca *NugetCommandArgs) SetArgs(args string) *NugetCommandArgs {
	nca.args = args
	return nca
}

func (nc *NugetCommand) Run() error {
	// Read config file.
	log.Debug("Preparing to read the config file", nc.configFilePath)
	vConfig, err := utils.ReadConfigFile(nc.configFilePath, utils.YAML)
	if err != nil {
		return err
	}
	// Extract resolution params.
	resolveParams, err := utils.GetRepoConfigByPrefix(nc.configFilePath, utils.ProjectConfigResolverPrefix, vConfig)
	if err != nil {
		return err
	}
	filteredNugetArgs, buildConfiguration, err := utils.ExtractBuildDetailsFromArgs(strings.Split(nc.args, " "))
	RtDetails, err := resolveParams.RtDetails()
	if err != nil {
		return err
	}
	nc.SetArgs(strings.Join(filteredNugetArgs, " ")).
		SetRepoName(resolveParams.TargetRepo()).
		SetBuildConfiguration(buildConfiguration).
		SetRtDetails(RtDetails)
	return nc.run()
}

// Exec all consume type nuget commands, install, update, add, restore.
func (nca *NugetCommandArgs) run() error {
	log.Info("Running nuget...")
	// Use temp dir to save config file, the config will be removed at the end.
	tempDirPath, err := fileutils.CreateTempDir()
	if err != nil {
		return err
	}
	defer fileutils.RemoveTempDir(tempDirPath)

	nca.solutionPath, err = changeWorkingDir(nca.solutionPath)
	if err != nil {
		return err
	}

	err = nca.prepareAndRunCmd(tempDirPath)
	if err != nil {
		return err
	}

	isCollectBuildInfo := len(nca.buildConfiguration.BuildName) > 0 && len(nca.buildConfiguration.BuildNumber) > 0
	if !isCollectBuildInfo {
		return nil
	}

	slnFile := ""
	flags := strings.Split(nca.flags, " ")
	if len(flags) > 0 && strings.HasSuffix(flags[0], ".sln") {
		slnFile = flags[0]
	}
	sol, err := solution.Load(nca.solutionPath, slnFile)
	if err != nil {
		return err
	}

	if err = utils.SaveBuildGeneralDetails(nca.buildConfiguration.BuildName, nca.buildConfiguration.BuildNumber); err != nil {
		return err
	}
	buildInfo, err := sol.BuildInfo(nca.buildConfiguration.Module)
	if err != nil {
		return err
	}
	return utils.SaveBuildInfo(nca.buildConfiguration.BuildName, nca.buildConfiguration.BuildNumber, buildInfo)
}

func (nca *NugetCommandArgs) RtDetails() (*config.ArtifactoryDetails, error) {
	return nca.rtDetails, nil
}

func (nca *NugetCommandArgs) CommandName() string {
	return "rt_nuget"
}

const sourceName = "JFrogCli"

func DependencyTreeCmd() error {
	workspace, err := os.Getwd()
	if err != nil {
		return errorutils.CheckError(err)
	}

	sol, err := solution.Load(workspace, "")
	if err != nil {
		return err
	}

	// Create the tree for each project
	for _, project := range sol.GetProjects() {
		err = project.CreateDependencyTree()
		if err != nil {
			return err
		}
	}
	// Build the tree.
	content, err := sol.Marshal()
	if err != nil {
		return errorutils.CheckError(err)
	}
	log.Output(clientutils.IndentJson(content))
	return nil
}

// Changes the working directory if provided.
// Returns the path to the solution
func changeWorkingDir(newWorkingDir string) (string, error) {
	var err error
	if newWorkingDir != "" {
		err = os.Chdir(newWorkingDir)
	} else {
		newWorkingDir, err = os.Getwd()
	}

	return newWorkingDir, errorutils.CheckError(err)
}

// Prepares the nuget configuration file within the temp directory
// Runs NuGet itself with the arguments and flags provided.
func (nca *NugetCommandArgs) prepareAndRunCmd(configDirPath string) error {
	cmd, err := nca.createNugetCmd()
	if err != nil {
		return err
	}
	// To prevent NuGet prompting for credentials
	err = os.Setenv("NUGET_EXE_NO_PROMPT", "true")
	if err != nil {
		return errorutils.CheckError(err)
	}

	err = nca.prepareConfigFile(cmd, configDirPath)
	if err != nil {
		return err
	}
	err = gofrogcmd.RunCmd(cmd)
	if err != nil {
		return err
	}

	return nil
}

// Checks if the user provided input such as -configfile flag or -Source flag.
// If those flags provided, NuGet will use the provided configs (default config file or the one with -configfile)
// If neither provided, we are initializing our own config.
func (nca *NugetCommandArgs) prepareConfigFile(cmd *nuget.Cmd, configDirPath string) error {
	currentConfigPath, err := getFlagValueIfExists("-configfile", cmd)
	if err != nil {
		return err
	}
	if currentConfigPath != "" {
		return nil
	}

	sourceCommandValue, err := getFlagValueIfExists("-source", cmd)
	if err != nil {
		return err
	}
	if sourceCommandValue != "" {
		return nil
	}

	err = nca.initNewConfig(cmd, configDirPath)
	return err
}

// Returns the value of the flag if exists
func getFlagValueIfExists(cmdFlag string, cmd *nuget.Cmd) (string, error) {
	for i := 0; i < len(cmd.CommandFlags); i++ {
		if !strings.EqualFold(cmd.CommandFlags[i], cmdFlag) {
			continue
		}
		if i+1 == len(cmd.CommandFlags) {
			return "", errorutils.CheckError(errorutils.CheckError(fmt.Errorf(cmdFlag, " flag was provided without value")))
		}
		return cmd.CommandFlags[i+1], nil
	}

	return "", nil
}

// Initializing a new NuGet config file that NuGet will use into a temp file
func (nca *NugetCommandArgs) initNewConfig(cmd *nuget.Cmd, configDirPath string) error {
	// Got to here, means that neither of the flags provided and we need to init our own config.
	configFile, err := writeToTempConfigFile(cmd, configDirPath)
	if err != nil {
		return err
	}

	return nca.addNugetAuthenticationToNewConfig(configFile)
}

// Runs nuget add sources and setapikey commands to authenticate with Artifactory server
func (nca *NugetCommandArgs) addNugetAuthenticationToNewConfig(configFile *os.File) error {
	sourceUrl, user, password, err := nca.getSourceDetails()
	if err != nil {
		return err
	}

	err = addNugetSource(configFile.Name(), sourceUrl, user, password)
	if err != nil {
		return err
	}

	err = addNugetApiKey(user, password, configFile.Name())
	return err
}

// Creates the temp file and writes the config template into the file for NuGet can use it.
func writeToTempConfigFile(cmd *nuget.Cmd, tempDirPath string) (*os.File, error) {
	configFile, err := ioutil.TempFile(tempDirPath, "jfrog.cli.nuget.")
	if err != nil {
		return nil, errorutils.CheckError(err)
	}
	log.Debug("Nuget config file created at:", configFile.Name())

	defer configFile.Close()

	cmd.CommandFlags = append(cmd.CommandFlags, "-ConfigFile", configFile.Name())

	// Set Artifactory repo as source
	content := nuget.ConfigFileTemplate
	_, err = configFile.WriteString(content)
	if err != nil {
		return nil, errorutils.CheckError(err)
	}
	return configFile, nil
}

// Runs nuget sources add command
func addNugetSource(configFileName, sourceUrl, user, password string) error {
	cmd, err := nuget.NewNugetCmd()
	if err != nil {
		return err
	}

	sourceCommand := "sources"
	cmd.Command = append(cmd.Command, sourceCommand)
	cmd.CommandFlags = append(cmd.CommandFlags, "-ConfigFile", configFileName)
	cmd.CommandFlags = append(cmd.CommandFlags, "Add")
	cmd.CommandFlags = append(cmd.CommandFlags, "-Name", sourceName)
	cmd.CommandFlags = append(cmd.CommandFlags, "-Source", sourceUrl)
	cmd.CommandFlags = append(cmd.CommandFlags, "-username", user)
	cmd.CommandFlags = append(cmd.CommandFlags, "-password", password)
	output, err := gofrogcmd.RunCmdOutput(cmd)
	log.Debug("Running command: Add sources. Output:", output)
	return err
}

// Runs nuget setapikey command
func addNugetApiKey(user, password, configFileName string) error {
	cmd, err := nuget.NewNugetCmd()
	if err != nil {
		return err
	}

	cmd.Command = append(cmd.Command, "setapikey")
	cmd.CommandFlags = append(cmd.CommandFlags, user+":"+password)
	cmd.CommandFlags = append(cmd.CommandFlags, "-Source", sourceName)
	cmd.CommandFlags = append(cmd.CommandFlags, "-ConfigFile", configFileName)

	output, err := gofrogcmd.RunCmdOutput(cmd)
	log.Debug("Running command: SetApiKey. Output:", output)
	return err
}

func (nca *NugetCommandArgs) getSourceDetails() (sourceURL, user, password string, err error) {
	var u *url.URL
	u, err = url.Parse(nca.rtDetails.Url)
	if errorutils.CheckError(err) != nil {
		return
	}
	u.Path = path.Join(u.Path, "api/nuget", nca.repoName)
	sourceURL = u.String()

	user = nca.rtDetails.User
	password = nca.rtDetails.Password
	// If access-token is defined, extract user from it.
	rtDetails, err := nca.RtDetails()
	if errorutils.CheckError(err) != nil {
		return
	}
	if rtDetails.AccessToken != "" {
		log.Debug("Using access-token details for nuget authentication.")
		user, err = auth.ExtractUsernameFromAccessToken(rtDetails.AccessToken)
		if err != nil {
			return
		}
		password = rtDetails.AccessToken
	}
	return
}

func (nca *NugetCommandArgs) createNugetCmd() (*nuget.Cmd, error) {
	c, err := nuget.NewNugetCmd()
	if err != nil {
		return nil, err
	}
	if nca.args != "" {
		c.Command, err = utils.ParseEnvVars(strings.Split(nca.args, " "))
		if err != nil {
			return nil, errorutils.CheckError(err)
		}
	}

	if nca.flags != "" {
		c.CommandFlags, err = utils.ParseEnvVars(strings.Split(nca.flags, " "))
	}

	return c, errorutils.CheckError(err)
}
