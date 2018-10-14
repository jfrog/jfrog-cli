package mvn

import (
	"errors"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/spf13/viper"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"fmt"
)

const mavenExtractorDependencyVersion = "2.11.1"
const ClasswordConfFileName = "classworlds.conf"
const MavenHome = "M2_HOME"

func Mvn(goals, configPath string, configuration *utils.BuildConfiguration) error {
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

	mvnRunConfig, err := createMvnRunConfig(goals, configPath, configuration, dependenciesPath)
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
	log.Debug("Checking prerequisites.")
	mavenHome := os.Getenv(MavenHome)
	if mavenHome == "" {
		return errorutils.CheckError(errors.New(MavenHome + " environment variable is not set"))
	}
	return nil
}

func downloadDependencies() (string, error) {
	dependenciesPath, err := config.GetJfrogDependenciesPath()
	if err != nil {
		return "", err
	}
	dependenciesPath = filepath.Join(dependenciesPath, "maven")

	filename := fmt.Sprintf("build-info-extractor-maven3-%s-uber.jar", mavenExtractorDependencyVersion)
	filePath := fmt.Sprintf("jfrog/jfrog-jars/org/jfrog/buildinfo/build-info-extractor-maven3/%s", mavenExtractorDependencyVersion)
	downloadPath := path.Join(filePath, filename)

	err = utils.DownloadFromBintrayIfNeeded(downloadPath, filename, dependenciesPath)
	if err != nil {
		return "", err
	}

	err = createClassworldsConfig(dependenciesPath)
	return dependenciesPath, err
}

func createClassworldsConfig(dependenciesPath string) error {
	classworldsPath := filepath.Join(dependenciesPath, ClasswordConfFileName)

	if fileutils.IsPathExists(classworldsPath) {
		return nil
	}
	return errorutils.CheckError(ioutil.WriteFile(classworldsPath, []byte(utils.ClassworldsConf), 0644))
}

func createMvnRunConfig(goals, configPath string, configuration *utils.BuildConfiguration, dependenciesPath string) (*mvnRunConfig, error) {
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
		return nil, errorutils.CheckError(errors.New("couldn't find plexus-classworlds-x.x.x.jar in Maven installation path, please check M2_HOME environment variable"))
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

	if len(configuration.BuildName) > 0 && len(configuration.BuildNumber) > 0 {
		vConfig.Set(utils.BUILD_NAME, configuration.BuildName)
		vConfig.Set(utils.BUILD_NUMBER, configuration.BuildNumber)
		err = utils.SaveBuildGeneralDetails(configuration.BuildName, configuration.BuildNumber)
		if err != nil {
			return nil, err
		}
	}

	buildInfoProperties, err := utils.CreateBuildInfoPropertiesFile(configuration.BuildName, configuration.BuildNumber, vConfig, utils.MAVEN)
	if err != nil {
		return nil, err
	}

	return &mvnRunConfig{
		java:                   javaExecPath,
		pluginDependencies:     dependenciesPath,
		plexusClassworlds:      plexusClassworlds[0],
		cleassworldsConfig:     filepath.Join(dependenciesPath, ClasswordConfFileName),
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

func (config *mvnRunConfig) GetStdWriter() io.WriteCloser {
	return nil
}

func (config *mvnRunConfig) GetErrWriter() io.WriteCloser {
	return nil
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
