package main

import (
	"errors"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"os"
	"path/filepath"
	"testing"

	"github.com/jfrog/jfrog-cli-core/utils/coreutils"

	"github.com/jfrog/jfrog-cli-core/utils/config"
	"github.com/jfrog/jfrog-cli/inttestutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/jfrog/jfrog-client-go/auth"
	clientDistUtils "github.com/jfrog/jfrog-client-go/distribution/services/utils"
	"github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/io/httputils"
	"github.com/stretchr/testify/assert"
)

const bundleVersion = "10"

var (
	distributionDetails *config.ServerDetails
	distAuth            auth.ServiceDetails
	distHttpDetails     httputils.HttpClientDetails
	// JFrog CLI for Distribution commands
	distributionCli *tests.JfrogCli
)

func InitDistributionTests() {
	initDistributionCli()
	inttestutils.CleanUpOldBundles(distHttpDetails, bundleVersion, distributionCli)
	InitArtifactoryTests()
	inttestutils.SendGpgKeys(artHttpDetails, distHttpDetails)
}

func CleanDistributionTests() {
	deleteCreatedRepos()
}

func authenticateDistribution() string {
	// Due to a bug in distribution when authenticate with a multi-scope token,
	// we must send a username as well as token or password.
	distributionDetails = &config.ServerDetails{
		DistributionUrl: *tests.RtDistributionUrl,
		User:            *tests.RtUser,
	}
	cred := "--dist-url=" + *tests.RtDistributionUrl + " --user=" + *tests.RtUser

	// Prefer the distribution token if provided.
	distributionDetails.AccessToken = *tests.RtDistributionAccessToken
	if distributionDetails.AccessToken == "" {
		distributionDetails.AccessToken = *tests.RtAccessToken
	}

	if distributionDetails.AccessToken != "" {
		cred += " --access-token=" + distributionDetails.AccessToken
	} else {
		distributionDetails.Password = *tests.RtPassword
		cred += " --password=" + *tests.RtPassword
	}

	var err error
	if distAuth, err = distributionDetails.CreateDistAuthConfig(); err != nil {
		coreutils.ExitOnErr(errors.New("Failed while attempting to authenticate with Artifactory: " + err.Error()))
	}
	distributionDetails.DistributionUrl = distAuth.GetUrl()
	distHttpDetails = distAuth.CreateHttpClientDetails()
	return cred
}

func initDistributionCli() {
	if distributionCli != nil {
		return
	}
	*tests.RtDistributionUrl = utils.AddTrailingSlashIfNeeded(*tests.RtDistributionUrl)
	cred := authenticateDistribution()
	distributionCli = tests.NewJfrogCli(execMain, "jfrog rt", cred)
}

func initDistributionTest(t *testing.T) {
	if !*tests.TestDistribution {
		t.Skip("Skipping distribution test. To run distribution test add the '-test.distribution=true' option.")
	}
}

func cleanDistributionTest(t *testing.T) {
	distributionCli.Exec("rbdel", tests.BundleName, bundleVersion, "--site=*", "--delete-from-dist", "--quiet")
	inttestutils.WaitForDeletion(t, tests.BundleName, bundleVersion, distHttpDetails)
	inttestutils.CleanDistributionRepositories(t, serverDetails)
	tests.CleanFileSystem()
}

func TestBundleAsyncDistDownload(t *testing.T) {
	initDistributionTest(t)

	// Upload files
	specFile, err := tests.CreateSpec(tests.DistributionUploadSpecB)
	assert.NoError(t, err)
	runRt(t, "u", "--spec="+specFile)

	// Create and distribute release bundle
	runRb(t, "rbc", tests.BundleName, bundleVersion, tests.DistRepo1+"/data/b1.in", "--sign")
	runRb(t, "rbd", tests.BundleName, bundleVersion, "--site=*")
	inttestutils.WaitForDistribution(t, tests.BundleName, bundleVersion, distHttpDetails)

	// Download by bundle version, b2 and b3 should not be downloaded, b1 should
	runRt(t, "dl", tests.DistRepo1+"/data/*", tests.Out+fileutils.GetFileSeparator()+"download"+fileutils.GetFileSeparator()+"simple_by_build"+fileutils.GetFileSeparator(), "--bundle="+tests.BundleName+"/"+bundleVersion)

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
	runRt(t, "u", "--spec="+specFile)

	// Create release bundle
	distributionRules, err := tests.CreateSpec(tests.DistributionRules)
	assert.NoError(t, err)
	runRb(t, "rbc", tests.BundleName, bundleVersion, tests.DistRepo1+"/data/b1.in", "--sign")
	runRb(t, "rbd", tests.BundleName, bundleVersion, "--dist-rules="+distributionRules, "--sync")

	// Download by bundle version, b2 and b3 should not be downloaded, b1 should
	specFile, err = tests.CreateSpec(tests.BundleDownloadSpec)
	assert.NoError(t, err)
	runRt(t, "dl", "--spec="+specFile)

	// Validate files are downloaded by bundle version
	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err = tests.ValidateListsIdentical(tests.GetBuildSimpleDownload(), paths)
	assert.NoError(t, err)

	// Cleanup
	cleanDistributionTest(t)
}

func TestBundleCreateByAql(t *testing.T) {
	initDistributionTest(t)

	// Upload files
	specFile, err := tests.CreateSpec(tests.DistributionUploadSpecB)
	assert.NoError(t, err)
	runRt(t, "u", "--spec="+specFile)

	// Create release bundle by AQL
	spec, err := tests.CreateSpec(tests.DistributionCreateByAql)
	assert.NoError(t, err)
	runRb(t, "rbc", tests.BundleName, bundleVersion, "--spec="+spec, "--sign")
	runRb(t, "rbd", tests.BundleName, bundleVersion, "--site=*", "--sync")

	// Download by bundle version, b2 and b3 should not be downloaded, b1 should
	specFile, err = tests.CreateSpec(tests.BundleDownloadSpec)
	assert.NoError(t, err)
	runRt(t, "dl", "--spec="+specFile)

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
	runRt(t, "u", "--spec="+specFile)

	// Create release bundle
	runRb(t, "rbc", tests.BundleName, bundleVersion, tests.DistRepo1+"/data/b1.in", "--sign")
	runRb(t, "rbd", tests.BundleName, bundleVersion, "--site=*", "--sync")

	// Download by bundle name and version with pattern "*", b2 and b3 should not be downloaded, b1 should
	runRt(t, "dl", "*", "out/download/simple_by_build/data/", "--bundle="+tests.BundleName+"/"+bundleVersion, "--flat")

	// Validate files are downloaded by bundle version
	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	err = tests.ValidateListsIdentical(tests.GetBuildSimpleDownload(), paths)
	assert.NoError(t, err)

	// Download by bundle name and version version without pattern, b2 and b3 should not be downloaded, b1 should
	tests.CleanFileSystem()
	specFile, err = tests.CreateSpec(tests.BundleDownloadSpecNoPattern)
	runRt(t, "dl", "--spec="+specFile, "--flat")

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
	runRt(t, "u", "--spec="+specFile)

	// Create release bundle. Include b1.in and b2.in. Exclude b3.in.
	runRb(t, "rbc", tests.BundleName, bundleVersion, tests.DistRepo1+"/data/b*.in", "--sign", "--exclusions=*b3.in")
	runRb(t, "rbd", tests.BundleName, bundleVersion, "--site=*", "--sync")

	// Download by bundle version, b2 and b3 should not be downloaded, b1 should
	runRt(t, "dl", tests.DistRepo1+"/data/*", tests.Out+fileutils.GetFileSeparator()+"download"+fileutils.GetFileSeparator()+"simple_by_build"+fileutils.GetFileSeparator(), "--bundle="+tests.BundleName+"/"+bundleVersion, "--exclusions=*b2.in")

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
	runRt(t, "u", "--spec="+specFileA)
	runRt(t, "u", "--spec="+specFileB)

	// Create release bundle
	runRb(t, "rbc", tests.BundleName, bundleVersion, tests.DistRepo1+"/data/a*", "--sign")
	runRb(t, "rbd", tests.BundleName, bundleVersion, "--site=*", "--sync")

	// Copy by bundle name and version
	specFile, err := tests.CreateSpec(tests.CopyByBundleSpec)
	assert.NoError(t, err)
	runRt(t, "cp", "--spec="+specFile)

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
	runRt(t, "u", "testdata/a/a1.in", tests.DistRepo1+"/a.in")

	// Create release bundle
	runRb(t, "rbc", tests.BundleName, bundleVersion, tests.DistRepo1+"/a.in", "--sign")
	runRb(t, "rbd", tests.BundleName, bundleVersion, "--site=*", "--sync")

	// Set the 'prop=red' property to the file.
	runRt(t, "sp", tests.DistRepo1+"/a.*", "prop=red", "--bundle="+tests.BundleName+"/"+bundleVersion)
	// Now let's change the property value, by searching for the 'prop=red'.
	specFile, err := tests.CreateSpec(tests.DistributionSetDeletePropsSpec)
	assert.NoError(t, err)
	runRt(t, "sp", "prop=green", "--spec="+specFile, "--bundle="+tests.BundleName+"/"+bundleVersion)

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
	runRt(t, "u", "--spec="+specFile)

	// Create a release bundle without --sign and make sure it is not signed
	runRb(t, "rbc", tests.BundleName, bundleVersion, tests.DistRepo1+"/data/b1.in")
	distributableResponse := inttestutils.GetLocalBundle(t, tests.BundleName, bundleVersion, distHttpDetails)
	inttestutils.AssertReleaseBundleOpen(t, distributableResponse)

	// Sign the release bundle and make sure it is signed
	runRb(t, "rbs", tests.BundleName, bundleVersion)
	distributableResponse = inttestutils.GetLocalBundle(t, tests.BundleName, bundleVersion, distHttpDetails)
	inttestutils.AssertReleaseBundleSigned(t, distributableResponse)

	// Cleanup
	cleanDistributionTest(t)
}

func TestBundleDeleteLocal(t *testing.T) {
	initDistributionTest(t)

	// Upload files
	specFile, err := tests.CreateSpec(tests.DistributionUploadSpecB)
	assert.NoError(t, err)
	runRt(t, "u", "--spec="+specFile)

	// Create a release bundle
	runRb(t, "rbc", tests.BundleName, bundleVersion, tests.DistRepo1+"/data/b1.in", "--sign")
	inttestutils.VerifyLocalBundleExistence(t, tests.BundleName, bundleVersion, true, distHttpDetails)

	// Delete release bundle locally
	runRb(t, "rbdel", tests.BundleName, bundleVersion, "--site=*", "--delete-from-dist", "--quiet")
	inttestutils.VerifyLocalBundleExistence(t, tests.BundleName, bundleVersion, false, distHttpDetails)

	// Cleanup
	cleanDistributionTest(t)
}

func TestUpdateReleaseBundle(t *testing.T) {
	initDistributionTest(t)

	// Upload files
	specFile, err := tests.CreateSpec(tests.DistributionUploadSpecB)
	assert.NoError(t, err)
	runRt(t, "u", "--spec="+specFile)

	// Create a release bundle with b2.in
	runRb(t, "rbc", tests.BundleName, bundleVersion, tests.DistRepo1+"/data/b2.in")
	inttestutils.VerifyLocalBundleExistence(t, tests.BundleName, bundleVersion, true, distHttpDetails)

	// Update release bundle to have b1.in
	runRb(t, "rbu", tests.BundleName, bundleVersion, tests.DistRepo1+"/data/b1.in", "--sign")

	// Distribute release bundle
	runRb(t, "rbd", tests.BundleName, bundleVersion, "--site=*", "--sync")

	// Download by bundle version, b2 and b3 should not be downloaded, b1 should
	runRt(t, "dl", tests.DistRepo1+"/data/*", tests.Out+fileutils.GetFileSeparator()+"download"+fileutils.GetFileSeparator()+"simple_by_build"+fileutils.GetFileSeparator(), "--bundle="+tests.BundleName+"/"+bundleVersion)

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
	runRt(t, "u", "--spec="+specFile)

	// Create a release bundle with release notes and description
	releaseNotesPath := filepath.Join(tests.GetTestResourcesPath(), "distribution", "releasenotes.md")
	description := "thisIsADescription"
	runRb(t, "rbc", tests.BundleName, bundleVersion, tests.DistRepo1+"/data/*", "--release-notes-path="+releaseNotesPath, "--desc="+description)

	// Validate release notes and description
	distributableResponse := inttestutils.GetLocalBundle(t, tests.BundleName, bundleVersion, distHttpDetails)
	if distributableResponse != nil {
		assert.Equal(t, description, distributableResponse.Description)
		releaseNotes, err := os.ReadFile(releaseNotesPath)
		assert.NoError(t, err)
		assert.Equal(t, string(releaseNotes), distributableResponse.ReleaseNotes.Content)
		assert.Equal(t, clientDistUtils.Markdown, distributableResponse.ReleaseNotes.Syntax)
	}

	cleanDistributionTest(t)
}

func TestCreateBundleProps(t *testing.T) {
	initDistributionTest(t)

	// Upload files
	specFile, err := tests.CreateSpec(tests.DistributionUploadSpecB)
	assert.NoError(t, err)
	runRt(t, "u", "--spec="+specFile)

	// Create and distribute release bundle with added props
	runRb(t, "rbc", tests.BundleName, bundleVersion, tests.DistRepo1+"/data/*", "--target-props=key1=val1;key2=val2,val3", "--sign")
	inttestutils.VerifyLocalBundleExistence(t, tests.BundleName, bundleVersion, true, distHttpDetails)
	runRb(t, "rbd", tests.BundleName, bundleVersion, "--site=*", "--sync")

	// Verify props are added to the distributes artifact
	verifyExistInArtifactoryByProps(tests.GetBundlePropsExpected(), tests.DistRepo1+"/data/", "key1=val1;key2=val2;key2=val3", t)

	cleanDistributionTest(t)
}

func TestUpdateBundleProps(t *testing.T) {
	initDistributionTest(t)

	// Upload files
	specFile, err := tests.CreateSpec(tests.DistributionUploadSpecB)
	assert.NoError(t, err)
	runRt(t, "u", "--spec="+specFile)

	// Create, update and distribute release bundle with added props
	runRb(t, "rbc", tests.BundleName, bundleVersion, tests.DistRepo1+"/data/*")
	runRb(t, "rbu", tests.BundleName, bundleVersion, tests.DistRepo1+"/data/*", "--target-props=key1=val1", "--sign")
	inttestutils.VerifyLocalBundleExistence(t, tests.BundleName, bundleVersion, true, distHttpDetails)
	runRb(t, "rbd", tests.BundleName, bundleVersion, "--site=*", "--sync")

	// Verify props are added to the distributes artifact
	verifyExistInArtifactoryByProps(tests.GetBundlePropsExpected(), tests.DistRepo1+"/data/", "key1=val1", t)

	cleanDistributionTest(t)
}

func TestBundlePathMapping(t *testing.T) {
	initDistributionTest(t)

	// Upload files
	specFile, err := tests.CreateSpec(tests.DistributionUploadSpecB)
	assert.NoError(t, err)
	runRt(t, "u", "--spec="+specFile)

	// Create and distribute release bundle with path mapping from <DistRepo1>/data/ to <DistRepo2>/target/
	runRb(t, "rbc", tests.BundleName, bundleVersion, tests.DistRepo1+"/data/(*)", "--sign", "--target="+tests.DistRepo2+"/target/{1}")
	runRb(t, "rbd", tests.BundleName, bundleVersion, "--site=*", "--sync")

	// Validate files are distributed to the target mapping
	spec, err := tests.CreateSpec(tests.DistributionMappingDownload)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetBundleMappingExpected(), spec, t)

	cleanDistributionTest(t)
}

func TestBundlePathMappingUsingSpec(t *testing.T) {
	initDistributionTest(t)

	// Upload files
	specFile, err := tests.CreateSpec(tests.DistributionUploadSpecB)
	assert.NoError(t, err)
	runRt(t, "u", "--spec="+specFile)

	// Create and distribute release bundle with path mapping from <DistRepo1>/data/ to <DistRepo2>/target/
	spec, err := tests.CreateSpec(tests.DistributionCreateWithMapping)
	assert.NoError(t, err)
	runRb(t, "rbc", tests.BundleName, bundleVersion, "--sign", "--spec="+spec)
	runRb(t, "rbd", tests.BundleName, bundleVersion, "--site=*", "--sync")

	// Validate files are distributed to the target mapping
	spec, err = tests.CreateSpec(tests.DistributionMappingDownload)
	assert.NoError(t, err)
	verifyExistInArtifactory(tests.GetBundleMappingExpected(), spec, t)

	cleanDistributionTest(t)
}

func TestReleaseBundleCreateDetailedSummary(t *testing.T) {
	initDistributionTest(t)

	// Upload files
	specFile, err := tests.CreateSpec(tests.DistributionUploadSpecB)
	assert.NoError(t, err)
	runRt(t, "u", "--spec="+specFile)

	buffer, previousLog := tests.RedirectLogOutputToBuffer()
	// Restore previous logger when the function returns
	defer log.SetLogger(previousLog)

	// Create a release bundle with b2.in
	runRb(t, "rbc", tests.BundleName, bundleVersion, tests.DistRepo1+"/data/b2.in", "--sign", "--detailed-summary")
	inttestutils.VerifyLocalBundleExistence(t, tests.BundleName, bundleVersion, true, distHttpDetails)

	tests.VerifySha256DetailedSummaryFromBuffer(t, buffer, previousLog)

	// Cleanup
	cleanDistributionTest(t)
}

func TestReleaseBundleUpdateDetailedSummary(t *testing.T) {
	initDistributionTest(t)

	// Upload files
	specFile, err := tests.CreateSpec(tests.DistributionUploadSpecB)
	assert.NoError(t, err)
	runRt(t, "u", "--spec="+specFile)

	buffer, previousLog := tests.RedirectLogOutputToBuffer()
	// Restore previous logger when the function returns
	defer log.SetLogger(previousLog)

	// Create a release bundle with b2.in
	runRb(t, "rbc", tests.BundleName, bundleVersion, tests.DistRepo1+"/data/b2.in")
	inttestutils.VerifyLocalBundleExistence(t, tests.BundleName, bundleVersion, true, distHttpDetails)

	// Update release bundle to have b1.in
	runRb(t, "rbu", tests.BundleName, bundleVersion, tests.DistRepo1+"/data/b1.in", "--sign", "--detailed-summary")

	tests.VerifySha256DetailedSummaryFromBuffer(t, buffer, previousLog)

	// Cleanup
	cleanDistributionTest(t)
}

func TestReleaseBundleSignDetailedSummary(t *testing.T) {
	initDistributionTest(t)

	// Upload files
	specFile, err := tests.CreateSpec(tests.DistributionUploadSpecB)
	assert.NoError(t, err)
	runRt(t, "u", "--spec="+specFile)

	buffer, previousLog := tests.RedirectLogOutputToBuffer()
	// Restore previous logger when the function returns
	defer log.SetLogger(previousLog)

	// Create a release bundle with b2.in
	runRb(t, "rbc", tests.BundleName, bundleVersion, tests.DistRepo1+"/data/b2.in")
	inttestutils.VerifyLocalBundleExistence(t, tests.BundleName, bundleVersion, true, distHttpDetails)

	// Update release bundle to have b1.in
	runRb(t, "rbs", tests.BundleName, bundleVersion, "--detailed-summary")

	tests.VerifySha256DetailedSummaryFromBuffer(t, buffer, previousLog)

	// Cleanup
	cleanDistributionTest(t)
}

// Run `jfrog rt rb*` command`. The first arg is the distribution command, such as 'rbc', 'rbu', etc.
func runRb(t *testing.T, args ...string) {
	err := distributionCli.Exec(args...)
	assert.NoError(t, err)
}

// Run `jfrog rt` command
func runRt(t *testing.T, args ...string) {
	err := artifactoryCli.Exec(args...)
	assert.NoError(t, err)
}
