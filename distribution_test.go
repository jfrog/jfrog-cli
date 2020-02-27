package main

import (
	"testing"

	"github.com/jfrog/jfrog-cli-go/inttestutils"
	"github.com/jfrog/jfrog-cli-go/utils/tests"
	"github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/stretchr/testify/assert"
)

func InitDistributionTests() {
	*tests.RtDistributionUrl = utils.AddTrailingSlashIfNeeded(*tests.RtDistributionUrl)
	InitArtifactoryTests()
	inttestutils.SendGpgKeys(artHttpDetails)
}

func CleanDistributionTests() {
	inttestutils.DeleteGpgKeys(artHttpDetails)
	CleanArtifactoryTests()
}

func initDistributionTest(t *testing.T) {
	if !*tests.TestDistribution {
		t.Skip("Distribution is not being tested, skipping...")
	}
}

func TestBundleDownload(t *testing.T) {
	initDistributionTest(t)
	bundleName, bundleVersion := "cli-test-bundle", "10"
	inttestutils.DeleteBundle(t, bundleName, bundleVersion, artHttpDetails)

	// Upload files
	specFile, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	artifactoryCli.Exec("u", "--spec="+specFile)

	// Create release bundle
	artifactoryCli.Exec("crb", bundleName, bundleVersion, tests.Repo1+"/data/b1.in", "--sign-immediately")
	inttestutils.DistributeBundle(t, bundleName, bundleVersion, artHttpDetails)

	// Download by bundle version, b2 and b3 should not be downloaded, b1 should
	artifactoryCli.Exec("dl "+tests.Repo1+"/data/* "+tests.Out+fileutils.GetFileSeparator()+"download"+fileutils.GetFileSeparator()+"simple_by_build"+fileutils.GetFileSeparator(), "--bundle="+bundleName+"/"+bundleVersion)

	// Validate files are downloaded by bundle version
	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err = tests.ValidateListsIdentical(tests.GetBuildSimpleDownload(), paths)
	assert.NoError(t, err)

	// Cleanup
	inttestutils.DeleteBundle(t, bundleName, bundleVersion, artHttpDetails)
	cleanArtifactoryTest()
}

func TestBundleDownloadUsingSpec(t *testing.T) {
	initDistributionTest(t)
	bundleName, bundleVersion := "cli-test-bundle", "10"
	inttestutils.DeleteBundle(t, bundleName, bundleVersion, artHttpDetails)

	// Upload files
	specFile, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	artifactoryCli.Exec("u", "--spec="+specFile)

	// Create release bundle
	artifactoryCli.Exec("crb", bundleName, bundleVersion, tests.Repo1+"/data/b1.in", "--sign-immediately")
	inttestutils.DistributeBundle(t, bundleName, bundleVersion, artHttpDetails)

	// Download by bundle version, b2 and b3 should not be downloaded, b1 should
	specFile, err = tests.CreateSpec(tests.BundleDownloadSpec)
	assert.NoError(t, err)
	artifactoryCli.Exec("dl", "--spec="+specFile)

	// Validate files are downloaded by bundle version
	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err = tests.ValidateListsIdentical(tests.GetBuildSimpleDownload(), paths)
	assert.NoError(t, err)

	// Cleanup
	inttestutils.DeleteBundle(t, bundleName, bundleVersion, artHttpDetails)
	cleanArtifactoryTest()
}

func TestBundleDownloadNoPattern(t *testing.T) {
	initDistributionTest(t)
	bundleName, bundleVersion := "cli-test-bundle", "10"
	inttestutils.DeleteBundle(t, bundleName, bundleVersion, artHttpDetails)

	// Upload files
	specFile, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	artifactoryCli.Exec("u", "--spec="+specFile)

	// Create release bundle
	artifactoryCli.Exec("crb", bundleName, bundleVersion, tests.Repo1+"/data/b1.in", "--sign-immediately")
	inttestutils.DistributeBundle(t, bundleName, bundleVersion, artHttpDetails)

	// Download by bundle name and version with pattern "*", b2 and b3 should not be downloaded, b1 should
	artifactoryCli.Exec("dl", "*", "out/download/simple_by_build/data/", "--bundle="+bundleName+"/"+bundleVersion, "--flat")

	// Validate files are downloaded by bundle version
	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err = tests.ValidateListsIdentical(tests.GetBuildSimpleDownload(), paths)
	assert.NoError(t, err)

	// Download by bundle name and version version without pattern, b2 and b3 should not be downloaded, b1 should
	tests.CleanFileSystem()
	specFile, err = tests.CreateSpec(tests.BundleDownloadSpecNoPattern)
	artifactoryCli.Exec("dl", "--spec="+specFile, "--flat")

	// Validate files are downloaded by bundle version
	paths, _ = fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err = tests.ValidateListsIdentical(tests.GetBuildSimpleDownload(), paths)
	assert.NoError(t, err)

	// Cleanup
	inttestutils.DeleteBundle(t, bundleName, bundleVersion, artHttpDetails)
	cleanArtifactoryTest()
}

func TestBundleExclusions(t *testing.T) {
	initDistributionTest(t)
	bundleName, bundleVersion := "cli-test-bundle", "10"
	inttestutils.DeleteBundle(t, bundleName, bundleVersion, artHttpDetails)

	// Upload files
	specFile, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	artifactoryCli.Exec("u", "--spec="+specFile)

	// Create release bundle. Include b1.in and b2.in. Exclude b3.in.
	artifactoryCli.Exec("crb", bundleName, bundleVersion, tests.Repo1+"/data/b*.in", "--sign-immediately", "--exclusions=*b3.in")
	inttestutils.DistributeBundle(t, bundleName, bundleVersion, artHttpDetails)

	// Download by bundle version, b2 and b3 should not be downloaded, b1 should
	artifactoryCli.Exec("dl "+tests.Repo1+"/data/* "+tests.Out+fileutils.GetFileSeparator()+"download"+fileutils.GetFileSeparator()+"simple_by_build"+fileutils.GetFileSeparator(), "--bundle="+bundleName+"/"+bundleVersion, "--exclusions=*b2.in")

	// Validate files are downloaded by bundle version
	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err = tests.ValidateListsIdentical(tests.GetBuildSimpleDownload(), paths)
	assert.NoError(t, err)

	// Cleanup
	inttestutils.DeleteBundle(t, bundleName, bundleVersion, artHttpDetails)
	cleanArtifactoryTest()
}

func TestBundleCopy(t *testing.T) {
	initDistributionTest(t)
	bundleName, bundleVersion := "cli-test-bundle", "10"
	inttestutils.DeleteBundle(t, bundleName, bundleVersion, artHttpDetails)

	// Upload files
	specFileA, err := tests.CreateSpec(tests.SplitUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.SplitUploadSpecB)
	assert.NoError(t, err)
	artifactoryCli.Exec("u", "--spec="+specFileB)
	artifactoryCli.Exec("u", "--spec="+specFileA)

	// Create release bundle
	artifactoryCli.Exec("crb", bundleName, bundleVersion, tests.Repo1+"/data/a*", "--sign-immediately")
	inttestutils.DistributeBundle(t, bundleName, bundleVersion, artHttpDetails)

	// Copy by bundle name and version
	specFile, err := tests.CreateSpec(tests.CopyByBundleSpec)
	assert.NoError(t, err)
	artifactoryCli.Exec("cp", "--spec="+specFile)

	// Validate files are moved by bundle version
	cpMvDlByBuildAssertSpec, err := tests.CreateSpec(tests.CpMvDlByBuildAssertSpec)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetBuildCopyExpected(), cpMvDlByBuildAssertSpec, t)

	// Cleanup
	inttestutils.DeleteBundle(t, bundleName, bundleVersion, artHttpDetails)
	cleanArtifactoryTest()
}

func TestBundleSetProperties(t *testing.T) {
	initDistributionTest(t)
	bundleName, bundleVersion := "cli-test-bundle", "10"
	inttestutils.DeleteBundle(t, bundleName, bundleVersion, artHttpDetails)

	// Upload a file.
	artifactoryCli.Exec("u", "testsdata/a/a1.in", tests.Repo1+"/a.in")

	// Create release bundle
	artifactoryCli.Exec("crb", bundleName, bundleVersion, tests.Repo1+"/a.in", "--sign-immediately")
	inttestutils.DistributeBundle(t, bundleName, bundleVersion, artHttpDetails)

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
	inttestutils.DeleteBundle(t, bundleName, bundleVersion, artHttpDetails)
	cleanArtifactoryTest()
}
