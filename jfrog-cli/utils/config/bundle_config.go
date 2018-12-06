package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/pkg/errors"
	"io/ioutil"
)

// Returns the configured bundles or error if the bundle id not found
func GetBundleConf(bundleId string) (*BundleDetails, error) {
	configs, err := GetAllBundleConfigs()
	if err != nil {
		return nil, err
	}
	return getBundleConfByBundleId(bundleId, configs)
}

// Returns the configured server or error if the server id not found
func getBundleConfByBundleId(serverId string, configs []*BundleDetails) (*BundleDetails, error) {
	for _, conf := range configs {
		if conf.ServerId == serverId {
			return conf, nil
		}
	}
	return nil, errorutils.CheckError(errors.New(fmt.Sprintf("Bundle config '%s' does not exist.", serverId)))
}

func GetAndRemoveBundleConfiguration(bundleId string, configs []*BundleDetails) (*BundleDetails, []*BundleDetails) {
	for i, conf := range configs {
		if conf.ServerId == bundleId {
			configs = append(configs[:i], configs[i+1:]...)
			return conf, configs
		}
	}
	return nil, configs
}

func SaveBundleConf(details []*BundleDetails) error {
	conf, err := readBundleConf()
	if err != nil {
		return err
	}
	conf.Bundle = details
	return saveBundleConfig(conf)
}

func GetAllBundleConfigs() ([]*BundleDetails, error) {
	conf, err := readBundleConf()
	if err != nil {
		return nil, err
	}
	details := conf.Bundle
	if details == nil {
		return make([]*BundleDetails, 0), nil
	}
	return details, nil
}

func GetBundleSpecificConfig(bundleConfigId string) (*BundleDetails, error) {
	conf, err := readBundleConf()
	if err != nil {
		return nil, err
	}
	details := conf.Bundle
	if details == nil || len(details) == 0 {
		return new(BundleDetails), nil
	}
	if len(bundleConfigId) == 0 {
		return GetDefaultBundleConf(details)
	}
	return getBundleConfByBundleId(bundleConfigId, details)
}

func GetDefaultBundleConf(configs []*BundleDetails) (*BundleDetails, error) {
	if len(configs) == 0 {
		details := new(BundleDetails)
		details.IsDefault = true
		return details, nil
	}
	for _, conf := range configs {
		if conf.IsDefault == true {
			return conf, nil
		}
	}
	return nil, errorutils.CheckError(errors.New("Couldn't find default server."))
}

func saveBundleConfig(configToSave *BundleConfigV1) error {
	configToSave.Version = cliutils.GetBundleConfigVersion()
	b, err := json.Marshal(&configToSave)
	if err != nil {
		return errorutils.CheckError(err)
	}
	var content bytes.Buffer
	err = json.Indent(&content, b, "", "  ")
	if err != nil {
		return errorutils.CheckError(err)
	}
	path, err := getConfFilePath(JfrogBundleConfigFile)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, []byte(content.String()), 0600)
	if err != nil {
		return errorutils.CheckError(err)
	}

	return nil
}

func readBundleConf() (*BundleConfigV1, error) {
	confFilePath, err := getConfFilePath(JfrogBundleConfigFile)
	if err != nil {
		return nil, err
	}
	res := new(BundleConfigV1)
	exists, err := fileutils.IsFileExists(confFilePath, false)
	if err != nil {
		return nil, err
	}
	if !exists {
		return res, nil
	}
	content, err := fileutils.ReadFile(confFilePath)
	if err != nil {
		return nil, err
	}
	if len(content) == 0 {
		return new(BundleConfigV1), nil
	}
	err = json.Unmarshal(content, &res)
	return res, err
}

type BundleConfigV1 struct {
	Version string           `json:"Version,omitempty"`
	Bundle  []*BundleDetails `json:"Bundle,omitempty"`
}

type BundleDetails struct {
	BundleConfigId string `json:"bundleConfigId,omitempty"`
	ServerId       string `json:"serverId,omitempty"`
	Name           string `json:"name,omitempty"`
	Version        string `json:"version,omitempty"`
	ScriptPath     string `json:"scriptPath,omitempty"`
	IsDefault      bool   `json:"isDefault,omitempty"`
}

func (artifactoryDetails *BundleDetails) IsEmpty() bool {
	return len(artifactoryDetails.Name) == 0
}

func (artifactoryDetails *BundleDetails) SetServerId(serverId string) {
	artifactoryDetails.ServerId = serverId
}

func (artifactoryDetails *BundleDetails) SetName(name string) {
	artifactoryDetails.Name = name
}

func (artifactoryDetails *BundleDetails) SetVersion(version string) {
	artifactoryDetails.Version = version
}

func (artifactoryDetails *BundleDetails) SetScriptPath(scriptPath string) {
	artifactoryDetails.ScriptPath = scriptPath
}
