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
	"os"
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

	isExistInArtifactory(tests.SimpleUploadExpected, tests.GetFilePath(tests.Search), t)
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

	paths, _ := ioutils.ListFilesRecursive(tests.Out + "/")
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
	isExistInArtifactoryByProps(tests.SimpleUploadExpected, tests.Repo1 + "/*", props, t)

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
	isExistInArtifactoryByProps(tests.SimpleUploadExpected2, tests.Repo2 + "/*", props, t)

	//cleanup
	deleteBuild(buildName)
	cleanArtifactoryTest()
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
	results, _ := searchInArtifactory(specFile);
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
	searchSpec := utils.CreateSpec(pattern, "", props, true, false, false)
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
