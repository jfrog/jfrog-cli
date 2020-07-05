package main

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/jfrog/jfrog-cli/inttestutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/stretchr/testify/assert"
)

const bundleVersion = "10"

func InitDistributionTests() {
	*tests.RtDistributionUrl = utils.AddTrailingSlashIfNeeded(*tests.RtDistributionUrl)
	initArtifactoryCli()
	inttestutils.CleanUpOldBundles(artHttpDetails, bundleVersion, artifactoryCli)
	InitArtifactoryTests()
	inttestutils.SendGpgKeys(artHttpDetails)
}

func CleanDistributionTests() {
	deleteCreatedRepos()
}

func initDistributionTest(t *testing.T) {
	if !*tests.TestDistribution {
		t.Skip("Skipping distribution test. To run distribution test add the '-test.distribution=true' option.")
	}
}

func cleanDistributionTest(t *testing.T) {
	artifactoryCli.Exec("rbdel", tests.BundleName, bundleVersion, "--site=*", "--delete-from-dist", "--quiet")
	inttestutils.WaitForDeletion(t, tests.BundleName, bundleVersion, artHttpDetails)
	inttestutils.CleanDistributionRepositories(t, artifactoryDetails)
	tests.CleanFileSystem()
}

func TestBundleAsyncDistDownload(t *testing.T) {
	initDistributionTest(t)

	// Upload files
	specFile, err := tests.CreateSpec(tests.DistributionUploadSpecB)
	assert.NoError(t, err)
	artifactoryCli.Exec("u", "--spec="+specFile)

	// Create and distribute release bundle
	artifactoryCli.Exec("rbc", tests.BundleName, bundleVersion, tests.DistRepo1+"/data/b1.in", "--sign")
	artifactoryCli.Exec("rbd", tests.BundleName, bundleVersion, "--site=*")
	inttestutils.WaitForDistribution(t, tests.BundleName, bundleVersion, artHttpDetails)

	// Download by bundle version, b2 and b3 should not be downloaded, b1 should
	artifactoryCli.Exec("dl "+tests.DistRepo1+"/data/* "+tests.Out+fileutils.GetFileSeparator()+"download"+fileutils.GetFileSeparator()+"simple_by_build"+fileutils.GetFileSeparator(), "--bundle="+tests.BundleName+"/"+bundleVersion)

	// Validate files are downloaded by bundle version
	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err = tests.ValidateListsIdentical(tests.GetBuildSimpleDownload(), paths)
	assert.NoError(t, err)

	// Cleanup
	cleanDistributionTest(t)
}

func TestBundleDownloadUsingSpec(t *testing.T) {
	initDistributionTest(t)

	// Upload files
	specFile, err := tests.CreateSpec(tests.DistributionUploadSpecB)
	assert.NoError(t, err)
	artifactoryCli.Exec("u", "--spec="+specFile)
	inttestutils.WaitForDeletion(t, tests.BundleName, bundleVersion, artHttpDetails)

	// Create release bundle
	distributionRules, err := tests.CreateSpec(tests.DistributionRules)
	assert.NoError(t, err)
	artifactoryCli.Exec("rbc", tests.BundleName, bundleVersion, tests.DistRepo1+"/data/b1.in", "--sign")
	artifactoryCli.Exec("rbd", tests.BundleName, bundleVersion, "--dist-rules="+distributionRules, "--sync")

	// Download by bundle version, b2 and b3 should not be downloaded, b1 should
	specFile, err = tests.CreateSpec(tests.BundleDownloadSpec)
	assert.NoError(t, err)
	artifactoryCli.Exec("dl", "--spec="+specFile)

	// Validate files are downloaded by bundle version
	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err = tests.ValidateListsIdentical(tests.GetBuildSimpleDownload(), paths)
	assert.NoError(t, err)

	// Cleanup
	cleanDistributionTest(t)
}

func TestBundleDownloadNoPattern(t *testing.T) {
	initDistributionTest(t)

	// Upload files
	specFile, err := tests.CreateSpec(tests.DistributionUploadSpecB)
	assert.NoError(t, err)
	artifactoryCli.Exec("u", "--spec="+specFile)

	// Create release bundle
	artifactoryCli.Exec("rbc", tests.BundleName, bundleVersion, tests.DistRepo1+"/data/b1.in", "--sign")
	artifactoryCli.Exec("rbd", tests.BundleName, bundleVersion, "--site=*", "--sync")

	// Download by bundle name and version with pattern "*", b2 and b3 should not be downloaded, b1 should
	artifactoryCli.Exec("dl", "*", "out/download/simple_by_build/data/", "--bundle="+tests.BundleName+"/"+bundleVersion, "--flat")

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
	cleanDistributionTest(t)
}

func TestBundleExclusions(t *testing.T) {
	initDistributionTest(t)

	// Upload files
	specFile, err := tests.CreateSpec(tests.DistributionUploadSpecB)
	assert.NoError(t, err)
	artifactoryCli.Exec("u", "--spec="+specFile)

	// Create release bundle. Include b1.in and b2.in. Exclude b3.in.
	artifactoryCli.Exec("rbc", tests.BundleName, bundleVersion, tests.DistRepo1+"/data/b*.in", "--sign", "--exclusions=*b3.in")
	artifactoryCli.Exec("rbd", tests.BundleName, bundleVersion, "--site=*", "--sync")

	// Download by bundle version, b2 and b3 should not be downloaded, b1 should
	artifactoryCli.Exec("dl "+tests.DistRepo1+"/data/* "+tests.Out+fileutils.GetFileSeparator()+"download"+fileutils.GetFileSeparator()+"simple_by_build"+fileutils.GetFileSeparator(), "--bundle="+tests.BundleName+"/"+bundleVersion, "--exclusions=*b2.in")

	// Validate files are downloaded by bundle version
	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err = tests.ValidateListsIdentical(tests.GetBuildSimpleDownload(), paths)
	assert.NoError(t, err)

	// Cleanup
	cleanDistributionTest(t)
}

func TestBundleCopy(t *testing.T) {
	initDistributionTest(t)

	// Upload files
	specFileA, err := tests.CreateSpec(tests.DistributionUploadSpecA)
	assert.NoError(t, err)
	specFileB, err := tests.CreateSpec(tests.DistributionUploadSpecB)
	assert.NoError(t, err)
	artifactoryCli.Exec("u", "--spec="+specFileA)
	artifactoryCli.Exec("u", "--spec="+specFileB)

	// Create release bundle
	artifactoryCli.Exec("rbc", tests.BundleName, bundleVersion, tests.DistRepo1+"/data/a*", "--sign")
	artifactoryCli.Exec("rbd", tests.BundleName, bundleVersion, "--site=*", "--sync")

	// Copy by bundle name and version
	specFile, err := tests.CreateSpec(tests.CopyByBundleSpec)
	assert.NoError(t, err)
	artifactoryCli.Exec("cp", "--spec="+specFile)

	// Validate files are copied by bundle version
	spec, err := tests.CreateSpec(tests.CopyByBundleAssertSpec)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetBundleCopyExpected(), spec, t)

	// Cleanup
	cleanDistributionTest(t)
}

func TestBundleSetProperties(t *testing.T) {
	initDistributionTest(t)

	// Upload a file.
	artifactoryCli.Exec("u", "testsdata/a/a1.in", tests.DistRepo1+"/a.in")

	// Create release bundle
	artifactoryCli.Exec("rbc", tests.BundleName, bundleVersion, tests.DistRepo1+"/a.in", "--sign")
	artifactoryCli.Exec("rbd", tests.BundleName, bundleVersion, "--site=*", "--sync")

	// Set the 'prop=red' property to the file.
	artifactoryCli.Exec("sp", tests.DistRepo1+"/a.*", "prop=red", "--bundle="+tests.BundleName+"/"+bundleVersion)
	// Now let's change the property value, by searching for the 'prop=red'.
	specFile, err := tests.CreateSpec(tests.DistributionSetDeletePropsSpec)
	assert.NoError(t, err)
	artifactoryCli.Exec("sp", "prop=green", "--spec="+specFile, "--bundle="+tests.BundleName+"/"+bundleVersion)

	resultItems := searchItemsInArtifactory(t, tests.SearchDistRepoByInSuffix)
	assert.NotZero(t, len(resultItems), "No artifacts were found.")
	for _, item := range resultItems {
		properties := item.Properties
		assert.Equal(t, 2, len(properties), "Failed setting properties on item:", item.GetItemRelativePath())
		for _, prop := range properties {
			if prop.Key == "sha256" {
				continue
			}
			assert.Equal(t, "prop", prop.Key, "Wrong property key")
			assert.Equal(t, "green", prop.Value, "Wrong property value")
		}
	}
	cleanDistributionTest(t)
}

func TestSignReleaseBundle(t *testing.T) {
	initDistributionTest(t)

	// Upload files
	specFile, err := tests.CreateSpec(tests.DistributionUploadSpecB)
	assert.NoError(t, err)
	artifactoryCli.Exec("u", "--spec="+specFile)

	// Create a release bundle without --sign and make sure it is not signed
	artifactoryCli.Exec("rbc", tests.BundleName, bundleVersion, tests.DistRepo1+"/data/b1.in")
	distributableResponse := inttestutils.GetLocalBundle(t, tests.BundleName, bundleVersion, artHttpDetails)
	assert.NotNil(t, distributableResponse)
	assert.Equal(t, inttestutils.Open, distributableResponse.State)

	// Sign the release bundle and make sure it is signed
	artifactoryCli.Exec("rbs", tests.BundleName, bundleVersion)
	distributableResponse = inttestutils.GetLocalBundle(t, tests.BundleName, bundleVersion, artHttpDetails)
	assert.NotNil(t, distributableResponse)
	assert.Equal(t, inttestutils.Signed, distributableResponse.State)

	// Cleanup
	cleanDistributionTest(t)
}

func TestBundleDeleteLocal(t *testing.T) {
	initDistributionTest(t)

	// Upload files
	specFile, err := tests.CreateSpec(tests.DistributionUploadSpecB)
	assert.NoError(t, err)
	artifactoryCli.Exec("u", "--spec="+specFile)

	// Create a release bundle
	artifactoryCli.Exec("rbc", tests.BundleName, bundleVersion, tests.DistRepo1+"/data/b1.in", "--sign")
	inttestutils.VerifyLocalBundleExistence(t, tests.BundleName, bundleVersion, true, artHttpDetails)

	// Delete release bundle locally
	artifactoryCli.Exec("rbdel", tests.BundleName, bundleVersion, "--site=*", "--delete-from-dist", "--quiet")
	inttestutils.VerifyLocalBundleExistence(t, tests.BundleName, bundleVersion, false, artHttpDetails)

	// Cleanup
	cleanDistributionTest(t)
}

func TestUpdateReleaseBundle(t *testing.T) {
	initDistributionTest(t)

	// Upload files
	specFile, err := tests.CreateSpec(tests.DistributionUploadSpecB)
	assert.NoError(t, err)
	artifactoryCli.Exec("u", "--spec="+specFile)

	// Create a release bundle with b2.in
	artifactoryCli.Exec("rbc", tests.BundleName, bundleVersion, tests.DistRepo1+"/data/b2.in")
	inttestutils.VerifyLocalBundleExistence(t, tests.BundleName, bundleVersion, true, artHttpDetails)

	// Update release bundle to have b1.in
	artifactoryCli.Exec("rbu", tests.BundleName, bundleVersion, tests.DistRepo1+"/data/b1.in", "--sign")

	// Distribute release bundle
	artifactoryCli.Exec("rbd", tests.BundleName, bundleVersion, "--site=*", "--sync")

	// Download by bundle version, b2 and b3 should not be downloaded, b1 should
	artifactoryCli.Exec("dl "+tests.DistRepo1+"/data/* "+tests.Out+fileutils.GetFileSeparator()+"download"+fileutils.GetFileSeparator()+"simple_by_build"+fileutils.GetFileSeparator(), "--bundle="+tests.BundleName+"/"+bundleVersion)

	// Validate files are downloaded by bundle version
	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err = tests.ValidateListsIdentical(tests.GetBuildSimpleDownload(), paths)
	assert.NoError(t, err)

	// Cleanup
	cleanDistributionTest(t)
}

func TestCreateBundleText(t *testing.T) {
	initDistributionTest(t)

	// Upload files
	specFile, err := tests.CreateSpec(tests.DistributionUploadSpecB)
	assert.NoError(t, err)
	artifactoryCli.Exec("u", "--spec="+specFile)

	// Create a release bundle with release notes and description
	releaseNotesPath := filepath.Join(tests.GetTestResourcesPath(), "distribution", "releasenotes.md")
	description := "thisIsADescription"
	artifactoryCli.Exec("rbc", tests.BundleName, bundleVersion, tests.DistRepo1+"/data/*", "--release-notes-path="+releaseNotesPath, "--desc="+description)

	// Validate release notes and description
	distributableResponse := inttestutils.GetLocalBundle(t, tests.BundleName, bundleVersion, artHttpDetails)
	if distributableResponse != nil {
		assert.Equal(t, description, distributableResponse.Description)
		releaseNotes, err := ioutil.ReadFile(releaseNotesPath)
		assert.NoError(t, err)
		assert.Equal(t, string(releaseNotes), distributableResponse.ReleaseNotes.Content)
		assert.Equal(t, "markdown", distributableResponse.ReleaseNotes.Syntax)
	}

	cleanDistributionTest(t)
}
