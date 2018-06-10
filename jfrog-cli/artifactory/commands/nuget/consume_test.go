package nuget

import (
	"github.com/jfrog/gofrog/io"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils/nuget"
)

func TestCopyExistingConfig(t *testing.T) {
	// Create temp files
	content := []byte("test file")
	err := ioutil.WriteFile("currentConfig", content, 0644)
	if err != nil {
		t.Error("Couldn't create file:", err)
	}
	defer os.Remove("currentConfig")

	newConfigFile, err := ioutil.TempFile(os.TempDir(), "newConfig")
	if err != nil {
		t.Error("Couldn't create file:", err)
	}

	err = copyExistingConfig(newConfigFile, "currentConfig")
	if err != nil {
		t.Error(err)
	}
	newConfigFile.Close()

	// check the new config is as the current
	newConfigContent, err := ioutil.ReadFile(newConfigFile.Name())
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(newConfigFile.Name())

	if string(content) != string(newConfigContent) {
		t.Errorf("Expecting: %s, Got: %s", string(content), string(newConfigContent))
	}
}

func Test(t *testing.T) {
	tests := []struct {
		name              string
		currentConfigPath string
		createConfig      bool
		expectErr         bool
		newConfigPath     string
		cmdFlags          []string
		expectedCmdFlags  []string
	}{
		{"simple", "file.config", true, false, "new.file.config",
			[]string{"-configFile", "file.config"}, []string{"-configFile", "new.file.config"}},

		{"simple2", "file.config", true, false, "new.file.config",
			[]string{"-before", "-configFile", "file.config", "after"}, []string{"-before", "-configFile", "new.file.config", "after"}},

		{"simple3", "file.config", true, false, "new.file.config",
			[]string{"-configFile", "file.config"}, []string{"-configFile", "new.file.config"}},

		{"configFileNotFound", "file.config", false, true, "new.file.config",
			[]string{"-before", "-configFile", "file.config", "after"}, []string{"-before", "-configFile", "file.config", "after"}},

		{"err", "file.config", false, true, "new.file.config",
			[]string{"-before", "-configFile", "after"}, []string{"-before", "-configFile", "after"}},

		{"err2", "file.config", false, true, "new.file.config",
			[]string{"-before", "-configFile"}, []string{"-before", "-configFile"}},

		{"err3", "file.config", false, true, "new.file.config",
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
			_, err := getAndReplaceCurrentConfigPath(test.newConfigPath, c)
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
	configFile, err := ioutil.TempFile("", "jfrog.cli.nuget.")
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(configFile.Name())

	c := &nuget.Cmd{}
	params := &Params{ArtifactoryDetails: &config.ArtifactoryDetails{Url: "some/url", User: "user", Password: "password"}}
	err = initNewConfig(configFile, params, c)
	if err != nil {
		t.Error(err)
	}
	err = configFile.Close()
	if err != nil {
		t.Error(err)
	}

	content, err := ioutil.ReadFile(configFile.Name())
	if err != nil {
		t.Error(err)
	}

	expectedContent := `<?xml version="1.0" encoding="utf-8"?>
<configuration>
  <packageSources>
    <add key="Artifactory" value="some/url/api/nuget" />
  </packageSources>
  <packageSourceCredentials>
    <Artifactory>
      <add key="Username" value="user" />
      <add key="ClearTextPassword" value="password" />
    </Artifactory>
  </packageSourceCredentials>
</configuration>`

	if expectedContent != string(content) {
		t.Errorf("Expecting: \n%s\n, Got: \n%s", expectedContent, content)
	}

}
