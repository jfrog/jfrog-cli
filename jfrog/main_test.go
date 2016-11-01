package main

import (
	"testing"
	"flag"
	"fmt"
	"os"
	"github.com/jfrogdev/jfrog-cli-go/artifactory/commands"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/tests"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/logger"
)

func TestMain(m *testing.M) {
	flag.Parse()
	err := tests.CreateReposIfNeeded()
	if err != nil {
		logger.Logger.Error(err)
		os.Exit(1)
	}
	result := m.Run()
	tests.CleanUp()
	os.Exit(result)
}

func TestSimpleUploadSpec(t *testing.T) {
	tests.CleanUp()
	specFile := tests.GetFile(tests.SimpleUploadSpec)
	resultSpecFile := tests.GetFile(tests.Search)
	os.Args = tests.GetSpecCommandAsArray("u", specFile)
	main()
	tests.IsExistInArtifactory(tests.SimpleUploadExpected, resultSpecFile, t)
}

func TestMoveSpec(t *testing.T) {
	tests.CleanUp()
	prepUploadFiles()
	prepCopyFiles()
	specFile := tests.GetFile(tests.MoveCopyDeleteSpec)
	os.Args = tests.GetSpecCommandAsArray("cp", specFile)
	main()
	tests.IsExistInArtifactory(tests.MassiveMoveExpected, tests.GetFile(tests.SearchMoveDeleteRepoSpec), t)
}

func TestDelete(t *testing.T) {
	tests.CleanUp()
	prepUploadFiles()
	prepCopyFiles()
	specFile := tests.GetFile(tests.MoveCopyDeleteSpec)
	os.Args = tests.GetSpecCommandAsArray("cp", specFile)
	main()
	os.Args = []string{"jfrog", "rt", "del", tests.Repo2 + "/nonflat_recursive_target/nonflat_recursive_source/a/b/*", "--url=" + *tests.Url, "--user=" + *tests.User, "--password=" + *tests.Password, "--quiet=true"}
	main()
	tests.IsExistInArtifactory(tests.Delete1, tests.GetFile(tests.SearchMoveDeleteRepoSpec), t)
}

func TestDeleteFolderWithWildcard(t *testing.T) {
	tests.CleanUp()
	prepUploadFiles()
	prepCopyFiles()
	specFile := tests.GetFile(tests.MoveCopyDeleteSpec)
	os.Args = tests.GetSpecCommandAsArray("cp", specFile)
	main()
	resp, _, _, _ := ioutils.SendGet(*tests.Url + "api/storage/" + tests.Repo2 + "/nonflat_recursive_target/nonflat_recursive_source/a/b/", true, ioutils.HttpClientDetails{User:*tests.User, Password:*tests.Password})
	if resp.StatusCode != 200 {
		t.Error("Missing folder in artifactory : " + tests.Repo2 + "/nonflat_recursive_target/nonflat_recursive_source/a/b/")
	}

	os.Args = []string{"jfrog", "rt", "del", tests.Repo2 + "/nonflat_recursive_target/nonflat_recursive_source/*/b/", "--url=" + *tests.Url, "--user=" + *tests.User, "--password=" + *tests.Password, "--quiet=true"}
	fmt.Println(os.Args)
	main()
	resp, _, _, _ = ioutils.SendGet(*tests.Url + "api/storage/" + tests.Repo2 + "/nonflat_recursive_target/nonflat_recursive_source/a/b/", true, ioutils.HttpClientDetails{User:*tests.User, Password:*tests.Password})
	if resp.StatusCode != 404 {
		t.Error("Couldn't delete folder in artifactory : " + tests.Repo2 + "/nonflat_recursive_target/nonflat_recursive_source/a/b/")
	}

	tests.IsExistInArtifactory(tests.Delete1, tests.GetFile(tests.SearchMoveDeleteRepoSpec), t)
}

func TestDeleteFolder(t *testing.T) {
	tests.CleanUp()
	prepUploadFiles()
	os.Args = []string{"jfrog", "rt", "del", tests.Repo1 + "/downloadTestResources/", "--quiet=true", "--url=" + *tests.Url, "--user=" + *tests.User, "--password=" + *tests.Password}
	fmt.Println(os.Args)
	main()
	resp, body, _, err := ioutils.SendGet(*tests.Url + "api/storage/" + tests.Repo1 + "/downloadTestResources", true, ioutils.HttpClientDetails{User:*tests.User, Password:*tests.Password})
	if err != nil || resp.StatusCode != 404 {
		t.Error("Coudln't delete path: " + tests.Repo1 + "/downloadTestResources/ " + string(body))
	}
}

func TestDeleteFoldersBySpec(t *testing.T) {
	tests.CleanUp()
	prepUploadFiles()
	prepCopyFiles()

	os.Args = []string{"jfrog", "rt", "del", "--spec=" + tests.GetFile(tests.DeleteSpec), "--quiet=true", "--url=" + *tests.Url, "--user=" + *tests.User, "--password=" + *tests.Password}
	fmt.Println(os.Args)
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

func TestDisplyedPathToDelete(t *testing.T) {
	tests.CleanUp()
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

func TestDeleteBySpec(t *testing.T) {
	tests.CleanUp()
	prepUploadFiles()
	prepCopyFiles()
	specFile := tests.GetFile(tests.DeleteComplexSpec)
	os.Args = tests.GetDeleteCommandAsArray(specFile)
	main()
	artifactsToDelete := tests.GetPathsToDelete(specFile)
	if len(artifactsToDelete) != 0 {
		t.Error("Couldn't delete paths")
	}
}

func TestMassiveDownloadSpec(t *testing.T) {
	tests.CleanUp()
	prepUploadFiles()
	specFile := tests.GetFile(tests.DownloadSpec)
	os.Args = tests.GetSpecCommandAsArray("dl", specFile)
	main()
	paths, _ := ioutils.ListFilesRecursive(tests.Out + "/")
	tests.IsExistLocally(tests.MassiveDownload, paths, t)
}

func TestMassiveUploadSpec(t *testing.T) {
	tests.CleanUp()
	specFile := tests.GetFile(tests.UploadSpec)
	resultSpecFile := tests.GetFile(tests.Search)
	os.Args = tests.GetSpecCommandAsArray("u", specFile)
	main()
	tests.IsExistInArtifactory(tests.MassiveUpload, resultSpecFile, t)
	tests.IsExistInArtifactoryByProps(tests.PropsExpected, tests.Repo1 + "/*/properties/*.in", "searchMe=true", t)
}

func TestSearchByApiKey(t *testing.T) {
	if len(*tests.ApiKey) > 0 {
		tests.CleanUp()
		specFile := tests.GetFile(tests.UploadSpec)
		resultSpecFile := tests.GetFile(tests.Search)
		os.Args = tests.GetSpecCommandAsArray("u", specFile)
		main()
		results, _ := tests.SearchInArtifactory(resultSpecFile, tests.CreateArtifactoryDetailsByApiKey);
		tests.CompareExpectedVsActuals(tests.MassiveUpload, results, t)
	}
}

func prepUploadFiles() {
	os.Args = []string{"jfrog", "rt", "u", tests.FixWinPath(tests.GetTestResourcesPath()) + "(.*)", tests.Repo1 + "/downloadTestResources/{1}", "--threads=10", "--regexp=true", "--props=searchMe=true", "--flat=false", "--url=" + *tests.Url, "--user=" + *tests.User, "--password=" + *tests.Password}
	main()
}

func prepCopyFiles() {
	specFile := tests.GetFile(tests.PrepareCopy)
	os.Args = tests.GetSpecCommandAsArray("cp", specFile)
	main()
}
