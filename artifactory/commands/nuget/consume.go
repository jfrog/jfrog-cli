package nuget

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"strings"

	gofrogcmd "github.com/jfrog/gofrog/io"
	"github.com/jfrog/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli/artifactory/utils/dotnet"
	"github.com/jfrog/jfrog-cli/artifactory/utils/dotnet/solution"
	"github.com/jfrog/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-client-go/auth"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

type dotnetCommandArgs struct {
	cmdType            dotnet.CmdType
	args               string
	flags              string
	repoName           string
	solutionPath       string
	buildConfiguration *utils.BuildConfiguration
	rtDetails          *config.ArtifactoryDetails
}

type NugetCommand struct {
	configFilePath string
	*dotnetCommandArgs
}

func NewNugetCommand() *NugetCommand {
	return &NugetCommand{"", &dotnetCommandArgs{cmdType: dotnet.Nuget}}
}

func (nc *NugetCommand) SetConfigFilePath(configFilePath string) *NugetCommand {
	nc.configFilePath = configFilePath
	return nc
}

func (dca *dotnetCommandArgs) SetRtDetails(rtDetails *config.ArtifactoryDetails) *dotnetCommandArgs {
	dca.rtDetails = rtDetails
	return dca
}

func (dca *dotnetCommandArgs) SetBuildConfiguration(buildConfiguration *utils.BuildConfiguration) *dotnetCommandArgs {
	dca.buildConfiguration = buildConfiguration
	return dca
}

func (dca *dotnetCommandArgs) SetSolutionPath(solutionPath string) *dotnetCommandArgs {
	dca.solutionPath = solutionPath
	return dca
}

func (dca *dotnetCommandArgs) SetRepoName(repoName string) *dotnetCommandArgs {
	dca.repoName = repoName
	return dca
}

func (dca *dotnetCommandArgs) SetFlags(flags string) *dotnetCommandArgs {
	dca.flags = flags
	return dca
}

func (dca *dotnetCommandArgs) SetArgs(args string) *dotnetCommandArgs {
	dca.args = args
	return dca
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
	RtDetails, err := resolveParams.RtDetails()
	if err != nil {
		return err
	}
	nc.SetRepoName(resolveParams.TargetRepo()).SetRtDetails(RtDetails)
	return nc.run()
}

// Exec all consume type nuget commands, install, update, add, restore.
func (dca *dotnetCommandArgs) run() error {
	log.Info("Running nuget...")
	// Use temp dir to save config file, the config will be removed at the end.
	tempDirPath, err := fileutils.CreateTempDir()
	if err != nil {
		return err
	}
	defer fileutils.RemoveTempDir(tempDirPath)

	dca.solutionPath, err = changeWorkingDir(dca.solutionPath)
	if err != nil {
		return err
	}

	err = dca.prepareAndRunCmd(tempDirPath)
	if err != nil {
		return err
	}

	isCollectBuildInfo := len(dca.buildConfiguration.BuildName) > 0 && len(dca.buildConfiguration.BuildNumber) > 0
	if !isCollectBuildInfo {
		return nil
	}

	slnFile := ""
	flags := strings.Split(dca.flags, " ")
	if len(flags) > 0 && strings.HasSuffix(flags[0], ".sln") {
		slnFile = flags[0]
	}
	sol, err := solution.Load(dca.solutionPath, slnFile)
	if err != nil {
		return err
	}

	if err = utils.SaveBuildGeneralDetails(dca.buildConfiguration.BuildName, dca.buildConfiguration.BuildNumber); err != nil {
		return err
	}
	buildInfo, err := sol.BuildInfo(dca.buildConfiguration.Module)
	if err != nil {
		return err
	}
	return utils.SaveBuildInfo(dca.buildConfiguration.BuildName, dca.buildConfiguration.BuildNumber, buildInfo)
}

func (dca *dotnetCommandArgs) RtDetails() (*config.ArtifactoryDetails, error) {
	return dca.rtDetails, nil
}

func (dca *dotnetCommandArgs) CommandName() string {
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
func (dca *dotnetCommandArgs) prepareAndRunCmd(configDirPath string) error {
	cmd, err := dca.createNugetCmd()
	if err != nil {
		return err
	}
	// To prevent NuGet prompting for credentials
	err = os.Setenv("NUGET_EXE_NO_PROMPT", "true")
	if err != nil {
		return errorutils.CheckError(err)
	}

	err = dca.prepareConfigFile(cmd, configDirPath)
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
func (dca *dotnetCommandArgs) prepareConfigFile(cmd *dotnet.Cmd, configDirPath string) error {
	cmdFlag := cmd.Type().GetTypeFlagPrefix() + "configfile"
	currentConfigPath, err := getFlagValueIfExists(cmdFlag, cmd)
	if err != nil {
		return err
	}
	if currentConfigPath != "" {
		return nil
	}

	cmdFlag = cmd.Type().GetTypeFlagPrefix() + "source"
	sourceCommandValue, err := getFlagValueIfExists(cmdFlag, cmd)
	if err != nil {
		return err
	}
	if sourceCommandValue != "" {
		return nil
	}

	err = dca.initNewConfig(cmd, configDirPath)
	return err
}

// Returns the value of the flag if exists
func getFlagValueIfExists(cmdFlag string, cmd *dotnet.Cmd) (string, error) {
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
func (dca *dotnetCommandArgs) initNewConfig(cmd *dotnet.Cmd, configDirPath string) error {
	// Got to here, means that neither of the flags provided and we need to init our own config.
	configFile, err := writeToTempConfigFile(cmd, configDirPath)
	if err != nil {
		return err
	}

	return dca.addNugetAuthenticationToNewConfig(cmd.Type(), configFile)
}

// Runs nuget add sources and setapikey commands to authenticate with Artifactory server
func (dca *dotnetCommandArgs) addNugetAuthenticationToNewConfig(cmdType dotnet.CmdType, configFile *os.File) error {
	sourceUrl, user, password, err := dca.getSourceDetails()
	if err != nil {
		return err
	}

	return addSourceToNugetConfig(cmdType, configFile.Name(), sourceUrl, user, password)
}

// Creates the temp file and writes the config template into the file for NuGet can use it.
func writeToTempConfigFile(cmd *dotnet.Cmd, tempDirPath string) (*os.File, error) {
	configFile, err := ioutil.TempFile(tempDirPath, "jfrog.cli.nuget.")
	if err != nil {
		return nil, errorutils.CheckError(err)
	}
	log.Debug("Nuget config file created at:", configFile.Name())

	defer configFile.Close()

	cmd.CommandFlags = append(cmd.CommandFlags, cmd.Type().GetTypeFlagPrefix()+"ConfigFile", configFile.Name())

	// Set Artifactory repo as source
	content := dotnet.ConfigFileTemplate
	_, err = configFile.WriteString(content)
	if err != nil {
		return nil, errorutils.CheckError(err)
	}
	return configFile, nil
}

// Runs nuget sources add command
func addSourceToNugetConfig(cmdType dotnet.CmdType, configFileName, sourceUrl, user, password string) error {
	cmd, err := dotnet.NewDotnetAddSourceCmd(cmdType)
	if err != nil {
		return err
	}

	flagPrefix := cmdType.GetTypeFlagPrefix()
	cmd.CommandFlags = append(cmd.CommandFlags, flagPrefix+"ConfigFile", configFileName)
	cmd.CommandFlags = append(cmd.CommandFlags, flagPrefix+"Name", sourceName)
	cmd.CommandFlags = append(cmd.CommandFlags, flagPrefix+"Source", sourceUrl)
	cmd.CommandFlags = append(cmd.CommandFlags, flagPrefix+"username", user)
	cmd.CommandFlags = append(cmd.CommandFlags, flagPrefix+"password", password)
	output, err := gofrogcmd.RunCmdOutput(cmd)
	log.Debug("Running command: Add sources. Output:", output)
	return err
}

func (dca *dotnetCommandArgs) getSourceDetails() (sourceURL, user, password string, err error) {
	var u *url.URL
	u, err = url.Parse(dca.rtDetails.Url)
	if errorutils.CheckError(err) != nil {
		return
	}
	u.Path = path.Join(u.Path, "api/nuget", dca.repoName)
	sourceURL = u.String()

	user = dca.rtDetails.User
	password = dca.rtDetails.Password
	// If access-token is defined, extract user from it.
	rtDetails, err := dca.RtDetails()
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

func (dca *dotnetCommandArgs) createNugetCmd() (*dotnet.Cmd, error) {
	c, err := dotnet.NewDotnetCmd(dca.cmdType)
	if err != nil {
		return nil, err
	}
	if dca.args != "" {
		c.Command, err = utils.ParseArgs(strings.Split(dca.args, " "))
		if err != nil {
			return nil, errorutils.CheckError(err)
		}
	}

	if dca.flags != "" {
		c.CommandFlags, err = utils.ParseArgs(strings.Split(dca.flags, " "))
	}

	return c, errorutils.CheckError(err)
}
