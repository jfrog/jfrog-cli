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
	params := &DotnetCommandArgs{rtDetails: &config.ArtifactoryDetails{Url: "http://some/url", User: "user", Password: "password"}}
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
		{"emptyFlags", "", "/path/to/solution/", "", "/path/to/solution/"},
		{"justFlags", "-flag1 value1 -flag2 value2", "/path/to/solution/", "", "/path/to/solution/"},
		{"relFileArgRelPath1", "testdata/slnDir/sol.sln", "rel/path/", "sol.sln", "rel/path/testdata/slnDir"},
		{"relDirArgRelPath2", "testdata/slnDir/", "rel/path", "", "rel/path/testdata/slnDir"},
		{"absFileArgRelPath1", workingDir + "/testdata/slnDir/sol.sln", "./rel/path/", "sol.sln", workingDir + "/testdata/slnDir"},
		{"absDirArgRelPath2", workingDir + "/testdata/slnDir/ -flag value", "./rel/path/", "", workingDir + "/testdata/slnDir/"},
		{"nonExistingFile", "./dir1/sol.sln", "/path/to/solution/", "", "/path/to/solution/"},
		{"nonExistingPath", "/non/existing/path/", "/path/to/solution/", "", "/path/to/solution/"},
		{"relCsprojFile", "testdata/slnDir/proj.csproj", "rel/path/", "", "rel/path/testdata/slnDir"},
		{"absCsprojFile", workingDir + "/testdata/slnDir/proj.csproj", "rel/path/", "", workingDir + "/testdata/slnDir"},
		{"relPackagesConfigFile", "testdata/slnDir/packages.config", "rel/path/", "", "rel/path/testdata/slnDir"},
		{"absPackagesConfigFile", workingDir + "/testdata/slnDir/packages.config", "rel/path/", "", workingDir + "/testdata/slnDir"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dca := DotnetCommandArgs{solutionPath: test.solutionPath, flags: test.flags}
			slnFile, err := dca.updateSolutionPathAndGetFileName()
			assert.NoError(t, err)
			assert.Equal(t, test.expectedSlnFile, slnFile)
			assert.Equal(t, test.expectedSolutionPath, dca.solutionPath)
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
