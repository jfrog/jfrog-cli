package commands

import (
	"os"
	"errors"
	"os/exec"
	"path/filepath"
	"io/ioutil"
	"strings"
	"github.com/spf13/viper"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/fileutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils"
)

const MAVEN_EXTRACTOR_DEPENDENCY_VERSION = "2.7.0"
const CLASSWORLD_CONF_FILE_NAME = "classworlds.conf"
const MAVEN_HOME = "M2_HOME"

func Mvn(goals, configPath string, flags *utils.BuildConfigFlags) error {
	log.Info("Running Mvn...")
	err := validateMavenInstallation()
	if err != nil {
		return err
	}

	var dependenciesPath string
	dependenciesPath, err = downloadDependencies()
	if err != nil {
		return err
	}

	mvnRunConfig, err := createMvnRunConfig(goals, configPath, flags, dependenciesPath)
	if err != nil {
		return err
	}

	defer os.Remove(mvnRunConfig.buildInfoProperties)
	if err := utils.RunCmd(mvnRunConfig); err != nil {
		return err
	}

	return nil
}

func validateMavenInstallation() error {
	log.Info("Checking prerequisites...")
	mavenHome := os.Getenv(MAVEN_HOME)
	if mavenHome == "" {
		return errorutils.CheckError(errors.New(MAVEN_HOME + " environment variable is not set"))
	}
	return nil
}

func downloadDependencies() (string, error) {
	dependenciesPath, err := config.GetJfrogDependenciesPath()
	if err != nil {
		return "", err
	}

	filename := "/build-info-extractor-maven3-${version}-uber.jar"
	downloadPath := filepath.Join("jfrog/jfrog-jars/org/jfrog/buildinfo/build-info-extractor-maven3/${version}/", filename)
	err = utils.DownloadFromBintray(downloadPath, filename, MAVEN_EXTRACTOR_DEPENDENCY_VERSION, dependenciesPath)
	if err != nil {
		return "", err
	}

	err = createClassworldsConfig(dependenciesPath)
	return dependenciesPath, err
}

func createClassworldsConfig(dependenciesPath string) error {
	classworldsPath := filepath.Join(dependenciesPath, CLASSWORLD_CONF_FILE_NAME)

	if fileutils.IsPathExists(classworldsPath) {
		return nil
	}
	return errorutils.CheckError(ioutil.WriteFile(classworldsPath, []byte(utils.ClassworldsConf), 0644))
}

func createMvnRunConfig(goals, configPath string, flags *utils.BuildConfigFlags, dependenciesPath string) (*mvnRunConfig, error) {
	var err error
	var javaExecPath string

	javaHome := os.Getenv("JAVA_HOME")
	if javaHome != "" {
		javaExecPath = filepath.Join(javaHome, "bin", "java")
	} else {
		javaExecPath, err = exec.LookPath("java")
		if err != nil {
			return nil, errorutils.CheckError(err)
		}
	}

	mavenHome := os.Getenv("M2_HOME")
	plexusClassworlds, err := filepath.Glob(filepath.Join(mavenHome, "boot", "plexus-classworlds*"))
	if err != nil {
		return nil, errorutils.CheckError(err)
	}

	if len(plexusClassworlds) != 1 {
		return nil, errorutils.CheckError(errors.New("Couldn't find plexus-classworlds-x.x.x.jar in Maven installation path, please check M2_HOME environmetn variable."))
	}

	var currentWorkdir string
	currentWorkdir, err = os.Getwd()
	if err != nil {
		return nil, errorutils.CheckError(err)
	}

	var vConfig *viper.Viper
	vConfig, err = utils.ReadConfigFile(configPath, utils.YAML)
	if err != nil {
		return nil, err
	}

	if len(flags.BuildName) > 0 && len(flags.BuildNumber) > 0 {
		vConfig.Set(utils.BUILD_NAME, flags.BuildName)
		vConfig.Set(utils.BUILD_NUMBER, flags.BuildNumber)
		err = utils.SaveBuildGeneralDetails(flags.BuildName, flags.BuildNumber)
		if err != nil {
			return nil, err
		}
	}

	buildInfoProperties, err := utils.CreateBuildInfoPropertiesFile(flags.BuildName, flags.BuildNumber, vConfig, utils.MAVEN)
	if err != nil {
		return nil, err
	}

	return &mvnRunConfig{
		java:                   javaExecPath,
		pluginDependencies:     dependenciesPath,
		plexusClassworlds:      plexusClassworlds[0],
		cleassworldsConfig:     filepath.Join(dependenciesPath, CLASSWORLD_CONF_FILE_NAME),
		mavenHome:              mavenHome,
		workspace:              currentWorkdir,
		goals:                  goals,
		buildInfoProperties:    buildInfoProperties,
		generatedBuildInfoPath: vConfig.GetString(utils.GENERATED_BUILD_INFO),
	}, nil
}

func (config *mvnRunConfig) GetCmd() *exec.Cmd {
	var cmd []string
	cmd = append(cmd, config.java)
	cmd = append(cmd, "-classpath", config.plexusClassworlds)
	cmd = append(cmd, "-Dmaven.home="+config.mavenHome)
	cmd = append(cmd, "-DbuildInfoConfig.propertiesFile="+config.buildInfoProperties)
	cmd = append(cmd, "-Dm3plugin.lib="+config.pluginDependencies)
	cmd = append(cmd, "-Dclassworlds.conf="+config.cleassworldsConfig)
	cmd = append(cmd, "-Dmaven.multiModuleProjectDirectory="+config.workspace)
	cmd = append(cmd, "org.codehaus.plexus.classworlds.launcher.Launcher")
	cmd = append(cmd, strings.Split(config.goals, " ")...)
	return exec.Command(cmd[0], cmd[1:]...)
}

func (config *mvnRunConfig) GetEnv() map[string]string {
	return map[string]string{}
}

type mvnRunConfig struct {
	java                   string
	plexusClassworlds      string
	cleassworldsConfig     string
	mavenHome              string
	pluginDependencies     string
	workspace              string
	pom                    string
	goals                  string
	buildInfoProperties    string
	generatedBuildInfoPath string
}
