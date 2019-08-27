package dependencies

import (
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"reflect"
	"testing"
)

func TestDependenciesCache(t *testing.T) {
	cache := make(DependenciesCache)
	csA := buildinfo.Checksum{Sha1: "sha1A", Md5: "md5A"}
	depenA := buildinfo.Dependency{
		Id:       "depenA-1.0-A.zip",
		Checksum: &csA,
	}
	cache["A"] = &depenA
	csC := buildinfo.Checksum{Sha1: "sha1C", Md5: "md5C"}
	depenC := buildinfo.Dependency{
		Id:       "depenC-3.4-C.gzip",
		Checksum: &csC,
	}
	cache["C"] = &depenC
	err := UpdateDependenciesCache(cache)
	if err != nil {
		t.Error("Failed creating dependencies cache!!!")
	}
	newCache, err := GetProjectDependenciesCache()
	if newCache == nil || err != nil {
		t.Error("Failed reading dependencies cache!!!")
	}

	if !reflect.DeepEqual(*newCache.GetDependency("A"), depenA) {
		t.Error("Failed retrieving dependency A!!!")
	}
	if newCache.GetDependency("B") != nil {
		t.Error("Retrieving non-existing dependency B should return nil!!!")
	}
	if !reflect.DeepEqual(*newCache.GetDependency("C"), depenC) {
		t.Error("Failed retrieving dependency C!!!")
	}
	if newCache.GetDependency("T") != nil {
		t.Error("Retrieving non-existing dependency T should return nil checksum!!!")
	}

	delete(*newCache, "A")
	csT := buildinfo.Checksum{Sha1: "sha1T", Md5: "md5T"}
	depenT := buildinfo.Dependency{
		Id:       "depenT-6.0.68-T.zip",
		Checksum: &csT,
	}
	(*newCache)["T"] = &depenT
	err = UpdateDependenciesCache(*newCache)
	if err != nil {
		t.Error("Failed creating dependencies cache!!!")
	}

	lastCache, err := GetProjectDependenciesCache()
	if lastCache == nil || err != nil {
		t.Error("Failed reading dependencies cache!!!")
	}
	if lastCache.GetDependency("A") != nil {
		t.Error("Retrieving non-existing dependency T should return nil checksum!!!")
	}
	if !reflect.DeepEqual(*lastCache.GetDependency("T"), depenT) {
		t.Error("Failed retrieving dependency T!!!")
	}
	if !reflect.DeepEqual(*lastCache.GetDependency("C"), depenC) {
		t.Error("Failed retrieving dependency C!!!")
	}

}
