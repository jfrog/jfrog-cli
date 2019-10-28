package main

import (
	"os"
	"path"
	"testing"

	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/utils/tests"
)

func TestGetJcenterRemoteDetails(t *testing.T) {
	initBuildToolsTest(t)
	createServerConfigAndReturnPassphrase()

	unsetEnvVars := func() {
		err := os.Unsetenv(utils.JCenterRemoteServerEnv)
		if err != nil {
			t.Error(err)
		}
		err = os.Unsetenv(utils.JCenterRemoteRepoEnv)
		if err != nil {
			t.Error(err)
		}
	}
	unsetEnvVars()
	defer unsetEnvVars()

	// The utils.JCenterRemoteServerEnv env var is not set, so extractor1.jar should be downloaded from jcenter.
	downloadPath := "org/jfrog/buildinfo/build-info-extractor/extractor1.jar"
	expectedRemotePath := path.Join("bintray/jcenter", downloadPath)
	validateJcenterRemoteDetails(t, downloadPath, expectedRemotePath)

	// Still, the utils.JCenterRemoteServerEnv env var is not set, so the download should be from jcenter.
	// Expecting a different download path this time.
	downloadPath = "org/jfrog/buildinfo/build-info-extractor/extractor2.jar"
	expectedRemotePath = path.Join("bintray/jcenter", downloadPath)
	validateJcenterRemoteDetails(t, downloadPath, expectedRemotePath)

	// Setting the utils.JCenterRemoteServerEnv env var now,
	// Expecting therefore the download to be from the the server ID configured by this env var.
	err := os.Setenv(utils.JCenterRemoteServerEnv, tests.RtServerId)
	if err != nil {
		t.Error(err)
	}
	downloadPath = "org/jfrog/buildinfo/build-info-extractor/extractor3.jar"
	expectedRemotePath = path.Join("jcenter", downloadPath)
	validateJcenterRemoteDetails(t, downloadPath, expectedRemotePath)

	// Still expecting the download to be from the same server ID, but this time, not through a remote repo named
	// jcenter, but through test-remote-repo.
	testRemoteRepo := "test-remote-repo"
	err = os.Setenv(utils.JCenterRemoteRepoEnv, testRemoteRepo)
	if err != nil {
		t.Error(err)
	}
	downloadPath = "1org/jfrog/buildinfo/build-info-extractor/extractor4.jar"
	expectedRemotePath = path.Join(testRemoteRepo, downloadPath)
	validateJcenterRemoteDetails(t, downloadPath, expectedRemotePath)

	cleanBuildToolsTest()
}

func validateJcenterRemoteDetails(t *testing.T, downloadPath, expectedRemotePath string) {
	artDetails, remotePath, err := utils.GetJcenterRemoteDetails(downloadPath)
	if err != nil {
		t.Error(err)
	}
	if remotePath != expectedRemotePath {
		t.Error("Expected remote path to be", expectedRemotePath, "but got", remotePath)
	}
	if os.Getenv(utils.JCenterRemoteServerEnv) != "" && artDetails == nil {
		t.Error("Expected a server to be returned")
	}
}
