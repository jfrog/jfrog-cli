package utils

import (
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/io/fileutils"
	"github.com/spf13/viper"
	"os"
	"testing"
)

func TestCreateDefaultPropertiesFile(t *testing.T) {
	for index := range BuildTypes {
		testCreateDefaultPropertiesFile(BuildType(index), t)
	}
}

func testCreateDefaultPropertiesFile(buildType BuildType, t *testing.T) {
	providedConfig := viper.New()
	providedConfig.Set("type", buildType.String())

	propsFile, err := CreateBuildInfoPropertiesFile("", "", providedConfig, buildType)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(propsFile)

	actualConfig, err := ReadConfigFile(propsFile, PROPERTIES)
	if err != nil {
		t.Error(err)
	}

	expectedConfig := viper.New()
	for _, partialMapping := range buildTypeConfigMapping[buildType] {
		for propertyKey := range *partialMapping {
			if defaultPropertiesValues[propertyKey] != "" {
				expectedConfig.Set(propertyKey, defaultPropertiesValues[propertyKey])
			}
		}
	}

	compareViperConfigs(t, actualConfig, expectedConfig, buildType)
}

func TestCreateSimplePropertiesFile(t *testing.T) {
	var yamlConfig = map[string]string{
		RESOLVER_PREFIX + URL: "http://some.url.com",
		DEPLOYER_PREFIX + URL: "http://some.other.url.com",
		BUILD_NAME:            "buildName",
	}
	var propertiesFileConfig = map[string]string{
		"artifactory.resolve.contextUrl": yamlConfig[RESOLVER_PREFIX+URL],
		"artifactory.publish.contextUrl": yamlConfig[DEPLOYER_PREFIX+URL],
		"artifactory.deploy.build.name":  yamlConfig[BUILD_NAME],
	}

	vConfig := viper.New()
	vConfig.Set("type", MAVEN.String())
	for k, v := range yamlConfig {
		vConfig.Set(k, v)
	}
	propsFilePath, err := CreateBuildInfoPropertiesFile("", "", vConfig, MAVEN)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(propsFilePath)

	actualConfig, err := ReadConfigFile(propsFilePath, PROPERTIES)
	if err != nil {
		t.Error(err)
	}

	expectedConfig := viper.New()
	for _, partialMapping := range buildTypeConfigMapping[MAVEN] {
		for propertyKey := range *partialMapping {
			if defaultPropertiesValues[propertyKey] != "" {
				expectedConfig.Set(propertyKey, defaultPropertiesValues[propertyKey])
			}
		}
	}

	for k, v := range propertiesFileConfig {
		expectedConfig.Set(k, v)
	}

	compareViperConfigs(t, actualConfig, expectedConfig, MAVEN)
}

func TestGeneratedBuildInfoFile(t *testing.T) {
	var yamlConfig = map[string]string{
		RESOLVER_PREFIX + URL: "http://some.url.com",
		DEPLOYER_PREFIX + URL: "http://some.other.url.com",
	}
	vConfig := viper.New()
	vConfig.Set("type", MAVEN.String())
	for k, v := range yamlConfig {
		vConfig.Set(k, v)
	}
	propsFilePath, err := CreateBuildInfoPropertiesFile("buildName", "buildNumber", vConfig, MAVEN)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(propsFilePath)

	actualConfig, err := ReadConfigFile(propsFilePath, PROPERTIES)
	if err != nil {
		t.Error(err)
	}

	generatedBuildInfoKey := "buildInfo.generated.build.info"
	if !actualConfig.IsSet(generatedBuildInfoKey) {
		t.Error(generatedBuildInfoKey, "key does not exists")
	}
	if !fileutils.IsPathExists(actualConfig.GetString(generatedBuildInfoKey)) {
		t.Error("Path: ", actualConfig.GetString(generatedBuildInfoKey), "does not exists")
	}
	defer os.Remove(actualConfig.GetString(generatedBuildInfoKey))
}

func compareViperConfigs(t *testing.T, actual, expected *viper.Viper, buildType BuildType) {
	for _, key := range expected.AllKeys() {
		value := expected.GetString(key)
		if !actual.IsSet(key) {
			t.Error("["+buildType.String()+"]: Expected key was not found:", "'"+key+"'")
			continue
		}
		if actual.GetString(key) != value {
			t.Error("["+buildType.String()+"]: Expected:", "('"+key+"','"+value+"'),", "found:", "('"+key+"','"+actual.GetString(key)+"').")
		}
	}

	for _, key := range actual.AllKeys() {
		value := actual.GetString(key)
		if !expected.IsSet(key) {
			t.Error("["+buildType.String()+"]: Unexpected key, value found:", "('"+key+"','"+value+"')")
		}
	}
}
