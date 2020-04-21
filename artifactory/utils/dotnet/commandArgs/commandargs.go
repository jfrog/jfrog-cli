package commandArgs

import (
	"fmt"
	"github.com/jfrog/gofrog/io"
	"github.com/jfrog/jfrog-cli/artifactory/utils"
	dotnet "github.com/jfrog/jfrog-cli/artifactory/utils/dotnet"
	"github.com/jfrog/jfrog-cli/artifactory/utils/dotnet/solution"
	"github.com/jfrog/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-client-go/auth"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"strings"
)

type DotnetCommandArgs struct {
	toolchainType      dotnet.ToolchainType
	args               string
	flags              string
	repoName           string
	solutionPath       string
	buildConfiguration *utils.BuildConfiguration
	rtDetails          *config.ArtifactoryDetails
}

func (dca *DotnetCommandArgs) SetRtDetails(rtDetails *config.ArtifactoryDetails) *DotnetCommandArgs {
	dca.rtDetails = rtDetails
	return dca
}

func (dca *DotnetCommandArgs) SetBuildConfiguration(buildConfiguration *utils.BuildConfiguration) *DotnetCommandArgs {
	dca.buildConfiguration = buildConfiguration
	return dca
}

func (dca *DotnetCommandArgs) SetToolchainType(toolchainType dotnet.ToolchainType) *DotnetCommandArgs {
	dca.toolchainType = toolchainType
	return dca
}

func (dca *DotnetCommandArgs) SetSolutionPath(solutionPath string) *DotnetCommandArgs {
	dca.solutionPath = solutionPath
	return dca
}

func (dca *DotnetCommandArgs) SetRepoName(repoName string) *DotnetCommandArgs {
	dca.repoName = repoName
	return dca
}

func (dca *DotnetCommandArgs) SetFlags(flags string) *DotnetCommandArgs {
	dca.flags = flags
	return dca
}

func (dca *DotnetCommandArgs) SetArgs(args string) *DotnetCommandArgs {
	dca.args = args
	return dca
}

func (dca *DotnetCommandArgs) RtDetails() (*config.ArtifactoryDetails, error) {
	return dca.rtDetails, nil
}

func (dca *DotnetCommandArgs) CommandName() string {
	return "rt_" + dca.toolchainType.String()
}

// Exec all consume type nuget commands, install, update, add, restore.
func (dca *DotnetCommandArgs) Exec() error {
	log.Info("Running " + dca.toolchainType.String() + "...")
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

const SourceName = "JFrogCli"

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
func (dca *DotnetCommandArgs) prepareAndRunCmd(configDirPath string) error {
	cmd, err := dca.createToolchainCmd()
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
	err = io.RunCmd(cmd)
	if err != nil {
		return err
	}

	return nil
}

// Checks if the user provided input such as -configfile flag or -Source flag.
// If those flags provided, NuGet will use the provided configs (default config file or the one with -configfile)
// If neither provided, we are initializing our own config.
func (dca *DotnetCommandArgs) prepareConfigFile(cmd *dotnet.Cmd, configDirPath string) error {
	cmdFlag := cmd.Toolchain().GetTypeFlagPrefix() + "configfile"
	currentConfigPath, err := getFlagValueIfExists(cmdFlag, cmd)
	if err != nil {
		return err
	}
	if currentConfigPath != "" {
		return nil
	}

	cmdFlag = cmd.Toolchain().GetTypeFlagPrefix() + "source"
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
func (dca *DotnetCommandArgs) initNewConfig(cmd *dotnet.Cmd, configDirPath string) error {
	// Got to here, means that neither of the flags provided and we need to init our own config.
	configFile, err := writeToTempConfigFile(cmd, configDirPath)
	if err != nil {
		return err
	}

	return dca.addNugetAuthenticationToNewConfig(cmd.Toolchain(), configFile)
}

// Runs nuget add sources and setapikey commands to authenticate with Artifactory server
func (dca *DotnetCommandArgs) addNugetAuthenticationToNewConfig(cmdType dotnet.ToolchainType, configFile *os.File) error {
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

	cmd.CommandFlags = append(cmd.CommandFlags, cmd.Toolchain().GetTypeFlagPrefix()+"ConfigFile", configFile.Name())

	// Set Artifactory repo as source
	content := dotnet.ConfigFileTemplate
	_, err = configFile.WriteString(content)
	if err != nil {
		return nil, errorutils.CheckError(err)
	}
	return configFile, nil
}

// Runs nuget sources add command
func addSourceToNugetConfig(cmdType dotnet.ToolchainType, configFileName, sourceUrl, user, password string) error {
	cmd, err := dotnet.CreateDotnetAddSourceCmd(cmdType, sourceUrl)
	if err != nil {
		return err
	}

	flagPrefix := cmdType.GetTypeFlagPrefix()
	cmd.CommandFlags = append(cmd.CommandFlags, flagPrefix+"configfile", configFileName)
	cmd.CommandFlags = append(cmd.CommandFlags, flagPrefix+"name", SourceName)
	cmd.CommandFlags = append(cmd.CommandFlags, flagPrefix+"username", user)
	cmd.CommandFlags = append(cmd.CommandFlags, flagPrefix+"password", password)
	var cmdd []string
	cmdd = append(cmdd, cmd.Command...)
	cmdd = append(cmdd, cmd.CommandFlags...)
	log.Info("running: ", cmdd)
	output, err := io.RunCmdOutput(cmd)
	log.Debug("Running command: Add sources. Output:", output)
	return err
}

func (dca *DotnetCommandArgs) getSourceDetails() (sourceURL, user, password string, err error) {
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

func (dca *DotnetCommandArgs) createToolchainCmd() (*dotnet.Cmd, error) {
	c, err := dotnet.NewToolchainCmd(dca.toolchainType)
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
