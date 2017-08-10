package services

import (
	"testing"
	"flag"
	"os"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/types/httpclient"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services/utils/auth"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services/utils"
)

var RtUrl *string
var RtUser *string
var RtPassword *string
var RtTargetRepo *string
var testsUploadService *UploadService
var testsSearchService *SearchService
var testsDeleteService *DeleteService
var testsDownloadService *DownloadService

func init() {
	RtUrl = flag.String("rt.url", "http://localhost:8081/artifactory/", "Artifactory url")
	RtUser = flag.String("rt.user", "admin", "Artifactory username")
	RtPassword = flag.String("rt.password", "password", "Artifactory password")
	RtTargetRepo = flag.String("rt.targetRepo", "Copy/", "Artifactory target repo")
}

func TestMain(m *testing.M) {
	flag.Parse()
	InitArtifactoryServiceManager()
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
	testsSearchService = NewSearchService(httpclient.NewDefaultJforgHttpClient())
	testsSearchService.ArtDetails = getArtDetails()
}

func createArtifactoryDeleteManager() {
	testsDeleteService = NewDeleteService(httpclient.NewDefaultJforgHttpClient())
	testsDeleteService.ArtDetails = getArtDetails()
}

func createArtifactoryUploadManager() {
	testsUploadService = NewUploadService(httpclient.NewDefaultJforgHttpClient())
	testsUploadService.ArtDetails = getArtDetails()
	testsUploadService.Threads = 3
}

func createArtifactoryDownloadManager() {
	testsDownloadService = NewDownloadService(httpclient.NewDefaultJforgHttpClient())
	testsDownloadService.ArtDetails = getArtDetails()
	testsDownloadService.SetThreads(3)
}

func getArtDetails() *auth.ArtifactoryDetails {
	return &auth.ArtifactoryDetails{Url: *RtUrl, User: *RtUser, Password: *RtPassword}
}

func artifactoryCleanUp(t *testing.T) {
	params := &utils.ArtifactoryCommonParams{Pattern: *RtTargetRepo}
	toDelete, err := testsDeleteService.GetPathsToDelete(&DeleteParamsImpl{ArtifactoryCommonParams: params})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	var deleteItems []DeleteItem = make([]DeleteItem, len(toDelete))
	for i, item := range toDelete {
		deleteItems[i] = item
	}
	err = testsDeleteService.DeleteFiles(deleteItems, testsDeleteService)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
}
