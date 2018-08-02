package services

import (
	"errors"
	"flag"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/artifactory/auth"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/artifactory/services/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/httpclient"
	clientutils "github.com/jfrog/jfrog-cli-go/jfrog-client/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/log"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/tests"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/io/fileutils"
)

var RtUrl *string
var RtUser *string
var RtPassword *string
var RtApiKey *string
var RtSshKeyPath *string
var RtSshPassphrase *string
var LogLevel *string
var testsUploadService *UploadService
var testsSearchService *SearchService
var testsDeleteService *DeleteService
var testsDownloadService *DownloadService

const (
	RtTargetRepo              = "jfrog-cli-tests-repo1/"
	SpecsTestRepositoryConfig = "specs_test_repository_config.json"
	RepoDetailsUrl            = "api/repositories/"
	ClientIntegrationTests    = "github.com/jfrog/jfrog-cli-go/jfrog-client/artifactory/services"
)

func init() {
	RtUrl = flag.String("rt.url", "http://localhost:8081/artifactory/", "Artifactory url")
	RtUser = flag.String("rt.user", "admin", "Artifactory username")
	RtPassword = flag.String("rt.password", "password", "Artifactory password")
	RtApiKey = flag.String("rt.apikey", "", "Artifactory user API key")
	RtSshKeyPath = flag.String("rt.sshKeyPath", "", "Ssh key file path")
	RtSshPassphrase = flag.String("rt.sshPassphrase", "", "Ssh key passphrase")
	LogLevel = flag.String("log-level", "INFO", "Sets the log level")
}

func TestMain(m *testing.M) {
	packages := tests.ExcludeTestsPackage(tests.GetTestPackages("../../..."), ClientIntegrationTests)
	tests.RunTests(packages)
	flag.Parse()
	log.Logger.SetLogLevel(log.GetCliLogLevel(*LogLevel))
	if *RtUrl != "" && !strings.HasSuffix(*RtUrl, "/") {
		*RtUrl += "/"
	}
	InitArtifactoryServiceManager()
	createReposIfNeeded()
	result := m.Run()
	os.Exit(result)
}

func InitArtifactoryServiceManager() {
	createArtifactoryUploadManager()
	createArtifactorySearchManager()
	createArtifactoryDeleteManager()
	createArtifactoryDownloadManager()
}

func createArtifactorySearchManager() {
	testsSearchService = NewSearchService(httpclient.NewDefaultHttpClient())
	testsSearchService.ArtDetails = getArtDetails()
}

func createArtifactoryDeleteManager() {
	testsDeleteService = NewDeleteService(httpclient.NewDefaultHttpClient())
	testsDeleteService.ArtDetails = getArtDetails()
}

func createArtifactoryUploadManager() {
	testsUploadService = NewUploadService(httpclient.NewDefaultHttpClient())
	testsUploadService.ArtDetails = getArtDetails()
	testsUploadService.Threads = 3
}

func createArtifactoryDownloadManager() {
	testsDownloadService = NewDownloadService(httpclient.NewDefaultHttpClient())
	testsDownloadService.ArtDetails = getArtDetails()
	testsDownloadService.SetThreads(3)
}

func getArtDetails() auth.ArtifactoryDetails {
	rtDetails := auth.NewArtifactoryDetails()
	rtDetails.SetUrl(*RtUrl)
	if !fileutils.IsSshUrl(rtDetails.GetUrl()) {
		if *RtApiKey != "" {
			rtDetails.SetApiKey(*RtApiKey)
		} else {
			rtDetails.SetUser(*RtUser)
			rtDetails.SetPassword(*RtPassword)
		}
		return rtDetails
	}

	sshKey, err := ioutil.ReadFile(clientutils.ReplaceTildeWithUserHome(*RtSshKeyPath))
	if err != nil {
		log.Error("Failed while attempting to read SSH key: " + err.Error())
		os.Exit(1)
	}

	err = rtDetails.AuthenticateSsh(sshKey, []byte(*RtSshPassphrase))
	if err != nil {
		log.Error("Failed while attempting to authenticate with Artifactory: " + err.Error())
		os.Exit(1)
	}
	return rtDetails
}

func artifactoryCleanUp(t *testing.T) {
	params := &utils.ArtifactoryCommonParams{Pattern: RtTargetRepo}
	toDelete, err := testsDeleteService.GetPathsToDelete(&DeleteParamsImpl{ArtifactoryCommonParams: params})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	deleteItems := make([]DeleteItem, len(toDelete))
	for i, item := range toDelete {
		deleteItems[i] = item
	}
	deletedCount, err := testsDeleteService.DeleteFiles(deleteItems, testsDeleteService)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if len(toDelete) != deletedCount {
		t.Errorf("Failed to delete files from Artifactory expected %d items to be deleted got %d.", len(toDelete), deletedCount)
	}
}

func createReposIfNeeded() error {
	var err error
	var repoConfig string
	repo := RtTargetRepo
	if strings.HasSuffix(repo, "/") {
		repo = repo[0:strings.LastIndex(repo, "/")]
	}
	if !isRepoExist(repo) {
		repoConfig = filepath.Join(getTestDataPath(), "reposconfig", SpecsTestRepositoryConfig)
		err = execCreateRepoRest(repoConfig, repo)
		if err != nil {
			log.Error(err)
			return err
		}
	}
	return nil
}

func isRepoExist(repoName string) bool {
	artDetails := getArtDetails()
	artHttpDetails := artDetails.CreateHttpClientDetails()
	client := httpclient.NewDefaultHttpClient()
	resp, _, _, err := client.SendGet(artDetails.GetUrl()+RepoDetailsUrl+repoName, true, artHttpDetails)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	if resp.StatusCode != http.StatusBadRequest {
		return true
	}
	return false
}

func execCreateRepoRest(repoConfig, repoName string) error {
	content, err := ioutil.ReadFile(repoConfig)
	if err != nil {
		return err
	}
	artHttpDetails := getArtDetails().CreateHttpClientDetails()

	artHttpDetails.Headers = map[string]string{"Content-Type": "application/json"}
	client := httpclient.NewDefaultHttpClient()
	resp, _, err := client.SendPut(*RtUrl+"api/repositories/"+repoName, content, artHttpDetails)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return errors.New("Fail to create repository. Reason local repository with key: " + repoName + " already exist\n")
	}
	log.Info("Repository", repoName, "was created.")
	return nil
}

func getTestDataPath() string {
	dir, _ := os.Getwd()
	return filepath.Join(dir, "testsdata")
}
