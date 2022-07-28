package utils

import (
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/httputils"
	"io"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
)

const (
	// This env var is mandatory for the 'publish' command.
	// The env var is optional for the install command - if provided, the plugin will be downloaded from a custom
	// plugins server, instead of the official registry.
	// The env var should store a server ID configured by JFrog CLI.
	PluginsServerEnv = "JFROG_CLI_PLUGINS_SERVER"
	// Used to set a custom plugins repo for the 'publish' & 'install' commands.
	PluginsRepoEnv     = "JFROG_CLI_PLUGINS_REPO"
	DefaultPluginsRepo = "jfrog-cli-plugins"

	PluginsOfficialRegistryUrl = "https://releases.jfrog.io/artifactory/"

	LatestVersionName = "latest"
)

var ArchitecturesMap = map[string]Architecture{
	"linux-386":     {"linux", "386", ""},
	"linux-amd64":   {"linux", "amd64", ""},
	"linux-s390x":   {"linux", "s390x", ""},
	"linux-arm64":   {"linux", "arm64", ""},
	"linux-arm":     {"linux", "arm", ""},
	"linux-ppc6":    {"linux", "ppc64", ""},
	"linux-ppc64le": {"linux", "ppc64le", ""},
	"mac-arm64":     {"darwin", "arm64", ""},
	"mac-386":       {"darwin", "amd64", ""},
	"windows-amd64": {"windows", "amd64", ".exe"},
}

// Returns plugin's directory path in Artifactory, corresponding to the local architecture.
// Example path: "repo-name/plugin-name/version/architecture-name
func GetPluginDirPath(pluginName, pluginVersion, architecture string) (pluginDirRtPath string) {
	pluginDirRtPath = path.Join(GetPluginVersionDirInArtifactory(pluginName, pluginVersion), architecture)
	return
}

// Returns plugin's executable name in Artifactory.
func GetPluginExecutableName(pluginName, architecture string) string {
	return pluginName + ArchitecturesMap[architecture].FileExtension
}

// Example path: "repo-name/plugin-name/v1.0.0/"
func GetPluginVersionDirInArtifactory(pluginName, pluginVersion string) string {
	return path.Join(GetPluginsRepo(), pluginName, pluginVersion)
}

// Returns a custom plugins repo if provided, default otherwise.
func GetPluginsRepo() string {
	repo := os.Getenv(PluginsRepoEnv)
	if repo != "" {
		return repo
	}
	return DefaultPluginsRepo
}

type Architecture struct {
	Goos          string
	Goarch        string
	FileExtension string
}

// Get the local architecture name corresponding to the architectures that exist in registry.
func GetLocalArchitecture() (string, error) {
	switch runtime.GOOS {
	case "windows":
		return "windows-amd64", nil
	case "darwin":
		if runtime.GOARCH == "arm64" {
			return "mac-arm64", nil
		} else {
			return "mac-386", nil
		}
	}
	// Assuming linux.
	switch runtime.GOARCH {
	case "amd64":
		return "linux-amd64", nil
	case "arm64":
		return "linux-arm64", nil
	case "arm":
		return "linux-arm", nil
	case "386":
		return "linux-386", nil
	case "s390x":
		return "linux-s390x", nil
	case "ppc64":
		return "linux-ppc64", nil
	case "ppc64le":
		return "linux-ppc64le", nil
	}
	return "", errorutils.CheckErrorf("no compatible plugin architecture was found for the architecture of this machine")
}

func CreatePluginsHttpDetails(rtDetails *config.ServerDetails) httputils.HttpClientDetails {
	if rtDetails.AccessToken != "" && rtDetails.ArtifactoryRefreshToken == "" {
		return httputils.HttpClientDetails{AccessToken: rtDetails.AccessToken}
	}
	return httputils.HttpClientDetails{
		User:     rtDetails.User,
		Password: rtDetails.Password}
}

// Asserts a plugin's version is as expected, by parsing the output of the version command.
func AssertPluginVersion(versionCmdOut string, expectedPluginVersion string) error {
	// Get the actual version which is after the last space. (expected output to -v for example: "plugin-name version v1.0.0")
	split := strings.Split(strings.TrimSpace(versionCmdOut), " ")
	if len(split) != 3 {
		return errorutils.CheckErrorf("failed verifying plugin version. Unexpected plugin output for version command: '" + versionCmdOut + "'")
	}
	if split[2] != expectedPluginVersion {
		return errorutils.CheckErrorf("provided version does not match the plugin's actual version. " +
			"Provided: '" + expectedPluginVersion + "', Actual: '" + split[2] + "'")
	}
	return nil
}

// Command used to build plugins.
type PluginBuildCmd struct {
	OutputFullPath string
	Env            map[string]string
}

func (buildCmd *PluginBuildCmd) GetCmd() *exec.Cmd {
	var cmd []string
	cmd = append(cmd, []string{"go", "build", "-o"}...)
	cmd = append(cmd, buildCmd.OutputFullPath)
	return exec.Command(cmd[0], cmd[1:]...)
}

func (buildCmd *PluginBuildCmd) GetEnv() map[string]string {
	buildCmd.Env["CGO_ENABLED"] = "0"
	return buildCmd.Env
}

func (buildCmd *PluginBuildCmd) GetStdWriter() io.WriteCloser {
	return nil
}

func (buildCmd *PluginBuildCmd) GetErrWriter() io.WriteCloser {
	return nil
}
