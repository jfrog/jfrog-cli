package gradle

import (
	"fmt"
	gofrogcmd "github.com/jfrog/gofrog/io"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/utils/cliutils"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

const gradleExtractorDependencyVersion = "4.15.0"
const gradleInitScriptTemplate = "gradle.init"

const usePlugin = "useplugin"
const useWrapper = "usewrapper"
const gradleBuildInfoProperties = "BUILDINFO_PROPFILE"

type GradleCommand struct {
	tasks         string
	configPath    string
	configuration *utils.BuildConfiguration
	rtDetails     *config.ArtifactoryDetails
	threads       int
}

func NewGradleCommand() *GradleCommand {
	return &GradleCommand{}
}

// Returns the ArtfiactoryDetails. The information returns from the config file provided.
func (gc *GradleCommand) RtDetails() (*config.ArtifactoryDetails, error) {
	// Get the rtDetails from the config file.
	var err error
	if gc.rtDetails == nil {
		vConfig, err := utils.ReadConfigFile(gc.configPath, utils.YAML)
		if err != nil {
			return nil, err
		}
		gc.rtDetails, err = utils.GetRtDetails(vConfig)
	}
	return gc.rtDetails, err
}

func (gc *GradleCommand) SetRtDetails(rtDetails *config.ArtifactoryDetails) *GradleCommand {
	gc.rtDetails = rtDetails
	return gc
}

func (gc *GradleCommand) Run() error {
	gradleDependenciesDir, gradlePluginFilename, err := downloadGradleDependencies()
	if err != nil {
		return err
	}
	gradleRunConfig, err := createGradleRunConfig(gc.tasks, gc.configPath, gc.configuration, gc.threads, gradleDependenciesDir, gradlePluginFilename)
	if err != nil {
		return err
	}
	defer os.Remove(gradleRunConfig.env[gradleBuildInfoProperties])
	if err := gofrogcmd.RunCmd(gradleRunConfig); err != nil {
		return err
	}
	return nil
}

func (gc *GradleCommand) CommandName() string {
	return "rt_gradle"
}

func (gc *GradleCommand) SetConfiguration(configuration *utils.BuildConfiguration) *GradleCommand {
	gc.configuration = configuration
	return gc
}

func (gc *GradleCommand) SetConfigPath(configPath string) *GradleCommand {
	gc.configPath = configPath
	return gc
}

func (gc *GradleCommand) SetTasks(tasks string) *GradleCommand {
	gc.tasks = tasks
	return gc
}

func (gc *GradleCommand) SetThreads(threads int) *GradleCommand {
	gc.threads = threads
	return gc
}

func downloadGradleDependencies() (gradleDependenciesDir, gradlePluginFilename string, err error) {
	dependenciesPath, err := config.GetJfrogDependenciesPath()
	if err != nil {
		return
	}
	gradleDependenciesDir = filepath.Join(dependenciesPath, "gradle", gradleExtractorDependencyVersion)
	gradlePluginFilename = fmt.Sprintf("build-info-extractor-gradle-%s-uber.jar", gradleExtractorDependencyVersion)

	filePath := fmt.Sprintf("org/jfrog/buildinfo/build-info-extractor-gradle/%s", gradleExtractorDependencyVersion)
	downloadPath := path.Join(filePath, gradlePluginFilename)

	filepath.Join(gradleDependenciesDir, gradlePluginFilename)
	err = utils.DownloadExtractorIfNeeded(downloadPath, filepath.Join(gradleDependenciesDir, gradlePluginFilename))
	return
}

func createGradleRunConfig(tasks, configPath string, configuration *utils.BuildConfiguration, threads int, gradleDependenciesDir, gradlePluginFilename string) (*gradleRunConfig, error) {
	runConfig := &gradleRunConfig{env: map[string]string{}}
	runConfig.tasks = tasks

	vConfig, err := utils.ReadConfigFile(configPath, utils.YAML)
	if err != nil {
		return nil, err
	}

	runConfig.gradle, err = getGradleExecPath(vConfig.GetBool(useWrapper))
	if err != nil {
		return nil, err
	}

	if threads > 0 {
		vConfig.Set(utils.FORK_COUNT, threads)
	}

	runConfig.env[gradleBuildInfoProperties], err = utils.CreateBuildInfoPropertiesFile(configuration.BuildName, configuration.BuildNumber, vConfig, utils.Gradle)
	if err != nil {
		return nil, err
	}

	if !vConfig.GetBool(usePlugin) {
		runConfig.initScript, err = getInitScript(gradleDependenciesDir, gradlePluginFilename)
		if err != nil {
			return nil, err
		}
	}

	return runConfig, nil
}

func getInitScript(gradleDependenciesDir, gradlePluginFilename string) (string, error) {
	gradleDependenciesDir, err := filepath.Abs(gradleDependenciesDir)
	if err != nil {
		return "", errorutils.CheckError(err)
	}
	initScriptPath := filepath.Join(gradleDependenciesDir, gradleInitScriptTemplate)

	exists, err := fileutils.IsFileExists(initScriptPath, false)
	if exists || err != nil {
		return initScriptPath, err
	}

	gradlePluginPath := filepath.Join(gradleDependenciesDir, gradlePluginFilename)
	gradlePluginPath = strings.Replace(gradlePluginPath, "\\", "\\\\", -1)
	initScriptContent := strings.Replace(utils.GradleInitScript, "${pluginLibDir}", gradlePluginPath, -1)
	if !fileutils.IsPathExists(gradleDependenciesDir, false) {
		err = os.MkdirAll(gradleDependenciesDir, 0777)
		if errorutils.CheckError(err) != nil {
			return "", err
		}
	}

	return initScriptPath, errorutils.CheckError(ioutil.WriteFile(initScriptPath, []byte(initScriptContent), 0644))
}

type gradleRunConfig struct {
	gradle     string
	tasks      string
	initScript string
	env        map[string]string
}

func (config *gradleRunConfig) GetCmd() *exec.Cmd {
	var cmd []string
	cmd = append(cmd, config.gradle)
	if config.initScript != "" {
		cmd = append(cmd, "--init-script", config.initScript)
	}
	cmd = append(cmd, strings.Split(config.tasks, " ")...)

	log.Info("Running gradle command:", strings.Join(cmd, " "))
	return exec.Command(cmd[0], cmd[1:]...)
}

func (config *gradleRunConfig) GetEnv() map[string]string {
	return config.env
}

func (config *gradleRunConfig) GetStdWriter() io.WriteCloser {
	return nil
}

func (config *gradleRunConfig) GetErrWriter() io.WriteCloser {
	return nil
}

func getGradleExecPath(useWrapper bool) (string, error) {
	if useWrapper {
		if cliutils.IsWindows() {
			return "gradlew.bat", nil
		}
		return "./gradlew", nil
	}
	gradleExec, err := exec.LookPath("gradle")
	if err != nil {
		return "", errorutils.CheckError(err)
	}
	return gradleExec, nil
}
