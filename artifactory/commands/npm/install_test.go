package npm

import (
	"io/ioutil"
	"strings"
	"testing"
)

func TestPrepareConfigData(t *testing.T) {
	configJson, err := ioutil.ReadFile("../testdata/config.json")
	if err != nil {
		t.Error(err)
	}

	expectedConfig :=
		[]string{
			"json = true",
			"allow-same-version = false",
			"user-agent = npm/5.5.1 node/v8.9.1 darwin x64",
			"@jfrog:registry = http://goodRegistry",
			"email = ddd@dd.dd",
			"cache-lock-retries = 10",
			"registry = http://goodRegistry",
			"_auth = YWRtaW46QVBCN1ZkZFMzN3NCakJiaHRGZThVb0JlZzFl"}

	npmi := NpmCommandArgs{registry: "http://goodRegistry", npmAuth: "_auth = YWRtaW46QVBCN1ZkZFMzN3NCakJiaHRGZThVb0JlZzFl"}
	actualConfig, err := npmi.prepareConfigData([]byte(configJson))
	if err != nil {
		t.Error(err)
	}
	actualConfigArray := strings.Split(string(actualConfig), "\n")
	if len(actualConfigArray) != len(expectedConfig) {
		t.Errorf("expeted:\n%s\n\ngot:\n%s", expectedConfig, actualConfigArray)
	}
	for _, eConfig := range expectedConfig {
		found := false
		for _, aConfig := range actualConfigArray {
			if aConfig == eConfig {
				found = true
				break
			}
		}
		if !found {
			t.Error("The expected config:", eConfig, "is missing from the actual configuration list:\n", actualConfigArray)
			t.Errorf("The expected config: %s is missing from the actual configuration list:\n %s", eConfig, actualConfigArray)
		}
	}
}

func TestPrepareConfigDataTypeRestriction(t *testing.T) {
	var typeRestrictions = map[string]string{
		`{"production": true}`:    "production",
		`{"only": "prod"}`:        "production",
		`{"only": "production"}`:  "production",
		`{"only": "development"}`: "development",
		`{"only": "dev"}`:         "development",
		`{"only": null}`:          "",
		`{"only": ""}`:            "",
		`{"kuku": true}`:          ""}

	for json, typeRestriction := range typeRestrictions {
		npmi := NpmCommandArgs{}
		npmi.prepareConfigData([]byte(json))
		if npmi.typeRestriction != typeRestriction {
			t.Errorf("Type restriction was supposed to be %s but set to: %s when using the json:\n%s", typeRestriction, npmi.typeRestriction, json)
		}
	}
}

func TestParseDependencies(t *testing.T) {
	dependenciesJsonList, err := ioutil.ReadFile("../testdata/dependenciesList.json")
	if err != nil {
		t.Error(err)
	}

	expectedDependenciesList := []string{
		"underscore-1.4.4",
		"@jfrog/npm_scoped-1.0.0",
		"xml-1.0.1",
		"xpm-0.1.1",
		"binary-search-tree-0.2.4",
		"nedb-1.0.2",
		"@ilg/es6-promisifier-0.1.9",
		"wscript-avoider-3.0.2",
		"yaml-0.2.3",
		"@ilg/cli-start-options-0.1.19",
		"async-0.2.10",
		"find-0.2.7",
		"jquery-3.2.0",
		"nub-1.0.0",
		"shopify-liquid-1.d7.9",
	}
	npmi := NpmCommandArgs{}
	npmi.dependencies = make(map[string]*dependency)
	err = npmi.parseDependencies([]byte(dependenciesJsonList), "myScope")
	if err != nil {
		t.Error(err)
	}
	if len(expectedDependenciesList) != len(npmi.dependencies) {
		t.Error("The expected dependencies list length is", len(expectedDependenciesList), "and should be:\n", expectedDependenciesList,
			"\nthe actual dependencies list length is", len(npmi.dependencies), "and the list is:\n", npmi.dependencies)
		t.Error("The expected dependencies list length is", len(expectedDependenciesList), "and should be:\n", expectedDependenciesList,
			"\nthe actual dependencies list length is", len(npmi.dependencies), "and the list is:\n", npmi.dependencies)
	}
	for _, eDependency := range expectedDependenciesList {
		found := false
		for aDependency := range npmi.dependencies {
			if aDependency == eDependency {
				found = true
				break
			}
		}
		if !found {
			t.Error("The expected dependency:", eDependency, "is missing from the actual dependencies list:\n", npmi.dependencies)
		}
	}
}

func TestGetRegistry(t *testing.T) {
	var getRegistryTest = []struct {
		repo     string
		url      string
		expected string
	}{
		{"repo", "http://url/art", "http://url/art/api/npm/repo"},
		{"repo", "http://url/art/", "http://url/art/api/npm/repo"},
		{"repo", "", "/api/npm/repo"},
		{"", "http://url/art", "http://url/art/api/npm/"},
	}

	for _, testCase := range getRegistryTest {
		if getNpmRepositoryUrl(testCase.repo, testCase.url) != testCase.expected {
			t.Errorf("The expected output of getRegistry(\"%s\", \"%s\") is %s. But the actual result is:%s", testCase.repo, testCase.url, testCase.expected, getNpmRepositoryUrl(testCase.repo, testCase.url))
		}
	}
}
