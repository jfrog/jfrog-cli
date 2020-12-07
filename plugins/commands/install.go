package commands

import (
	"errors"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/jfrog/jfrog-cli-core/utils/coreutils"
	pluginsutils "github.com/jfrog/jfrog-cli/plugins/utils"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	logUtils "github.com/jfrog/jfrog-cli/utils/log"
	"github.com/jfrog/jfrog-cli/utils/progressbar"
	"github.com/jfrog/jfrog-client-go/httpclient"
	"github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/io/httputils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

const pluginsRegistryUrl = "https://releases.jfrog.io/artifactory"
const pluginsRegistryRepo = "jfrog-cli-plugins"
const latestVersionName = "latest"

func InstallCmd(c *cli.Context) error {
	if c.NArg() != 1 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	return runInstallCmd(c.Args().Get(0))
}

func runInstallCmd(requestedPlugin string) error {
	pluginName, version, err := getNameAndVersion(requestedPlugin)
	if err != nil {
		return err
	}
	srcPath, err := buildSrcPath(pluginName, version)
	if err != nil {
		return err
	}
	downloadUrl := utils.AddTrailingSlashIfNeeded(pluginsRegistryUrl) + srcPath

	pluginsDir, err := coreutils.GetJfrogPluginsDir()
	if err != nil {
		return err
	}

	exists, err := fileutils.IsDirExists(pluginsDir, false)
	if err != nil {
		return err
	}
	if exists {
		should, err := shouldDownloadPlugin(pluginsDir, pluginName, downloadUrl)
		if err != nil {
			return err
		}
		if !should {
			return errors.New("requested plugin already exists locally")
		}
	} else {
		err = createPluginsDir(pluginsDir)
		if err != nil {
			return err
		}
	}

	return downloadPlugin(pluginsDir, pluginName, downloadUrl)
}

func shouldDownloadPlugin(pluginsDir, pluginName, downloadUrl string) (bool, error) {
	log.Debug("Verifying plugin download is needed...")
	client, err := httpclient.ClientBuilder().Build()
	if err != nil {
		return false, err
	}
	log.Debug("Fetching plugin details from: ", downloadUrl)

	details, resp, err := client.GetRemoteFileDetails(downloadUrl, httputils.HttpClientDetails{})
	if err != nil {
		return false, err
	}
	log.Debug("Artifactory response: ", resp.Status)
	err = errorutils.CheckResponseStatus(resp, http.StatusOK)
	if err != nil {
		return false, err
	}
	isEqual, err := fileutils.IsEqualToLocalFile(filepath.Join(pluginsDir, pluginName), details.Checksum.Md5, details.Checksum.Sha1)
	return !isEqual, err
}

func buildSrcPath(pluginName, version string) (string, error) {
	arc, err := getArchitecture()
	if err != nil {
		return "", err
	}
	return path.Join(pluginsRegistryRepo, pluginName, version, arc, pluginsutils.GetPluginExecutableName(pluginName)), nil
}

func createPluginsDir(pluginsDir string) error {
	return os.MkdirAll(pluginsDir, 0777)
}

func downloadPlugin(pluginsDir, pluginName, downloadUrl string) error {
	exeName := pluginsutils.GetPluginExecutableName(pluginName)
	log.Debug("Downloading plugin from: ", downloadUrl)
	downloadDetails := &httpclient.DownloadFileDetails{
		FileName:      pluginName,
		DownloadPath:  downloadUrl,
		LocalPath:     pluginsDir,
		LocalFileName: exeName,
		RelativePath:  exeName,
	}

	client, err := httpclient.ClientBuilder().Build()
	if err != nil {
		return err
	}
	// Init progress bar.
	progressMgr, logFile, err := progressbar.InitProgressBarIfPossible()
	if err != nil {
		return err
	}
	if progressMgr != nil {
		progressMgr.IncGeneralProgressTotalBy(1)
		defer logUtils.CloseLogFile(logFile)
		defer progressMgr.Quit()
	}
	log.Info("Downloading plugin: " + pluginName)

	resp, err := client.DownloadFileWithProgress(downloadDetails, "", httputils.HttpClientDetails{}, 3, false, progressMgr)
	if err != nil {
		return err
	}
	log.Debug("Artifactory response: ", resp.Status)
	err = errorutils.CheckResponseStatus(resp, http.StatusOK)
	if err != nil {
		return err
	}
	log.Debug("Plugin downloaded successfully.")
	return os.Chmod(filepath.Join(pluginsDir, exeName), 0777)
}

func getNameAndVersion(requested string) (name, version string, err error) {
	split := strings.Split(requested, "@")
	if len(split) == 1 || (len(split) == 2 && split[1] == "") {
		return split[0], latestVersionName, nil
	}
	if len(split) > 2 {
		return "", "", errors.New("unexpected number of '@' separators in provided argument")
	}
	return split[0], split[1], nil
}

// Get the architecture name corresponding to the architectures that exist in registry.
func getArchitecture() (string, error) {
	switch runtime.GOOS {
	case "windows":
		return "windows-amd64", nil
	case "darwin":
		return "mac-386", nil
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
	}
	return "", errors.New("no compatible plugin architecture was found for the architecture of this machine")
}
