package dependencies

import (
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDependenciesCache(t *testing.T) {
	// Change test's work directory, rollback after function returns.
	wd, _ := os.Getwd()
	tmpTestPath := filepath.Join(os.TempDir(), "cacheTest")
	err := os.MkdirAll(tmpTestPath, os.ModePerm)
	if err != nil {
		t.Error("Failed mkDirAll: " + err.Error())
	}
	err = os.Chdir(tmpTestPath)
	if err != nil {
		t.Error("Failed Chdir: " + err.Error())
	}
	defer func() {
		os.RemoveAll(tmpTestPath)
		os.Chdir(wd)
	}()

	cacheMap := make(map[string]*buildinfo.Dependency)
	csA := buildinfo.Checksum{Sha1: "sha1A", Md5: "md5A"}
	depenA := buildinfo.Dependency{
		Id:       "depenA-1.0-A.zip",
		Checksum: &csA,
	}
	cacheMap["A"] = &depenA
	csC := buildinfo.Checksum{Sha1: "sha1C", Md5: "md5C"}
	depenC := buildinfo.Dependency{
		Id:       "depenC-3.4-C.gzip",
		Checksum: &csC,
	}
	cacheMap["C"] = &depenC
	err = UpdateDependenciesCache(cacheMap)
	if err != nil {
		t.Error("Failed creating dependencies cache: " + err.Error())
	}
	cache, err := readCacheAndCheckError()
	if err != nil {
		t.Error("Failed reading dependencies cache: " + err.Error())
	}

	if !reflect.DeepEqual(*cache.GetDependency("A"), depenA) {
		t.Error("Failed retrieving dependency A!!!")
	}
	if cache.GetDependency("B") != nil {
		t.Error("Retrieving non-existing dependency B should return nil.")
	}
	if !reflect.DeepEqual(*cache.GetDependency("C"), depenC) {
		t.Error("Failed retrieving dependency C!!!")
	}
	if cache.GetDependency("T") != nil {
		t.Error("Retrieving non-existing dependency T should return nil checksum.")
	}

	delete(cacheMap, "A")
	csT := buildinfo.Checksum{Sha1: "sha1T", Md5: "md5T"}
	depenT := buildinfo.Dependency{
		Id:       "depenT-6.0.68-T.zip",
		Checksum: &csT,
	}
	cacheMap["T"] = &depenT
	err = UpdateDependenciesCache(cacheMap)
	if err != nil {
		t.Error("Failed creating dependencies cache: " + err.Error())
	}

	cache, err = readCacheAndCheckError()
	if err != nil {
		t.Error("Failed reading dependencies cache: " + err.Error())
	}
	if cache.GetDependency("A") != nil {
		t.Error("Retrieving non-existing dependency T should return nil checksum.")
	}
	if !reflect.DeepEqual(*cache.GetDependency("T"), depenT) {
		t.Error("Failed retrieving dependency T.")
	}
	if !reflect.DeepEqual(*cache.GetDependency("C"), depenC) {
		t.Error("Failed retrieving dependency C.")
	}
}

func readCacheAndCheckError() (cache *DependenciesCache, err error) {
	cache, err = GetProjectDependenciesCache()
	if err != nil {
		return
	}
	if cache == nil {
		err = errors.New("Cache file does not exist.")
	}
	return
}
