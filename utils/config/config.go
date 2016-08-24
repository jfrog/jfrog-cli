package config

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"errors"
	"os"
	"net/http"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
)

func IsArtifactoryConfExists() (bool, error) {
    conf, err := readConf()
    if err != nil {
        return false, err
    }
	return conf.Artifactory != nil, nil
}

func IsMissionControlConfExists() (bool, error) {
    conf, err := readConf()
    if err != nil {
        return false, err
    }
	return conf.MissionControl != nil, nil
}

func IsBintrayConfExists() (bool, error) {
    conf, err := readConf()
    if err != nil {
        return false, err
    }
	return conf.Bintray != nil, nil
}

func ReadArtifactoryConf() (*ArtifactoryDetails, error) {
    conf, err := readConf()
    if err != nil {
        return nil, err
    }
	details := conf.Artifactory
	if details == nil {
		return new(ArtifactoryDetails), nil
	}
	return details, nil
}

func ReadMissionControlConf() (*MissionControlDetails, error) {
    conf, err := readConf()
    if err != nil {
        return nil, err
    }
	details := conf.MissionControl
	if details == nil {
		return new(MissionControlDetails), nil
	}
	return details, nil
}

func ReadBintrayConf() (*BintrayDetails, error) {
    conf, err := readConf()
    if err != nil {
        return nil, err
    }
	details := conf.Bintray
	if details == nil {
		return new(BintrayDetails), nil
	}
	return details, nil
}

func SaveArtifactoryConf(details *ArtifactoryDetails) error {
    conf, err := readConf()
    if err != nil {
        return err
    }
	conf.Artifactory = details
	return saveConfig(conf)
	return nil
}

func SaveMissionControlConf(details *MissionControlDetails) error {
    conf, err := readConf()
    if err != nil {
        return err
    }
	conf.MissionControl = details
	return saveConfig(conf)
	return nil
}

func SaveBintrayConf(details *BintrayDetails) error {
	config, err := readConf()
    if err != nil {
        return err
    }
	config.Bintray = details
	return saveConfig(config)
}

func saveConfig(config *Config) error {
	b, err := json.Marshal(&config)
	err = cliutils.CheckError(err)
	if err != nil {
	    return err
	}
	var content bytes.Buffer
	err = json.Indent(&content, b, "", "  ")
	err = cliutils.CheckError(err)
	if err != nil {
	    return err
	}
	path, err := getConFilePath()
	if err != nil {
	    return err
	}
	var exists bool
	exists, err = ioutils.IsFileExists(path)
	if err != nil {
	    return err
	}
	if exists {
		err := os.Remove(path)
		err = cliutils.CheckError(err)
        if err != nil {
            return err
        }
	}
	path, err = getConFilePath()
    if err != nil {
        return err
    }
	ioutil.WriteFile(path, []byte(content.String()), 0600)
	return nil
}

func readConf() (*Config, error) {
	confFilePath, err := getConFilePath()
	if err != nil {
	    return nil, err
	}
	config := new(Config)
	exists, err := ioutils.IsFileExists(confFilePath)
	if err != nil {
	    return nil, err
	}
	if !exists {
		return config, nil
	}
	content, err := ioutils.ReadFile(confFilePath)
	if err != nil {
	    return nil, err
	}
	json.Unmarshal(content, &config)

	return config, nil
}

func GetJfrogHomeDir() (string, error) {
	userDir := ioutils.GetHomeDir()
	if userDir == "" {
		err := cliutils.CheckError(errors.New("Couldn't find home directory. Make sure your HOME environment variable is set."))
        if err != nil {
            return "", err
        }
	}
	return userDir + "/.jfrog/", nil
}

func getConFilePath() (string, error) {
	confPath, err := GetJfrogHomeDir()
    if err != nil {
        return "", err
    }
	os.MkdirAll(confPath, 0777)
	return confPath + "jfrog-cli.conf", nil
}

type Config struct {
	Artifactory    *ArtifactoryDetails    `json:"artifactory,omitempty"`
	Bintray        *BintrayDetails        `json:"bintray,omitempty"`
	MissionControl *MissionControlDetails `json:"MissionControl,omitempty"`
}

type ArtifactoryDetails struct {
	Url            string            `json:"url,omitempty"`
	User           string            `json:"user,omitempty"`
	Password       string            `json:"password,omitempty"`
	ApiKey         string            `json:"apiKey,omitempty"`
	SshKeyPath     string            `json:"sshKeyPath,omitempty"`
	SshAuthHeaders map[string]string `json:"-"`
	Transport      *http.Transport   `json:"-"`
}

type BintrayDetails struct {
	ApiUrl             string `json:"-"`
	DownloadServerUrl  string `json:"-"`
	User               string `json:"user,omitempty"`
	Key                string `json:"key,omitempty"`
	DefPackageLicenses string `json:"defPackageLicense,omitempty"`
}

type MissionControlDetails struct {
	Url      string `json:"url,omitempty"`
	User     string `json:"user,omitempty"`
	Password string `json:"password,omitempty"`
}

func (artifactoryDetails *ArtifactoryDetails) SetApiKey(apiKey string) {
	artifactoryDetails.ApiKey = apiKey
}

func (artifactoryDetails *ArtifactoryDetails) SetUser(username string) {
	artifactoryDetails.User = username
}

func (artifactoryDetails *ArtifactoryDetails) SetPassword(password string) {
	artifactoryDetails.Password = password
}

func (artifactoryDetails *ArtifactoryDetails) GetApiKey() string {
	return artifactoryDetails.ApiKey
}

func (artifactoryDetails *ArtifactoryDetails) GetUser() string {
	return artifactoryDetails.User
}

func (artifactoryDetails *ArtifactoryDetails) GetPassword() string {
	return artifactoryDetails.Password
}

func (missionControlDetails *MissionControlDetails) SetUser(username string) {
	missionControlDetails.User = username
}

func (missionControlDetails *MissionControlDetails) SetPassword(password string) {
	missionControlDetails.Password = password
}

func (missionControlDetails *MissionControlDetails) GetUser() string {
	return missionControlDetails.User
}

func (missionControlDetails *MissionControlDetails) GetPassword() string {
	return missionControlDetails.Password
}