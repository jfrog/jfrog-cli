package utils

import (
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/spf13/viper"
	"os"
	"testing"
)

const (
	host  = "localhost"
	port  = "8888"
	proxy = "http://" + host + ":" + port
)

func TestCreateDefaultPropertiesFile(t *testing.T) {
	proxyOrg := getOriginalProxyValue()
	setProxy("", t)

	for index := range BuildTypes {
		testCreateDefaultPropertiesFile(BuildType(index), t)
	}
	setProxy(proxyOrg, t)
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

func TestCreateSimplePropertiesFileWithProxy(t *testing.T) {
	proxyOrg := getOriginalProxyValue()
	setProxy(proxy, t)
	var propertiesFileConfig = map[string]string{
		"artifactory.resolve.contextUrl": "http://some.url.com",
		"artifactory.publish.contextUrl": "http://some.other.url.com",
		"artifactory.deploy.build.name":  "buildName",
		"artifactory.proxy.host":         host,
		"artifactory.proxy.port":         port,
	}
	createSimplePropertiesFile(t, propertiesFileConfig)
	setProxy(proxyOrg, t)
}

func TestCreateSimplePropertiesFileWithoutProxy(t *testing.T) {
	proxyOrg := getOriginalProxyValue()
	setProxy("", t)
	var propertiesFileConfig = map[string]string{
		"artifactory.resolve.contextUrl": "http://some.url.com",
		"artifactory.publish.contextUrl": "http://some.other.url.com",
		"artifactory.deploy.build.name":  "buildName",
	}
	createSimplePropertiesFile(t, propertiesFileConfig)
	setProxy(proxyOrg, t)

}

func createSimplePropertiesFile(t *testing.T, propertiesFileConfig map[string]string) {
	var yamlConfig = map[string]string{
		RESOLVER_PREFIX + URL: "http://some.url.com",
		DEPLOYER_PREFIX + URL: "http://some.other.url.com",
		BUILD_NAME:            "buildName",
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
	if !fileutils.IsPathExists(actualConfig.GetString(generatedBuildInfoKey), false) {
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

func TestSetProxyIfNeeded(t *testing.T) {
	proxyOrg := getOriginalProxyValue()
	setProxy(proxy, t)
	vConfig := viper.New()

	err := setProxyIfDefined(vConfig)
	if err != nil {
		t.Error(err)
	}

	expectedConfig := viper.New()
	expectedConfig.Set(PROXY+HOST, host)
	expectedConfig.Set(PROXY+PORT, port)
	compareViperConfigs(t, vConfig, expectedConfig, MAVEN)

	setProxy(proxyOrg, t)
}

func getOriginalProxyValue() string {
	return os.Getenv(HttpProxy)
}

func setProxy(proxy string, t *testing.T) {
	err := os.Setenv(HttpProxy, proxy)
	if err != nil {
		t.Error(err)
	}
}
