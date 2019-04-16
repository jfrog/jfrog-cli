package nuget

import (
	"fmt"
	gofrogcmd "github.com/jfrog/gofrog/io"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils/nuget"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils/nuget/solution"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/config"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/mattn/go-shellwords"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"strings"
)

type Params struct {
	Args               string
	Flags              string
	RepoName           string
	BuildName          string
	BuildNumber        string
	ArtifactoryDetails *config.ArtifactoryDetails
}

const SOURCE_NAME = "JFrogCli"

// Exec all consume type nuget commands, install, update, add, restore.
func ConsumeCmd(params *Params, solutionPath string) error {
	log.Info("Running nuget...")
	// Use temp dir to save config file, the config will be removed at the end.
	tempDirPath, err := fileutils.CreateTempDir()
	if err != nil {
		return err
	}
	defer fileutils.RemoveTempDir(tempDirPath)


	solutionPath, err = changeWorkingDir(solutionPath)
	if err != nil {
		return err
	}

	err = prepareAndRunCmd(params, tempDirPath)
	if err != nil {
		return err
	}

	isCollectBuildInfo := len(params.BuildName) > 0 && len(params.BuildNumber) > 0
	if !isCollectBuildInfo {
		return nil
	}

	sol, err := solution.Load(solutionPath)
	if err != nil {
		return err
	}

	if err = utils.SaveBuildGeneralDetails(params.BuildName, params.BuildNumber); err != nil {
		return err
	}
	buildInfo, err := sol.BuildInfo()
	if err != nil {
		return err
	}
	return utils.SaveBuildInfo(params.BuildName, params.BuildNumber, buildInfo)
}

func DependencyTreeCmd() error {
	workspace, err := os.Getwd()
	if err != nil {
		return errorutils.CheckError(err)
	}

	sol, err := solution.Load(workspace)
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
func prepareAndRunCmd(params *Params, configDirPath string) error {
	cmd, err := createNugetCmd(params)
	if err != nil {
		return err
	}
	// To prevent NuGet prompting for credentials
	err = os.Setenv("NUGET_EXE_NO_PROMPT", "true")
	if err != nil {
		return errorutils.CheckError(err)
	}

	err = prepareConfigFile(params, cmd, configDirPath)
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
func prepareConfigFile(params *Params, cmd *nuget.Cmd, configDirPath string) error {
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

	err = initNewConfig(params, cmd, configDirPath)
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
func initNewConfig(params *Params, cmd *nuget.Cmd, configDirPath string) error {
	// Got to here, means that neither of the flags provided and we need to init our own config.
	configFile, err := writeToTempConfigFile(cmd, configDirPath)
	if err != nil {
		return err
	}

	return addNugetAuthenticationToNewConfig(params, configFile)
}

// Runs nuget add sources and setapikey commands to authenticate with Artifactory server
func addNugetAuthenticationToNewConfig(params *Params, configFile *os.File) error {
	sourceUrl, user, password, err := getSourceDetails(params)
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
	cmd.CommandFlags = append(cmd.CommandFlags, "-Name", SOURCE_NAME)
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
	cmd.CommandFlags = append(cmd.CommandFlags, "-Source", SOURCE_NAME)
	cmd.CommandFlags = append(cmd.CommandFlags, "-ConfigFile", configFileName)

	output, err := gofrogcmd.RunCmdOutput(cmd)
	log.Debug("Running command: SetApiKey. Output:", output)
	return err
}

func getSourceDetails(params *Params) (sourceURL, user, password string, err error) {
	var u *url.URL
	u, err = url.Parse(params.ArtifactoryDetails.Url)
	if errorutils.CheckError(err) != nil {
		return
	}
	u.Path = path.Join(u.Path, "api/nuget", params.RepoName)
	sourceURL = u.String()
	user = params.ArtifactoryDetails.User
	password = params.ArtifactoryDetails.Password
	return
}

func createNugetCmd(params *Params) (*nuget.Cmd, error) {
	c, err := nuget.NewNugetCmd()
	if err != nil {
		return nil, err
	}
	if params.Args != "" {
		c.Command, err = shellwords.Parse(params.Args)
		if err != nil {
			return nil, errorutils.CheckError(err)
		}
	}

	if params.Flags != "" {
		c.CommandFlags, err = shellwords.Parse(params.Flags)
	}

	return c, errorutils.CheckError(err)
}
