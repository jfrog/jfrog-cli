package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestCovertConfigV0ToV1(t *testing.T) {
	configV0 := `
		{
		  "artifactory": {
			  "url": "http://localhost:8080/artifactory/",
			  "user": "user",
			  "password": "password"
		  },
		  "bintray": {
			"user": "user",
			"key": "api-key",
			"defPackageLicense": "Apache-2.0"
		  }
		}
	`
	content, err := convertIfNecessary([]byte(configV0))
	if err != nil {
		t.Error(err.Error())
	}
	configV1 := new(ConfigV1)
	err = json.Unmarshal(content, &configV1)
	if err != nil {
		t.Error(err.Error())
	}
	assertionHelper(configV1, t)
}

func TestCovertConfigV0ToV1EmptyArtifactory(t *testing.T) {
	configV0 := `
		{
		  "bintray": {
			"user": "user",
			"key": "api-key",
			"defPackageLicense": "Apache-2.0"
		  }
		}
	`
	content, err := convertIfNecessary([]byte(configV0))
	if err != nil {
		t.Error(err.Error())
	}
	configV1 := new(ConfigV1)
	err = json.Unmarshal(content, &configV1)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestConfigV1EmptyArtifactory(t *testing.T) {
	configV0 := `
		{
		  "bintray": {
			"user": "user",
			"key": "api-key",
			"defPackageLicense": "Apache-2.0"
		  },
		  "Version": "1"
		}
	`
	content, err := convertIfNecessary([]byte(configV0))
	if err != nil {
		t.Error(err.Error())
	}
	configV1 := new(ConfigV1)
	err = json.Unmarshal(content, &configV1)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestConfigV1Covert(t *testing.T) {
	config := `
		{
		  "artifactory": [
			{
			  "url": "http://localhost:8080/artifactory/",
			  "user": "user",
			  "password": "password",
			  "serverId": "` + DefaultServerId + `",
			  "isDefault": true
			}
		  ],
		  "bintray": {
			"user": "user",
			"key": "api-key",
			"defPackageLicense": "Apache-2.0"
		  },
		  "Version": "1"
		}
	`
	content, err := convertIfNecessary([]byte(config))
	if err != nil {
		t.Error(err.Error())
	}
	configV1 := new(ConfigV1)
	err = json.Unmarshal(content, &configV1)
	if err != nil {
		t.Error(err.Error())
	}
	assertionHelper(configV1, t)
}

func TestGetArtifactoriesFromConfig(t *testing.T) {
	config := `
		{
		  "artifactory": [
			{
			  "url": "http://localhost:8080/artifactory/",
			  "user": "user",
			  "password": "password",
			  "serverId": "name",
			  "isDefault": true
			},
			{
			  "url": "http://localhost:8080/artifactory/",
			  "user": "user",
			  "password": "password",
			  "serverId": "notDefault"
			}
		  ],
		  "bintray": {
			"user": "user",
			"key": "api-key",
			"defPackageLicense": "Apache-2.0"
		  },
		  "Version": "1"
		}
	`
	content, err := convertIfNecessary([]byte(config))
	if err != nil {
		t.Error(err.Error())
	}
	configV1 := new(ConfigV1)
	err = json.Unmarshal(content, &configV1)
	if err != nil {
		t.Error(err.Error())
	}
	serverDetails, err := GetDefaultConfiguredArtifactoryConf(configV1.Artifactory)
	if err != nil {
		t.Error(err.Error())
	}
	if serverDetails.ServerId != "name" {
		t.Error(errors.New("Failed to get default server."))
	}

	serverDetails, err = getArtifactoryConfByServerId("notDefault", configV1.Artifactory)
	if err != nil {
		t.Error(err.Error())
	}
	if serverDetails.ServerId != "notDefault" {
		t.Error(errors.New("Failed to get server by name."))
	}
}

func TestGetJfrogDependenciesPath(t *testing.T) {
	// Check default value of dependencies path, should be JFROG_CLI_HOME/dependencies
	dependenciesPath, err := GetJfrogDependenciesPath()
	if err != nil {
		t.Error(err.Error())
	}
	jfrogHomeDir, err := cliutils.GetJfrogHomeDir()
	if err != nil {
		t.Error(err.Error())
	}
	expectedDependenciesPath := filepath.Join(jfrogHomeDir, JfrogDependencies)
	if expectedDependenciesPath != dependenciesPath {
		t.Error(errors.New(fmt.Sprintf("Dependencies Path should be %s (actual path: %s)", expectedDependenciesPath, dependenciesPath)))
	}
	// Check dependencies path when JFROG_DEPENDENCIES_DIR is set, should be JFROG_DEPENDENCIES_DIR/
	previousDependenciesDirEnv := os.Getenv(cliutils.DependenciesDir)
	expectedDependenciesPath = "/tmp/my-dependencies/"
	err = os.Setenv(cliutils.DependenciesDir, expectedDependenciesPath)
	if err != nil {
		t.Error(err.Error())
	}
	defer os.Setenv(cliutils.DependenciesDir, previousDependenciesDirEnv)
	dependenciesPath, err = GetJfrogDependenciesPath()
	if expectedDependenciesPath != dependenciesPath {
		t.Error(errors.New(fmt.Sprintf("Dependencies Path should be %s (actual path: %s)", expectedDependenciesPath, dependenciesPath)))
	}
}

func assertionHelper(configV1 *ConfigV1, t *testing.T) {
	if configV1.Version != "1" {
		t.Error(errors.New("Failed to convert config version."))
	}
	rtConverted := configV1.Artifactory
	if rtConverted == nil {
		t.Error(errors.New("Empty Artifactory config!."))
	}
	if len(rtConverted) != 1 {
		t.Error(errors.New("Conversion failed!"))
	}
	rtConfigType := reflect.TypeOf(rtConverted)
	if rtConfigType.String() != "[]*config.ArtifactoryDetails" {
		t.Error(errors.New("Couldn't convert to Array."))
	}
	if rtConverted[0].IsDefault != true {
		t.Error(errors.New("Should be default."))
	}
	if rtConverted[0].ServerId != DefaultServerId {
		t.Error(errors.New("serverId should be " + DefaultServerId + "."))
	}
	if rtConverted[0].Url != "http://localhost:8080/artifactory/" {
		t.Error(errors.New("Url shouldn't change."))
	}
	if rtConverted[0].User != "user" {
		t.Error(errors.New("User shouldn't change."))
	}
	if rtConverted[0].Password != "password" {
		t.Error(errors.New("Password shouldn't change."))
	}
}

func TestEncryption(t *testing.T) {
	EncryptionFile = "{\"Keys\":[\"randomkeywithlengthofexactly32!!\",\"anotherkeywithlengthofexactly32!\"],\"Version\":\"1\"}"
	details := map[string]string{
		"password":     "password",
		"accessToken":  "accessToken",
		"refreshToken": "refreshToken",
		"apiKEY":       "apiKEY",
		"sshPass":      "sshPass",
		"bintrayKey":   "bintrayKey",
		"mcToken":      "mcToken",
	}

	config := new(ConfigV1)
	config.Artifactory = []*ArtifactoryDetails{{User: "user", Password: details["password"], Url: "http://localhost:8080/artifactory/", AccessToken: details["accessToken"],
		RefreshToken: details["refreshToken"], ApiKey: details["apiKEY"], SshPassphrase: details["sshPass"]}}
	config.Bintray = &BintrayDetails{ApiUrl: "APIurl", Key: details["bintrayKey"]}
	config.MissionControl = &MissionControlDetails{Url: "url", AccessToken: details["mcToken"]}

	// Encrypt decrypted
	assert.NoError(t, handleSecrets(config, encrypt, 0))
	verifyAllEncrypted(t, config, details, true)

	// Decrypt encrypted
	assert.NoError(t, handleSecrets(config, decrypt, 0))
	verifyAllEncrypted(t, config, details, false)
}

func verifyAllEncrypted(t *testing.T, config *ConfigV1, originalValues map[string]string, checkEncrypted bool) {
	equals := []bool{
		config.Artifactory[0].Password == originalValues["password"],
		config.Artifactory[0].AccessToken == originalValues["accessToken"],
		config.Artifactory[0].RefreshToken == originalValues["refreshToken"],
		config.Artifactory[0].ApiKey == originalValues["apiKEY"],
		config.Artifactory[0].SshPassphrase == originalValues["sshPass"],
		config.Bintray.Key == originalValues["bintrayKey"],
		config.MissionControl.AccessToken == originalValues["mcToken"],
	}

	if checkEncrypted {
		// Verify non match.
		assert.Zero(t, cliutils.SumTrueValues(equals))
	} else {
		// Verify all match.
		assert.Equal(t, cliutils.SumTrueValues(equals), len(equals))
	}
}
