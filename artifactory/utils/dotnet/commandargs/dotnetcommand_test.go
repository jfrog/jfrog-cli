package commandargs

import (
	"encoding/xml"
	"github.com/jfrog/gofrog/io"
	"github.com/jfrog/jfrog-cli/artifactory/utils/dotnet"
	"github.com/jfrog/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-cli/utils/log"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestGetFlagValueExists(t *testing.T) {
	tests := []struct {
		name              string
		currentConfigPath string
		createConfig      bool
		expectErr         bool
		cmdFlags          []string
		expectedCmdFlags  []string
	}{
		{"simple", "file.config", true, false,
			[]string{"-configFile", "file.config"}, []string{"-configFile", "file.config"}},

		{"simple2", "file.config", true, false,
			[]string{"-before", "-configFile", "file.config", "after"}, []string{"-before", "-configFile", "file.config", "after"}},

		{"err", "file.config", false, true,
			[]string{"-before", "-configFile"}, []string{"-before", "-configFile"}},

		{"err2", "file.config", false, true,
			[]string{"-configFile"}, []string{"-configFile"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.createConfig {
				_, err := io.CreateRandFile(test.currentConfigPath, 0)
				if err != nil {
					t.Error(err)
				}
				defer os.Remove(test.currentConfigPath)
			}
			c := &dotnet.Cmd{CommandFlags: test.cmdFlags}
			_, err := getFlagValueIfExists("-configfile", c)
			if err != nil && !test.expectErr {
				t.Error(err)
			}
			if err == nil && test.expectErr {
				t.Errorf("Expecting: error, Got: nil")
			}
			if !reflect.DeepEqual(c.CommandFlags, test.expectedCmdFlags) {
				t.Errorf("Expecting: %s, Got: %s", test.expectedCmdFlags, c.CommandFlags)
			}
		})
	}
}

func TestInitNewConfig(t *testing.T) {
	log.SetDefaultLogger()

	tempDirPath, err := fileutils.CreateTempDir()
	if err != nil {
		t.Error(err)
	}
	defer fileutils.RemoveTempDir(tempDirPath)

	c := &dotnet.Cmd{}
	params := &DotnetCommand{rtDetails: &config.ArtifactoryDetails{Url: "http://some/url", User: "user", Password: "password"}}
	configFile, err := writeToTempConfigFile(c, tempDirPath)
	if err != nil {
		t.Error(err)
	}

	// Prepare the config file with NuGet authentication
	err = params.addNugetAuthenticationToNewConfig(dotnet.Nuget, configFile)
	if err != nil {
		t.Error(err)
	}

	content, err := ioutil.ReadFile(configFile.Name())
	if err != nil {
		t.Error(err)
	}

	nugetConfig := NugetConfig{}
	err = xml.Unmarshal(content, &nugetConfig)
	if err != nil {
		t.Error("Unmarshalling failed with an error:", err.Error())
	}

	if len(nugetConfig.PackageSources) != 1 {
		t.Error("Expected one package sources, got", len(nugetConfig.PackageSources))
	}

	source := "http://some/url/api/nuget"

	for _, packageSource := range nugetConfig.PackageSources {
		if packageSource.Key != SourceName {
			t.Error("Expected", SourceName, ",got", packageSource.Key)
		}

		if packageSource.Value != source {
			t.Error("Expected", source, ", got", packageSource.Value)
		}
	}

	if len(nugetConfig.PackageSourceCredentials) != 1 {
		t.Error("Expected one packageSourceCredentials, got", len(nugetConfig.PackageSourceCredentials))
	}

	if len(nugetConfig.PackageSourceCredentials[0].JFrogCli) != 2 {
		t.Error("Expected two fields in the JFrogCli credentials, got", len(nugetConfig.PackageSourceCredentials[0].JFrogCli))
	}
}

func TestUpdateSolutionPathAndGetFileName(t *testing.T) {
	workingDir, err := os.Getwd()
	assert.NoError(t, err)
	tests := []struct {
		name                 string
		flags                string
		solutionPath         string
		expectedSlnFile      string
		expectedSolutionPath string
	}{
		{"emptyFlags", "", workingDir, "", workingDir},
		{"justFlags", "-flag1 value1 -flag2 value2", workingDir, "", workingDir},
		{"relFileArgRelPath1", filepath.Join("testdata", "slnDir", "sol.sln"), filepath.Join("rel", "path"), "sol.sln", filepath.Join("rel", "path", "testdata", "slnDir")},
		{"relDirArgRelPath2", filepath.Join("testdata", "slnDir"), filepath.Join("rel", "path"), "", filepath.Join("rel", "path", "testdata", "slnDir")},
		{"absFileArgRelPath1", filepath.Join(workingDir, "testdata", "slnDir", "sol.sln"), filepath.Join(".", "rel", "path"), "sol.sln", filepath.Join(workingDir, "testdata", "slnDir")},
		{"absDirArgRelPath2", filepath.Join(workingDir, "testdata", "slnDir") + " -flag value", filepath.Join(".", "rel", "path"), "", filepath.Join(workingDir, "testdata", "slnDir")},
		{"nonExistingFile", filepath.Join(".", "dir1", "sol.sln"), workingDir, "", workingDir},
		{"nonExistingPath", filepath.Join(workingDir, "non", "existing", "path"), workingDir, "", workingDir},
		{"relCsprojFile", filepath.Join("testdata", "slnDir", "proj.csproj"), filepath.Join("rel", "path"), "", filepath.Join("rel", "path", "testdata", "slnDir")},
		{"absCsprojFile", filepath.Join(workingDir, "testdata", "slnDir", "proj.csproj"), filepath.Join("rel", "path"), "", filepath.Join(workingDir, "testdata", "slnDir")},
		{"relPackagesConfigFile", filepath.Join("testdata", "slnDir", "packages.config"), filepath.Join("rel", "path"), "", filepath.Join("rel", "path", "testdata", "slnDir")},
		{"absPackagesConfigFile", filepath.Join(workingDir, "testdata", "slnDir", "packages.config"), filepath.Join("rel", "path"), "", filepath.Join(workingDir, "testdata", "slnDir")},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dc := DotnetCommand{solutionPath: test.solutionPath, flags: test.flags}
			slnFile, err := dc.updateSolutionPathAndGetFileName()
			assert.NoError(t, err)
			assert.Equal(t, test.expectedSlnFile, slnFile)
			assert.Equal(t, test.expectedSolutionPath, dc.solutionPath)
		})
	}
}

type NugetConfig struct {
	XMLName                  xml.Name                   `xml:"configuration"`
	PackageSources           []PackageSources           `xml:"packageSources>add"`
	PackageSourceCredentials []PackageSourceCredentials `xml:"packageSourceCredentials"`
	Apikeys                  []PackageSources           `xml:"apikeys>add"`
}

type PackageSources struct {
	Key   string `xml:"key,attr"`
	Value string `xml:"value,attr"`
}

type PackageSourceCredentials struct {
	JFrogCli []PackageSources `xml:">add"`
}
