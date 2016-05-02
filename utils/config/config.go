package config

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"net/http"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
)

func IsArtifactoryConfExists() bool {
	return readConf().Artifactory != nil
}

func IsMissionControlConfExists() bool {
	return readConf().MissionControl != nil
}

func IsBintrayConfExists() bool {
	return readConf().Bintray != nil
}

func ReadArtifactoryConf() *ArtifactoryDetails {
	details := readConf().Artifactory
	if details == nil {
		return new(ArtifactoryDetails)
	}
	return details
}

func ReadMissionControlConf() *MissionControlDetails {
	details := readConf().MissionControl
	if details == nil {
		return new(MissionControlDetails)
	}
	return details
}

func ReadBintrayConf() *BintrayDetails {
	details := readConf().Bintray
	if details == nil {
		return new(BintrayDetails)
	}
	return details
}

func SaveArtifactoryConf(details *ArtifactoryDetails) {
	config := readConf()
	config.Artifactory = details
	saveConfig(config)
}

func SaveMissionControlConf(details *MissionControlDetails) {
	config := readConf()
	config.MissionControl = details
	saveConfig(config)
}

func SaveBintrayConf(details *BintrayDetails) {
	config := readConf()
	config.Bintray = details
	saveConfig(config)
}

func saveConfig(config *Config) {
	b, err := json.Marshal(&config)
	cliutils.CheckError(err)
	var content bytes.Buffer
	err = json.Indent(&content, b, "", "  ")
	cliutils.CheckError(err)
	path := getConFilePath()
	if ioutils.IsFileExists(path) {
        err := os.Remove(path)
		cliutils.CheckError(err)
	}
	ioutil.WriteFile(getConFilePath(), []byte(content.String()), 0600)
}

func readConf() *Config {
	confFilePath := getConFilePath()
	config := new(Config)
	if !ioutils.IsFileExists(confFilePath) {
		return config
	}
	content := ioutils.ReadFile(confFilePath)
	json.Unmarshal(content, &config)

	return config
}

func GetJfrogHomeFolder() string {
	userDir := ioutils.GetHomeDir()
	if userDir == "" {
		cliutils.Exit(cliutils.ExitCodeError, "Couldn't find home directory. Make sure your HOME environment variable is set.")
	}
	return userDir + "/.jfrog/"
}

func getConFilePath() string {
	confPath := GetJfrogHomeFolder()
	os.MkdirAll(confPath, 0777)
	return confPath + "jfrog-cli.conf"
}

type Config struct {
	Artifactory    *ArtifactoryDetails 		 `json:"artifactory,omitempty"`
	Bintray        *BintrayDetails     		 `json:"bintray,omitempty"`
	MissionControl *MissionControlDetails    `json:"MissionControl,omitempty"`
}

type ArtifactoryDetails struct {
	Url            string            `json:"url,omitempty"`
	User           string            `json:"user,omitempty"`
	Password       string            `json:"password,omitempty"`
	SshKeyPath     string            `json:"sshKeyPath,omitempty"`
	SshAuthHeaders map[string]string `json:"-"`
	Transport      *http.Transport	 `json:"-"`
}

type BintrayDetails struct {
	ApiUrl             string `json:"-"`
	DownloadServerUrl  string `json:"-"`
	User               string `json:"user,omitempty"`
	Key                string `json:"key,omitempty"`
	DefPackageLicenses string `json:"defPackageLicense,omitempty"`
}

type MissionControlDetails struct {
	Url            string            `json:"url,omitempty"`
	User           string            `json:"user,omitempty"`
	Password       string            `json:"password,omitempty"`
}

func (artifactoryDetails *ArtifactoryDetails) SetUser(username string){
	artifactoryDetails.User = username
}

func (artifactoryDetails *ArtifactoryDetails) SetPassword(password string){
	artifactoryDetails.Password = password
}

func (artifactoryDetails *ArtifactoryDetails) GetUser() string{
	return artifactoryDetails.User
}

func (artifactoryDetails *ArtifactoryDetails) GetPassword() string{
	return artifactoryDetails.Password
}

func (missionControlDetails *MissionControlDetails) SetUser(username string){
	missionControlDetails.User = username
}

func (missionControlDetails *MissionControlDetails) SetPassword(password string){
	missionControlDetails.Password = password
}

func (missionControlDetails *MissionControlDetails) GetUser() string{
	return missionControlDetails.User
}

func (missionControlDetails *MissionControlDetails) GetPassword() string{
	return missionControlDetails.Password
}