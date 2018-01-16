package services

import (
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/bintray/services/utils/tests"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/bintray/services/versions"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/httpclient"
	"path/filepath"
	"strconv"
	"testing"
)

func TestSingleFileUpload(t *testing.T) {
	uploadService := newDryRunUploadService()
	params, err := createUploadParams()
	if err != nil {
		t.Error(err.Error())
	}
	params.Pattern = "testdata/a.txt"
	uploaded1, _, err := uploadService.Upload(params)
	if err != nil {
		t.Error(err.Error())
	}

	params.Pattern = "testdata/aa.txt"
	uploaded2, _, err := uploadService.Upload(params)
	if err != nil {
		t.Error(err.Error())
	}

	params.Pattern = "testdata/aa1*.txt"
	uploaded3, _, err := uploadService.Upload(params)
	if err != nil {
		t.Error(err.Error())
	}
	if uploaded1 != 1 {
		t.Error("Expected 1 file to be uploaded. Got " + strconv.Itoa(uploaded1) + ".")
	}
	if uploaded2 != 1 {
		t.Error("Expected 1 file to be uploaded. Got " + strconv.Itoa(uploaded2) + ".")
	}
	if uploaded3 != 0 {
		t.Error("Expected 0 file to be uploaded. Got " + strconv.Itoa(uploaded3) + ".")
	}
}

func TestPatternRecursiveUpload(t *testing.T) {
	params, err := createUploadParams()
	if err != nil {
		t.Error(err.Error())
	}
	params.Recursive = true
	testPatternUpload(t, params)
}

func TestPatternNonRecursiveUpload(t *testing.T) {
	params, err := createUploadParams()
	if err != nil {
		t.Error(err.Error())
	}
	params.Recursive = false
	testPatternUpload(t, params)
}

func testPatternUpload(t *testing.T, params *UploadParams) {
	uploadService := newDryRunUploadService()

	sep := string(filepath.Separator)
	params.Pattern = "testdata" + sep + "*"
	uploaded1, _, err := uploadService.Upload(params)
	if err != nil {
		t.Error(err.Error())
	}

	params.Pattern = "testdata" + sep + "a*"
	uploaded2, _, err := uploadService.Upload(params)
	if err != nil {
		t.Error(err.Error())
	}

	params.Pattern = "testdata" + sep + "b*"
	uploaded3, _, err := uploadService.Upload(params)
	if err != nil {
		t.Error(err.Error())
	}

	if uploaded1 != 3 {
		t.Error("Expected 3 file to be uploaded. Got " + strconv.Itoa(uploaded1) + ".")
	}
	if uploaded2 != 2 {
		t.Error("Expected 2 file to be uploaded. Got " + strconv.Itoa(uploaded2) + ".")
	}
	if uploaded3 != 1 {
		t.Error("Expected 1 file to be uploaded. Got " + strconv.Itoa(uploaded3) + ".")
	}
}

func createUploadParams() (*UploadParams, error) {
	versionPath, err := versions.CreatePath("test-subject/test-repo/test-package/ver-1.2")
	if err != nil {
		return nil, err
	}
	params := &UploadParams{Params: &versions.Params{}}
	params.Path = versionPath
	params.TargetPath = "/a/b/"
	params.Recursive = true
	params.Flat = true
	return params, nil
}

func newDryRunUploadService() *UploadService {
	uploadService := NewUploadService(httpclient.NewDefaultHttpClient())
	uploadService.DryRun = true
	uploadService.BintrayDetails = tests.CreateBintrayDetails()
	uploadService.Threads = 1
	return uploadService
}
