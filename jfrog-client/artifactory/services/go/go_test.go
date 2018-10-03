package _go

import (
	"strings"
	"testing"
)

func TestCreateUrlPath(t *testing.T) {

	tests := []struct {
		name        string
		params      GoParams
		url         string
		expectedUrl string
	}{
		{"withBuildProperties", &GoParamsImpl{ZipPath: "path/to/zip/file", Version: "v1.1.1", TargetRepo: "ArtiRepo", ModuleId: "github.com/jfrog/test", Props: "build.name=a;build.number=1"}, "http://test.url/", "http://test.url//github.com/jfrog/test/@v/v1.1.1.zip;build.name=a;build.number=1"},
		{"withoutBuildProperties", &GoParamsImpl{ZipPath: "path/to/zip/file", Version: "v1.1.1", TargetRepo: "ArtiRepo", ModuleId: "github.com/jfrog/test"}, "http://test.url/", "http://test.url//github.com/jfrog/test/@v/v1.1.1.zip"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			createUrlPath(test.params, &test.url)
			if !strings.EqualFold(test.url, test.expectedUrl) {
				t.Error("Expected:", test.expectedUrl, "Got:", test.url)
			}
		})
	}
}

func TestShouldUseHeaders(t *testing.T) {
	tests := []struct {
		artifactoryVersion string
		expectedResult     bool
	}{
		{"6.5.0", false},
		{"6.2.0", true},
		{"5.9.0", true},
		{"6.0.0", true},
		{"6.6.0", false},
		{"development", false},
		{"6.10.2", false},
	}
	for _, test := range tests {
		t.Run(test.artifactoryVersion, func(t *testing.T) {
			result := shouldUseHeaders(test.artifactoryVersion)
			if result && !test.expectedResult {
				t.Error("Expected:", test.expectedResult, "Got:", result)
			}

			if !result && test.expectedResult {
				t.Error("Expected:", test.expectedResult, "Got:", result)
			}
		})
	}
}