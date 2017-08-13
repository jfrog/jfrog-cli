package services

import (
	"testing"
	"path/filepath"
	"os"
	"strings"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/fileutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services/utils/tests"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services/utils"
)

func TestArtifactoryUpload(t *testing.T) {
	t.Run("flat", flatUpload)
	t.Run("recursive", recursiveUpload)
	t.Run("placeholder", placeholderUpload)
	t.Run("includeDirs", includeDirsUpload)
	t.Run("explode", explodeUpload)
}

func flatUpload(t *testing.T) {
	workingDir, _, err := tests.CreateFileWithContent("a.in", "/out/")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	defer os.RemoveAll(workingDir)
	pattern := filepath.Join(workingDir, "out", "*")
	pattern = strings.Replace(pattern, "\\", "\\\\", -1)
	up := &UploadParamsImp{}
	up.ArtifactoryCommonParams = &utils.ArtifactoryCommonParams{Pattern: pattern, Recursive: true, Target: RtTargetRepo}
	up.Flat = true
	_, uploaded, failed, err := testsUploadService.UploadFiles(up)
	if uploaded != 1 {
		t.Error("Expected to upload 1 file.")
	}
	if failed != 0 {
		t.Error("Failed to upload", failed, "files.")
	}
	if err != nil {
		t.Error(err)
	}
	items, err := testsSearchService.Search(&utils.SearchParamsImpl{ArtifactoryCommonParams: &utils.ArtifactoryCommonParams{Pattern: RtTargetRepo}})
	if err != nil {
		t.Error(err)
	}
	if len(items) > 1 {
		t.Error("Expected single file.")
	}
	for _, item := range items {
		if item.Path != "." {
			t.Error("Expected path to be root due to using the flat flag.", "Got:", item.Path)
		}
	}
	artifactoryCleanUp(t)
}

func recursiveUpload(t *testing.T) {
	workingDir, _, err := tests.CreateFileWithContent("a.in", "/out/")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	defer os.RemoveAll(workingDir)
	pattern := filepath.Join(workingDir, "*")
	pattern = strings.Replace(pattern, "\\", "\\\\", -1)
	up := &UploadParamsImp{}
	up.ArtifactoryCommonParams = &utils.ArtifactoryCommonParams{Pattern: pattern, Recursive: true, Target: RtTargetRepo}
	up.Flat = true
	_, uploaded, failed, err := testsUploadService.UploadFiles(up)
	if uploaded != 1 {
		t.Error("Expected to upload 1 file.")
	}
	if failed != 0 {
		t.Error("Failed to upload", failed, "files.")
	}
	if err != nil {
		t.Error(err)
	}
	items, err := testsSearchService.Search(&utils.SearchParamsImpl{ArtifactoryCommonParams: &utils.ArtifactoryCommonParams{Pattern: RtTargetRepo}})
	if err != nil {
		t.Error(err)
	}
	if len(items) > 1 {
		t.Error("Expected single file.")
	}
	for _, item := range items {
		if item.Path != "." {
			t.Error("Expected path to be root(flat by default).", "Got:", item.Path)
		}
		if item.Name != "a.in" {
			t.Error("Missing File a.in")
		}
	}
	artifactoryCleanUp(t)
}

func placeholderUpload(t *testing.T) {
	workingDir, _, err := tests.CreateFileWithContent("a.in", "/out/")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	defer os.RemoveAll(workingDir)
	pattern := filepath.Join(workingDir, "(*).in")
	pattern = strings.Replace(pattern, "\\", "\\\\", -1)
	up := &UploadParamsImp{}
	up.ArtifactoryCommonParams = &utils.ArtifactoryCommonParams{Pattern: pattern, Recursive: true, Target: RtTargetRepo + "{1}"}
	up.Flat = true
	_, uploaded, failed, err := testsUploadService.UploadFiles(up)
	if uploaded != 1 {
		t.Error("Expected to upload 1 file.")
	}
	if failed != 0 {
		t.Error("Failed to upload", failed, "files.")
	}
	if err != nil {
		t.Error(err)
	}
	items, err := testsSearchService.Search(&utils.SearchParamsImpl{ArtifactoryCommonParams: &utils.ArtifactoryCommonParams{Pattern: RtTargetRepo}})
	if err != nil {
		t.Error(err)
	}
	if len(items) > 1 {
		t.Error("Expected single file.")
	}
	for _, item := range items {
		if item.Path != "out" {
			t.Error("Expected path to be out.", "Got:", item.Path)
		}
		if item.Name != "a" {
			t.Error("Missing File a")
		}
	}
	artifactoryCleanUp(t)
}

func includeDirsUpload(t *testing.T) {
	workingDir, _, err := tests.CreateFileWithContent("a.in", "/out/")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	defer os.RemoveAll(workingDir)
	pattern := filepath.Join(workingDir, "*")
	pattern = strings.Replace(pattern, "\\", "\\\\", -1)
	up := &UploadParamsImp{}
	up.ArtifactoryCommonParams = &utils.ArtifactoryCommonParams{Pattern: pattern, IncludeDirs: true, Recursive: false, Target: RtTargetRepo}
	up.Flat = true
	_, uploaded, failed, err := testsUploadService.UploadFiles(up)
	if uploaded != 0 {
		t.Error("Expected to upload 1 file.")
	}
	if failed != 0 {
		t.Error("Failed to upload", failed, "files.")
	}
	if err != nil {
		t.Error(err)
	}
	items, err := testsSearchService.Search(&utils.SearchParamsImpl{ArtifactoryCommonParams: &utils.ArtifactoryCommonParams{Pattern: RtTargetRepo, IncludeDirs: true}})
	if err != nil {
		t.Error(err)
	}
	if len(items) < 2 {
		t.Error("Expected to get at least two items, default and the out folder.")
	}
	for _, item := range items {
		if item.Name == "." {
			continue
		}
		if item.Path != "." {
			t.Error("Expected path to be root(flat by default).", "Got:", item.Path)
		}
		if item.Name != "out" {
			t.Error("Missing directory out.")
		}
	}
	artifactoryCleanUp(t)
}

func explodeUpload(t *testing.T) {
	workingDir, filePath, err := tests.CreateFileWithContent("a.in", "/out/")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	defer os.RemoveAll(workingDir)
	err = fileutils.ZipFolderFiles(filePath, filepath.Join(workingDir, "zipFile.zip"))
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	err = os.Remove(filePath)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	pattern := filepath.Join(workingDir, "*.zip")
	pattern = strings.Replace(pattern, "\\", "\\\\", -1)
	up := &UploadParamsImp{}
	up.ArtifactoryCommonParams = &utils.ArtifactoryCommonParams{Pattern: pattern, IncludeDirs: true, Recursive: false, Target: RtTargetRepo}
	up.Flat = true
	up.ExplodeArchive = true
	_, uploaded, failed, err := testsUploadService.UploadFiles(up)
	if uploaded != 1 {
		t.Error("Expected to upload 1 file.")
	}
	if failed != 0 {
		t.Error("Failed to upload", failed, "files.")
	}
	if err != nil {
		t.Error(err)
	}
	items, err := testsSearchService.Search(&utils.SearchParamsImpl{ArtifactoryCommonParams: &utils.ArtifactoryCommonParams{Pattern: RtTargetRepo, IncludeDirs: true}})
	if err != nil {
		t.Error(err)
	}
	if len(items) < 2 {
		t.Error("Expected to get at least two items, default and the out folder.")
	}
	for _, item := range items {
		if item.Name == "." {
			continue
		}
		if item.Name != "a.in" {
			t.Error("Missing file a.in")
		}
	}
	artifactoryCleanUp(t)
}

