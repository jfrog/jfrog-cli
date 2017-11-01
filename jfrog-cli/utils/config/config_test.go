package config

import (
	"testing"
	"errors"
	"encoding/json"
	"reflect"
	"os"
)

func TestCovertConfigV0ToV1(t *testing.T) {
	setJfrogHome(t)
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
	setJfrogHome(t)
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
	setJfrogHome(t)
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
	setJfrogHome(t)
	config := `
		{
		  "artifactory": [
			{
			  "url": "http://localhost:8080/artifactory/",
			  "user": "user",
			  "password": "password",
			  "serverId": "` + DEFAULT_SERVER_ID + `",
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
	setJfrogHome(t)
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
	serverDetails, err := GetDefaultArtifactoryConf(configV1.Artifactory)
	if err != nil {
		t.Error(err.Error())
	}
	if serverDetails.ServerId != "name" {
		t.Error(errors.New("Failed to get default server."))
	}

	serverDetails = GetArtifactoryConfByServerId("notDefault", configV1.Artifactory)
	if serverDetails.ServerId != "notDefault" {
		t.Error(errors.New("Failed to get server by name."))
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
	if rtConverted[0].ServerId != DEFAULT_SERVER_ID {
		t.Error(errors.New("serverId should be " + DEFAULT_SERVER_ID + "."))
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

func setJfrogHome(t *testing.T) {
	err := os.Setenv(JFROG_HOME_ENV, ".jfrogTest")
	if err != nil {
		t.Error(err.Error())
	}
}