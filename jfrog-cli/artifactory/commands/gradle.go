package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/fileutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const gradleExtractorDependencyVersion = "4.6.0"
const gradleInitScriptTemplate = "gradle.init.v2"

const usePlugin = "useplugin"
const useWrapper = "usewrapper"
const gradleBuildInfoProperties = "buildInfoConfig.propertiesFile"

func Gradle(tasks, configPath string, flags *utils.BuildConfigFlags) error {
	log.Info("Running Gradle...")
	dependenciesPath, err := downloadGradleDependencies()
	if err != nil {
		return err
	}

	gradleRunConfig, err := createGradleRunConfig(tasks, configPath, flags, dependenciesPath)
	if err != nil {
		return err
	}

	defer os.Remove(gradleRunConfig.env[gradleBuildInfoProperties])
	if err := utils.RunCmd(gradleRunConfig); err != nil {
		return err
	}

	return nil
}

func downloadGradleDependencies() (string, error) {
	dependenciesPath, err := config.GetJfrogDependenciesPath()
	if err != nil {
		return "", err
	}

	filename := "build-info-extractor-gradle-${version}-uber.jar"
	downloadPath := filepath.Join("jfrog/jfrog-jars/org/jfrog/buildinfo/build-info-extractor-gradle/${version}/", filename)
	err = utils.DownloadFromBintray(downloadPath, filename, gradleExtractorDependencyVersion, dependenciesPath)
	if err != nil {
		return "", err
	}

	return dependenciesPath, err
}

func createGradleRunConfig(tasks, configPath string, flags *utils.BuildConfigFlags, dependenciesPath string) (*gradleRunConfig, error) {
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

	runConfig.env[gradleBuildInfoProperties], err = utils.CreateBuildInfoPropertiesFile(flags.BuildName, flags.BuildNumber, vConfig, utils.GRADLE)
	if err != nil {
		return nil, err
	}

	if !vConfig.GetBool(usePlugin) {
		runConfig.initScript, err = getInitScript(dependenciesPath)
		if err != nil {
			return nil, err
		}
	}

	return runConfig, nil
}

func getInitScript(dependenciesPath string) (string, error) {
	initScript := filepath.Join(dependenciesPath, gradleInitScriptTemplate)
	dependenciesPath, err := filepath.Abs(dependenciesPath)
	if err != nil {
		return "", errorutils.CheckError(err)
	}

	if fileutils.IsPathExists(initScript) {
		return initScript, nil
	}

	initScriptContent := strings.Replace(utils.GradleInitScript, "${pluginLibDir}", dependenciesPath, -1)
	return initScript, errorutils.CheckError(ioutil.WriteFile(initScript, []byte(initScriptContent), 0644))
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
		if runtime.GOOS == "windows" {
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
