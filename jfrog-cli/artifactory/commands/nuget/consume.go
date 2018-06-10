package nuget

import (
	"fmt"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils/nuget/solution"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/config"
	clientutils "github.com/jfrog/jfrog-cli-go/jfrog-client/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/errorutils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/io/fileutils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/log"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"strings"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils/nuget"
)

type Params struct {
	Args        string
	Flags       string
	RepoName    string
	BuildName   string
	BuildNumber string

	ArtifactoryDetails *config.ArtifactoryDetails
}

// Exec all consume type nuget commands, install, update, add, restore.
func ConsumeCmd(params *Params) error {
	log.Info("Running nuget...")
	// Use temp dir to save config file, the config will be removed at the end.
	err := fileutils.CreateTempDirPath()
	if err != nil {
		return err
	}
	defer fileutils.RemoveTempDir()

	solutionPath, err := prepareAndRunCmd(params)
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

	if isCollectBuildInfo {
		if err = utils.SaveBuildGeneralDetails(params.BuildName, params.BuildNumber); err != nil {
			return err
		}
		return utils.SaveBuildInfo(params.BuildName, params.BuildNumber, sol.BuildInfo())
	}
	return nil
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

	content, err := sol.Marshal()
	if err != nil {
		return err
	}
	log.Output(clientutils.IndentJson(content))
	return nil
}

func prepareAndRunCmd(params *Params) (string, error) {
	cmd, err := createNugetCmd(params)
	if err != nil {
		return "", err
	}

	err = prepareConfigFile(params, cmd)
	if err != nil {
		return "", err
	}

	err = utils.RunCmd(cmd)
	if err != nil {
		return "", errorutils.CheckError(err)
	}

	// Return current solution path.
	return os.Getwd()
}

func prepareConfigFile(params *Params, cmd *nuget.Cmd) error {
	tempDir, err := fileutils.GetTempDirPath()
	if err != nil {
		return err
	}
	configFile, err := ioutil.TempFile(tempDir, "jfrog.cli.nuget.")
	if err != nil {
		return err
	}
	log.Debug("Nuget config file created at:", configFile.Name())

	currentConfigPath, err := getAndReplaceCurrentConfigPath(configFile.Name(), cmd)
	if err != nil {
		return err
	}
	if currentConfigPath != "" {
		err = copyExistingConfig(configFile, currentConfigPath)
	} else {
		err = initNewConfig(configFile, params, cmd)
	}
	if err != nil {
		return err
	}

	err = configFile.Close()
	if err != nil {
		return err
	}

	return nil
}

func copyExistingConfig(newConfigFile *os.File, currentConfigPath string) error {
	// Copy the content of the current config file to the temp config.
	// To preserve the behaviour of the given config.
	log.Debug("Copying config file content from:", currentConfigPath)
	currentConfigFile, err := os.Open(currentConfigPath)
	if err != nil {
		return errorutils.CheckError(err)
	}
	defer currentConfigFile.Close()

	_, err = io.Copy(newConfigFile, currentConfigFile)
	if err != nil {
		return errorutils.CheckError(err)
	}

	return nil
}
func getAndReplaceCurrentConfigPath(newConfigPath string, cmd *nuget.Cmd) (string, error) {
	for i := 0; i < len(cmd.CommandFlags); i++ {
		if !strings.Contains(strings.ToLower(cmd.CommandFlags[i]), "-configfile") {
			continue
		}
		if i+1 == len(cmd.CommandFlags) {
			return "", errorutils.CheckError(fmt.Errorf("-ConfigFile flag was provided without file path"))
		}
		currentConfigPath := cmd.CommandFlags[i+1]
		exists, err := fileutils.IsFileExists(currentConfigPath)
		if err != nil {
			return "", errorutils.CheckError(err)
		}
		if !exists {
			return "", errorutils.CheckError(fmt.Errorf("Could not find config file at: %s", currentConfigPath))
		}

		// Override config file in the cmd
		cmd.CommandFlags[i+1] = newConfigPath
		return currentConfigPath, nil
	}

	return "", nil
}

func initNewConfig(newConfigFile *os.File, params *Params, cmd *nuget.Cmd) error {
	cmd.CommandFlags = append(cmd.CommandFlags, "-ConfigFile", newConfigFile.Name())

	// Set Artifactory repo as source
	sourceUrl, user, password, err := getSourceDetails(params)
	if err != nil {
		return err
	}
	content := fmt.Sprintf(nuget.ConfigFileTemplate, sourceUrl, user, password)
	_, err = newConfigFile.WriteString(content)
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
		c.Command = strings.Split(params.Args, " ")
	}

	if params.Flags != "" {
		c.CommandFlags = strings.Split(params.Flags, " ")
	}

	return c, nil
}
