package main

import (
	"testing"
	"os"
	"github.com/jfrogdev/jfrog-cli-go/artifactory/commands"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/tests"
	"strings"
	"fmt"
)

func TestArtifactorySimpleUploadSpec(t *testing.T) {
	tests.InitTest()
	specFile := tests.GetFile(tests.SimpleUploadSpec)
	resultSpecFile := tests.GetFile(tests.Search)
	os.Args = tests.GetSpecCommandAsArray("u", specFile)
	tests.LogCommand()
	main()
	tests.IsExistInArtifactory(tests.SimpleUploadExpected, resultSpecFile, t)
}

func TestArtifactoryMoveSpec(t *testing.T) {
	tests.InitTest()
	prepUploadFiles()
	prepCopyFiles()
	specFile := tests.GetFile(tests.MoveCopyDeleteSpec)
	os.Args = tests.GetSpecCommandAsArray("cp", specFile)
	tests.LogCommand()
	main()
	tests.IsExistInArtifactory(tests.MassiveMoveExpected, tests.GetFile(tests.SearchMoveDeleteRepoSpec), t)
}

func TestArtifactoryDelete(t *testing.T) {
	tests.InitTest()
	prepUploadFiles()
	prepCopyFiles()
	specFile := tests.GetFile(tests.MoveCopyDeleteSpec)
	os.Args = tests.GetSpecCommandAsArray("cp", specFile)
	tests.LogCommand()
	main()
	os.Args = []string{"jfrog", "rt", "del", tests.Repo2 + "/nonflat_recursive_target/nonflat_recursive_source/a/b/*", "--url=" + *tests.Url, "--user=" + *tests.User, "--password=" + *tests.Password, "--quiet=true"}
	tests.LogCommand()
	main()
	tests.IsExistInArtifactory(tests.Delete1, tests.GetFile(tests.SearchMoveDeleteRepoSpec), t)
}

func TestArtifactoryDeleteFolderWithWildcard(t *testing.T) {
	tests.InitTest()
	prepUploadFiles()
	prepCopyFiles()
	specFile := tests.GetFile(tests.MoveCopyDeleteSpec)
	os.Args = tests.GetSpecCommandAsArray("cp", specFile)
	tests.LogCommand()
	main()
	resp, _, _, _ := ioutils.SendGet(*tests.Url + "api/storage/" + tests.Repo2 + "/nonflat_recursive_target/nonflat_recursive_source/a/b/", true, ioutils.HttpClientDetails{User:*tests.User, Password:*tests.Password})
	if resp.StatusCode != 200 {
		t.Error("Missing folder in artifactory : " + tests.Repo2 + "/nonflat_recursive_target/nonflat_recursive_source/a/b/")
	}

	os.Args = []string{"jfrog", "rt", "del", tests.Repo2 + "/nonflat_recursive_target/nonflat_recursive_source/*/b/", "--url=" + *tests.Url, "--user=" + *tests.User, "--password=" + *tests.Password, "--quiet=true"}
	tests.LogCommand()
	main()
	resp, _, _, _ = ioutils.SendGet(*tests.Url + "api/storage/" + tests.Repo2 + "/nonflat_recursive_target/nonflat_recursive_source/a/b/", true, ioutils.HttpClientDetails{User:*tests.User, Password:*tests.Password})
	if resp.StatusCode != 404 {
		t.Error("Couldn't delete folder in artifactory : " + tests.Repo2 + "/nonflat_recursive_target/nonflat_recursive_source/a/b/")
	}

	tests.IsExistInArtifactory(tests.Delete1, tests.GetFile(tests.SearchMoveDeleteRepoSpec), t)
}

func TestArtifactoryDeleteFolder(t *testing.T) {
	tests.InitTest()
	prepUploadFiles()
	os.Args = []string{"jfrog", "rt", "del", tests.Repo1 + "/downloadTestResources/", "--quiet=true", "--url=" + *tests.Url, "--user=" + *tests.User, "--password=" + *tests.Password}
	tests.LogCommand()
	main()
	resp, body, _, err := ioutils.SendGet(*tests.Url + "api/storage/" + tests.Repo1 + "/downloadTestResources", true, ioutils.HttpClientDetails{User:*tests.User, Password:*tests.Password})
	if err != nil || resp.StatusCode != 404 {
		t.Error("Coudln't delete path: " + tests.Repo1 + "/downloadTestResources/ " + string(body))
	}
}

func TestArtifactoryDeleteFoldersBySpec(t *testing.T) {
	tests.InitTest()
	prepUploadFiles()
	prepCopyFiles()

	os.Args = []string{"jfrog", "rt", "del", "--spec=" + tests.GetFile(tests.DeleteSpec), "--quiet=true", "--url=" + *tests.Url, "--user=" + *tests.User, "--password=" + *tests.Password}
	tests.LogCommand()
	main()
	resp, body, _, err := ioutils.SendGet(*tests.Url + "api/storage/" + tests.Repo1 + "/downloadTestResources", true, ioutils.HttpClientDetails{User:*tests.User, Password:*tests.Password})
	if err != nil || resp.StatusCode != 404 {
		t.Error("Coudln't delete path: " + tests.Repo1 + "/downloadTestResources/ " + string(body))
	}
	resp, body, _, err = ioutils.SendGet(*tests.Url + "api/storage/" + tests.Repo2 + "/downloadTestResources", true, ioutils.HttpClientDetails{User:*tests.User, Password:*tests.Password})
	if err != nil || resp.StatusCode != 404 {
		t.Error("Coudln't delete path: " + tests.Repo2 + "/downloadTestResources/ " + string(body))
	}
}

func TestArtifactoryDisplyedPathToDelete(t *testing.T) {
	tests.InitTest()
	prepUploadFiles()
	prepCopyFiles()
	specFile := tests.GetFile(tests.DeleteComplexSpec)
	artifactsToDelete := tests.GetPathsToDelete(specFile)
	var displyedPaths []commands.SearchResult
	for _, v := range artifactsToDelete {
		displyedPaths = append(displyedPaths, commands.SearchResult{Path:v.Repo + "/" + v.Path + "/" + v.Name})
	}
	tests.CompareExpectedVsActuals(tests.DeleteDisplyedFiles, displyedPaths, t)
}

func TestArtifactoryDeleteBySpec(t *testing.T) {
	tests.InitTest()
	prepUploadFiles()
	prepCopyFiles()
	specFile := tests.GetFile(tests.DeleteComplexSpec)
	os.Args = tests.GetDeleteCommandAsArray(specFile)
	tests.LogCommand()
	main()
	artifactsToDelete := tests.GetPathsToDelete(specFile)
	if len(artifactsToDelete) != 0 {
		t.Error("Couldn't delete paths")
	}
}

func TestArtifactoryMassiveDownloadSpec(t *testing.T) {
	tests.InitTest()
	prepUploadFiles()
	specFile := tests.GetFile(tests.DownloadSpec)
	os.Args = tests.GetSpecCommandAsArray("dl", specFile)
	tests.LogCommand()
	main()
	paths, _ := ioutils.ListFilesRecursive(tests.Out + "/")
	tests.IsExistLocally(tests.MassiveDownload, paths, t)
}

func TestArtifactoryMassiveUploadSpec(t *testing.T) {
	tests.InitTest()
	specFile := tests.GetFile(tests.UploadSpec)
	resultSpecFile := tests.GetFile(tests.Search)
	os.Args = tests.GetSpecCommandAsArray("u", specFile)
	tests.LogCommand()
	main()
	tests.IsExistInArtifactory(tests.MassiveUpload, resultSpecFile, t)
	tests.IsExistInArtifactoryByProps(tests.PropsExpected, tests.Repo1 + "/*/properties/*.in", "searchMe=true", t)
}

func TestArtifactorySearchByApiKey(t *testing.T) {
	if len(*tests.ApiKey) > 0 {
		tests.InitTest()
		specFile := tests.GetFile(tests.UploadSpec)
		resultSpecFile := tests.GetFile(tests.Search)
		os.Args = tests.GetSpecCommandAsArray("u", specFile)
		tests.LogCommand()
		main()
		results, _ := tests.SearchInArtifactory(resultSpecFile, tests.CreateArtifactoryDetailsByApiKey);
		tests.CompareExpectedVsActuals(tests.MassiveUpload, results, t)
	}
}

func TestArtifactoryPublishBuildInfo(t *testing.T) {
	tests.InitTest()
	buildName, buildNumber := "cli-test-build", "10"

	//upload files with buildName and buildNumber
	specFile := tests.GetFile(tests.SimpleUploadSpec)
	os.Args = tests.AppendBuildInfoParams(tests.GetSpecCommandAsArray("u", specFile), buildName, buildNumber)
	tests.LogCommand()
	main()

	//publish buildInfo
	os.Args = tests.AppendCredentials(strings.Split(fmt.Sprintf("jfrog rt bp %v %v", buildName, buildNumber), " "))
	tests.LogCommand()
	main()

	//validate files are uploaded with the build info name and number
	props := fmt.Sprintf("build.name=%v;build.number=%v", buildName, buildNumber)
	tests.IsExistInArtifactoryByProps(tests.SimpleUploadExpected, tests.Repo1 + "/*", props, t)

	//cleanup
	tests.DeleteBuild(buildName)
}

func TestArtifactoryCleanBuildInfo(t *testing.T) {
	tests.InitTest()
	buildName, buildNumber := "cli-test-build", "11"

	//upload files with buildName and buildNumber
	specFile := tests.GetFile(tests.UploadSpec)
	os.Args = tests.AppendBuildInfoParams(tests.GetSpecCommandAsArray("u", specFile), buildName, buildNumber)
	tests.LogCommand()
	main()

	//cleanup buildInfo
	os.Args = strings.Split(fmt.Sprintf("jfrog rt bc %v %v", buildName, buildNumber), " ")
	tests.LogCommand()
	main()

	//upload files with buildName and buildNumber
	specFile = tests.GetFile(tests.SimpleUploadSpec)
	os.Args = tests.AppendBuildInfoParams(tests.GetSpecCommandAsArray("u", specFile), buildName, buildNumber)
	tests.LogCommand()
	main()

	//publish buildInfo
	os.Args = tests.AppendCredentials(strings.Split(fmt.Sprintf("jfrog rt bp %v %v", buildName, buildNumber), " "))
	tests.LogCommand()
	main()

	//promote buildInfo
	os.Args = tests.AppendCredentials(strings.Split(fmt.Sprintf("jfrog rt bpr %v %v %v", buildName, buildNumber, tests.Repo2), " "))
	tests.LogCommand()
	main()

	//validate files are uploaded with the build info name and number
	props := fmt.Sprintf("build.name=%v;build.number=%v", buildName, buildNumber)
	tests.IsExistInArtifactoryByProps(tests.SimpleUploadExpected2, tests.Repo2 + "/*", props, t)

	//cleanup
	tests.DeleteBuild(buildName)
}

func prepUploadFiles() {
	os.Args = []string{"jfrog", "rt", "u", tests.FixWinPath(tests.GetTestResourcesPath()) + "(.*)", tests.Repo1 + "/downloadTestResources/{1}", "--threads=10", "--regexp=true", "--props=searchMe=true", "--flat=false", "--url=" + *tests.Url, "--user=" + *tests.User, "--password=" + *tests.Password}
	tests.LogCommand()
	main()
}

func prepCopyFiles() {
	specFile := tests.GetFile(tests.PrepareCopy)
	os.Args = tests.GetSpecCommandAsArray("cp", specFile)
	tests.LogCommand()
	main()
}
