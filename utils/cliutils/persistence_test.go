package cliutils

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestSetAndGetLatestVersionCheckTime tests setting and getting the LatestCliVersionCheckTime
func TestSetAndGetLatestVersionCheckTime(t *testing.T) {
	// Setup temporary directory
	persistenceFilePath = filepath.Join(t.TempDir(), persistenceFileName)

	// Set the timestamp
	timestamp := time.Now().UnixMilli()
	err := SetCliLatestVersionCheckTime(timestamp)
	assert.NoError(t, err, "Failed to set LatestCliVersionCheckTime")

	// Get the timestamp
	storedTimestamp, err := GetLatestCliVersionCheckTime()
	assert.NoError(t, err, "Failed to get LatestCliVersionCheckTime")

	// Assert equality
	assert.Equal(t, timestamp, *storedTimestamp, "Stored timestamp does not match the set timestamp")
}

// TestSetAndGetAiTermsVersion tests setting and getting the LatestAiTermsRevision
func TestSetAndGetAiTermsVersion(t *testing.T) {
	// Setup temporary directory
	persistenceFilePath = filepath.Join(t.TempDir(), persistenceFileName)

	// Set the AI terms version
	version := 42
	err := SetLatestAiTermsRevision(version)
	assert.NoError(t, err, "Failed to set LatestAiTermsRevision")

	// Get the AI terms version
	storedVersion, err := GetLatestAiTermsRevision()
	assert.NoError(t, err, "Failed to get LatestAiTermsRevision")

	// Assert equality
	assert.Equal(t, version, *storedVersion, "Stored AI terms version does not match the set version")
}

// TestPersistenceFileCreation tests if the persistence file is created when it doesn't exist
func TestPersistenceFileCreation(t *testing.T) {
	// Setup temporary directory
	persistenceFilePath = filepath.Join(t.TempDir(), persistenceFileName)

	// Ensure the persistence file doesn't exist
	_, err := os.Stat(persistenceFilePath)
	assert.ErrorIs(t, err, os.ErrNotExist, "Expected error to be os.ErrNotExist")

	// Trigger file creation by setting version check time
	timestamp := time.Now().UnixMilli()
	err = SetCliLatestVersionCheckTime(timestamp)
	assert.NoError(t, err, "Failed to set LatestCliVersionCheckTime")

	// Verify the persistence file was created
	_, err = os.Stat(persistenceFilePath)
	assert.False(t, os.IsNotExist(err), "Expected file to exist, but it does not")
}
