package npm

import (
	"testing"
	"reflect"
)

func TestReadPackageInfoFromPackageJson(t *testing.T) {
	tests := []struct {
		json string
		pi   *PackageInfo
	}{
		{`{ "name": "jfrog-cli-tests", "version": "1.0.0", "description": "test package"}`,
			&PackageInfo{Name: "jfrog-cli-tests", Version: "1.0.0", Scope: ""}},
		{`{ "name": "@jfrog/jfrog-cli-tests", "version": "1.0.0", "description": "test package"}`,
			&PackageInfo{Name: "jfrog-cli-tests", Version: "1.0.0", Scope: "@jfrog"}},
	}
	for _, test := range tests {
		t.Run(test.json, func(t *testing.T) {
			packInfo, err := ReadPackageInfo([]byte(test.json))
			if err != nil {
				t.Error("No error was expected in this test", err)
			}

			equals := reflect.DeepEqual(test.pi, packInfo)
			if !equals {
				t.Error("expeted:", test.pi, "got:", packInfo)
			}
		})
	}
}

func TestGetDeployPath(t *testing.T) {
	tests := []struct {
		expectedPath string
		pi           *PackageInfo
	}{
		{`jfrog-cli-tests/-/jfrog-cli-tests-1.0.0.tgz`, &PackageInfo{Name: "jfrog-cli-tests", Version: "1.0.0", Scope: ""}},
		{`@jfrog/jfrog-cli-tests/-/@jfrog/jfrog-cli-tests-1.0.0.tgz`, &PackageInfo{Name: "jfrog-cli-tests", Version: "1.0.0", Scope: "@jfrog"}},
	}
	for _, test := range tests {
		t.Run(test.expectedPath, func(t *testing.T) {
			actualPath := test.pi.GetDeployPath()
			if actualPath != test.expectedPath {
				t.Error("expeted:", test.expectedPath, "got:", actualPath)
			}
		})
	}
}

func TestGetExpectedPackedFileName(t *testing.T) {
	tests := []struct {
		fileName string
		pi       *PackageInfo
	}{
		{`jfrog-cli-tests-1.0.0.tgz`, &PackageInfo{Name: "jfrog-cli-tests", Version: "1.0.0", Scope: ""}},
		{`jfrog-jfrog-cli-tests-1.0.0.tgz`, &PackageInfo{Name: "jfrog-cli-tests", Version: "1.0.0", Scope: "@jfrog"}},
	}
	for _, test := range tests {
		t.Run(test.fileName, func(t *testing.T) {
			actualFileName := test.pi.GetExpectedPackedFileName()
			if actualFileName != test.fileName {
				t.Error("expeted:", test.fileName, "got:", actualFileName)
			}
		})
	}
}
