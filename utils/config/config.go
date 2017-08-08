package config

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"errors"
	"os"
	"net/http"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/io/fileutils"
	"github.com/buger/jsonparser"
)

// This is the default server id. It is used when adding a server config without providing a server ID
const DefaultServerId = "Default-Server"

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

func GetArtifactorySpecificConfig(serverId string) (*ArtifactoryDetails, error) {
	conf, err := readConf()
	if err != nil {
		return nil, err
	}
	details := conf.Artifactory
	if details == nil || len(details) == 0 {
		return new(ArtifactoryDetails), nil
	}
	var artifactoryDetails *ArtifactoryDetails
	if len(serverId) == 0 {
		artifactoryDetails, err = GetDefaultArtifactoryConf(details)
	} else {
		artifactoryDetails = GetArtifactoryConfByServerId(serverId, details)
	}
	return artifactoryDetails, err
}

func GetDefaultArtifactoryConf(configs []*ArtifactoryDetails) (*ArtifactoryDetails, error) {
	if len(configs) == 0 {
		details := new(ArtifactoryDetails)
		details.IsDefault = true
		return details, nil
	}
	for _, conf := range configs {
		if conf.IsDefault == true {
			return conf, nil
		}
	}
	return nil, cliutils.CheckError(errors.New("Couldn't find default server."))
}

func GetArtifactoryConfByServerId(serverName string, configs []*ArtifactoryDetails) (*ArtifactoryDetails) {
	for _, conf := range configs {
		if conf.ServerId == serverName {
			return conf
		}
	}
	return new(ArtifactoryDetails)
}

func GetAllArtifactoryConfigs() ([]*ArtifactoryDetails, error) {
	conf, err := readConf()
	if err != nil {
		return nil, err
	}
	details := conf.Artifactory
	if details == nil {
		return make([]*ArtifactoryDetails, 0), nil
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

func SaveArtifactoryConf(details []*ArtifactoryDetails) error {
	conf, err := readConf()
	if err != nil {
		return err
	}
	conf.Artifactory = details
	return saveConfig(conf)
}

func SaveMissionControlConf(details *MissionControlDetails) error {
    conf, err := readConf()
    if err != nil {
        return err
    }
	conf.MissionControl = details
	return saveConfig(conf)
}

func SaveBintrayConf(details *BintrayDetails) error {
	config, err := readConf()
    if err != nil {
        return err
    }
	config.Bintray = details
	return saveConfig(config)
}

func saveConfig(config *ConfigV1) error {
	config.Version = cliutils.GetConfigVersion()
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
	exists, err = fileutils.IsFileExists(path)
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

func readConf() (*ConfigV1, error) {
	confFilePath, err := getConFilePath()
	if err != nil {
		return nil, err
	}
	config := new(ConfigV1)
	exists, err := fileutils.IsFileExists(confFilePath)
	if err != nil {
		return nil, err
	}
	if !exists {
		return config, nil
	}
	content, err := fileutils.ReadFile(confFilePath)
	if err != nil {
		return nil, err
	}
	if len(content) == 0 {
		return new(ConfigV1), nil
	}
	content, err = convertIfNecessary(content)
	err = json.Unmarshal(content, &config)
	return config, err
}

// The configuration schema can change between versions, therefore we need to convert old versions to the new schema.
func convertIfNecessary(content []byte) ([]byte, error) {
	version, err := jsonparser.GetString(content, "Version")
	if err != nil {
		if err.Error() == "Key path not found" {
			version = "0"
		} else {
			return nil, cliutils.CheckError(err)
		}
	}
	switch version {
	case "0":
		result := new(ConfigV1)
		configV0 := new(ConfigV0)
		err = json.Unmarshal(content, &configV0)
		if cliutils.CheckError(err) != nil {
			return nil, err
		}
		result = configV0.Convert()
		err = saveConfig(result)
		content, err = json.Marshal(&result)
	}
	return content, err
}

func GetJfrogHomeDir() (string, error) {
	userDir := fileutils.GetHomeDir()
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

type ConfigV1 struct {
	Artifactory    []*ArtifactoryDetails  `json:"artifactory"`
	Bintray        *BintrayDetails        `json:"bintray,omitempty"`
	MissionControl *MissionControlDetails `json:"MissionControl,omitempty"`
	Version        string                 `json:"Version,omitempty"`
}

type ConfigV0 struct {
	Artifactory    *ArtifactoryDetails    `json:"artifactory,omitempty"`
	Bintray        *BintrayDetails        `json:"bintray,omitempty"`
	MissionControl *MissionControlDetails `json:"MissionControl,omitempty"`
}

func (o *ConfigV0) Convert() *ConfigV1 {
	config := new(ConfigV1)
	config.Bintray = o.Bintray
	config.MissionControl = o.MissionControl
	if o.Artifactory != nil {
		o.Artifactory.IsDefault = true
		o.Artifactory.ServerId = DefaultServerId
		config.Artifactory = []*ArtifactoryDetails{o.Artifactory}
	}
	return config
}

type ArtifactoryDetails struct {
	Url            string            `json:"url,omitempty"`
	User           string            `json:"user,omitempty"`
	Password       string            `json:"password,omitempty"`
	ApiKey         string            `json:"apiKey,omitempty"`
	Ssh	           bool              `json:"ssh,omitempty"`
	SshKeyPath     string            `json:"sshKeyPath,omitempty"`
	ServerId       string            `json:"serverId,omitempty"`
	IsDefault      bool              `json:"isDefault,omitempty"`
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

func (artifactoryDetails *ArtifactoryDetails) IsEmpty() bool {
	return len(artifactoryDetails.Url) == 0
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
