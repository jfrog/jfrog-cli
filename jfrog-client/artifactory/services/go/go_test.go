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
