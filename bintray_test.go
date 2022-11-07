package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jfrog/jfrog-cli-core/utils/coreutils"

	"github.com/buger/jsonparser"
	"github.com/stretchr/testify/assert"

	"github.com/jfrog/jfrog-cli-core/utils/config"
	"github.com/jfrog/jfrog-cli-core/utils/ioutils"
	"github.com/jfrog/jfrog-cli/utils/tests"
	"github.com/jfrog/jfrog-client-go/http/httpclient"
	"github.com/jfrog/jfrog-client-go/utils"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/io/httputils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

var bintrayConfig *config.BintrayDetails
var bintrayCli *tests.JfrogCli
var bintrayOrganization string

func InitBintrayTests() {
	initBintrayCredentials()
	initBintrayOrg()
	cleanUpOldBintrayRepositories()
	tests.AddTimestampToGlobalVars()
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

	packagePath := path.Join(bintrayOrganization, tests.BintrayRepo, "testPackage")
	bintrayCli.Exec("package-create", packagePath, "--licenses=Apache-2.0", "--vcs-url=vcs.url.com")
	bintrayCli.Exec("package-show", packagePath)
	bintrayCli.Exec("package-update", packagePath, "--licenses=GPL-3.0", "--vcs-url=other.url.com")
	bintrayCli.Exec("package-delete", packagePath)

	cleanBintrayTest()
}

func TestBintrayVersions(t *testing.T) {
	initBintrayTest(t)

	packagePath := path.Join(bintrayOrganization, tests.BintrayRepo, "testPackage")
	versionPath := packagePath + "/1.0"

	bintrayCli.Exec("package-create", packagePath, "--licenses=Apache-2.0", "--vcs-url=vcs.url.com")
	bintrayCli.Exec("version-create", versionPath, "--desc=versionDescription", "--vcs-tag=vcs.tag")
	bintrayCli.Exec("version-show", versionPath)
	bintrayCli.Exec("version-update", versionPath, "--desc=newVersionDescription", "--vcs-tag=new.vcs.tag")
	bintrayCli.Exec("version-publish", versionPath)
	bintrayCli.Exec("version-delete", versionPath)
	bintrayCli.Exec("package-delete", packagePath)

	cleanBintrayTest()
}

func TestBintraySimpleUpload(t *testing.T) {
	initBintrayTest(t)

	packageName := "simpleUploadPackage"
	packagePath := path.Join(bintrayOrganization, tests.BintrayRepo, packageName)
	versionName := "1.0"
	versionPath := path.Join(packagePath, versionName)

	createPackageAndVersion(packagePath, versionPath)
	// Upload file
	fileName := "a1.in"
	path := "some/path/in/bintray/"
	uploadFilePath := tests.GetFilePathForBintray(fileName, tests.GetTestResourcesPath(), "a")
	bintrayCli.Exec("upload", uploadFilePath, versionPath, path)

	// Check file uploaded
	expected := []tests.PackageSearchResultItem{{
		Repo:    tests.BintrayRepo,
		Path:    path + fileName,
		Package: packageName,
		Name:    fileName,
		Version: "1.0",
		Sha1:    "507ac63c6b0f650fb6f36b5621e70ebca3b0965c"}}
	assertPackageFiles(expected, getPackageFiles(packageName), t)

	bintrayCli.Exec("package-delete", packagePath)
	cleanBintrayTest()
}

func TestBintrayUploadNoVersion(t *testing.T) {
	initBintrayTest(t)

	packageName := "simpleUploadPackage"
	packagePath := path.Join(bintrayOrganization, tests.BintrayRepo, packageName)
	versionName := "1.0"
	versionPath := path.Join(packagePath, versionName)
	bintrayCli.Exec("package-create", packagePath, "--licenses=Apache-2.0", "--vcs-url=vcs.url.com")

	//Upload file
	fileName := "a1.in"
	uploadFilePath := tests.GetFilePathForBintray(fileName, tests.GetTestResourcesPath(), "a")
	bintrayCli.Exec("upload", uploadFilePath, versionPath)

	//Check file uploaded
	expected := []tests.PackageSearchResultItem{{
		Repo:    tests.BintrayRepo,
		Path:    fileName,
		Package: packageName,
		Name:    fileName,
		Version: "1.0",
		Sha1:    "507ac63c6b0f650fb6f36b5621e70ebca3b0965c"}}
	assertPackageFiles(expected, getPackageFiles(packageName), t)

	bintrayCli.Exec("package-delete", packagePath)
	cleanBintrayTest()
}

func TestBintrayUploadFromHomeDir(t *testing.T) {
	initBintrayTest(t)
	filename := "cliTestFile.txt"
	testFileRel := filepath.ToSlash(fileutils.GetHomeDir()) + "/cliTestFile.*"
	testFileAbs := fileutils.GetHomeDir() + fileutils.GetFileSeparator() + filename

	d1 := []byte("test file")
	assert.NoError(t, os.WriteFile(testFileAbs, d1, 0644), "Couldn't create file")

	packageName := "simpleUploadHomePackage"
	packagePath := path.Join(bintrayOrganization, tests.BintrayRepo, packageName)
	versionName := "1.0"
	versionPath := path.Join(packagePath, versionName)
	bintrayCli.Exec("package-create", packagePath, "--licenses=Apache-2.0", "--vcs-url=vcs.url.com")
	bintrayCli.Exec("upload", testFileRel, versionPath, "--recursive=false")

	//Check file uploaded
	expected := []tests.PackageSearchResultItem{{
		Repo:    tests.BintrayRepo,
		Path:    filename,
		Package: packageName,
		Name:    filename,
		Version: "1.0",
		Sha1:    "8f93542443e98f41fe98e97d6d2a147193b1b005"}}
	assertPackageFiles(expected, getPackageFiles(packageName), t)

	os.Remove(testFileAbs)
	bintrayCli.Exec("package-delete", packagePath)
	cleanBintrayTest()
}

func TestBintrayUploadOverride(t *testing.T) {
	initBintrayTest(t)

	packagePath := path.Join(bintrayOrganization, tests.BintrayRepo, tests.BintrayUploadTestPackageName)
	versionPath := path.Join(packagePath, tests.BintrayUploadTestVersion)
	createPackageAndVersion(packagePath, versionPath)

	testResourcePath := tests.GetTestResourcesPath()
	path := tests.GetFilePathForBintray("*", testResourcePath, "a")
	bintrayCli.Exec("upload", path, versionPath, "--flat=true", "--recursive=false", "--publish=true")
	assertPackageFiles(tests.GetBintrayExpectedUploadFlatNonRecursive(), getPackageFiles(tests.BintrayUploadTestPackageName), t)
	path = tests.GetFilePathForBintray("b1.in", testResourcePath, "a", "b")
	bintrayCli.Exec("upload", path, versionPath, "a1.in", "--flat=true", "--recursive=false", "--override=true")
	assertPackageFiles(tests.GetBintrayExpectedUploadFlatNonRecursiveModified(), getPackageFiles(tests.BintrayUploadTestPackageName), t)

	bintrayCli.Exec("package-delete", packagePath)
	cleanBintrayTest()
}

func TestBintrayUploads(t *testing.T) {
	initBintrayTest(t)

	path := tests.GetFilePathForBintray("*", "", "a")

	bintrayExpectedUploadNonFlatRecursive := tests.GetBintrayExpectedUploadNonFlatRecursive()
	bintrayExpectedUploadNonFlatNonRecursive := tests.GetBintrayExpectedUploadNonFlatNonRecursive()
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

	testBintrayUpload(t, path, tests.GetBintrayExpectedUploadFlatRecursive(), "--flat=true", "--recursive=true")
	testBintrayUpload(t, path, tests.GetBintrayExpectedUploadFlatNonRecursive(), "--flat=true", "--recursive=false")
	testBintrayUpload(t, path, bintrayExpectedUploadNonFlatRecursive, "--flat=false", "--recursive=true")
	testBintrayUpload(t, path, bintrayExpectedUploadNonFlatNonRecursive, "--flat=false", "--recursive=false")

	path = tests.GetFilePathForBintray("(.*)", "", "a")
	testBintrayUpload(t, path, tests.GetBintrayExpectedUploadFlatRecursive(), "--flat=true", "--recursive=true", "--regexp=true")
	testBintrayUpload(t, path, tests.GetBintrayExpectedUploadFlatNonRecursive(), "--flat=true", "--recursive=false", "--regexp=true")
	testBintrayUpload(t, path, bintrayExpectedUploadNonFlatRecursive, "--flat=false", "--recursive=true", "--regexp=true")
	testBintrayUpload(t, path, bintrayExpectedUploadNonFlatNonRecursive, "--flat=false", "--recursive=false", "--regexp=true")

	cleanBintrayTest()
}

func TestBintrayLogs(t *testing.T) {
	initBintrayTest(t)

	packagePath := path.Join(bintrayOrganization, tests.BintrayRepo, tests.BintrayUploadTestPackageName)
	versionPath := path.Join(packagePath, tests.BintrayUploadTestVersion)
	createPackageAndVersion(packagePath, versionPath)

	path := tests.GetTestResourcesPath() + "a/*"
	bintrayCli.Exec("upload", path, versionPath, "--flat=true", "--recursive=true", "--publish=true")
	assertPackageFiles(tests.GetBintrayExpectedUploadFlatRecursive(), getPackageFiles(tests.BintrayUploadTestPackageName), t)
	bintrayCli.Exec("logs", packagePath)

	bintrayCli.Exec("package-delete", packagePath)
	cleanBintrayTest()
}

func TestBintrayFileDownloads(t *testing.T) {
	initBintrayTest(t)

	repositoryPath := path.Join(bintrayOrganization, tests.BintrayRepo)
	packagePath := path.Join(repositoryPath, tests.BintrayUploadTestPackageName)
	versionPath := path.Join(packagePath, tests.BintrayUploadTestVersion)
	createPackageAndVersion(packagePath, versionPath)
	path := tests.GetFilePathForBintray("*", tests.GetTestResourcesPath(), "a")
	bintrayCli.Exec("upload", path, versionPath, "--flat=true", "--recursive=true")
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
	assert.NoError(t, retryExecutor.Execute())

	// File b1.in
	args = []string{"download-file", repositoryPath + "/b1.in", tests.Out + "/bintray/x.in", "--unpublished=true"}
	assert.NoError(t, retryExecutor.Execute())

	// File c1.in
	args = []string{"download-file", repositoryPath + "/(c)1.in", tests.Out + "/bintray/z{1}.in", "--unpublished=true"}
	assert.NoError(t, retryExecutor.Execute())

	// File a/a1.in
	args = []string{"download-file", repositoryPath + "/" + tests.GetTestResourcesPath() + "(a)/a1.in", tests.Out + "/bintray/{1}/fullpatha1.in", "--flat=true", "--unpublished=true"}
	assert.NoError(t, retryExecutor.Execute())

	//Validate that files were downloaded as expected
	expected := []string{
		filepath.Join(tests.Out, "bintray", "a1.in"),
		filepath.Join(tests.Out, "bintray", "x.in"),
		filepath.Join(tests.Out, "bintray", "zc.in"),
		filepath.Join(tests.Out, "bintray", "a", "fullpatha1.in"),
	}
	paths, _ := fileutils.ListFilesRecursiveWalkIntoDirSymlink(tests.Out+"/bintray/", false)
	tests.VerifyExistLocally(expected, paths, t)

	// Cleanup
	bintrayCli.Exec("package-delete", packagePath)
	cleanBintrayTest()
}

func TestBintrayVersionDownloads(t *testing.T) {
	initBintrayTest(t)

	repositoryPath := path.Join(bintrayOrganization, tests.BintrayRepo)
	packagePath := path.Join(repositoryPath, tests.BintrayUploadTestPackageName)
	versionPath := path.Join(packagePath, tests.BintrayUploadTestVersion)
	createPackageAndVersion(packagePath, versionPath)

	path := tests.GetFilePathForBintray("*", tests.GetTestResourcesPath(), "a")
	bintrayCli.Exec("upload", path, versionPath, "--flat=true", "--recursive=true")
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
	tests.VerifyExistLocally(expected, paths, t)
	bintrayCli.Exec("package-delete", packagePath)
	cleanBintrayTest()
}

// Tests compatibility to file paths with windows separators.
func TestBintrayUploadWindowsCompatibility(t *testing.T) {
	initBintrayTest(t)
	if !coreutils.IsWindows() {
		t.Skip("Not running on Windows, skipping...")
	}

	packageName := "simpleUploadPackage"
	packagePath := path.Join(bintrayOrganization, tests.BintrayRepo, packageName)
	versionName := "1.0"
	versionPath := path.Join(packagePath, versionName)

	createPackageAndVersion(packagePath, versionPath)
	// Upload file
	fileName := "a1.in"
	path := "some/path/in/bintray/"
	uploadFilePath := ioutils.UnixToWinPathSeparator(tests.GetFilePathForBintray(fileName, tests.GetTestResourcesPath(), "a"))
	bintrayCli.Exec("upload", uploadFilePath, versionPath, path)

	//Check file uploaded
	expected := []tests.PackageSearchResultItem{{
		Repo:    tests.BintrayRepo,
		Path:    path + fileName,
		Package: packageName,
		Name:    fileName,
		Version: "1.0",
		Sha1:    "507ac63c6b0f650fb6f36b5621e70ebca3b0965c"}}
	assertPackageFiles(expected, getPackageFiles(packageName), t)

	bintrayCli.Exec("package-delete", packagePath)
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

func testBintrayUpload(t *testing.T, relPath string, expected []tests.PackageSearchResultItem, flags ...string) {
	packagePath := path.Join(bintrayOrganization, tests.BintrayRepo, tests.BintrayUploadTestPackageName)
	versionPath := path.Join(packagePath, tests.BintrayUploadTestVersion)
	createPackageAndVersion(packagePath, versionPath)

	args := []string{"upload", tests.GetTestResourcesPath() + relPath, versionPath}
	args = append(args, flags...)
	bintrayCli.Exec(args...)
	assertPackageFiles(expected, getPackageFiles(tests.BintrayUploadTestPackageName), t)
	bintrayCli.Exec("package-delete", packagePath)
}

func createHttpClientDetails() httputils.HttpClientDetails {
	return httputils.HttpClientDetails{
		User:     bintrayConfig.User,
		Password: bintrayConfig.Key,
		Headers:  map[string]string{"Content-Type": "application/json"}}
}

func getPackageFiles(packageName string) []tests.PackageSearchResultItem {
	apiUrl := bintrayConfig.ApiUrl + path.Join("packages", bintrayOrganization, tests.BintrayRepo, packageName, "files?include_unpublished=1")

	client, err := httpclient.ClientBuilder().Build()
	if err != nil {
		os.Exit(1)
	}
	resp, body, _, err := client.SendGet(apiUrl, true, createHttpClientDetails(), "")
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
	assert.Equal(t, len(expected), len(actual))

	expectedMap := make(map[string]tests.PackageSearchResultItem)
	for _, v := range expected {
		expectedMap[packageFileHash(v)] = v
	}

	actualMap := make(map[string]tests.PackageSearchResultItem)
	for _, v := range actual {
		actualMap[packageFileHash(v)] = v
	}

	for _, v := range actual {
		_, ok := expectedMap[packageFileHash(v)]
		assert.True(t, ok, "Unexpected file:", v)
	}
	for _, v := range expected {
		_, ok := actualMap[packageFileHash(v)]
		assert.True(t, ok, "File not found:", v)
	}
}

func packageFileHash(item tests.PackageSearchResultItem) string {
	return item.Repo + item.Path + item.Package + item.Version + item.Name + item.Sha1
}

func createPackageAndVersion(packagePath, versionPath string) {
	bintrayCli.Exec("package-create", packagePath, "--licenses=Apache-2.0", "--vcs-url=vcs.url.com")
	bintrayCli.Exec("version-create", versionPath, "--desc=versionDescription", "--vcs-tag=vcs.tag")
}

func createBintrayRepo() {
	content, err := os.ReadFile(tests.GetTestResourcesPath() + tests.BintrayTestRepositoryConfig)
	if errorutils.CheckError(err) != nil {
		os.Exit(1)
	}

	apiUrl := bintrayConfig.ApiUrl + path.Join("repos", bintrayOrganization, tests.BintrayRepo)
	client, err := httpclient.ClientBuilder().Build()
	if err != nil {
		os.Exit(1)
	}
	resp, body, err := client.SendPost(apiUrl, content, createHttpClientDetails(), "")
	if errorutils.CheckError(err) != nil {
		os.Exit(1)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusConflict {
		log.Error(resp.Status)
		log.Error(string(body))
		os.Exit(1)
	}
}

func getAllBintrayRepositories() ([]string, error) {
	var bintrayRepositories []string
	apiUrl := bintrayConfig.ApiUrl + path.Join("repos", bintrayOrganization)
	client, err := httpclient.ClientBuilder().Build()
	if err != nil {
		return nil, err
	}
	resp, body, _, err := client.SendGet(apiUrl, true, createHttpClientDetails(), "")
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("Bintray response: " + resp.Status + "\n" + clientutils.IndentJson(body))
	}

	// Extract repository keys from the json response
	var keyError error
	_, err = jsonparser.ArrayEach(body, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil || keyError != nil {
			return
		}
		repoKey, err := jsonparser.GetString(value, "name")
		if err != nil {
			keyError = err
			return
		}
		bintrayRepositories = append(bintrayRepositories, repoKey)
	})
	if keyError != nil {
		return nil, err
	}
	return bintrayRepositories, err
}

func deleteBintrayRepoByName(repoName string) {
	apiUrl := bintrayConfig.ApiUrl + path.Join("repos", bintrayOrganization, repoName)
	client, err := httpclient.ClientBuilder().Build()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	resp, body, err := client.SendDelete(apiUrl, nil, createHttpClientDetails(), "")
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

func deleteBintrayRepo() {
	deleteBintrayRepoByName(tests.BintrayRepo)
}

func cleanUpOldBintrayRepositories() {
	getActualItems := func() ([]string, error) { return getAllBintrayRepositories() }
	deleteItem := func(repoName string) {
		deleteBintrayRepoByName(repoName)
		log.Info("Repo", repoName, "deleted.")
	}
	tests.CleanUpOldItems([]string{tests.BintrayRepo}, getActualItems, deleteItem)
}
