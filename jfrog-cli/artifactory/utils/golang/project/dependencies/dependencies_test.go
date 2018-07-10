package dependencies

import (
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/tests"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseListOutput(t *testing.T) {
	content := []byte(`github.com/vgo-example 
golang.org/x/text       v0.0.0-20170915032832-14c0d48ead0c
rsc.io/quote            v1.5.2
rsc.io/sampler          v1.3.0
	`)

	actual, err := parseListOutput(content)
	if err != nil {
		t.Error(err)
	}
	expected := map[string]string{
		"golang.org/x/text": "v0.0.0-20170915032832-14c0d48ead0c",
		"rsc.io/quote":      "v1.5.2",
		"rsc.io/sampler":    "v1.3.0",
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expecting: \n%s \nGot: \n%s", expected, actual)
	}
}

func TestGetPackageZipLocation(t *testing.T) {
	cachePath := filepath.Join(tests.GetBaseDir(true), "zip", "download")
	tests := []struct {
		dependencyName string
		version        string
		expectedPath   string
	}{
		{"rsc.io/quote", "v1.5.2", filepath.Join(filepath.Dir(cachePath), "rsc.io", "quote", "@v", "v1.5.2.zip")},
		{"rsc/quote", "v1.5.3", filepath.Join(cachePath, "rsc", "quote", "@v", "v1.5.3.zip")},
		{"rsc.io/quote", "v1.5.4", ""},
	}

	for _, test := range tests {
		t.Run(test.dependencyName+":"+test.version, func(t *testing.T) {
			actual, err := getPackageZipLocation(cachePath, test.dependencyName, test.version)
			if err != nil {
				t.Error(err.Error())
			}

			if test.expectedPath != actual {
				t.Errorf("Test name: %s:%s: Expected: %s, Got: %s", test.dependencyName, test.version, test.expectedPath, actual)
			}
		})
	}
}
