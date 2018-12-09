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

// Returns the configured bundle config or error if not found
func getBundleConfByBundleId(bundleConfigId string, configs []*BundleDetails) (*BundleDetails, error) {
	for _, conf := range configs {
		if conf.BundleConfigId == bundleConfigId {
			return conf, nil
		}
	}
	return nil, errorutils.CheckError(errors.New(fmt.Sprintf("Bundle config '%s' does not exist.", bundleConfigId)))
}

func GetAndRemoveBundleConfiguration(bundleId string, configs []*BundleDetails) (*BundleDetails, []*BundleDetails) {
	for i, conf := range configs {
		if conf.BundleConfigId == bundleId {
			configs = append(configs[:i], configs[i+1:]...)
			return conf, configs
		}
	}
	return nil, configs
}

func SaveBundleConfigs(details []*BundleDetails) error {
	conf, err := readBundleConfig()
	if err != nil {
		return errorutils.CheckError(err)
	}
	conf.Bundle = details
	return saveBundleConfig(conf)
}

func EditBundleConfig(bundleDetails *BundleDetails) error {
	bundleConfigs, err := GetAllBundleConfigs()
	if err != nil {
		return err
	}
	for i, config := range bundleConfigs {
		if config.BundleConfigId == bundleDetails.BundleConfigId {
			bundleConfigs[i] = bundleDetails
			return SaveBundleConfigs(bundleConfigs)
		}
	}
	return errors.New(fmt.Sprintf("Bundle config id %s not found", bundleDetails.BundleConfigId))
}

func GetAllBundleConfigs() ([]*BundleDetails, error) {
	conf, err := readBundleConfig()
	if err != nil {
		return nil, err
	}
	details := conf.Bundle
	if details == nil {
		return make([]*BundleDetails, 0), nil
	}
	return details, nil
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
	return nil, errorutils.CheckError(errors.New("Couldn't find default bundle config."))
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

func readBundleConfig() (*BundleConfigV1, error) {
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
	BundleConfigId     string `json:"bundleConfigId,omitempty"`
	ServerId           string `json:"serverId,omitempty"`
	Name               string `json:"name,omitempty"`
	VersionConstraints string `json:"versionConstraints,omitempty"`
	CurrentVersion     string `json:"currentVersion,omitempty"`
	ScriptPath         string `json:"scriptPath,omitempty"`
	IsDefault          bool   `json:"isDefault,omitempty"`
}

func (bundleDetails *BundleDetails) IsEmpty() bool {
	return len(bundleDetails.Name) == 0
}

func (bundleDetails *BundleDetails) SetServerId(serverId string) {
	bundleDetails.ServerId = serverId
}

func (bundleDetails *BundleDetails) SetName(name string) {
	bundleDetails.Name = name
}

func (bundleDetails *BundleDetails) SetVersionConstraints(versionConstraints string) {
	bundleDetails.VersionConstraints = versionConstraints
}

func (bundleDetails *BundleDetails) SetCurrentVersion(currentVersion string) {
	bundleDetails.CurrentVersion = currentVersion
}

func (bundleDetails *BundleDetails) SetScriptPath(scriptPath string) {
	bundleDetails.ScriptPath = scriptPath
}
