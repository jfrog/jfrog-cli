package utils

import (
	"errors"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"path/filepath"
)

type BuildType int

const (
	MAVEN BuildType = iota
	GRADLE
)

var BuildTypes = []string{
	"maven",
	"gradle",
}

func (buildType BuildType) String() string {
	return BuildTypes[buildType]
}

type BuildConfigMapping map[BuildType][]*map[string]string

var buildTypeConfigMapping = BuildConfigMapping{
	MAVEN:  {&commonConfigMapping, &mavenConfigMapping},
	GRADLE: {&commonConfigMapping, &gradleConfigMapping},
}

type ConfigType string

const (
	YAML       ConfigType = "yaml"
	PROPERTIES ConfigType = "properties"
)

// For key/value binding
const BUILD_NAME = "build.name"
const BUILD_NUMBER = "build.number"
const GENERATED_BUILD_INFO = "buildInfo.generated"

const RESOLVER_PREFIX = "resolver."
const DEPLOYER_PREFIX = "deployer."

const REPO = "repo"
const SNAPSHOT_REPO = "snapshotRepo"
const RELEASE_REPO = "releaseRepo"

const SERVER_ID = "serverId"
const URL = "url"
const USERNAME = "username"
const PASSWORD = "password"

const MAVEN_DESCRIPTOR = "deployMavenDescriptors"
const IVY_DESCRIPTOR = "deployIvyDescriptors"
const IVY_PATTERN = "ivyPattern"
const ARTIFACT_PATTERN = "artifactPattern"

// For path and temp files
const PROPERTIES_TEMP_PREFIX = "buildInfoProperties"
const PROPERTIES_TEMP_PATH = "jfrog/properties/"
const GENERATED_BUILD_INFO_TEMP_PREFIX = "generatedBuildInfo"

// Config mapping are used to create buildInfo properties file to be used by BuildInfo extractors.
// Build config provided by the user may contain other properties that will not be included in the properties file.
var defaultPropertiesValues = map[string]string{
	"artifactory.publish.artifacts":                      "true",
	"artifactory.publish.buildInfo":                      "false",
	"artifactory.publish.unstable":                       "false",
	"artifactory.publish.maven":                          "false",
	"artifactory.publish.ivy":                            "false",
	"buildInfoConfig.includeEnvVars":                     "false",
	"buildInfoConfig.envVarsExcludePatterns":             "*password*,*secret*,*key*,*token*",
	"buildInfo.agent.name":                               utils.ClientAgent + "/" + utils.GetVersion(),
	"buildInfo.licenseControl.autoDiscover":              "true",
	"buildInfo.licenseControl.includePublishedArtifacts": "false",
	"buildInfo.licenseControl.runChecks":                 "false",
	"org.jfrog.build.extractor.maven.recorder.activate":  "true",
	"buildInfo.env.extractor.used":                       "true",
}

var commonConfigMapping = map[string]string{
	"artifactory.publish.buildInfo":                      "",
	"artifactory.publish.unstable":                       "",
	"buildInfoConfig.includeEnvVars":                     "",
	"buildInfoConfig.envVarsExcludePatterns":             "",
	"buildInfo.agent.name":                               "",
	"buildInfo.licenseControl.autoDiscover":              "",
	"buildInfo.licenseControl.includePublishedArtifacts": "",
	"buildInfo.licenseControl.runChecks":                 "",
	"artifactory.resolve.contextUrl":                     RESOLVER_PREFIX + URL,
	"artifactory.resolve.username":                       RESOLVER_PREFIX + USERNAME,
	"artifactory.resolve.password":                       RESOLVER_PREFIX + PASSWORD,
	"artifactory.publish.contextUrl":                     DEPLOYER_PREFIX + URL,
	"artifactory.publish.username":                       DEPLOYER_PREFIX + USERNAME,
	"artifactory.publish.password":                       DEPLOYER_PREFIX + PASSWORD,
	"artifactory.publish.artifacts":                      "",
	"artifactory.deploy.build.name":                      BUILD_NAME,
	"artifactory.deploy.build.number":                    BUILD_NUMBER,
	"buildInfo.generated.build.info":                     GENERATED_BUILD_INFO,
}

var mavenConfigMapping = map[string]string{
	"org.jfrog.build.extractor.maven.recorder.activate": "",
	"artifactory.resolve.repoKey":                       RESOLVER_PREFIX + RELEASE_REPO,
	"artifactory.resolve.downSnapshotRepoKey":           RESOLVER_PREFIX + SNAPSHOT_REPO,
	"artifactory.publish.repoKey":                       DEPLOYER_PREFIX + RELEASE_REPO,
	"artifactory.publish.snapshot.repoKey":              DEPLOYER_PREFIX + SNAPSHOT_REPO,
}

var gradleConfigMapping = map[string]string{
	"buildInfo.env.extractor.used":                      "",
	"org.jfrog.build.extractor.maven.recorder.activate": "",
	"artifactory.resolve.repoKey":                       RESOLVER_PREFIX + REPO,
	"artifactory.resolve.downSnapshotRepoKey":           RESOLVER_PREFIX + REPO,
	"artifactory.publish.repoKey":                       DEPLOYER_PREFIX + REPO,
	"artifactory.publish.snapshot.repoKey":              DEPLOYER_PREFIX + REPO,
	"artifactory.publish.maven":                         DEPLOYER_PREFIX + MAVEN_DESCRIPTOR,
	"artifactory.publish.ivy":                           DEPLOYER_PREFIX + IVY_DESCRIPTOR,
	"artifactory.publish.ivy.ivyPattern":                DEPLOYER_PREFIX + IVY_PATTERN,
	"artifactory.publish.ivy.artPattern":                DEPLOYER_PREFIX + ARTIFACT_PATTERN,
}

func ReadConfigFile(configPath string, configType ConfigType) (*viper.Viper, error) {
	config := viper.New()
	config.SetConfigType(string(configType))

	f, err := os.Open(configPath)
	if err != nil {
		return config, errorutils.CheckError(err)
	}
	err = config.ReadConfig(f)
	if err != nil {
		return config, errorutils.CheckError(err)
	}

	return config, nil
}

func CreateBuildInfoPropertiesFile(buildName, buildNumber string, config *viper.Viper, buildType BuildType) (string, error) {
	if config.GetString("type") != buildType.String() {
		return "", errorutils.CheckError(errors.New("Incompatible build config, expected: " + buildType.String() + " got: " + config.GetString("type")))
	}

	tempDir := os.TempDir()
	propertiesPath := filepath.Join(tempDir, PROPERTIES_TEMP_PATH)
	err := os.MkdirAll(propertiesPath, 0777)
	if errorutils.CheckError(err) != nil {
		return "", err
	}
	propertiesFile, err := ioutil.TempFile(propertiesPath, PROPERTIES_TEMP_PREFIX)
	defer propertiesFile.Close()
	if err != nil {
		return "", errorutils.CheckError(err)
	}

	err = setServerDetailsToConfig(RESOLVER_PREFIX, config)
	if err != nil {
		return "", err
	}
	err = setServerDetailsToConfig(DEPLOYER_PREFIX, config)
	if err != nil {
		return "", err
	}

	err = createGeneratedBuildInfoFile(buildName, buildNumber, config)
	if err != nil {
		return "", err
	}

	// Iterate all the required properties keys according to the buildType and create properties file.
	// If a value is provided by the build config file write it,
	// otherwise use the default value from defaultPropertiesValues map.
	for _, partialMapping := range buildTypeConfigMapping[buildType] {
		for propertyKey, configKey := range *partialMapping {
			if config.IsSet(configKey) {
				_, err = propertiesFile.WriteString(propertyKey + "=" + config.GetString(configKey) + "\n")
			} else if defaultVal, ok := defaultPropertiesValues[propertyKey]; ok {
				_, err = propertiesFile.WriteString(propertyKey + "=" + defaultVal + "\n")
			}
			if err != nil {
				return "", errorutils.CheckError(err)
			}
		}
	}
	return propertiesFile.Name(), nil
}

func setServerDetailsToConfig(contextPrefix string, vConfig *viper.Viper) error {
	if !vConfig.IsSet(contextPrefix + SERVER_ID) {
		return nil
	}

	serverId := vConfig.GetString(contextPrefix + SERVER_ID)
	artDetails, err := config.GetArtifactorySpecificConfig(serverId)
	if err != nil {
		return err
	}
	if artDetails.GetUrl() == "" {
		return errorutils.CheckError(errors.New("Server ID " + serverId + " URL is required"))
	}
	vConfig.Set(contextPrefix+URL, artDetails.GetUrl())

	if artDetails.GetApiKey() != "" {
		return errorutils.CheckError(errors.New("Server ID " + serverId + " API key authentication is not supported"))
	}

	if artDetails.GetUser() != "" && artDetails.GetPassword() != "" {
		vConfig.Set(contextPrefix+USERNAME, artDetails.GetUser())
		vConfig.Set(contextPrefix+PASSWORD, artDetails.GetPassword())
	}
	return nil
}

// Generated build info file is template file where build-info will be written to during the
// Maven or Gradle build.
// Creating this file only if build name and number is provided.
func createGeneratedBuildInfoFile(buildName, buildNumber string, config *viper.Viper) error {
	if buildName == "" && buildNumber == "" {
		return nil
	}

	config.Set(BUILD_NAME, buildName)
	config.Set(BUILD_NUMBER, buildNumber)
	err := SaveBuildGeneralDetails(buildName, buildNumber)
	if err != nil {
		return err
	}

	buildPath, err := getBuildDir(config.GetString(BUILD_NAME), config.GetString(BUILD_NUMBER))
	if err != nil {
		return err
	}
	var tempFile *os.File
	tempFile, err = ioutil.TempFile(buildPath, GENERATED_BUILD_INFO_TEMP_PREFIX)
	defer tempFile.Close()
	if err != nil {
		return err
	}
	config.Set(GENERATED_BUILD_INFO, tempFile.Name())
	return nil
}
