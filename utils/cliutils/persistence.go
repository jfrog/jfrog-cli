package cliutils

import (
	"encoding/json"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"os"
	"path/filepath"
	gosync "sync"
)

const persistenceFileName = "persistence.json"

// PersistenceFile holds various indicators that need to be persisted between CLI runs
type PersistenceFile struct {
	LatestCliVersionCheckTime *int64 `json:"latestCliVersionCheckTime,omitempty"`
	LatestAiTermsRevision     *int   `json:"latestAiTermsRevision,omitempty"`
}

var (
	persistenceFilePath string
	fileLock            gosync.Mutex
)

// getPersistenceFilePath ensures that the persistence file path is initialized
func getPersistenceFilePath() error {
	if persistenceFilePath == "" {
		homeDir, err := coreutils.GetJfrogHomeDir()
		if err != nil {
			return errorutils.CheckErrorf("failed to get JFrog home directory: " + err.Error())
		}
		persistenceFilePath = filepath.Join(homeDir, persistenceFileName)
	}
	return nil
}

// setCliLatestVersionCheckTime updates the latest version check time in the persistence file
func setCliLatestVersionCheckTime(timestamp int64) error {
	fileLock.Lock()
	defer fileLock.Unlock()

	info, err := getPersistenceInfo()
	if err != nil {
		return err
	}

	info.LatestCliVersionCheckTime = &timestamp
	return setPersistenceInfo(info)
}

// getLatestCliVersionCheckTime retrieves the latest version check time from the persistence file
func getLatestCliVersionCheckTime() (*int64, error) {
	fileLock.Lock()
	defer fileLock.Unlock()

	info, err := getPersistenceInfo()
	if err != nil {
		return nil, err
	}

	return info.LatestCliVersionCheckTime, nil
}

// SetLatestAiTermsRevision updates the AI terms version in the persistence file
func SetLatestAiTermsRevision(version int) error {
	fileLock.Lock()
	defer fileLock.Unlock()

	info, err := getPersistenceInfo()
	if err != nil {
		return err
	}

	info.LatestAiTermsRevision = &version
	return setPersistenceInfo(info)
}

// GetLatestAiTermsRevision retrieves the AI terms version from the persistence file
func GetLatestAiTermsRevision() (*int, error) {
	fileLock.Lock()
	defer fileLock.Unlock()

	info, err := getPersistenceInfo()
	if err != nil {
		return nil, err
	}

	return info.LatestAiTermsRevision, nil
}

// getPersistenceInfo reads the persistence file, creates it if it doesn't exist, and returns the persisted info
func getPersistenceInfo() (*PersistenceFile, error) {
	if err := getPersistenceFilePath(); err != nil {
		return nil, err
	}
	if exists, err := fileutils.IsFileExists(persistenceFilePath, false); err != nil || !exists {
		if err != nil {
			return nil, err
		}
		// Create an empty persistence file if it doesn't exist
		pFile := &PersistenceFile{}
		if err = setPersistenceInfo(pFile); err != nil {
			return nil, errorutils.CheckErrorf("failed while attempting to initialize persistence file: " + err.Error())
		}
		return pFile, nil
	}

	data, err := os.ReadFile(persistenceFilePath)
	if err != nil {
		return nil, errorutils.CheckErrorf("failed while attempting to read persistence file: " + err.Error())
	}

	var info PersistenceFile
	if err = json.Unmarshal(data, &info); err != nil {
		return nil, errorutils.CheckErrorf("failed while attempting to parse persistence file: " + err.Error())
	}

	return &info, nil
}

// setPersistenceInfo writes the given info to the persistence file
func setPersistenceInfo(info *PersistenceFile) error {
	if err := getPersistenceFilePath(); err != nil {
		return err
	}
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return errorutils.CheckErrorf("failed while attempting to create persistence file: " + err.Error())
	}

	if err = os.WriteFile(persistenceFilePath, data, 0644); err != nil {
		return errorutils.CheckErrorf("failed while attempting to write persistence file: " + err.Error())
	}
	return nil
}
