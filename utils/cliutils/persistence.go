package cliutils

import (
	"encoding/json"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"os"
	"path/filepath"
	gosync "sync"
)

const persistenceFileName = "persistence.json"

// PersistenceInfo represents the fields we are persisting
type PersistenceInfo struct {
	LatestVersionCheckTime *int64 `json:"latestVersionCheckTime,omitempty"`
	AiTermsVersion         *int   `json:"aiTermsVersion,omitempty"`
}

var (
	persistenceFilePath string
	fileLock            gosync.Mutex
)

// init initializes the persistence file path once, and stores it for future use
func init() {
	homeDir, err := coreutils.GetJfrogHomeDir()
	if err != nil {
		panic("Failed to get JFrog home directory: " + err.Error())
	}
	persistenceFilePath = filepath.Join(homeDir, persistenceFileName)
}

// SetLatestVersionCheckTime updates the latest version check time in the persistence file
func SetLatestVersionCheckTime(timestamp int64) error {
	info, err := getPersistenceInfo()
	if err != nil {
		return err
	}

	info.LatestVersionCheckTime = &timestamp
	return setPersistenceInfo(info)
}

// GetLatestVersionCheckTime retrieves the latest version check time from the persistence file
func GetLatestVersionCheckTime() (*int64, error) {
	info, err := getPersistenceInfo()
	if err != nil {
		return nil, err
	}

	return info.LatestVersionCheckTime, nil
}

// SetAiTermsVersion updates the AI terms version in the persistence file
func SetAiTermsVersion(version int) error {
	info, err := getPersistenceInfo()
	if err != nil {
		return err
	}

	info.AiTermsVersion = &version
	return setPersistenceInfo(info)
}

// GetAiTermsVersion retrieves the AI terms version from the persistence file
func GetAiTermsVersion() (*int, error) {
	info, err := getPersistenceInfo()
	if err != nil {
		return nil, err
	}

	return info.AiTermsVersion, nil
}

// getPersistenceInfo reads the persistence file, creates it if it doesn't exist, and returns the persisted info
func getPersistenceInfo() (*PersistenceInfo, error) {
	if _, err := os.Stat(persistenceFilePath); os.IsNotExist(err) {
		// Create an empty persistence file if it doesn't exist
		pFile := &PersistenceInfo{}
		if err = setPersistenceInfo(pFile); err != nil {
			return nil, errorutils.CheckErrorf("failed while attempting to initialize persistence file" + err.Error())
		}
		return pFile, nil
	}

	fileLock.Lock()
	defer fileLock.Unlock()
	data, err := os.ReadFile(persistenceFilePath)
	if err != nil {
		return nil, errorutils.CheckErrorf("failed while attempting to read persistence file" + err.Error())
	}

	var info PersistenceInfo
	if err = json.Unmarshal(data, &info); err != nil {
		return nil, errorutils.CheckErrorf("failed while attempting to parse persistence file" + err.Error())
	}

	return &info, nil
}

// setPersistenceInfo writes the given info to the persistence file
func setPersistenceInfo(info *PersistenceInfo) error {
	fileLock.Lock()
	defer fileLock.Unlock()

	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return errorutils.CheckErrorf("failed while attempting to create persistence file" + err.Error())
	}

	err = os.WriteFile(persistenceFilePath, data, 0644)
	if err != nil {
		return errorutils.CheckErrorf("failed while attempting to write persistence file" + err.Error())
	}
	return nil
}
