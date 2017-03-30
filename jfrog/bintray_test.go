package main

import (
	"testing"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"os"
	"github.com/jfrogdev/jfrog-cli-go/utils/tests"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/log"
	"github.com/jfrogdev/jfrog-cli-go/utils/io/httputils"
	"github.com/jfrogdev/jfrog-cli-go/utils/io/fileutils"
	"path"
	"io/ioutil"
	"encoding/json"
	"strconv"
)

var bintrayConfig *config.BintrayDetails
var bintrayCli *tests.JfrogCli
var bintrayOrganization string

func InitBintrayTests() {
	if !*tests.TestBintray {
		return
	}
	initBintrayCredentials()
	initBintrayOrg()
	deleteBintrayRepo()
	createBintrayRepo()
	bintrayCli = tests.NewJfrogCli(main, "jfrog bt", "--user=" + bintrayConfig.User + " --key=" + bintrayConfig.Key)
}

func initBintrayOrg() {
	bintrayOrganization = bintrayConfig.User
	if *tests.BtOrganization != "" {
		bintrayOrganization = *tests.BtOrganization
	}
}

func TestBintrayPackages(t *testing.T) {
	initBintrayTest(t)

	packagePath := bintrayOrganization + "/" + tests.BintrayRepo1 + "/testPackage"
	bintrayCli.Exec("package-create", packagePath, "--licenses=Apache-2.0 --vcs-url=vcs.url.com")
	bintrayCli.Exec("package-show", packagePath)
	bintrayCli.Exec("package-update", packagePath, "--licenses=GPL-3.0 --vcs-url=other.url.com")
	bintrayCli.Exec("package-delete", packagePath, "--quiet=true")

	cleanBintrayTest()
}

func TestBintrayVersions(t *testing.T) {
	initBintrayTest(t)

	packagePath := bintrayOrganization + "/" + tests.BintrayRepo1 + "/testPackage"
	versionPath := packagePath + "/1.0"

	bintrayCli.Exec("package-create", packagePath, "--licenses=Apache-2.0 --vcs-url=vcs.url.com")
	bintrayCli.Exec("version-create", versionPath, "--desc=versionDescription --vcs-tag=vcs.tag")
	bintrayCli.Exec("version-show", versionPath)
	bintrayCli.Exec("version-update", versionPath, "--desc=newVersionDescription --vcs-tag=new.vcs.tag")
	bintrayCli.Exec("version-publish", versionPath)
	bintrayCli.Exec("version-delete", versionPath, "--quiet=true")
	bintrayCli.Exec("package-delete", packagePath, "--quiet=true")

	cleanBintrayTest()
}

func TestBintraySimpleUpload(t *testing.T) {
	initBintrayTest(t)

	packageName := "simpleUploadPackage"
	packagePath := bintrayOrganization + "/" + tests.BintrayRepo1 + "/" + packageName
	versionName := "1.0"
	versionPath := packagePath + "/" + versionName

	createPackageAndVersion(packagePath, versionPath)
	//Upload file
	fileName := "a1.in"
	uploadFilePath := tests.GetTestResourcesPath() + "a/" + fileName
	bintrayCli.Exec("upload", uploadFilePath, versionPath)

	//Check file uploaded
	expected := []tests.PackageSearchResultItem{{
		Repo:       tests.BintrayRepo1,
		Path:       fileName,
		Package:    packageName,
		Name:       fileName,
		Version:    "1.0",
		Sha1:       "507ac63c6b0f650fb6f36b5621e70ebca3b0965c"}}
	assertPackageFiles(expected, getPackageFiles(packageName), t)

	bintrayCli.Exec("package-delete", packagePath, "--quiet=true")
	cleanBintrayTest()
}

func TestBintrayUploadNoVersion(t *testing.T) {
	initBintrayTest(t)

	packageName := "simpleUploadPackage"
	packagePath := bintrayOrganization + "/" + tests.BintrayRepo1 + "/" + packageName
	versionName := "1.0"
	versionPath := packagePath + "/" + versionName
	bintrayCli.Exec("package-create", packagePath, "--licenses=Apache-2.0 --vcs-url=vcs.url.com")

	//Upload file
	fileName := "a1.in"
	uploadFilePath := tests.GetTestResourcesPath() + "a/" + fileName
	bintrayCli.Exec("upload", uploadFilePath, versionPath)

	//Check file uploaded
	expected := []tests.PackageSearchResultItem{{
		Repo:       tests.BintrayRepo1,
		Path:       fileName,
		Package:    packageName,
		Name:       fileName,
		Version:    "1.0",
		Sha1:       "507ac63c6b0f650fb6f36b5621e70ebca3b0965c"}}
	assertPackageFiles(expected, getPackageFiles(packageName), t)

	bintrayCli.Exec("package-delete", packagePath, "--quiet=true")
	cleanBintrayTest()
}

func TestBintrayUploadFromHomeDir(t *testing.T) {
	initBintrayTest(t)
	filename := "cliTestFile.txt"
	testFileRel := "~/cliTestFile.*"
	testFileAbs := fileutils.GetHomeDir() + "/" + filename

	d1 := []byte("test file")
	err := ioutil.WriteFile(testFileAbs, d1, 0644)
	if err != nil {
		t.Error("Coudln't create file:", err)
	}

	packageName := "simpleUploadHomePackage"
	packagePath := bintrayOrganization + "/" + tests.BintrayRepo1 + "/" + packageName
	versionName := "1.0"
	versionPath := packagePath + "/" + versionName
	bintrayCli.Exec("package-create", packagePath, "--licenses=Apache-2.0 --vcs-url=vcs.url.com")
	bintrayCli.Exec("upload", testFileRel, versionPath, "--recursive=false")

	//Check file uploaded
	expected := []tests.PackageSearchResultItem{{
		Repo:       tests.BintrayRepo1,
		Path:       filename,
		Package:    packageName,
		Name:       filename,
		Version:    "1.0",
		Sha1:       "8f93542443e98f41fe98e97d6d2a147193b1b005"}}
	assertPackageFiles(expected, getPackageFiles(packageName), t)

	os.Remove(testFileAbs)
	bintrayCli.Exec("package-delete", packagePath, "--quiet=true")
	cleanBintrayTest()
}

func TestBintrayUploadOverride(t *testing.T) {
	initBintrayTest(t)

	packagePath := bintrayOrganization + "/" + tests.BintrayRepo1 + "/" + tests.BintrayUploadTestPackageName
	versionPath := packagePath + "/" + tests.BintrayUploadTestVersion
	createPackageAndVersion(packagePath, versionPath)

	bintrayCli.Exec("upload", tests.GetTestResourcesPath() + "a/*", versionPath, "--flat=true --recursive=false --publish=true")
	assertPackageFiles(tests.BintrayExpectedUploadFlatNonRecursive, getPackageFiles(tests.BintrayUploadTestPackageName), t)
	bintrayCli.Exec("upload", tests.GetTestResourcesPath() + "a/b/b1.in", versionPath, "a1.in", "--flat=true --recursive=false --override=true")
	assertPackageFiles(tests.BintrayExpectedUploadFlatNonRecursiveModified, getPackageFiles(tests.BintrayUploadTestPackageName), t)

	bintrayCli.Exec("package-delete", packagePath, "--quiet=true")
	cleanBintrayTest()
}

func TestBintrayUploads(t *testing.T) {
	initBintrayTest(t)

	testBintrayUpload(t, "a/*", "--flat=true --recursive=true", tests.BintrayExpectedUploadFlatRecursive)
	testBintrayUpload(t, "a/*", "--flat=true --recursive=false", tests.BintrayExpectedUploadFlatNonRecursive)
	testBintrayUpload(t, "a/*", "--flat=false --recursive=true", tests.BintrayExpectedUploadNonFlatRecursive)
	testBintrayUpload(t, "a/*", "--flat=false --recursive=false", tests.BintrayExpectedUploadNonFlatNonRecursive)
	testBintrayUpload(t, "a/(.*)", "--flat=true --recursive=true --regexp=true", tests.BintrayExpectedUploadFlatRecursive)
	testBintrayUpload(t, "a/(.*)", "--flat=true --recursive=false --regexp=true", tests.BintrayExpectedUploadFlatNonRecursive)
	testBintrayUpload(t, "a/(.*)", "--flat=false --recursive=true --regexp=true", tests.BintrayExpectedUploadNonFlatRecursive)
	testBintrayUpload(t, "a/(.*)", "--flat=false --recursive=false --regexp=true", tests.BintrayExpectedUploadNonFlatNonRecursive)

	cleanBintrayTest()
}

func TestBintrayLogs(t *testing.T) {
	initBintrayTest(t)

	packagePath := bintrayOrganization + "/" + tests.BintrayRepo1 + "/" + tests.BintrayUploadTestPackageName
	versionPath := packagePath + "/" + tests.BintrayUploadTestVersion
	createPackageAndVersion(packagePath, versionPath)

	bintrayCli.Exec("upload", tests.GetTestResourcesPath() + "a/*", versionPath, "--flat=true --recursive=true --publish=true")
	assertPackageFiles(tests.BintrayExpectedUploadFlatRecursive, getPackageFiles(tests.BintrayUploadTestPackageName), t)
	bintrayCli.Exec("logs", packagePath)

	bintrayCli.Exec("package-delete", packagePath, "--quiet=true")
	cleanBintrayTest()
}

func TestBintrayFileDownloads(t *testing.T) {
	initBintrayTest(t)

	repositoryPath := bintrayOrganization + "/" + tests.BintrayRepo1
	packagePath := repositoryPath + "/" + tests.BintrayUploadTestPackageName
	versionPath := packagePath + "/" + tests.BintrayUploadTestVersion
	createPackageAndVersion(packagePath, versionPath)
	bintrayCli.Exec("upload", tests.GetTestResourcesPath() + "a/*", versionPath, "--flat=true --recursive=true")
	bintrayCli.Exec("upload", tests.GetTestResourcesPath() + "a/a1.in", versionPath, "--flat=false")

	bintrayCli.Exec("download-file", repositoryPath + "/a1.in", tests.Out + "/bintray/", "--unpublished=true")
	bintrayCli.Exec("download-file", repositoryPath + "/b1.in", tests.Out + "/bintray/x.in", "--unpublished=true")
	bintrayCli.Exec("download-file", repositoryPath + "/(c)1.in", tests.Out + "/bintray/z{1}.in", "--unpublished=true")
	bintrayCli.Exec("download-file", repositoryPath + "/" + tests.GetTestResourcesPath()[1:] + "(a)/a1.in", tests.Out + "/bintray/{1}/fullpatha1.in", "--flat=true --unpublished=true")

	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out + "/bintray/", false)
	expected := []string{
		tests.Out + fileutils.GetFileSeperator() + "bintray" + fileutils.GetFileSeperator() + "a1.in",
		tests.Out + fileutils.GetFileSeperator() + "bintray" + fileutils.GetFileSeperator() + "x.in",
		tests.Out + fileutils.GetFileSeperator() + "bintray" + fileutils.GetFileSeperator() + "zc.in",
		tests.Out + fileutils.GetFileSeperator() + "bintray" + fileutils.GetFileSeperator() + "a" + fileutils.GetFileSeperator() + "fullpatha1.in",
	}
	tests.IsExistLocally(expected, paths, t)
	bintrayCli.Exec("package-delete", packagePath, "--quiet=true")
	cleanBintrayTest()
}

func TestBintrayVersionDownloads(t *testing.T) {
	initBintrayTest(t)

	repositoryPath := bintrayOrganization + "/" + tests.BintrayRepo1
	packagePath := repositoryPath + "/" + tests.BintrayUploadTestPackageName
	versionPath := packagePath + "/" + tests.BintrayUploadTestVersion
	createPackageAndVersion(packagePath, versionPath)

	bintrayCli.Exec("upload", tests.GetTestResourcesPath() + "a/*", versionPath, "--flat=true --recursive=true")
	bintrayCli.Exec("download-ver", versionPath, tests.Out + "/bintray/", "--unpublished=true")

	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out + "/bintray/", false)
	expected := []string{
		tests.Out + fileutils.GetFileSeperator() + "bintray" + fileutils.GetFileSeperator() + "a1.in",
		tests.Out + fileutils.GetFileSeperator() + "bintray" + fileutils.GetFileSeperator() + "a2.in",
		tests.Out + fileutils.GetFileSeperator() + "bintray" + fileutils.GetFileSeperator() + "a3.in",
		tests.Out + fileutils.GetFileSeperator() + "bintray" + fileutils.GetFileSeperator() + "b1.in",
		tests.Out + fileutils.GetFileSeperator() + "bintray" + fileutils.GetFileSeperator() + "b2.in",
		tests.Out + fileutils.GetFileSeperator() + "bintray" + fileutils.GetFileSeperator() + "b3.in",
		tests.Out + fileutils.GetFileSeperator() + "bintray" + fileutils.GetFileSeperator() + "c1.in",
		tests.Out + fileutils.GetFileSeperator() + "bintray" + fileutils.GetFileSeperator() + "c2.in",
		tests.Out + fileutils.GetFileSeperator() + "bintray" + fileutils.GetFileSeperator() + "c3.in",
	}
	tests.IsExistLocally(expected, paths, t)
	bintrayCli.Exec("package-delete", packagePath, "--quiet=true")
	cleanBintrayTest()
}

func CleanBintrayTests() {
	if !*tests.TestBintray {
		return
	}
	deleteBintrayRepo()
}

func initBintrayTest(t *testing.T) {
	if !*tests.TestBintray {
		t.Skip("Bintray is not beeing tested, skipping...")
	}
}

func initBintrayCredentials() {
	if bintrayConfig != nil {
		return
	}

	var err error
	bintrayConfig, err = config.ReadBintrayConf()
	if cliutils.CheckError(err) != nil {
		os.Exit(1)
	}
	if *tests.BtUser != "" {
		bintrayConfig.User = *tests.BtUser
	}
	if *tests.BtKey != "" {
		bintrayConfig.Key = *tests.BtKey
	}

	if bintrayConfig.User == "" || bintrayConfig.Key == "" {
		log.Error("To test Bintray credentials must be configured.")
		os.Exit(1)
	}

	apiUrl := os.Getenv("JFROG_CLI_BINTRAY_API_URL")
	if apiUrl == "" {
		apiUrl = "https://bintray.com/api/v1/"
	}

	downloadServerUrl := os.Getenv("JFROG_CLI_BINTRAY_DOWNLOAD_URL")
	if downloadServerUrl == "" {
		downloadServerUrl = "https://dl.bintray.com/"
	}

	apiUrl = cliutils.AddTrailingSlashIfNeeded(apiUrl)
	downloadServerUrl = cliutils.AddTrailingSlashIfNeeded(downloadServerUrl)

	bintrayConfig.ApiUrl = apiUrl
	bintrayConfig.DownloadServerUrl = downloadServerUrl
}

func cleanBintrayTest() {
	tests.CleanFileSystem()
}

func testBintrayUpload(t *testing.T, relPath, flags string, expected []tests.PackageSearchResultItem) {
	packagePath := bintrayOrganization + "/" + tests.BintrayRepo1 + "/" + tests.BintrayUploadTestPackageName
	versionPath := packagePath + "/" + tests.BintrayUploadTestVersion
	createPackageAndVersion(packagePath, versionPath)

	bintrayCli.Exec("upload", tests.GetTestResourcesPath() + relPath, versionPath, flags)
	assertPackageFiles(expected, getPackageFiles(tests.BintrayUploadTestPackageName), t)
	bintrayCli.Exec("package-delete", packagePath, "--quiet=true")
}

func getPackageFiles(packageName string) []tests.PackageSearchResultItem {
	apiUrl := bintrayConfig.ApiUrl + path.Join("packages", bintrayOrganization, tests.BintrayRepo1, packageName, "files?include_unpublished=1")
	clientDetails := httputils.HttpClientDetails{
		User:       bintrayConfig.User,
		Password:   bintrayConfig.Key,
		Headers:    map[string]string{"Content-Type": "application/json"}}

	resp, body, _, err := httputils.SendGet(apiUrl, true, clientDetails)
	if cliutils.CheckError(err) != nil {
		os.Exit(1)
	}
	if (resp.StatusCode != 200) {
		log.Error(resp.Status)
		log.Error(string(body))
		os.Exit(1)
	}

	var result []tests.PackageSearchResultItem
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	return result
}

func assertPackageFiles(expected, actual []tests.PackageSearchResultItem, t *testing.T) {
	if len(actual) != len(expected) {
		t.Error("Expected: " + strconv.Itoa(len(expected)) + ", Got: " + strconv.Itoa(len(actual)) + " files.")
	}

	expectedMap := make(map[string]tests.PackageSearchResultItem)
	for _, v := range expected {
		expectedMap[packageFileHash(v)] = v
	}

	actualMap := make(map[string]tests.PackageSearchResultItem)
	for _, v := range actual {
		actualMap[packageFileHash(v)] = v
	}

	for _, v := range actual {
		if _, ok := expectedMap[packageFileHash(v)]; !ok {
			t.Error("Unexpected file:", v)
		}
	}
	for _, v := range expected {
		if _, ok := actualMap[packageFileHash(v)]; !ok {
			t.Error("File not found:", v)
		}
	}
}

func packageFileHash(item tests.PackageSearchResultItem) string {
	return item.Repo + item.Path + item.Package + item.Version + item.Name + item.Sha1
}

func createPackageAndVersion(packagePath, versionPath string) {
	bintrayCli.Exec("package-create", packagePath, "--licenses=Apache-2.0 --vcs-url=vcs.url.com")
	bintrayCli.Exec("version-create", versionPath, "--desc=versionDescription --vcs-tag=vcs.tag")
}

func createBintrayRepo() {
	content, err := ioutil.ReadFile(tests.GetTestResourcesPath() + tests.BintrayTestRepositoryConfig)
	if cliutils.CheckError(err) != nil {
		os.Exit(1)
	}

	apiUrl := bintrayConfig.ApiUrl + path.Join("repos", bintrayOrganization, tests.BintrayRepo1)
	clientDetails := httputils.HttpClientDetails{
		User:       bintrayConfig.User,
		Password:   bintrayConfig.Key,
		Headers:    map[string]string{"Content-Type": "application/json"}}

	resp, body, err := httputils.SendPost(apiUrl, content, clientDetails)
	if cliutils.CheckError(err) != nil {
		os.Exit(1)
	}

	if (resp.StatusCode != 200 && resp.StatusCode != 201 && resp.StatusCode != 409) {
		log.Error(resp.Status)
		log.Error(string(body))
		os.Exit(1)
	}
}

func deleteBintrayRepo() {
	apiUrl := bintrayConfig.ApiUrl + path.Join("repos", bintrayOrganization, tests.BintrayRepo1)
	clientDetails := httputils.HttpClientDetails{
		User:       bintrayConfig.User,
		Password:   bintrayConfig.Key,
		Headers:    map[string]string{"Content-Type": "application/json"}}

	resp, body, err := httputils.SendDelete(apiUrl, nil, clientDetails)
	if cliutils.CheckError(err) != nil {
		os.Exit(1)
	}
	if (resp.StatusCode != 200 && resp.StatusCode != 404) {
		log.Error(resp.Status)
		log.Error(string(body))
		os.Exit(1)
	}
}


