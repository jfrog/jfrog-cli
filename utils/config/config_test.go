package config

import (
	"encoding/json"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-cli/utils/log"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
)

func init() {
	log.SetDefaultLogger()
}

const certsConversionResources = "../../testsdata/config/configconversion"
const encryptionResources = "../../testsdata/config/encryption"

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
	content, err := convertConfigV0toV1([]byte(configV0))
	assert.NoError(t, err)
	configV1 := new(ConfigV2)
	assert.NoError(t, json.Unmarshal(content, &configV1))
	assertionHelper(t, configV1, 1, false)
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
	content, err := convertConfigV0toV1([]byte(configV0))
	assert.NoError(t, err)
	configV1 := new(ConfigV2)
	assert.NoError(t, json.Unmarshal(content, &configV1))
}

func TestConvertConfigV1ToV2(t *testing.T) {
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

	tempDirPath, oldHomeDir := createTempEnv(t)
	defer os.RemoveAll(tempDirPath)
	defer os.Setenv(cliutils.HomeDir, oldHomeDir)
	copyResources(t, certsConversionResources, tempDirPath)

	content, err := convertIfNeeded([]byte(config))
	assert.NoError(t, err)
	configV2 := new(ConfigV2)
	assert.NoError(t, json.Unmarshal(content, &configV2))
	assertionHelper(t, configV2, 2, false)

	assertCertsMigrationAndBackupCreation(t)
}

func assertCertsMigrationAndBackupCreation(t *testing.T) {
	assertCertsMigration(t)
	backupDir, err := cliutils.GetJfrogBackupDir()
	assert.NoError(t, err)
	assert.DirExists(t, backupDir)
}

func TestConvertConfigV0ToV2(t *testing.T) {
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

	tempDirPath, oldHomeDir := createTempEnv(t)
	defer os.RemoveAll(tempDirPath)
	defer os.Setenv(cliutils.HomeDir, oldHomeDir)
	copyResources(t, certsConversionResources, tempDirPath)

	content, err := convertIfNeeded([]byte(configV0))
	assert.NoError(t, err)
	configV2 := new(ConfigV2)
	assert.NoError(t, json.Unmarshal(content, &configV2))
	assertionHelper(t, configV2, 2, false)
	assertCertsMigrationAndBackupCreation(t)
}

func TestConfigEncryption(t *testing.T) {
	tempDirPath, oldHomeDir := createTempEnv(t)
	defer os.RemoveAll(tempDirPath)
	defer os.Setenv(cliutils.HomeDir, oldHomeDir)
	copyResources(t, encryptionResources, tempDirPath)

	// Original decrypted config, read directly from file
	originalConfig := readConfFromFile(t)

	// Reading through this function will update encryption, and encrypt the config file.
	readConfig, err := readConf()
	assert.NoError(t, err)

	// Config file encryption should be updated, so Enc=true. Secrets should be decrypted to be used in the rest of the execution.
	assert.True(t, readConfig.Enc)
	verifyEncryptionStatus(t, originalConfig, readConfig, false)
	// Config file should be encrypted.
	encryptedConfig := readConfFromFile(t)
	verifyEncryptionStatus(t, originalConfig, encryptedConfig, true)

	// Verify successfully decrypting.
	readConfig, err = readConf()
	assert.NoError(t, err)
	assert.True(t, readConfig.Enc)
	verifyEncryptionStatus(t, originalConfig, readConfig, false)
}

func readConfFromFile(t *testing.T) *ConfigV2 {
	confFilePath, err := getConfFilePath()
	assert.NoError(t, err)
	config := new(ConfigV2)
	assert.FileExists(t, confFilePath)
	content, err := fileutils.ReadFile(confFilePath)
	assert.NoError(t, err)
	assert.NoError(t, json.Unmarshal(content, &config))
	return config
}

// Set JFROG_CLI_HOME_DIR environment variable to be a new temp directory
func createTempEnv(t *testing.T) (newHomeDir, oldHomeDir string) {
	tmpDir, err := ioutil.TempDir("", "config_test")
	assert.NoError(t, err)
	oldHome := os.Getenv(cliutils.HomeDir)
	assert.NoError(t, os.Setenv(cliutils.HomeDir, tmpDir))
	return tmpDir, oldHome
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
		  "version": "2"
		}
	`
	content, err := convertIfNeeded([]byte(config))
	assert.NoError(t, err)
	configV1 := new(ConfigV2)
	assert.NoError(t, json.Unmarshal(content, &configV1))
	serverDetails, err := GetDefaultConfiguredArtifactoryConf(configV1.Artifactory)
	assert.NoError(t, err)
	assert.Equal(t, serverDetails.ServerId, "name")

	serverDetails, err = getArtifactoryConfByServerId("notDefault", configV1.Artifactory)
	assert.NoError(t, err)
	assert.Equal(t, serverDetails.ServerId, "notDefault")
}

func TestGetJfrogDependenciesPath(t *testing.T) {
	// Check default value of dependencies path, should be JFROG_CLI_HOME/dependencies
	dependenciesPath, err := GetJfrogDependenciesPath()
	assert.NoError(t, err)
	jfrogHomeDir, err := cliutils.GetJfrogHomeDir()
	assert.NoError(t, err)
	expectedDependenciesPath := filepath.Join(jfrogHomeDir, cliutils.JfrogDependenciesDirName)
	assert.Equal(t, expectedDependenciesPath, dependenciesPath)

	// Check dependencies path when JFROG_DEPENDENCIES_DIR is set, should be JFROG_DEPENDENCIES_DIR/
	previousDependenciesDirEnv := os.Getenv(cliutils.DependenciesDir)
	expectedDependenciesPath = "/tmp/my-dependencies/"
	err = os.Setenv(cliutils.DependenciesDir, expectedDependenciesPath)
	assert.NoError(t, err)
	defer os.Setenv(cliutils.DependenciesDir, previousDependenciesDirEnv)
	dependenciesPath, err = GetJfrogDependenciesPath()
	assert.Equal(t, expectedDependenciesPath, dependenciesPath)
}

func assertionHelper(t *testing.T, convertedConfig *ConfigV2, expectedVersion int, expectedEnc bool) {
	assert.Equal(t, strconv.Itoa(expectedVersion), convertedConfig.Version)
	assert.Equal(t, expectedEnc, convertedConfig.Enc)

	rtConverted := convertedConfig.Artifactory
	if rtConverted == nil {
		assert.Fail(t, "empty Artifactory config!")
		return
	}
	assert.Len(t, rtConverted, 1)
	rtConfigType := reflect.TypeOf(rtConverted)
	assert.Equal(t, "[]*config.ArtifactoryDetails", rtConfigType.String())
	assert.True(t, rtConverted[0].IsDefault)
	assert.Equal(t, DefaultServerId, rtConverted[0].ServerId)
	assert.Equal(t, "http://localhost:8080/artifactory/", rtConverted[0].Url)
	assert.Equal(t, "user", rtConverted[0].User)
	assert.Equal(t, "password", rtConverted[0].Password)
}

func TestHandleSecrets(t *testing.T) {
	masterKey := "randomkeywithlengthofexactly32!!"

	original := new(ConfigV2)
	original.Artifactory = []*ArtifactoryDetails{{User: "user", Password: "password", Url: "http://localhost:8080/artifactory/", AccessToken: "accessToken",
		RefreshToken: "refreshToken", ApiKey: "apiKEY", SshPassphrase: "sshPass"}}
	original.Bintray = &BintrayDetails{ApiUrl: "APIurl", Key: "bintrayKey"}
	original.MissionControl = &MissionControlDetails{Url: "url", AccessToken: "mcToken"}

	newConf := copyConfig(t, original)

	// Encrypt decrypted
	assert.NoError(t, handleSecrets(original, encrypt, masterKey))
	verifyEncryptionStatus(t, original, newConf, true)

	// Decrypt encrypted
	assert.NoError(t, handleSecrets(original, decrypt, masterKey))
	verifyEncryptionStatus(t, original, newConf, false)
}

func copyConfig(t *testing.T, original *ConfigV2) *ConfigV2 {
	b, err := json.Marshal(&original)
	assert.NoError(t, err)
	newConf := new(ConfigV2)
	err = json.Unmarshal(b, &newConf)
	assert.NoError(t, err)
	return newConf
}

func verifyEncryptionStatus(t *testing.T, original, actual *ConfigV2, encryptionExpected bool) {
	var equals []bool
	for i := range actual.Artifactory {
		if original.Artifactory[i].Password != "" {
			equals = append(equals, original.Artifactory[i].Password == actual.Artifactory[i].Password)
		}
		if original.Artifactory[i].AccessToken != "" {
			equals = append(equals, original.Artifactory[i].AccessToken == actual.Artifactory[i].AccessToken)
		}
		if original.Artifactory[i].RefreshToken != "" {
			equals = append(equals, original.Artifactory[i].RefreshToken == actual.Artifactory[i].RefreshToken)
		}
		if original.Artifactory[i].ApiKey != "" {
			equals = append(equals, original.Artifactory[i].ApiKey == actual.Artifactory[i].ApiKey)
		}
	}
	if actual.Bintray != nil {
		equals = append(equals, original.Bintray.Key == actual.Bintray.Key)
	}
	if actual.MissionControl != nil {
		equals = append(equals, original.MissionControl.AccessToken == actual.MissionControl.AccessToken)
	}

	if encryptionExpected {
		// Verify non match.
		assert.Zero(t, cliutils.SumTrueValues(equals))
	} else {
		// Verify all match.
		assert.Equal(t, cliutils.SumTrueValues(equals), len(equals))
	}
}

func copyResources(t *testing.T, sourcePath string, destPath string) {
	assert.NoError(t, fileutils.CopyDir(sourcePath, destPath, true))
}

func assertCertsMigration(t *testing.T) {
	certsDir, err := cliutils.GetJfrogCertsDir()
	assert.NoError(t, err)
	assert.DirExists(t, certsDir)
	secFile, err := cliutils.GetJfrogSecurityConfFilePath()
	assert.NoError(t, err)
	assert.FileExists(t, secFile)
	files, err := ioutil.ReadDir(certsDir)
	assert.NoError(t, err)
	// Verify only the certs were moved
	assert.Len(t, files, 2)
}
