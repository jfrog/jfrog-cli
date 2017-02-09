package main

import (
	"testing"
	"github.com/jfrogdev/jfrog-cli-go/artifactory/commands"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/tests"
	"fmt"
	"github.com/jfrogdev/jfrog-cli-go/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"io/ioutil"
	"errors"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/log"
	"github.com/buger/jsonparser"
	"os"
	"path/filepath"
	"runtime"
)

var artifactoryCli *tests.JfrogCli
var artifactoryDetails *config.ArtifactoryDetails

func InitArtifactoryTests() {
	if !*tests.TestArtifactory {
		return
	}

	artifactoryDetails = new(config.ArtifactoryDetails)
	cred := "--url=" + *tests.RtUrl
	artifactoryDetails.Url = *tests.RtUrl
	if *tests.RtApiKey != "" {
		cred += " --apikey=" + *tests.RtApiKey
		artifactoryDetails.ApiKey = *tests.RtApiKey
	} else {
		cred += " --user=" + *tests.RtUser + " --password=" + *tests.RtPassword
		artifactoryDetails.User = *tests.RtUser
		artifactoryDetails.Password = *tests.RtPassword
	}

	artifactoryCli = tests.NewJfrogCli(main, "jfrog rt", cred)
	createReposIfNeeded()
	cleanArtifactoryTest()
}

func TestArtifactorySimpleUploadSpec(t *testing.T) {
	initArtifactoryTest(t)
	specFile := tests.GetFilePath(tests.SimpleUploadSpec)
	artifactoryCli.Exec("upload", "--spec=" + specFile)

	isExistInArtifactory(tests.SimpleUploadExpectedRepo1, tests.GetFilePath(tests.Search), t)
	cleanArtifactoryTest()
}

func TestArtifactoryUploadandExplode(t *testing.T) {
	initArtifactoryTest(t)
	artifactoryCli.Exec("upload", "../testsdata/a.zip", "jfrog-cli-tests-repo1", "--explode=true")
	isExistInArtifactory(tests.ExplodeUploadExpectedRepo1, tests.GetFilePath(tests.Search), t)
	cleanArtifactoryTest()
}

func TestArtifactoryUploadFromHomeDir(t *testing.T) {
	initArtifactoryTest(t)

	testFileRel := "~" + ioutils.GetFileSeperator() + "cliTestFile.txt"
	testFileAbs := ioutils.GetHomeDir() + "/cliTestFile.txt"

	d1 := []byte("test file")
	err := ioutil.WriteFile(testFileAbs, d1, 0644)
	if err != nil {
		t.Error("Coudln't create file:", err)
	}

	artifactoryCli.Exec("upload", testFileRel, tests.Repo1, "--recursive=false")
	isExistInArtifactory(tests.TxtUploadExpectedRepo1, tests.GetFilePath(tests.SearchTxt), t)

	os.Remove(testFileAbs)
	cleanArtifactoryTest()
}

func TestArtifactoryCopySpec(t *testing.T) {
	initArtifactoryTest(t)
	prepUploadFiles()
	prepCopyFiles()

	specFile := tests.GetFilePath(tests.MoveCopyDeleteSpec)
	artifactoryCli.Exec("copy", "--spec=" + specFile)

	isExistInArtifactory(tests.MassiveMoveExpected, tests.GetFilePath(tests.SearchMoveDeleteRepoSpec), t)
	cleanArtifactoryTest()
}

// Upload symlink by full path to Artifactory and the link content checksum
// Download the symlink which was uploaded.
// validate the symlink content checksum.
func TestSimpleSymlinkHandling(t *testing.T) {
	if runtime.GOOS == "windows" {
		return
	}
	initArtifactoryTest(t)
	localFile := filepath.Join(tests.GetTestResourcesPath() + "a/", "a1.in")
	link := filepath.Join(tests.GetTestResourcesPath() + "a/", "link")
	err := os.Symlink(localFile, link)
	if err != nil {
		t.Error(err.Error())
	}
	artifactoryCli.Exec("u", link + " " + tests.Repo1 + " --symlinks=true")
	err = os.Remove(link)
	if err != nil {
		t.Error(err.Error())
	}
	artifactoryCli.Exec("dl", tests.Repo1 + "/link " + tests.GetTestResourcesPath() + "a/ --validate-symlinks=true")
	validateSymLink(link, localFile, t)
	os.Remove(link)
	cleanArtifactoryTest()
}

// Upload symlink to Artifactory using wildcard pattern and the link content checksum
// Download the symlink which was uploaded.
// validate the symlink content checksum.
func TestSymlinkWildcardPathHandling(t *testing.T) {
	if runtime.GOOS == "windows" {
		return
	}
	initArtifactoryTest(t)
	localFile := filepath.Join(tests.GetTestResourcesPath() + "a/", "a1.in")
	link := filepath.Join(tests.GetTestResourcesPath() + "a/", "link")
	err := os.Symlink(localFile, link)
	if err != nil {
		t.Error(err.Error())
	}
	link1 := filepath.Join(tests.GetTestResourcesPath() + "a/", "link*")
	artifactoryCli.Exec("u", link1 + " " + tests.Repo1 + " --symlinks=true")
	err = os.Remove(link)
	if err != nil {
		t.Error(err.Error())
	}
	artifactoryCli.Exec("dl", tests.Repo1 + "/link " + tests.GetTestResourcesPath() + "a/ --validate-symlinks=true")
	validateSymLink(link, localFile, t)
	os.Remove(link)
	cleanArtifactoryTest()
}

// Upload symlink pointing to directory to Artifactory.
// Download the symlink which was uploaded.
func TestSymlinkToDirHandling(t *testing.T) {
	if runtime.GOOS == "windows" {
		return
	}
	initArtifactoryTest(t)
	localFile := filepath.Join(tests.GetTestResourcesPath(), "a")
	link := filepath.Join(tests.GetTestResourcesPath() + "a/", "link")
	err := os.Symlink(localFile, link)
	if err != nil {
		t.Error(err.Error())
	}
	artifactoryCli.Exec("u", link + " " + tests.Repo1 + " --symlinks=true --recursive=true")
	err = os.Remove(link)
	if err != nil {
		t.Error(err.Error())
	}
	artifactoryCli.Exec("dl", tests.Repo1 + "/link " + tests.GetTestResourcesPath() + "a/")
	validateSymLink(link, localFile, t)
	os.Remove(link)
	cleanArtifactoryTest()
}

// Upload symlink pointing to directory using wildcard path to Artifactory.
// Download the symlink which was uploaded.
func TestSymlinkToDirWilcardHandling(t *testing.T) {
	if runtime.GOOS == "windows" {
		return
	}
	initArtifactoryTest(t)
	localFile := filepath.Join(tests.GetTestResourcesPath(), "a")
	link := filepath.Join(tests.GetTestResourcesPath() + "a/", "link")
	err := os.Symlink(localFile, link)
	if err != nil {
		t.Error(err.Error())
	}
	link1 := filepath.Join(tests.GetTestResourcesPath() + "a/", "lin*")
	artifactoryCli.Exec("u", link1 + " " + tests.Repo1 + " --symlinks=true --recursive=true")
	err = os.Remove(link)
	if err != nil {
		t.Error(err.Error())
	}
	artifactoryCli.Exec("dl", tests.Repo1 + "/link " + tests.GetTestResourcesPath() + "a/")
	validateSymLink(link, localFile, t)
	os.Remove(link)
	cleanArtifactoryTest()
}

// Upload symlink pointing to directory using wildcard path to Artifactory.
// Download the symlink which was uploaded.
// The test create circular links and the test suppose to prune the circular searching.
func TestSymlinkInsideSymlinkDirWithRecursionIssueUpload(t *testing.T) {
	if runtime.GOOS == "windows" {
		return
	}
	initArtifactoryTest(t)
	localDirPath := filepath.Join(tests.GetTestResourcesPath(), "a")
	link1 := filepath.Join(tests.GetTestResourcesPath() + "a/", "link1")
	err := os.Symlink(localDirPath, link1)
	if err != nil {
		t.Error(err.Error())
	}
	localFilePath := filepath.Join(tests.GetTestResourcesPath() + "a/", "a1.in")
	link2 := filepath.Join(tests.GetTestResourcesPath() + "a/", "link2")
	err = os.Symlink(localFilePath, link2)
	if err != nil {
		t.Error(err.Error())
	}

	artifactoryCli.Exec("u", localDirPath + "/link* " + tests.Repo1 + " --symlinks=true --recursive=true")
	err = os.Remove(link1)
	if err != nil {
		t.Error(err.Error())
	}

	err = os.Remove(link2)
	if err != nil {
		t.Error(err.Error())
	}

	artifactoryCli.Exec("dl", tests.Repo1 + "/link* " + tests.GetTestResourcesPath() + "a/")
	validateSymLink(link1, localDirPath, t)
	os.Remove(link1)
	validateSymLink(link2, localFilePath, t)
	os.Remove(link2)
	cleanArtifactoryTest()
}

func validateSymLink(localLinkPath, localFilePath string, t *testing.T) {
	exists := ioutils.IsPathSymlink(localLinkPath)
	if !exists {
		t.Error(errors.New("Faild to download symlinks from artifactory"))
	}
	symlinks, err := filepath.EvalSymlinks(localLinkPath)
	if err != nil {
		t.Error(errors.New("Can't eval symlinks"))
	}
	if symlinks != localFilePath {
		t.Error(errors.New("Symlinks wasn't created as expected. expected:" + localFilePath + " actual: " + symlinks))
	}
}

func TestArtifactoryDelete(t *testing.T) {
	initArtifactoryTest(t)
	prepUploadFiles()
	prepCopyFiles()

	specFile := tests.GetFilePath(tests.MoveCopyDeleteSpec)
	artifactoryCli.Exec("copy", "--spec=" + specFile)
	artifactoryCli.Exec("delete", tests.Repo2 + "/nonflat_recursive_target/nonflat_recursive_source/a/b/*", "--quiet=true")

	isExistInArtifactory(tests.Delete1, tests.GetFilePath(tests.SearchMoveDeleteRepoSpec), t)
	cleanArtifactoryTest()
}

func TestArtifactoryDeleteFolderWithWildcard(t *testing.T) {
	initArtifactoryTest(t)
	prepUploadFiles()
	prepCopyFiles()

	specFile := tests.GetFilePath(tests.MoveCopyDeleteSpec)
	artifactoryCli.Exec("copy", "--spec=" + specFile)

	artHttpDetails := utils.GetArtifactoryHttpClientDetails(artifactoryDetails)
	resp, _, _, _ := ioutils.SendGet(*tests.RtUrl + "api/storage/" + tests.Repo2 + "/nonflat_recursive_target/nonflat_recursive_source/a/b/", true, artHttpDetails)
	if resp.StatusCode != 200 {
		t.Error("Missing folder in artifactory : " + tests.Repo2 + "/nonflat_recursive_target/nonflat_recursive_source/a/b/")
	}

	artifactoryCli.Exec("delete", tests.Repo2 + "/nonflat_recursive_target/nonflat_recursive_source/*/b/", "--quiet=true")
	resp, _, _, _ = ioutils.SendGet(*tests.RtUrl + "api/storage/" + tests.Repo2 + "/nonflat_recursive_target/nonflat_recursive_source/a/b/", true, artHttpDetails)
	if resp.StatusCode != 404 {
		t.Error("Couldn't delete folder in artifactory : " + tests.Repo2 + "/nonflat_recursive_target/nonflat_recursive_source/a/b/")
	}

	isExistInArtifactory(tests.Delete1, tests.GetFilePath(tests.SearchMoveDeleteRepoSpec), t)
	cleanArtifactoryTest()
}

func TestArtifactoryDeleteFolder(t *testing.T) {
	initArtifactoryTest(t)
	prepUploadFiles()
	artifactoryCli.Exec("delete", tests.Repo1 + "/downloadTestResources/", "--quiet=true")

	artHttpDetails := utils.GetArtifactoryHttpClientDetails(artifactoryDetails)
	resp, body, _, err := ioutils.SendGet(*tests.RtUrl + "api/storage/" + tests.Repo1 + "/downloadTestResources", true, artHttpDetails)
	if err != nil || resp.StatusCode != 404 {
		t.Error("Coudln't delete path: " + tests.Repo1 + "/downloadTestResources/ " + string(body))
	}

	cleanArtifactoryTest()
}

func TestArtifactoryDeleteFoldersBySpec(t *testing.T) {
	initArtifactoryTest(t)
	prepUploadFiles()
	prepCopyFiles()

	artifactoryCli.Exec("delete", "--spec=" + tests.GetFilePath(tests.DeleteSpec), "--quiet=true")

	artHttpDetails := utils.GetArtifactoryHttpClientDetails(artifactoryDetails)
	resp, body, _, err := ioutils.SendGet(*tests.RtUrl + "api/storage/" + tests.Repo1 + "/downloadTestResources", true, artHttpDetails)
	if err != nil || resp.StatusCode != 404 {
		t.Error("Coudln't delete path: " + tests.Repo1 + "/downloadTestResources/ " + string(body))
	}
	resp, body, _, err = ioutils.SendGet(*tests.RtUrl + "api/storage/" + tests.Repo2 + "/downloadTestResources", true, artHttpDetails)
	if err != nil || resp.StatusCode != 404 {
		t.Error("Coudln't delete path: " + tests.Repo2 + "/downloadTestResources/ " + string(body))
	}

	cleanArtifactoryTest()
}

func TestArtifactoryDisplyedPathToDelete(t *testing.T) {
	initArtifactoryTest(t)
	prepUploadFiles()
	prepCopyFiles()

	specFile := tests.GetFilePath(tests.DeleteComplexSpec)
	artifactsToDelete := getPathsToDelete(specFile)
	var displyedPaths []commands.SearchResult
	for _, v := range artifactsToDelete {
		displyedPaths = append(displyedPaths, commands.SearchResult{Path:v.Repo + "/" + v.Path + "/" + v.Name})
	}

	tests.CompareExpectedVsActuals(tests.DeleteDisplyedFiles, displyedPaths, t)
	cleanArtifactoryTest()
}

func TestArtifactoryDeleteBySpec(t *testing.T) {
	initArtifactoryTest(t)
	prepUploadFiles()
	prepCopyFiles()

	specFile := tests.GetFilePath(tests.DeleteComplexSpec)
	artifactoryCli.Exec("delete", "--spec=" + specFile, "--quiet=true")

	artifactsToDelete := getPathsToDelete(specFile)
	if len(artifactsToDelete) != 0 {
		t.Error("Couldn't delete paths")
	}

	cleanArtifactoryTest()
}

func TestArtifactoryMassiveDownloadSpec(t *testing.T) {
	initArtifactoryTest(t)
	prepUploadFiles()
	specFile := tests.GetFilePath(tests.DownloadSpec)
	artifactoryCli.Exec("download", "--spec=" + specFile)

	paths, _ := ioutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	tests.IsExistLocally(tests.MassiveDownload, paths, t)
	cleanArtifactoryTest()
}

func TestArtifactoryMassiveUploadSpec(t *testing.T) {
	initArtifactoryTest(t)
	specFile := tests.GetFilePath(tests.UploadSpec)
	resultSpecFile := tests.GetFilePath(tests.Search)
	artifactoryCli.Exec("upload", "--spec=" + specFile)

	isExistInArtifactory(tests.MassiveUpload, resultSpecFile, t)
	isExistInArtifactoryByProps(tests.PropsExpected, tests.Repo1 + "/*/properties/*.in", "searchMe=true", t)

	cleanArtifactoryTest()
}

func TestArtifactoryPublishBuildInfo(t *testing.T) {
	initArtifactoryTest(t)
	buildName, buildNumber := "cli-test-build", "10"

	//upload files with buildName and buildNumber
	specFile := tests.GetFilePath(tests.SimpleUploadSpec)
	artifactoryCli.Exec("upload", "--spec=" + specFile, "--build-name=" + buildName, "--build-number=" + buildNumber)

	//publish buildInfo
	artifactoryCli.Exec("build-publish", buildName, buildNumber)

	//validate files are uploaded with the build info name and number
	props := fmt.Sprintf("build.name=%v;build.number=%v", buildName, buildNumber)
	isExistInArtifactoryByProps(tests.SimpleUploadExpectedRepo1, tests.Repo1 + "/*", props, t)

	//cleanup
	deleteBuild(buildName)
	cleanArtifactoryTest()
}

func TestArtifactoryDownloadByBuildUsingSpec(t *testing.T) {
	initArtifactoryTest(t)
	buildName, buildNumberA, buildNumberB := "cli-test-build", "10", "11"
	specFile := tests.GetFilePath(tests.BuildDownloadSpec)

	//upload files with buildName and buildNumber
	specFileA := tests.GetFilePath(tests.SplittedUploadSpecA)
	specFileB := tests.GetFilePath(tests.SplittedUploadSpecB)
	artifactoryCli.Exec("upload", "--spec=" + specFileA, "--build-name=" + buildName, "--build-number=" + buildNumberA)
	artifactoryCli.Exec("upload", "--spec=" + specFileB, "--build-name=" + buildName, "--build-number=" + buildNumberB)

	//publish buildInfo
	artifactoryCli.Exec("build-publish", buildName, buildNumberA)
	artifactoryCli.Exec("build-publish", buildName, buildNumberB)

	// Download by build number
	artifactoryCli.Exec("download", "--spec=" + specFile)

	//validate files are downloaded by build number
	paths, _ := ioutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	tests.IsLocalExactAsExpected(tests.BuildDownload, paths, t)

	//cleanup
	deleteBuild(buildName)
	cleanArtifactoryTest()
}

func TestArtifactoryDownloadByBuildUsingSimpleDownload(t *testing.T) {
	initArtifactoryTest(t)
	buildName, buildNumberA, buildNumberB := "cli-test-build", "10", "11"

	//upload files with buildName and buildNumber
	specFileA := tests.GetFilePath(tests.SplittedUploadSpecA)
	specFileB := tests.GetFilePath(tests.SplittedUploadSpecB)
	artifactoryCli.Exec("upload", "--spec=" + specFileA, "--build-name=" + buildName, "--build-number=" + buildNumberA)
	artifactoryCli.Exec("upload", "--spec=" + specFileB, "--build-name=" + buildName, "--build-number=" + buildNumberB)

	//publish buildInfo
	artifactoryCli.Exec("build-publish", buildName, buildNumberA)
	artifactoryCli.Exec("build-publish", buildName, buildNumberB)

	// Download by build number, a1 should not be downloaded, b1 should
	artifactoryCli.Exec("download jfrog-cli-tests-repo1/data/a1.in " + tests.Out + ioutils.GetFileSeperator() + "download" + ioutils.GetFileSeperator() + "simple_by_build" + ioutils.GetFileSeperator(), "--build=" + buildName)
	artifactoryCli.Exec("download jfrog-cli-tests-repo1/data/b1.in " +  tests.Out + ioutils.GetFileSeperator() + "download" + ioutils.GetFileSeperator() + "simple_by_build" + ioutils.GetFileSeperator(), "--build=" + buildName)

	//validate files are downloaded by build number
	paths, _ := ioutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out, false)
	tests.IsLocalExactAsExpected(tests.BuildSimpleDownload, paths, t)

	//cleanup
	deleteBuild(buildName)
	cleanArtifactoryTest()
}

func TestArtifactoryCopyByBuildUsingSpec(t *testing.T) {
	initArtifactoryTest(t)
	buildName, buildNumberA, buildNumberB := "cli-test-build", "10", "11"
	specFile := tests.GetFilePath(tests.CopyByBuildSpec)


	// Upload files with buildName and buildNumber: a* uploaded with build number "10", b* uploaded with build number "11"
	specFileA := tests.GetFilePath(tests.SplittedUploadSpecA)
	specFileB := tests.GetFilePath(tests.SplittedUploadSpecB)
	artifactoryCli.Exec("upload", "--spec=" + specFileA, "--build-name=" + buildName, "--build-number=" + buildNumberA)
	artifactoryCli.Exec("upload", "--spec=" + specFileB, "--build-name=" + buildName, "--build-number=" + buildNumberB)

	//publish buildInfo
	artifactoryCli.Exec("build-publish", buildName, buildNumberA)
	artifactoryCli.Exec("build-publish", buildName, buildNumberB)

	// Copy by build build number "10" from spec, a* should be copied
	artifactoryCli.Exec("copy", "--spec=" + specFile)

	//validate files are Copied by build number
	isExistInArtifactory(tests.BuildCopyExpected, tests.GetFilePath(tests.CpMvDlByBuildAssertSpec), t)

	//cleanup
	deleteBuild(buildName)
	cleanArtifactoryTest()
}

func TestArtifactoryCopyByBuildOverridingByInlineFlag(t *testing.T) {
	initArtifactoryTest(t)
	buildName, buildNumberA, buildNumberB := "cli-test-build", "10", "11"
	specFile := tests.GetFilePath(tests.CopyByBuildSpec)


	// Upload files with buildName and buildNumber: b* uploaded with build number "10", a* uploaded with build number "11"
	specFileA := tests.GetFilePath(tests.SplittedUploadSpecA)
	specFileB := tests.GetFilePath(tests.SplittedUploadSpecB)
	artifactoryCli.Exec("upload", "--spec=" + specFileB, "--build-name=" + buildName, "--build-number=" + buildNumberA)
	artifactoryCli.Exec("upload", "--spec=" + specFileA, "--build-name=" + buildName, "--build-number=" + buildNumberB)

	//publish buildInfo
	artifactoryCli.Exec("build-publish", buildName, buildNumberA)
	artifactoryCli.Exec("build-publish", buildName, buildNumberB)

	// Copy by build number: using override of build by flag from inline (no number set so LATEST build should be copied), a* should be copied
	artifactoryCli.Exec("copy", "--build=" + buildName + " --spec=" + specFile)

	//validate files are Copied by build number
	isExistInArtifactory(tests.BuildCopyExpected, tests.GetFilePath(tests.CpMvDlByBuildAssertSpec), t)

	//cleanup
	deleteBuild(buildName)
	cleanArtifactoryTest()
}

func TestArtifactoryMoveByBuildUsingFlags(t *testing.T) {
	initArtifactoryTest(t)
	buildName, buildNumberA, buildNumberB := "cli-test-build", "10", "11"
	specFile := tests.GetFilePath(tests.CopyByBuildSpec)


	//upload files with buildName and buildNumber
	specFileA := tests.GetFilePath(tests.SplittedUploadSpecA)
	specFileB := tests.GetFilePath(tests.SplittedUploadSpecB)
	artifactoryCli.Exec("upload", "--spec=" + specFileB, "--build-name=" + buildName, "--build-number=" + buildNumberA)
	artifactoryCli.Exec("upload", "--spec=" + specFileA, "--build-name=" + buildName, "--build-number=" + buildNumberB)

	//publish buildInfo
	artifactoryCli.Exec("build-publish", buildName, buildNumberA)
	artifactoryCli.Exec("build-publish", buildName, buildNumberB)

	// Move by build name and number
	artifactoryCli.Exec("move", "--build=" + buildName + "/11 --spec=" + specFile)

	//validate files are moved by build number
	isExistInArtifactory(tests.BuildMoveExpected, tests.GetFilePath(tests.CpMvDlByBuildAssertSpec), t)

	//cleanup
	deleteBuild(buildName)
	cleanArtifactoryTest()
}

func TestArtifactoryDeleteByLatestBuild(t *testing.T) {
	initArtifactoryTest(t)
	buildName, buildNumberA, buildNumberB := "cli-test-build", "10", "11"
	specFile := tests.GetFilePath(tests.CopyByBuildSpec)


	//upload files with buildName and buildNumber
	specFileA := tests.GetFilePath(tests.SplittedUploadSpecA)
	specFileB := tests.GetFilePath(tests.SplittedUploadSpecB)
	artifactoryCli.Exec("upload", "--spec=" + specFileB, "--build-name=" + buildName, "--build-number=" + buildNumberA)
	artifactoryCli.Exec("upload", "--spec=" + specFileA, "--build-name=" + buildName, "--build-number=" + buildNumberB)

	//publish buildInfo
	artifactoryCli.Exec("build-publish", buildName, buildNumberA)
	artifactoryCli.Exec("build-publish", buildName, buildNumberB)

	// Delete by build name and LATEST
	artifactoryCli.Exec("delete", "--build=" + buildName + "/LATEST --quiet=true --spec=" + specFile)

	//validate files are deleted by build number
	isExistInArtifactory(tests.BuildDeleteExpected, tests.GetFilePath(tests.CpMvDlByBuildAssertSpec), t)

	//cleanup
	deleteBuild(buildName)
	cleanArtifactoryTest()
}

func TestArtifactoryCleanBuildInfo(t *testing.T) {
	initArtifactoryTest(t)
	buildName, buildNumber := "cli-test-build", "11"

	//upload files with buildName and buildNumber
	specFile := tests.GetFilePath(tests.UploadSpec)
	artifactoryCli.Exec("upload", "--spec=" + specFile, "--build-name=" + buildName, "--build-number=" + buildNumber)

	//cleanup buildInfo
	artifactoryCli.WithSuffix("").Exec("build-clean", buildName, buildNumber)

	//upload files with buildName and buildNumber
	specFile = tests.GetFilePath(tests.SimpleUploadSpec)
	artifactoryCli.Exec("upload", "--spec=" + specFile, "--build-name=" + buildName, "--build-number=" + buildNumber)

	//publish buildInfo
	artifactoryCli.Exec("build-publish", buildName, buildNumber)

	//promote buildInfo
	artifactoryCli.Exec("build-promote", buildName, buildNumber, tests.Repo2)

	//validate files are uploaded with the build info name and number
	props := fmt.Sprintf("build.name=%v;build.number=%v", buildName, buildNumber)
	isExistInArtifactoryByProps(tests.SimpleUploadExpectedRepo2, tests.Repo2 + "/*", props, t)

	//cleanup
	deleteBuild(buildName)
	cleanArtifactoryTest()
}

func TestCollectGitBuildInfo(t *testing.T) {
	initArtifactoryTest(t)
	buildName, buildNumber := "cli-test-build", "13"
	dotGitPath := getCliDotGitPath(t)
	artifactoryCli.Exec("build-collect-git-info", buildName, buildNumber, dotGitPath)

	//publish buildInfo
	artifactoryCli.Exec("build-publish", buildName, buildNumber)

	artHttpDetails := utils.GetArtifactoryHttpClientDetails(artifactoryDetails)
	_, body, _, err := ioutils.SendGet(*tests.RtUrl + "api/build/" + buildName + "/" + buildNumber, false, artHttpDetails)
	if err != nil {
		t.Error(err)
	}
	buildInfoVcsRevision, err := jsonparser.GetString(body, "buildInfo", "vcsRevision")
	if err != nil {
		t.Error(err)
	}
	buildInfoVcsUrl, err := jsonparser.GetString(body, "buildInfo", "vcsUrl")
	if err != nil {
		t.Error(err)
	}
	if buildInfoVcsRevision == "" {
		t.Error("failed to get git revision.")
	}

	if buildInfoVcsUrl == "" {
		t.Error("failed to get git remote url.")
	}

	gitManager := commands.NewGitManager(dotGitPath)
	if err = gitManager.ReadGitConfig(); err != nil {
		t.Error("failed read .git config file.")
	}
	if gitManager.GetRevision() != buildInfoVcsRevision {
		t.Error("Wrong revision", "expected: " + gitManager.GetRevision(), "Got: " + buildInfoVcsRevision)
	}

	gitConfigUrl := gitManager.GetUrl() + ".git"
	if gitConfigUrl != buildInfoVcsUrl {
		t.Error("Wrong url", "expected: " + gitConfigUrl, "Got: " + buildInfoVcsUrl)
	}

	deleteBuild(buildName)
	cleanArtifactoryTest()
}

func TestReadGitConfig(t *testing.T) {
	dotGitPath := getCliDotGitPath(t)
	gitManager := commands.NewGitManager(dotGitPath)
	err := gitManager.ReadGitConfig()
	if err != nil {
		t.Error("failed read .git config file.")
	}

	workingDir, err := os.Getwd()
	if err != nil {
		t.Error("failed to get current dir.")
	}
	gitExecutor := tests.GitExecutor(workingDir)
	revision, _, err := gitExecutor.GetRevision()
	if err != nil {
		t.Error(err)
		return
	}

	if gitManager.GetRevision() != revision {
		t.Error("Wrong revision", "expected: " + revision, "Got: " + gitManager.GetRevision())
	}

	url, _, err := gitExecutor.GetUrl()
	if err != nil {
		t.Error(err)
		return
	}

	if gitManager.GetUrl() != url {
		t.Error("Wrong revision", "expected: " + url, "Got: " + gitManager.GetUrl())
	}
}

func CleanArtifactoryTests() {
	cleanArtifactoryTest()
}

func initArtifactoryTest(t *testing.T) {
	if !*tests.TestArtifactory {
		t.Skip("Artifactory is not beeing tested, skipping...")
	}
}

func cleanArtifactoryTest() {
	if !*tests.TestArtifactory {
		return
	}
	cleanArtifactory()
	tests.CleanFileSystem()
}

func prepUploadFiles() {
	uploadPath := tests.FixWinPath(tests.GetTestResourcesPath()) + "(.*)"
	targetPath := tests.Repo1 + "/downloadTestResources/{1}"
	flags := "--threads=10 --regexp=true --props=searchMe=true --flat=false"
	artifactoryCli.Exec("upload", uploadPath, targetPath, flags)
}

func prepCopyFiles() {
	specFile := tests.GetFilePath(tests.PrepareCopy)
	artifactoryCli.Exec("copy", "--spec=" + specFile)
}

func getPathsToDelete(specFile string) []utils.AqlSearchResultItem {
	flags := new(commands.DeleteFlags)
	flags.ArtDetails = artifactoryDetails
	deleteSpec, _ := utils.CreateSpecFromFile(specFile)
	artifactsToDelete, _ := commands.GetPathsToDelete(deleteSpec, flags)
	return artifactsToDelete
}

func deleteBuild(buildName string) {
	artHttpDetails := utils.GetArtifactoryHttpClientDetails(artifactoryDetails)
	resp, body, err := ioutils.SendDelete(*tests.RtUrl + "api/build/" + buildName + "?deleteAll=1", nil, artHttpDetails)
	if err != nil {
		log.Error(err)
	}
	if resp.StatusCode != 200 {
		log.Error(resp.Status)
		log.Error(string(body))
	}
}

func execCreateRepoRest(repoConfig, repoName string) error {
	content, err := ioutil.ReadFile(repoConfig)
	if err != nil {
		return err
	}
	artHttpDetails := utils.GetArtifactoryHttpClientDetails(artifactoryDetails)
	artHttpDetails.Headers = map[string]string{"Content-Type": "application/json"}
	resp, _, err := ioutils.SendPut(*tests.RtUrl + "api/repositories/" + repoName, content, artHttpDetails)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return errors.New("Fail to create repository. Reason local repository with key: " + repoName + " already exist\n")
	}
	log.Info("Repository", repoName, "was created.")
	return nil
}

func createReposIfNeeded() error {
	var err error
	var repoConfig string
	if !isRepoExist(tests.Repo1) {
		repoConfig = tests.GetTestResourcesPath() + tests.SpecsTestRepositoryConfig
		err = execCreateRepoRest(repoConfig, tests.Repo1)
		if err != nil {
			return err
		}
	}

	if !isRepoExist(tests.Repo2) {
		repoConfig = tests.GetTestResourcesPath() + tests.MoveRepositoryConfig
		err = execCreateRepoRest(repoConfig, tests.Repo2)
		if err != nil {
			return err
		}
	}
	return nil
}

func cleanArtifactory() {
	deleteFlags := new(commands.DeleteFlags)
	deleteFlags.ArtDetails = artifactoryDetails
	deleteSpec, _ := utils.CreateSpecFromFile(tests.GetFilePath(tests.DeleteSpec))
	commands.Delete(deleteSpec, deleteFlags)
}

func searchInArtifactory(specFile string) (result []commands.SearchResult, err error) {
	searchFlags := new(commands.SearchFlags)
	searchFlags.ArtDetails = artifactoryDetails
	searchSpec, _ := utils.CreateSpecFromFile(specFile)
	result, err = commands.Search(searchSpec, searchFlags)
	return
}

func isExistInArtifactory(expected []string, specFile string, t *testing.T) {
	results, _ := searchInArtifactory(specFile)
	if *tests.PrintSearchResult {
		for _, v := range results {
			fmt.Print("\"")
			fmt.Print(v.Path)
			fmt.Print("\"")
			fmt.Print(",")
			fmt.Println("")
		}
	}
	tests.CompareExpectedVsActuals(expected, results, t)
}

func isExistInArtifactoryByProps(expected []string, pattern, props string, t *testing.T) {
	searchFlags := new(commands.SearchFlags)
	searchFlags.ArtDetails = artifactoryDetails
	searchSpec := utils.CreateSpec(pattern, "", props, "", true, false, false)
	results, err := commands.Search(searchSpec, searchFlags)
	if err != nil {
		t.Error(err)
	}
	tests.CompareExpectedVsActuals(expected, results, t)
}

func isRepoExist(repoName string) bool {
	artHttpDetails := utils.GetArtifactoryHttpClientDetails(artifactoryDetails)
	resp, _, _, err := ioutils.SendGet(*tests.RtUrl + tests.RepoDetailsUrl + repoName, true, artHttpDetails)
	if err != nil {
		os.Exit(1)
	}

	if resp.StatusCode != 400 {
		return true
	}
	return false
}

func getCliDotGitPath(t *testing.T) string {
	workingDir, err := os.Getwd()
	if err != nil {
		t.Error("failed to get current dir.")
	}
	dotGitPath := filepath.Join(workingDir, "..")
	dotGitExists, err := ioutils.IsDirExists(filepath.Join(dotGitPath, ".git"))
	if err != nil {
		t.Error(err)
	}
	if !dotGitExists {
		t.Error("can't find .git")
	}
	return dotGitPath
}