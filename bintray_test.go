package main

import (
	"encoding/json"
	"github.com/jfrog/jfrog-cli-go/utils/cliutils"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-cli-go/utils/ioutils"
	"github.com/jfrog/jfrog-cli-go/utils/tests"
	"github.com/jfrog/jfrog-client-go/httpclient"
	"github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/io/httputils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
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
	bintrayCli = tests.NewJfrogCli(execMain, "jfrog bt", "--user="+bintrayConfig.User+" --key="+bintrayConfig.Key)
}

func initBintrayOrg() {
	bintrayOrganization = bintrayConfig.User
	if *tests.BtOrg != "" {
		bintrayOrganization = *tests.BtOrg
	}
}

func TestBintrayPackages(t *testing.T) {
	initBintrayTest(t)

	packagePath := bintrayOrganization + "/" + tests.BintrayRepo1 + "/testPackage"
	bintrayCli.Exec("package-create", packagePath, "--licenses=Apache-2.0", "--vcs-url=vcs.url.com")
	bintrayCli.Exec("package-show", packagePath)
	bintrayCli.Exec("package-update", packagePath, "--licenses=GPL-3.0", "--vcs-url=other.url.com")
	bintrayCli.Exec("package-delete", packagePath, "--quiet=true")

	cleanBintrayTest()
}

func TestBintrayVersions(t *testing.T) {
	initBintrayTest(t)

	packagePath := bintrayOrganization + "/" + tests.BintrayRepo1 + "/testPackage"
	versionPath := packagePath + "/1.0"

	bintrayCli.Exec("package-create", packagePath, "--licenses=Apache-2.0", "--vcs-url=vcs.url.com")
	bintrayCli.Exec("version-create", versionPath, "--desc=versionDescription", "--vcs-tag=vcs.tag")
	bintrayCli.Exec("version-show", versionPath)
	bintrayCli.Exec("version-update", versionPath, "--desc=newVersionDescription", "--vcs-tag=new.vcs.tag")
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
	path := "some/path/in/bintray/"
	uploadFilePath := tests.GetFilePathForBintray(fileName, tests.GetTestResourcesPath(), "a")
	bintrayCli.Exec("upload", uploadFilePath, versionPath, path)

	//Check file uploaded
	expected := []tests.PackageSearchResultItem{{
		Repo:    tests.BintrayRepo1,
		Path:    path + fileName,
		Package: packageName,
		Name:    fileName,
		Version: "1.0",
		Sha1:    "507ac63c6b0f650fb6f36b5621e70ebca3b0965c"}}
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
	bintrayCli.Exec("package-create", packagePath, "--licenses=Apache-2.0", "--vcs-url=vcs.url.com")

	//Upload file
	fileName := "a1.in"
	uploadFilePath := tests.GetFilePathForBintray(fileName, tests.GetTestResourcesPath(), "a")
	bintrayCli.Exec("upload", uploadFilePath, versionPath)

	//Check file uploaded
	expected := []tests.PackageSearchResultItem{{
		Repo:    tests.BintrayRepo1,
		Path:    fileName,
		Package: packageName,
		Name:    fileName,
		Version: "1.0",
		Sha1:    "507ac63c6b0f650fb6f36b5621e70ebca3b0965c"}}
	assertPackageFiles(expected, getPackageFiles(packageName), t)

	bintrayCli.Exec("package-delete", packagePath, "--quiet=true")
	cleanBintrayTest()
}

func TestBintrayUploadFromHomeDir(t *testing.T) {
	initBintrayTest(t)
	filename := "cliTestFile.txt"
	testFileRel := filepath.ToSlash(fileutils.GetHomeDir()) + "/cliTestFile.*"
	testFileAbs := fileutils.GetHomeDir() + fileutils.GetFileSeparator() + filename

	d1 := []byte("test file")
	err := ioutil.WriteFile(testFileAbs, d1, 0644)
	if err != nil {
		t.Error("Couldn't create file:", err)
	}

	packageName := "simpleUploadHomePackage"
	packagePath := bintrayOrganization + "/" + tests.BintrayRepo1 + "/" + packageName
	versionName := "1.0"
	versionPath := packagePath + "/" + versionName
	bintrayCli.Exec("package-create", packagePath, "--licenses=Apache-2.0", "--vcs-url=vcs.url.com")
	bintrayCli.Exec("upload", testFileRel, versionPath, "--recursive=false")

	//Check file uploaded
	expected := []tests.PackageSearchResultItem{{
		Repo:    tests.BintrayRepo1,
		Path:    filename,
		Package: packageName,
		Name:    filename,
		Version: "1.0",
		Sha1:    "8f93542443e98f41fe98e97d6d2a147193b1b005"}}
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

	testResourcePath := tests.GetTestResourcesPath()
	path := tests.GetFilePathForBintray("*", testResourcePath, "a")
	bintrayCli.Exec("upload", path, versionPath, "--flat=true", "--recursive=false", "--publish=true")
	assertPackageFiles(tests.BintrayExpectedUploadFlatNonRecursive, getPackageFiles(tests.BintrayUploadTestPackageName), t)
	path = tests.GetFilePathForBintray("b1.in", testResourcePath, "a", "b")
	bintrayCli.Exec("upload", path, versionPath, "a1.in", "--flat=true", "--recursive=false", "--override=true")
	assertPackageFiles(tests.BintrayExpectedUploadFlatNonRecursiveModified, getPackageFiles(tests.BintrayUploadTestPackageName), t)

	bintrayCli.Exec("package-delete", packagePath, "--quiet=true")
	cleanBintrayTest()
}

func TestBintrayUploads(t *testing.T) {
	initBintrayTest(t)

	path := tests.GetFilePathForBintray("*", "", "a")

	bintrayExpectedUploadNonFlatRecursive := tests.BintrayExpectedUploadNonFlatRecursive
	bintrayExpectedUploadNonFlatNonRecursive := tests.BintrayExpectedUploadNonFlatNonRecursive
	for i := range bintrayExpectedUploadNonFlatRecursive {
		if strings.HasPrefix(bintrayExpectedUploadNonFlatRecursive[i].Path, "/") {
			bintrayExpectedUploadNonFlatRecursive[i].Path = bintrayExpectedUploadNonFlatRecursive[i].Path[1:]
		}
	}

	for i := range bintrayExpectedUploadNonFlatNonRecursive {
		if strings.HasPrefix(bintrayExpectedUploadNonFlatNonRecursive[i].Path, "/") {
			bintrayExpectedUploadNonFlatNonRecursive[i].Path = bintrayExpectedUploadNonFlatNonRecursive[i].Path[1:]
		}
	}

	testBintrayUpload(t, path, "--flat=true --recursive=true", tests.BintrayExpectedUploadFlatRecursive)
	testBintrayUpload(t, path, "--flat=true --recursive=false", tests.BintrayExpectedUploadFlatNonRecursive)
	testBintrayUpload(t, path, "--flat=false --recursive=true", bintrayExpectedUploadNonFlatRecursive)
	testBintrayUpload(t, path, "--flat=false --recursive=false", bintrayExpectedUploadNonFlatNonRecursive)

	path = tests.GetFilePathForBintray("(.*)", "", "a")
	testBintrayUpload(t, path, "--flat=true --recursive=true --regexp=true", tests.BintrayExpectedUploadFlatRecursive)
	testBintrayUpload(t, path, "--flat=true --recursive=false --regexp=true", tests.BintrayExpectedUploadFlatNonRecursive)
	testBintrayUpload(t, path, "--flat=false --recursive=true --regexp=true", bintrayExpectedUploadNonFlatRecursive)
	testBintrayUpload(t, path, "--flat=false --recursive=false --regexp=true", bintrayExpectedUploadNonFlatNonRecursive)

	cleanBintrayTest()
}

func TestBintrayLogs(t *testing.T) {
	initBintrayTest(t)

	packagePath := bintrayOrganization + "/" + tests.BintrayRepo1 + "/" + tests.BintrayUploadTestPackageName
	versionPath := packagePath + "/" + tests.BintrayUploadTestVersion
	createPackageAndVersion(packagePath, versionPath)

	path := tests.GetTestResourcesPath() + "a/*"
	bintrayCli.Exec("upload", path, versionPath, "--flat=true --recursive=true --publish=true")
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
	path := tests.GetFilePathForBintray("*", tests.GetTestResourcesPath(), "a")
	bintrayCli.Exec("upload", path, versionPath, "--flat=true --recursive=true")
	path = tests.GetFilePathForBintray("a1.in", tests.GetTestResourcesPath(), "a")
	bintrayCli.Exec("upload", path, versionPath, "--flat=false")

	// Define executor for testing with retries
	var args []string
	retryExecutor := utils.RetryExecutor{
		MaxRetries:      25,
		RetriesInterval: 5,
		ErrorMessage:    "Waiting for bintray to index files...",
		ExecutionHandler: func() (bool, error) {
			// Execute Bintray downloads
			err := bintrayCli.Exec(args...)
			if err != nil {
				return true, err
			}
			return false, nil
		},
	}

	// File a1.in
	args = []string{"download-file", repositoryPath + "/a1.in", tests.Out + "/bintray/", "--unpublished=true"}
	if err := retryExecutor.Execute(); err != nil {
		t.Error(err.Error())
	}

	// File b1.in
	args = []string{"download-file", repositoryPath + "/b1.in", tests.Out + "/bintray/x.in", "--unpublished=true"}
	if err := retryExecutor.Execute(); err != nil {
		t.Error(err.Error())
	}

	// File c1.in
	args = []string{"download-file", repositoryPath + "/(c)1.in", tests.Out + "/bintray/z{1}.in", "--unpublished=true"}
	if err := retryExecutor.Execute(); err != nil {
		t.Error(err.Error())
	}

	// File a/a1.in
	args = []string{"download-file", repositoryPath + "/" + tests.GetTestResourcesPath() + "(a)/a1.in", tests.Out + "/bintray/{1}/fullpatha1.in", "--flat=true --unpublished=true"}
	if err := retryExecutor.Execute(); err != nil {
		t.Error(err.Error())
	}

	//Validate that files were downloaded as expected
	expected := []string{
		filepath.Join(tests.Out, "bintray", "a1.in"),
		filepath.Join(tests.Out, "bintray", "x.in"),
		filepath.Join(tests.Out, "bintray", "zc.in"),
		filepath.Join(tests.Out, "bintray", "a", "fullpatha1.in"),
	}
	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out+"/bintray/", false)
	tests.IsExistLocally(expected, paths, t)

	// Cleanup
	bintrayCli.Exec("package-delete", packagePath, "--quiet=true")
	cleanBintrayTest()
}

func TestBintrayVersionDownloads(t *testing.T) {
	initBintrayTest(t)

	repositoryPath := bintrayOrganization + "/" + tests.BintrayRepo1
	packagePath := repositoryPath + "/" + tests.BintrayUploadTestPackageName
	versionPath := packagePath + "/" + tests.BintrayUploadTestVersion
	createPackageAndVersion(packagePath, versionPath)

	path := tests.GetFilePathForBintray("*", tests.GetTestResourcesPath(), "a")
	bintrayCli.Exec("upload", path, versionPath, "--flat=true --recursive=true")
	bintrayCli.Exec("download-ver", versionPath, tests.Out+"/bintray/", "--unpublished=true")

	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out+"/bintray/", false)
	expected := []string{
		filepath.Join(tests.Out, "bintray", "a1.in"),
		filepath.Join(tests.Out, "bintray", "a2.in"),
		filepath.Join(tests.Out, "bintray", "a3.in"),
		filepath.Join(tests.Out, "bintray", "b1.in"),
		filepath.Join(tests.Out, "bintray", "b2.in"),
		filepath.Join(tests.Out, "bintray", "b3.in"),
		filepath.Join(tests.Out, "bintray", "c1.in"),
		filepath.Join(tests.Out, "bintray", "c2.in"),
		filepath.Join(tests.Out, "bintray", "c3.in"),
	}
	tests.IsExistLocally(expected, paths, t)
	bintrayCli.Exec("package-delete", packagePath, "--quiet=true")
	cleanBintrayTest()
}

// Tests compatibility to file paths with windows separators.
func TestBintrayUploadWindowsCompatibility(t *testing.T) {
	if !cliutils.IsWindows() {
		return
	}
	initBintrayTest(t)

	packageName := "simpleUploadPackage"
	packagePath := bintrayOrganization + "/" + tests.BintrayRepo1 + "/" + packageName
	versionName := "1.0"
	versionPath := packagePath + "/" + versionName

	createPackageAndVersion(packagePath, versionPath)
	//Upload file
	fileName := "a1.in"
	path := "some/path/in/bintray/"
	uploadFilePath := ioutils.UnixToWinPathSeparator(tests.GetFilePathForBintray(fileName, tests.GetTestResourcesPath(), "a"))
	bintrayCli.Exec("upload", uploadFilePath, versionPath, path)

	//Check file uploaded
	expected := []tests.PackageSearchResultItem{{
		Repo:    tests.BintrayRepo1,
		Path:    path + fileName,
		Package: packageName,
		Name:    fileName,
		Version: "1.0",
		Sha1:    "507ac63c6b0f650fb6f36b5621e70ebca3b0965c"}}
	assertPackageFiles(expected, getPackageFiles(packageName), t)

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
		t.Skip("Bintray is not being tested, skipping...")
	}
}

func initBintrayCredentials() {
	if bintrayConfig != nil {
		return
	}

	var err error
	bintrayConfig, err = config.ReadBintrayConf()
	if errorutils.CheckError(err) != nil {
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

	apiUrl = utils.AddTrailingSlashIfNeeded(apiUrl)
	downloadServerUrl = utils.AddTrailingSlashIfNeeded(downloadServerUrl)

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

	bintrayCli.Exec("upload", tests.GetTestResourcesPath()+relPath, versionPath, flags)
	assertPackageFiles(expected, getPackageFiles(tests.BintrayUploadTestPackageName), t)
	bintrayCli.Exec("package-delete", packagePath, "--quiet=true")
}

func getPackageFiles(packageName string) []tests.PackageSearchResultItem {
	apiUrl := bintrayConfig.ApiUrl + path.Join("packages", bintrayOrganization, tests.BintrayRepo1, packageName, "files?include_unpublished=1")
	clientDetails := httputils.HttpClientDetails{
		User:     bintrayConfig.User,
		Password: bintrayConfig.Key,
		Headers:  map[string]string{"Content-Type": "application/json"}}

	client, err := httpclient.ClientBuilder().Build()
	if err != nil {
		os.Exit(1)
	}
	resp, body, _, err := client.SendGet(apiUrl, true, clientDetails)
	if errorutils.CheckError(err) != nil {
		os.Exit(1)
	}
	if resp.StatusCode != http.StatusOK {
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
	if errorutils.CheckError(err) != nil {
		os.Exit(1)
	}

	apiUrl := bintrayConfig.ApiUrl + path.Join("repos", bintrayOrganization, tests.BintrayRepo1)
	clientDetails := httputils.HttpClientDetails{
		User:     bintrayConfig.User,
		Password: bintrayConfig.Key,
		Headers:  map[string]string{"Content-Type": "application/json"}}

	client, err := httpclient.ClientBuilder().Build()
	if err != nil {
		os.Exit(1)
	}
	resp, body, err := client.SendPost(apiUrl, content, clientDetails)
	if errorutils.CheckError(err) != nil {
		os.Exit(1)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusConflict {
		log.Error(resp.Status)
		log.Error(string(body))
		os.Exit(1)
	}
}

func deleteBintrayRepo() {
	apiUrl := bintrayConfig.ApiUrl + path.Join("repos", bintrayOrganization, tests.BintrayRepo1)
	clientDetails := httputils.HttpClientDetails{
		User:     bintrayConfig.User,
		Password: bintrayConfig.Key,
		Headers:  map[string]string{"Content-Type": "application/json"}}

	client, err := httpclient.ClientBuilder().Build()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	resp, body, err := client.SendDelete(apiUrl, nil, clientDetails)
	if errorutils.CheckError(err) != nil {
		log.Error(err)
		os.Exit(1)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		log.Error(resp.Status)
		log.Error(string(body))
		os.Exit(1)
	}
}
