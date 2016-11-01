package tests

import (
	"testing"
	"strconv"
	"flag"
	"strings"
	"os"
	"errors"
	"fmt"
	"io/ioutil"
	"runtime"
	"github.com/jfrogdev/jfrog-cli-go/artifactory/commands"
	"github.com/jfrogdev/jfrog-cli-go/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/logger"
)

var PrintSearchResult *bool
var Url *string
var User *string
var Password *string
var ApiKey *string

func init() {
	PrintSearchResult = flag.Bool("printSearchResult", false, "Set to true for printing search results")
	Url = flag.String("url", "http://localhost:8081/artifactory/", "Artifactory url")
	User = flag.String("user", "admin", "Artifactory username")
	Password = flag.String("password", "password", "Artifactory password")
	ApiKey = flag.String("apikey", "", "Artifactory user API key")
}

func CreateReposIfNeeded() error {
	var err error
	var repoConfig string
	if !isRepoExist(Repo1) {
		repoConfig = GetTestResourcesPath() + SpecsTestRepositoryConfig
		err = execCreateRepoRest(repoConfig, Repo1)
		if err != nil {
			return err
		}
	}

	if !isRepoExist(Repo2) {
		repoConfig = GetTestResourcesPath() + MoveRepositoryConfig
		err = execCreateRepoRest(repoConfig, Repo2)
		if err != nil {
			return err
		}
	}
	return nil
}

func execCreateRepoRest(repoConfig, repoName string) error {
	content, err := ioutil.ReadFile(repoConfig)
	if err != nil {
		return err
	}
	resp, _, err := ioutils.SendPut(*Url + "api/repositories/" + repoName, content, ioutils.HttpClientDetails{User:*User, Password:*Password, Headers: map[string]string{"Content-Type": "application/json"}})
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return errors.New("Fail to create repository. Reason local repository with key: " + repoName + " already exist\n")
	}
	return nil
}

func isRepoExist(repoName string) bool {
	resp, _, _, _ := ioutils.SendGet(*Url + RepoDetailsUrl + repoName, true, ioutils.HttpClientDetails{User:*User, Password:*Password})
	println(*Url + RepoDetailsUrl + repoName)
	if resp.StatusCode != 400 {
		return true
	}
	return false
}

func CleanUp() {
	cleanArtifactory()
	cleanFileSystem()
}

func cleanFileSystem() {
	isExist, err := ioutils.IsDirExists(Out)
	if err != nil {
		logger.Logger.Error(err)
	}
	if isExist {
		os.RemoveAll(Out)
	}
}

func cleanArtifactory() {
	details := createArtifactoryDetailsByUserPassowrd()
	deleteFlags := new(commands.DeleteFlags)
	deleteFlags.ArtDetails = details
	deleteSpec, _ := utils.CreateSpecFromFile(GetFile(DeleteSpec))
	commands.Delete(deleteSpec, deleteFlags)
}

func IsExistLocally(expected, actual []string, t *testing.T) {
	if len(actual) == 0 && len(expected) != 0 {
		t.Error("Couldn't find all expected files, expected: " + strconv.Itoa(len(expected)) + ", found: " + strconv.Itoa(len(actual)))
	}
	for _, v := range expected {
		for i, r := range actual {
			if v == r {
				break
			}
			if i == len(actual) - 1 {
				t.Error("Missing file : " + v)
			}
		}
	}
}

type createArtifactoryDetails func() *config.ArtifactoryDetails

func createArtifactoryDetailsByUserPassowrd() *config.ArtifactoryDetails {
	details := new(config.ArtifactoryDetails)
	details.Url = *Url
	details.User = *User
	details.Password = *Password
	return details
}

func CreateArtifactoryDetailsByApiKey() *config.ArtifactoryDetails {
	details := new(config.ArtifactoryDetails)
	details.Url = *Url
	details.ApiKey = *ApiKey
	return details
}

func SearchInArtifactory(specFile string, rtDetailsCreatorFunc createArtifactoryDetails) (result []commands.SearchResult, err error) {
	details := rtDetailsCreatorFunc()
	searchFlags := new(commands.SearchFlags)
	searchFlags.ArtDetails = details
	searchSpec, _ := utils.CreateSpecFromFile(specFile)
	result, err = commands.Search(searchSpec, searchFlags)
	return
}

func IsExistInArtifactory(expected []string, specFile string, t *testing.T) {
	results, _ := SearchInArtifactory(specFile, createArtifactoryDetailsByUserPassowrd);
	if *PrintSearchResult {
		for _, v := range results {
			fmt.Print("\"")
			fmt.Print(v.Path)
			fmt.Print("\"")
			fmt.Print(",")
			fmt.Println("")
		}
	}
	CompareExpectedVsActuals(expected, results, t)
}

func IsExistInArtifactoryByProps(expected []string, pattern, props string, t *testing.T) {
	details := createArtifactoryDetailsByUserPassowrd()
	searchFlags := new(commands.SearchFlags)
	searchFlags.ArtDetails = details
	searchSpec := utils.CreateSpec(pattern, "", props, true, false, false)
	results, _ := commands.Search(searchSpec, searchFlags)
	CompareExpectedVsActuals(expected, results, t)
}

func CompareExpectedVsActuals(expected []string, actual []commands.SearchResult, t *testing.T) {
	if len(actual) != len(expected) {
		t.Error("Couldn't find all expected files, expected: " + strconv.Itoa(len(expected)) + ", found: " + strconv.Itoa(len(actual)))
	}
	for _, v := range expected {
		for i, r := range actual {
			if v == r.Path {
				break
			}
			if i == len(actual) - 1 {
				t.Error("Missing file: " + v)
			}
		}
	}
}

func GetTestResourcesPath() string {
	dir, _ := os.Getwd()
	fileSepatatr := ioutils.GetFileSeperator()
	index := strings.LastIndex(dir, fileSepatatr)
	dir = dir[:index]
	return dir + ioutils.GetFileSeperator() + "testsdata" + ioutils.GetFileSeperator()
}

func getFileByOs(fileName string) string {
	var currentOs string;
	fileSepatatr := ioutils.GetFileSeperator()
	if runtime.GOOS == "windows" {
		currentOs = "win"
	} else {
		currentOs = "unix"
	}
	return GetTestResourcesPath() + "specs" + fileSepatatr + currentOs + fileSepatatr + fileName
}

func GetFile(fileName string) string {
	filePath := GetTestResourcesPath() + "specs/common" + ioutils.GetFileSeperator() + fileName
	isExists, _ := ioutils.IsFileExists(filePath)
	if isExists {
		return filePath
	}
	return getFileByOs(fileName)
}

func FixWinPath(filePath string) string {
	fixedPath := strings.Replace(filePath, "\\", "\\\\", -1)
	return fixedPath
}

func GetSpecCommandAsArray(command, spec string) []string {
	parsedCommand := fmt.Sprintf(SpecsCommand, command, spec)
	fmt.Println(parsedCommand)
	return AppendCredentials(strings.Split(parsedCommand, " "))
}

func GetDeleteCommandAsArray(spec string) []string {
	parsedCommand := fmt.Sprintf(SpecsCommand, "del", spec)
	parsedCommand += " --quiet=true"
	fmt.Println(parsedCommand)
	return AppendCredentials(strings.Split(parsedCommand, " "))
}

func AppendCredentials(args []string) []string {
	credentialsParameters := fmt.Sprintf(CredentialsParameters, *Url, *User, *Password)
	fmt.Println(credentialsParameters)
	return append(args, strings.Split(credentialsParameters, " ")...)
}

func GetPathsToDelete(specFile string) []utils.AqlSearchResultItem {
	details := createArtifactoryDetailsByUserPassowrd()
	flags := new(commands.DeleteFlags)
	flags.ArtDetails = details
	deleteSpec, _ := utils.CreateSpecFromFile(specFile)
	artifactsToDelete, _ := commands.GetPathsToDelete(deleteSpec, flags)
	return artifactsToDelete
}