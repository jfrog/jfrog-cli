package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jfrog/jfrog-cli-go/utils/cliutils"
	"os"
	"path/filepath"
	"reflect"
	"strings"
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
	jfrogHomeDir, err := GetJfrogHomeDir()
	if err != nil {
		t.Error(err.Error())
	}
	expectedDependenciesPath := filepath.Join(jfrogHomeDir, JfrogDependencies)
	if strings.Compare(expectedDependenciesPath, dependenciesPath) != 0 {
		t.Error(errors.New(fmt.Sprintf("Dependencies Path should be %s (actual path: %s)", expectedDependenciesPath, dependenciesPath)))
	}
	// Check dependencies path when JFROG_DEPENDENCIES_DIR is set, should be JFROG_DEPENDENCIES_DIR/
	err = os.Setenv(cliutils.JFrogCliDependenciesDir, "/tmp/my-dependencies")
	if err != nil {
		t.Error(err.Error())
	}
	expectedDependenciesPath = "/tmp/my-dependencies/"
	dependenciesPath, err = GetJfrogDependenciesPath()
	if strings.Compare(expectedDependenciesPath, dependenciesPath) != 0 {
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
