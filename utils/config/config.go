package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	artifactoryAuth "github.com/jfrog/jfrog-client-go/artifactory/auth"
	"github.com/jfrog/jfrog-client-go/auth"
	distributionAuth "github.com/jfrog/jfrog-client-go/distribution/auth"
	"github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// This is the default server id. It is used when adding a server config without providing a server ID
const DefaultServerId = "Default-Server"

func IsArtifactoryConfExists() (bool, error) {
	conf, err := readConf()
	if err != nil {
		return false, err
	}
	return conf.Artifactory != nil && len(conf.Artifactory) > 0, nil
}

func IsMissionControlConfExists() (bool, error) {
	conf, err := readConf()
	if err != nil {
		return false, err
	}
	return conf.MissionControl != nil && *conf.MissionControl != MissionControlDetails{}, nil
}

func IsBintrayConfExists() (bool, error) {
	conf, err := readConf()
	if err != nil {
		return false, err
	}
	return conf.Bintray != nil, nil
}

// Returns the configured server or error if the server id was not found.
// If defaultOrEmpty: return empty details if no configurations found, or default conf for empty serverId.
// Exclude refreshable tokens when working with external tools (build tools, curl, etc) or when sending requests not via ArtifactoryHttpClient.
func GetArtifactorySpecificConfig(serverId string, defaultOrEmpty bool, excludeRefreshableTokens bool) (*ArtifactoryDetails, error) {
	configs, err := GetAllArtifactoryConfigs()
	if err != nil {
		return nil, err
	}
	if defaultOrEmpty {
		if len(configs) == 0 {
			return new(ArtifactoryDetails), nil
		}
		if len(serverId) == 0 {
			details, err := GetDefaultConfiguredArtifactoryConf(configs)
			return details, errorutils.CheckError(err)
		}
	}

	details, err := getArtifactoryConfByServerId(serverId, configs)
	if err != nil {
		return nil, err
	}
	if excludeRefreshableTokens {
		if details.AccessToken != "" && details.RefreshToken != "" {
			details.AccessToken = ""
			details.RefreshToken = ""
		}
		details.TokenRefreshInterval = cliutils.TokenRefreshDisabled
	}
	return details, nil
}

// Returns the default server configuration or error if not found.
// Caller should perform the check error if required.
func GetDefaultConfiguredArtifactoryConf(configs []*ArtifactoryDetails) (*ArtifactoryDetails, error) {
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
	return nil, errors.New("Couldn't find default server.")
}

// Returns default artifactory conf. Returns nil if default server doesn't exists.
func GetDefaultArtifactoryConf() (*ArtifactoryDetails, error) {
	configurations, err := GetAllArtifactoryConfigs()
	if err != nil {
		return nil, err
	}

	if len(configurations) == 0 {
		log.Debug("No servers were configured.")
		return nil, err
	}

	return GetDefaultConfiguredArtifactoryConf(configurations)
}

// Returns the configured server or error if the server id not found
func getArtifactoryConfByServerId(serverId string, configs []*ArtifactoryDetails) (*ArtifactoryDetails, error) {
	for _, conf := range configs {
		if conf.ServerId == serverId {
			return conf, nil
		}
	}
	return nil, errorutils.CheckError(errors.New(fmt.Sprintf("Server ID '%s' does not exist.", serverId)))
}

func GetAndRemoveConfiguration(serverName string, configs []*ArtifactoryDetails) (*ArtifactoryDetails, []*ArtifactoryDetails) {
	for i, conf := range configs {
		if conf.ServerId == serverName {
			configs = append(configs[:i], configs[i+1:]...)
			return conf, configs
		}
	}
	return nil, configs
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

func saveConfig(config *ConfigV2) error {
	config.Version = cliutils.GetConfigVersion()
	err := config.encrypt()
	if err != nil {
		return err
	}

	content, err := config.getContent()
	if err != nil {
		return err
	}

	path, err := getConfFilePath()
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, content, 0600)
	if err != nil {
		return errorutils.CheckError(err)
	}
	return nil
}

func readConf() (*ConfigV2, error) {
	confFilePath, err := getConfFilePath()
	if err != nil {
		return nil, err
	}
	config := new(ConfigV2)
	exists, err := fileutils.IsFileExists(confFilePath, false)
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
		return new(ConfigV2), nil
	}
	content, err = convertIfNeeded(content)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(content, &config)
	if err != nil {
		return nil, errorutils.CheckError(err)
	}

	err = config.decrypt()
	return config, err
}

func (config *ConfigV2) getContent() ([]byte, error) {
	b, err := json.Marshal(&config)
	if err != nil {
		return []byte{}, errorutils.CheckError(err)
	}
	var content bytes.Buffer
	err = json.Indent(&content, b, "", "  ")
	if err != nil {
		return []byte{}, errorutils.CheckError(err)
	}
	return []byte(content.String()), nil
}

// Move SSL certificates from the old location in security dir to certs dir.
func convertCertsDir() error {
	securityDir, err := cliutils.GetJfrogSecurityDir()
	if err != nil {
		return err
	}
	exists, err := fileutils.IsDirExists(securityDir, false)
	// Security dir doesn't exist, no conversion needed.
	if err != nil || !exists {
		return err
	}

	certsDir, err := cliutils.GetJfrogCertsDir()
	if err != nil {
		return err
	}
	exists, err = fileutils.IsDirExists(certsDir, false)
	// Certs dir already exists, no conversion needed.
	if err != nil || exists {
		return err
	}

	// Move certs to the new location.
	files, err := ioutil.ReadDir(securityDir)
	if err != nil {
		return errorutils.CheckError(err)
	}

	log.Debug("Migrating SSL certificates to the new location at: " + certsDir)
	for _, f := range files {
		// Skip directories and the security configuration file
		if !f.IsDir() && f.Name() != cliutils.JfrogSecurityConfFile {
			err = fileutils.CreateDirIfNotExist(certsDir)
			if err != nil {
				return err
			}
			err = os.Rename(filepath.Join(securityDir, f.Name()), filepath.Join(certsDir, f.Name()))
			if err != nil {
				return errorutils.CheckError(err)
			}
		}
	}
	return nil
}

// The configuration schema can change between versions, therefore we need to convert old versions to the new schema.
func convertIfNeeded(content []byte) ([]byte, error) {
	version, exists, err := getKeyFromConfig(content, "version")
	if err != nil {
		return nil, err
	}

	// If lower case "version" exists, version is 2 or higher
	if !exists {
		version, exists, err = getKeyFromConfig(content, "Version")
		if err != nil {
			return nil, err
		}
		// Config version 0 is before introducing the "Version" field
		if !exists {
			version = "0"
		}
	}

	// Switch contains FALLTHROUGH to convert from a certain version to the latest.
	switch version {
	case "2":
		return content, nil
	case "0":
		content, err = convertConfigV0toV1(content)
		if err != nil {
			return nil, err
		}
		fallthrough
	case "1":
		log.Debug("Converting JFrog CLI's config to the latest version...")
		err = createHomeDirBackup()
		if err != nil {
			return nil, err
		}
		err = convertCertsDir()
		if err != nil {
			return nil, err
		}
	}

	// Save config after all conversions (also updates version).
	result := new(ConfigV2)
	err = json.Unmarshal(content, &result)
	if errorutils.CheckError(err) != nil {
		return nil, err
	}
	err = saveConfig(result)
	if err != nil {
		return nil, err
	}
	content, err = json.Marshal(&result)
	if err != nil {
		return nil, errorutils.CheckError(err)
	}
	return content, err
}

// Creating a homedir backup prior to converting.
func createHomeDirBackup() error {
	homeDir, err := cliutils.GetJfrogHomeDir()
	if err != nil {
		return err
	}
	backupDir, err := cliutils.GetJfrogBackupDir()
	if err != nil {
		return err
	}

	// Copy to temp dir before creating back up dir
	tempDirPath, err := fileutils.CreateTempDir()
	if err != nil {
		return err
	}
	defer fileutils.RemoveTempDir(tempDirPath)
	err = fileutils.CopyDir(homeDir, tempDirPath, true)
	if err != nil {
		return err
	}

	// Create backup dir and copy contents from temp dir
	backupName := ".jfrog-" + strconv.FormatInt(time.Now().Unix(), 10)
	curBackupPath := filepath.Join(backupDir, backupName)
	log.Debug("Creating a homedir backup at: " + curBackupPath)
	return fileutils.CopyDir(tempDirPath, curBackupPath, true)
}

func getKeyFromConfig(content []byte, key string) (value string, exists bool, err error) {
	value, err = jsonparser.GetString(content, key)
	if err != nil {
		if err.Error() == "Key path not found" {
			return "", false, nil
		} else {
			return "", false, errorutils.CheckError(err)
		}
	}
	return value, true, nil
}

func convertConfigV0toV1(content []byte) ([]byte, error) {
	result := new(ConfigV2)
	configV0 := new(ConfigV0)
	err := json.Unmarshal(content, &configV0)
	if errorutils.CheckError(err) != nil {
		return nil, err
	}
	result = configV0.Convert()
	result.Version = "1"
	content, err = json.Marshal(&result)
	return content, errorutils.CheckError(err)
}

func GetJfrogDependenciesPath() (string, error) {
	dependenciesDir := os.Getenv(cliutils.DependenciesDir)
	if dependenciesDir != "" {
		return utils.AddTrailingSlashIfNeeded(dependenciesDir), nil
	}
	jfrogHome, err := cliutils.GetJfrogHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(jfrogHome, cliutils.JfrogDependenciesDirName), nil
}

func getConfFilePath() (string, error) {
	confPath, err := cliutils.GetJfrogHomeDir()
	if err != nil {
		return "", err
	}
	os.MkdirAll(confPath, 0777)
	return filepath.Join(confPath, cliutils.JfrogConfigFile), nil
}

// This struct is suitable for both version 1 and 2.
type ConfigV2 struct {
	Artifactory    []*ArtifactoryDetails  `json:"artifactory"`
	Bintray        *BintrayDetails        `json:"bintray,omitempty"`
	MissionControl *MissionControlDetails `json:"missionControl,omitempty"`
	Version        string                 `json:"version,omitempty"`
	Enc            bool                   `json:"enc,omitempty"`
}

// This struct was created before the version property was added to the config.
type ConfigV0 struct {
	Artifactory    *ArtifactoryDetails    `json:"artifactory,omitempty"`
	Bintray        *BintrayDetails        `json:"bintray,omitempty"`
	MissionControl *MissionControlDetails `json:"MissionControl,omitempty"`
}

func (o *ConfigV0) Convert() *ConfigV2 {
	config := new(ConfigV2)
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
	Url                  string `json:"url,omitempty"`
	SshUrl               string `json:"-"`
	DistributionUrl      string `json:"distributionUrl,omitempty"`
	User                 string `json:"user,omitempty"`
	Password             string `json:"password,omitempty"`
	SshKeyPath           string `json:"sshKeyPath,omitempty"`
	SshPassphrase        string `json:"SshPassphrase,omitempty"`
	AccessToken          string `json:"accessToken,omitempty"`
	RefreshToken         string `json:"refreshToken,omitempty"`
	TokenRefreshInterval int    `json:"tokenRefreshInterval,omitempty"`
	ClientCertPath       string `json:"clientCertPath,omitempty"`
	ClientCertKeyPath    string `json:"clientCertKeyPath,omitempty"`
	ServerId             string `json:"serverId,omitempty"`
	IsDefault            bool   `json:"isDefault,omitempty"`
	InsecureTls          bool   `json:"-"`
	// Deprecated, use password option instead.
	ApiKey string `json:"apiKey,omitempty"`
}

type BintrayDetails struct {
	ApiUrl            string `json:"-"`
	DownloadServerUrl string `json:"-"`
	User              string `json:"user,omitempty"`
	Key               string `json:"key,omitempty"`
	DefPackageLicense string `json:"defPackageLicense,omitempty"`
}

type MissionControlDetails struct {
	Url         string `json:"url,omitempty"`
	AccessToken string `json:"accessToken,omitempty"`
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

func (artifactoryDetails *ArtifactoryDetails) SetAccessToken(accessToken string) {
	artifactoryDetails.AccessToken = accessToken
}

func (artifactoryDetails *ArtifactoryDetails) SetRefreshToken(refreshToken string) {
	artifactoryDetails.RefreshToken = refreshToken
}

func (artifactoryDetails *ArtifactoryDetails) SetClientCertPath(certificatePath string) {
	artifactoryDetails.ClientCertPath = certificatePath
}

func (artifactoryDetails *ArtifactoryDetails) SetClientCertKeyPath(certificatePath string) {
	artifactoryDetails.ClientCertKeyPath = certificatePath
}

func (artifactoryDetails *ArtifactoryDetails) GetApiKey() string {
	return artifactoryDetails.ApiKey
}

func (artifactoryDetails *ArtifactoryDetails) GetUrl() string {
	return artifactoryDetails.Url
}

func (artifactoryDetails *ArtifactoryDetails) GetDistributionUrl() string {
	return artifactoryDetails.DistributionUrl
}

func (artifactoryDetails *ArtifactoryDetails) GetUser() string {
	return artifactoryDetails.User
}

func (artifactoryDetails *ArtifactoryDetails) GetPassword() string {
	return artifactoryDetails.Password
}

func (artifactoryDetails *ArtifactoryDetails) GetAccessToken() string {
	return artifactoryDetails.AccessToken
}

func (artifactoryDetails *ArtifactoryDetails) GetRefreshToken() string {
	return artifactoryDetails.RefreshToken
}

func (artifactoryDetails *ArtifactoryDetails) GetClientCertPath() string {
	return artifactoryDetails.ClientCertPath
}

func (artifactoryDetails *ArtifactoryDetails) GetClientCertKeyPath() string {
	return artifactoryDetails.ClientCertKeyPath
}

func (artifactoryDetails *ArtifactoryDetails) CreateArtAuthConfig() (auth.ServiceDetails, error) {
	artAuth := artifactoryAuth.NewArtifactoryDetails()
	artAuth.SetUrl(artifactoryDetails.Url)
	return artifactoryDetails.createArtAuthConfig(artAuth)
}

func (artifactoryDetails *ArtifactoryDetails) CreateDistAuthConfig() (auth.ServiceDetails, error) {
	artAuth := distributionAuth.NewDistributionDetails()
	artAuth.SetUrl(artifactoryDetails.DistributionUrl)
	return artifactoryDetails.createArtAuthConfig(artAuth)
}

func (artifactoryDetails *ArtifactoryDetails) createArtAuthConfig(details auth.ServiceDetails) (auth.ServiceDetails, error) {
	details.SetSshUrl(artifactoryDetails.SshUrl)
	details.SetAccessToken(artifactoryDetails.AccessToken)
	// If refresh token is not empty, set a refresh handler and skip other credentials
	if artifactoryDetails.RefreshToken != "" {
		tokenRefreshServerId = artifactoryDetails.ServerId
		details.AppendPreRequestInterceptor(AccessTokenRefreshPreRequestInterceptor)
	} else {
		details.SetApiKey(artifactoryDetails.ApiKey)
		details.SetUser(artifactoryDetails.User)
		details.SetPassword(artifactoryDetails.Password)
	}
	details.SetClientCertPath(artifactoryDetails.ClientCertPath)
	details.SetClientCertKeyPath(artifactoryDetails.ClientCertKeyPath)
	details.SetSshKeyPath(artifactoryDetails.SshKeyPath)
	details.SetSshPassphrase(artifactoryDetails.SshPassphrase)
	return details, nil
}

func (missionControlDetails *MissionControlDetails) GetAccessToken() string {
	return missionControlDetails.AccessToken
}

func (missionControlDetails *MissionControlDetails) SetAccessToken(accessToken string) {
	missionControlDetails.AccessToken = accessToken
}
