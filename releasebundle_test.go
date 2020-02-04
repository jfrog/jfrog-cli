package main

import (
	"testing"

	"github.com/jfrog/jfrog-cli-go/inttestutils"
	"github.com/jfrog/jfrog-cli-go/utils/tests"
	"github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/stretchr/testify/assert"
)

func InitReleaseBundleTests() {
	*tests.RtDistributionUrl = utils.AddTrailingSlashIfNeeded(*tests.RtDistributionUrl)
	initArtifactoryCli()
	InitArtifactoryTests()
	inttestutils.SendGpgKeys(artHttpDetails)
}

func CleanReleaseBundleTests() {
	inttestutils.DeleteGpgKeys(artHttpDetails)
}

func initReleaseBundleTest(t *testing.T) {
	if !*tests.TestReleaseBundle {
		t.Skip("Release bundle is not being tested, skipping...")
	}
}

func TestBundleDownload(t *testing.T) {
	initReleaseBundleTest(t)
	bundleName, bundleVersion := "cli-test-bundle", "10"
	inttestutils.DeleteBundle(t, bundleName, artHttpDetails)

	// Upload files
	specFile, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", "--spec="+specFile)

	// Create release bundle
	triples := []inttestutils.RepoPathName{{Repo: tests.Repo1, Path: "data", Name: "b1.in"}}
	inttestutils.CreateBundle(t, bundleName, bundleVersion, triples, artHttpDetails)
	defer inttestutils.DeleteBundle(t, bundleName, artHttpDetails)

	// Download by bundle version, b2 and b3 should not be downloaded, b1 should
	artifactoryCli.Exec("download "+tests.Repo1+"/data/* "+tests.Out+fileutils.GetFileSeparator()+"download"+fileutils.GetFileSeparator()+"simple_by_build"+fileutils.GetFileSeparator(), "--bundle="+bundleName+"/"+bundleVersion)

	// Validate files are downloaded by bundle version
	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err = tests.ValidateListsIdentical(tests.GetBuildSimpleDownload(), paths)
	assert.NoError(t, err)

	// Cleanup
	cleanArtifactoryTest()
}

func TestBundleDownloadUsingSpec(t *testing.T) {
	initReleaseBundleTest(t)
	bundleName, bundleVersion := "cli-test-bundle", "10"
	inttestutils.DeleteBundle(t, bundleName, artHttpDetails)

	// Upload files
	specFile, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", "--spec="+specFile)

	// Create release bundle
	triples := []inttestutils.RepoPathName{{Repo: tests.Repo1, Path: "data", Name: "b1.in"}}
	inttestutils.CreateBundle(t, bundleName, bundleVersion, triples, artHttpDetails)
	defer inttestutils.DeleteBundle(t, bundleName, artHttpDetails)

	// Download by bundle version, b2 and b3 should not be downloaded, b1 should
	specFile, err = tests.CreateSpec(tests.BundleDownloadSpec)
	assert.NoError(t, err)
	artifactoryCli.Exec("download", "--spec="+specFile)

	// Validate files are downloaded by bundle version
	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err = tests.ValidateListsIdentical(tests.GetBuildSimpleDownload(), paths)
	assert.NoError(t, err)

	// Cleanup
	cleanArtifactoryTest()
}

func TestBundleMove(t *testing.T) {
	initReleaseBundleTest(t)
	bundleName, bundleVersion := "cli-test-bundle", "10"
	inttestutils.DeleteBundle(t, bundleName, artHttpDetails)

	// Upload files
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", "--spec="+specFileB)
	artifactoryCli.Exec("upload", "--spec="+specFileA)

	// Create release bundle
	triples := []inttestutils.RepoPathName{{Repo: tests.Repo1, Path: "data", Name: "a*"}}
	inttestutils.CreateBundle(t, bundleName, bundleVersion, triples, artHttpDetails)
	defer inttestutils.DeleteBundle(t, bundleName, artHttpDetails)

	// Move by bundle name and version
	specFile, err := tests.CreateSpec(tests.CopyByBundleSpec)
	assert.NoError(t, err)
	artifactoryCli.Exec("move", "--spec="+specFile)

	// Validate files are moved by bundle version
	cpMvDlByBuildAssertSpec, err := tests.CreateSpec(tests.CpMvDlByBuildAssertSpec)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetBuildMoveExpected(), cpMvDlByBuildAssertSpec, t)

	// Cleanup
	cleanArtifactoryTest()
}

func TestBundleCopy(t *testing.T) {
	initReleaseBundleTest(t)
	bundleName, bundleVersion := "cli-test-bundle", "10"
	inttestutils.DeleteBundle(t, bundleName, artHttpDetails)

	// Upload files
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", "--spec="+specFileB)
	artifactoryCli.Exec("upload", "--spec="+specFileA)

	// Create release bundle
	triples := []inttestutils.RepoPathName{{Repo: tests.Repo1, Path: "data", Name: "a*"}}
	inttestutils.CreateBundle(t, bundleName, bundleVersion, triples, artHttpDetails)
	defer inttestutils.DeleteBundle(t, bundleName, artHttpDetails)

	// Copy by bundle name and version
	specFile, err := tests.CreateSpec(tests.CopyByBundleSpec)
	assert.NoError(t, err)
	artifactoryCli.Exec("copy", "--spec="+specFile)

	// Validate files are moved by bundle version
	cpMvDlByBuildAssertSpec, err := tests.CreateSpec(tests.CpMvDlByBuildAssertSpec)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetBuildCopyExpected(), cpMvDlByBuildAssertSpec, t)

	// Cleanup
	cleanArtifactoryTest()
}

func TestBundleDelete(t *testing.T) {
	initReleaseBundleTest(t)
	bundleName, bundleVersion := "cli-test-bundle", "10"
	inttestutils.DeleteBundle(t, bundleName, artHttpDetails)

	// Upload files
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	artifactoryCli.Exec("upload", "--spec="+specFileB)
	artifactoryCli.Exec("upload", "--spec="+specFileA)

	// Create release bundle
	triples := []inttestutils.RepoPathName{{Repo: tests.Repo1, Path: "data", Name: "a*"}}
	inttestutils.CreateBundle(t, bundleName, bundleVersion, triples, artHttpDetails)
	defer inttestutils.DeleteBundle(t, bundleName, artHttpDetails)

	// Delete by bundle name and version
	specFile, err := tests.CreateSpec(tests.CopyByBundleSpec)
	assert.NoError(t, err)
	artifactoryCli.Exec("del", "--quiet", "--spec="+specFile)

	// Validate files are moved by bundle version
	cpMvDlByBuildAssertSpec, err := tests.CreateSpec(tests.CpMvDlByBuildAssertSpec)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetBuildDeleteExpected(), cpMvDlByBuildAssertSpec, t)

	// Cleanup
	cleanArtifactoryTest()
}

func TestBundleSetProperties(t *testing.T) {
	initReleaseBundleTest(t)
	bundleName, bundleVersion := "cli-test-bundle", "10"
	inttestutils.DeleteBundle(t, bundleName, artHttpDetails)

	// Upload a file.
	artifactoryCli.Exec("upload", "testsdata/a/a1.in", tests.Repo1+"/a.in")

	// Create release bundle
	triples := []inttestutils.RepoPathName{{Repo: tests.Repo1, Path: "*", Name: "a.in"}}
	inttestutils.CreateBundle(t, bundleName, bundleVersion, triples, artHttpDetails)
	defer inttestutils.DeleteBundle(t, bundleName, artHttpDetails)

	// Set the 'prop=red' property to the file.
	artifactoryCli.Exec("sp", tests.Repo1+"/a.*", "prop=red", "--bundle="+bundleName+"/"+bundleVersion)
	// Now let's change the property value, by searching for the 'prop=red'.
	specFile, err := tests.CreateSpec(tests.SetDeletePropsSpec)
	assert.NoError(t, err)
	artifactoryCli.Exec("sp", "prop=green", "--spec="+specFile, "--bundle="+bundleName+"/"+bundleVersion)

	resultItems := searchItemsInArtifactory(t)
	assert.NotZero(t, len(resultItems), "No artifacts were found.")
	for _, item := range resultItems {
		properties := item.Properties
		assert.Equal(t, len(properties), 2, "Failed setting properties on item:", item.GetItemRelativePath())
		for _, prop := range properties {
			if prop.Key == "sha256" {
				continue
			}
			assert.Equal(t, "prop", prop.Key, "Wrong property key")
			assert.Equal(t, "green", prop.Value, "Wrong property value")
		}
	}
	cleanArtifactoryTest()
}
