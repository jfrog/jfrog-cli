package nuget

import (
	"encoding/xml"
	"github.com/jfrog/gofrog/io"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils/nuget"
	"github.com/jfrog/jfrog-cli-go/utils/cliutils"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-cli-go/utils/log"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
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
			c := &nuget.Cmd{CommandFlags: test.cmdFlags}
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
	if !cliutils.IsWindows() {
		t.Skip("Skipping nuget tests, since this is not a Windows machine.")
	}

	tempDirPath, err := fileutils.CreateTempDir()
	if err != nil {
		t.Error(err)
	}
	defer fileutils.RemoveTempDir(tempDirPath)

	c := &nuget.Cmd{}
	params := &NugetCommandArgs{rtDetails: &config.ArtifactoryDetails{Url: "http://some/url", User: "user", Password: "password"}}
	configFile, err := writeToTempConfigFile(c, tempDirPath)
	if err != nil {
		t.Error(err)
	}

	// Prepare the config file with NuGet authentication
	err = params.addNugetAuthenticationToNewConfig(configFile)
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
		if packageSource.Key != sourceName {
			t.Error("Expected", sourceName, ",got", packageSource.Key)
		}

		if packageSource.Value != source {
			t.Error("Expected", source, ", got", packageSource.Value)
		}
	}

	if len(nugetConfig.Apikeys) != 1 {
		t.Error("Expected one api key, got", len(nugetConfig.Apikeys))
	}

	apiKey := nugetConfig.Apikeys[0]
	if apiKey.Key != source {
		t.Error("Expected", source, ", got", apiKey.Key)
	}
	if apiKey.Value == "" {
		t.Error("Expected apiKey with value, got", apiKey.Value)
	}

	if len(nugetConfig.PackageSourceCredentials) != 1 {
		t.Error("Expected one packageSourceCredentials, got", len(nugetConfig.PackageSourceCredentials))
	}

	if len(nugetConfig.PackageSourceCredentials[0].JFrogCli) != 2 {
		t.Error("Expected two fields in the JFrogCli credentials, got", len(nugetConfig.PackageSourceCredentials[0].JFrogCli))
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
