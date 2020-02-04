package utils

import (
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codegangsta/cli"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/utils/cliutils"
	"github.com/jfrog/jfrog-cli-go/utils/log"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func init() {
	os.Setenv(cliutils.LogLevel, "WARN") // Disable "[Info] *** build config successfully created." messages
	log.SetDefaultLogger()
}

func TestGoConfigFile(t *testing.T) {
	// Set JFROG_CLI_HOME_DIR environment variable
	tempDirPath := createTempEnv(t)
	defer os.RemoveAll(tempDirPath)

	// Create build config
	context := createContext(ResolutionServerId+"=relServer", ResolutionRepo+"=repo", DeploymentServerId+"=depServer", DeploymentRepo+"=repo-local")
	err := CreateBuildConfig(context, utils.Go)
	assert.NoError(t, err)

	// Check configuration
	config := checkCommonAndGetConfiguration(t, utils.Go.String(), tempDirPath)
	assert.Equal(t, "relServer", config.GetString("resolver.serverId"))
	assert.Equal(t, "repo", config.GetString("resolver.repo"))
	assert.Equal(t, "depServer", config.GetString("deployer.serverId"))
	assert.Equal(t, "repo-local", config.GetString("deployer.repo"))
}

func TestPipConfigFile(t *testing.T) {
	// Set JFROG_CLI_HOME_DIR environment variable
	tempDirPath := createTempEnv(t)
	defer os.RemoveAll(tempDirPath)

	// Create build config
	context := createContext(ResolutionServerId+"=relServer", ResolutionRepo+"=repo", DeploymentServerId+"=depServer", DeploymentRepo+"=repo-local")
	err := CreateBuildConfig(context, utils.Pip)
	assert.NoError(t, err)

	// Check configuration
	config := checkCommonAndGetConfiguration(t, utils.Pip.String(), tempDirPath)
	assert.Equal(t, "relServer", config.GetString("resolver.serverId"))
	assert.Equal(t, "repo", config.GetString("resolver.repo"))
	assert.Equal(t, "depServer", config.GetString("deployer.serverId"))
	assert.Equal(t, "repo-local", config.GetString("deployer.repo"))
}

func TestNpmConfigFile(t *testing.T) {
	// Set JFROG_CLI_HOME_DIR environment variable
	tempDirPath := createTempEnv(t)
	defer os.RemoveAll(tempDirPath)

	// Create build config
	context := createContext(ResolutionServerId+"=relServer", ResolutionRepo+"=repo", DeploymentServerId+"=depServer", DeploymentRepo+"=repo-local")
	err := CreateBuildConfig(context, utils.Npm)
	assert.NoError(t, err)

	// Check configuration
	config := checkCommonAndGetConfiguration(t, utils.Npm.String(), tempDirPath)
	assert.Equal(t, "relServer", config.GetString("resolver.serverId"))
	assert.Equal(t, "repo", config.GetString("resolver.repo"))
	assert.Equal(t, "depServer", config.GetString("deployer.serverId"))
	assert.Equal(t, "repo-local", config.GetString("deployer.repo"))
}

func TestNugetConfigFile(t *testing.T) {
	// Set JFROG_CLI_HOME_DIR environment variable
	tempDirPath := createTempEnv(t)
	defer os.RemoveAll(tempDirPath)

	// Create build config
	context := createContext(ResolutionServerId+"=relServer", ResolutionRepo+"=repo")
	err := CreateBuildConfig(context, utils.Nuget)
	assert.NoError(t, err)

	// Check configuration
	config := checkCommonAndGetConfiguration(t, utils.Nuget.String(), tempDirPath)
	assert.Equal(t, "relServer", config.GetString("resolver.serverId"))
	assert.Equal(t, "repo", config.GetString("resolver.repo"))
}

func TestMavenConfigFile(t *testing.T) {
	// Set JFROG_CLI_HOME_DIR environment variable
	tempDirPath := createTempEnv(t)
	defer os.RemoveAll(tempDirPath)

	// Create build config
	context := createContext(ResolutionServerId+"=relServer", ResolutionReleasesRepo+"=release-repo", ResolutionSnapshotsRepo+"=snapshot-repo",
		DeploymentServerId+"=depServer", DeploymentReleasesRepo+"=release-repo-local", DeploymentSnapshotsRepo+"=snapshot-repo-local")
	err := CreateBuildConfig(context, utils.Maven)
	assert.NoError(t, err)

	// Check configuration
	config := checkCommonAndGetConfiguration(t, utils.Maven.String(), tempDirPath)
	assert.Equal(t, "relServer", config.GetString("resolver.serverId"))
	assert.Equal(t, "snapshot-repo", config.GetString("resolver.snapshotRepo"))
	assert.Equal(t, "release-repo", config.GetString("resolver.releaseRepo"))
	assert.Equal(t, "depServer", config.GetString("deployer.serverId"))
	assert.Equal(t, "snapshot-repo-local", config.GetString("deployer.snapshotRepo"))
	assert.Equal(t, "release-repo-local", config.GetString("deployer.releaseRepo"))
}

func TestGradleConfigFile(t *testing.T) {
	// Set JFROG_CLI_HOME_DIR environment variable
	tempDirPath := createTempEnv(t)
	defer os.RemoveAll(tempDirPath)

	// Create build config
	context := createContext(ResolutionServerId+"=relServer", ResolutionRepo+"=repo", DeploymentServerId+"=depServer", DeploymentRepo+"=repo-local",
		IvyDescPattern+"=[ivy]/[pattern]", IvyArtifactsPattern+"=[artifact]/[pattern]")
	err := CreateBuildConfig(context, utils.Gradle)
	assert.NoError(t, err)

	// Check configuration
	config := checkCommonAndGetConfiguration(t, utils.Gradle.String(), tempDirPath)
	assert.Equal(t, "relServer", config.GetString("resolver.serverId"))
	assert.Equal(t, "repo", config.GetString("resolver.repo"))
	assert.Equal(t, "depServer", config.GetString("deployer.serverId"))
	assert.Equal(t, "repo-local", config.GetString("deployer.repo"))
	assert.Equal(t, true, config.GetBool("deployer.deployMavenDescriptors"))
	assert.Equal(t, true, config.GetBool("deployer.deployIvyDescriptors"))
	assert.Equal(t, "[ivy]/[pattern]", config.GetString("deployer.ivyPattern"))
	assert.Equal(t, "[artifact]/[pattern]", config.GetString("deployer.artifactPattern"))
	assert.Equal(t, true, config.GetBool("usePlugin"))
	assert.Equal(t, true, config.GetBool("useWrapper"))
}

// Set JFROG_CLI_HOME_DIR environment variable to be a new temp directory
func createTempEnv(t *testing.T) string {
	tmpDir, err := ioutil.TempDir("", "configfile_test")
	assert.NoError(t, err)
	err = os.Setenv(cliutils.HomeDir, tmpDir)
	assert.NoError(t, err)
	return tmpDir
}

// Create new Codegangsta context with all required flags.
func createContext(stringFlags ...string) *cli.Context {
	flagSet := flag.NewFlagSet("TestFlagSet", flag.ContinueOnError)
	flags := setBoolFlags(flagSet, Global, UsesPlugin, UseWrapper, DeployMavenDesc, DeployIvyDesc)
	flags = append(flags, setStringFlags(flagSet, stringFlags...)...)
	flagSet.Parse(flags)
	return cli.NewContext(nil, flagSet, nil)
}

// Set boolean flags and initialize them to true. Return a slice of them.
func setBoolFlags(flagSet *flag.FlagSet, flags ...string) []string {
	cmdFlags := []string{}
	for _, flag := range flags {
		flagSet.Bool(flag, true, "")
		cmdFlags = append(cmdFlags, "--"+flag)
	}
	return cmdFlags
}

// Set string flags. Return a slice of their values.
func setStringFlags(flagSet *flag.FlagSet, flags ...string) []string {
	cmdFlags := []string{}
	for _, flag := range flags {
		flagSet.String(strings.Split(flag, "=")[0], "", "")
		cmdFlags = append(cmdFlags, "--"+flag)
	}
	return cmdFlags
}

// Read yaml configuration from disk, check version and type.
func checkCommonAndGetConfiguration(t *testing.T, projectType string, tempDirPath string) *viper.Viper {
	config, err := utils.ReadConfigFile(filepath.Join(tempDirPath, "projects", projectType+".yaml"), utils.YAML)
	assert.NoError(t, err)
	assert.Equal(t, BUILD_CONF_VERSION, config.GetInt("version"))
	assert.Equal(t, projectType, config.GetString("type"))
	return config
}
