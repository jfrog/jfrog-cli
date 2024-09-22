package cliutils

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestSetAndGetLatestVersionCheckTime tests setting and getting the LatestCliVersionCheckTime
func TestSetAndGetLatestVersionCheckTime(t *testing.T) {
	// Setup temporary directory
	persistenceFilePath = filepath.Join(t.TempDir(), persistenceFileName)

	// Set the timestamp
	timestamp := time.Now().UnixMilli()
	err := SetCliLatestVersionCheckTime(timestamp)
	if err != nil {
		t.Fatalf("Failed to set LatestCliVersionCheckTime: %v", err)
	}

	// Get the timestamp
	storedTimestamp, err := GetLatestCliVersionCheckTime()
	if err != nil {
		t.Fatalf("Failed to get LatestCliVersionCheckTime: %v", err)
	}

	// Assert equality
	if *storedTimestamp != timestamp {
		t.Fatalf("Expected %d, got %d", timestamp, *storedTimestamp)
	}
}

// TestSetAndGetAiTermsVersion tests setting and getting the LatestAiTermsRevision
func TestSetAndGetAiTermsVersion(t *testing.T) {
	// Setup temporary directory
	persistenceFilePath = filepath.Join(t.TempDir(), persistenceFileName)

	// Set the AI terms version
	version := 42
	err := SetLatestAiTermsRevision(version)
	if err != nil {
		t.Fatalf("Failed to set LatestAiTermsRevision: %v", err)
	}

	// Get the AI terms version
	storedVersion, err := GetLatestAiTermsRevision()
	if err != nil {
		t.Fatalf("Failed to get LatestAiTermsRevision: %v", err)
	}

	// Assert equality
	if *storedVersion != version {
		t.Fatalf("Expected %d, got %d", version, *storedVersion)
	}
}

// TestPersistenceFileCreation tests if the persistence file is created when it doesn't exist
func TestPersistenceFileCreation(t *testing.T) {
	// Setup temporary directory
	persistenceFilePath = filepath.Join(t.TempDir(), persistenceFileName)

	// Ensure the persistence file doesn't exist
	_, err := os.Stat(persistenceFilePath)
	if !os.IsNotExist(err) {
		t.Fatalf("Expected file to not exist, but it does: %v", err)
	}

	// Trigger file creation by setting version check time
	timestamp := time.Now().UnixMilli()
	err = SetCliLatestVersionCheckTime(timestamp)
	if err != nil {
		t.Fatalf("Failed to set LatestCliVersionCheckTime: %v", err)
	}

	// Verify the persistence file was created
	_, err = os.Stat(persistenceFilePath)
	if os.IsNotExist(err) {
		t.Fatalf("Expected file to exist, but it does not: %v", err)
	}
}
