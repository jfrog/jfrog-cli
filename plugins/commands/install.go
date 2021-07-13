package commands

import (
	"errors"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	commandsUtils "github.com/jfrog/jfrog-cli/plugins/commands/utils"
	clientUtils "github.com/jfrog/jfrog-client-go/utils"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	logUtils "github.com/jfrog/jfrog-cli/utils/log"
	"github.com/jfrog/jfrog-cli/utils/progressbar"
	"github.com/jfrog/jfrog-client-go/http/httpclient"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/io/httputils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

func InstallCmd(c *cli.Context) error {
	if c.NArg() != 1 {
		return cliutils.PrintHelpAndReturnError("Wrong number of arguments.", c)
	}
	err := assertValidEnv(c)
	if err != nil {
		return err
	}
	return runInstallCmd(c.Args().Get(0))
}

func runInstallCmd(requestedPlugin string) error {
	pluginName, version, err := getNameAndVersion(requestedPlugin)
	if err != nil {
		return err
	}

	pluginsDir, err := createPluginsDirIfNeeded()
	if err != nil {
		return err
	}

	url, httpDetails, err := getServerDetails()
	if err != nil {
		return err
	}

	pluginRtPath, err := getRequiredPluginRtPath(pluginName, version)
	if err != nil {
		return err
	}
	downloadUrl := clientUtils.AddTrailingSlashIfNeeded(url) + pluginRtPath

	should, err := shouldDownloadPlugin(pluginsDir, pluginName, downloadUrl, httpDetails)
	if err != nil {
		return err
	}
	if !should {
		return errorutils.CheckError(errors.New("the plugin with the requested version already exists locally"))
	}

	return downloadPlugin(pluginsDir, pluginName, downloadUrl, httpDetails)
}

// Assert repo env is not passed without server env.
func assertValidEnv(c *cli.Context) error {
	repo := os.Getenv(commandsUtils.PluginsRepoEnv)
	serverId := os.Getenv(commandsUtils.PluginsServerEnv)
	if repo != "" && serverId == "" {
		return cliutils.PrintHelpAndReturnError(commandsUtils.PluginsRepoEnv+" should not be provided without "+commandsUtils.PluginsServerEnv, c)
	}
	return nil
}

func createPluginsDirIfNeeded() (string, error) {
	pluginsDir, err := coreutils.GetJfrogPluginsDir()
	if err != nil {
		return "", err
	}

	exists, err := fileutils.IsDirExists(pluginsDir, false)
	if err != nil {
		return "", err
	}

	if exists {
		return pluginsDir, nil
	}

	err = createPluginsDir(pluginsDir)
	if err != nil {
		return "", err
	}
	return pluginsDir, nil
}

// Use the server ID if provided, else use the official registry.
func getServerDetails() (string, httputils.HttpClientDetails, error) {
	serverId := os.Getenv(commandsUtils.PluginsServerEnv)
	if serverId == "" {
		return commandsUtils.PluginsOfficialRegistryUrl, httputils.HttpClientDetails{}, nil
	}

	rtDetails, err := config.GetSpecificConfig(serverId, false, true)
	if err != nil {
		return "", httputils.HttpClientDetails{}, err
	}
	return rtDetails.ArtifactoryUrl, commandsUtils.CreatePluginsHttpDetails(rtDetails), nil
}

// Checks if the requested plugin exists in registry and does not exists locally.
func shouldDownloadPlugin(pluginsDir, pluginName, downloadUrl string, httpDetails httputils.HttpClientDetails) (bool, error) {
	exists, err := fileutils.IsFileExists(filepath.Join(pluginsDir, commandsUtils.GetLocalPluginExecutableName(pluginName)), false)
	if err != nil {
		return false, err
	}
	if !exists {
		return true, nil
	}
	log.Debug("Verifying plugin download is needed...")
	client, err := httpclient.ClientBuilder().Build()
	if err != nil {
		return false, err
	}
	log.Debug("Fetching plugin details from: ", downloadUrl)

	details, resp, err := client.GetRemoteFileDetails(downloadUrl, httpDetails)
	if err != nil {
		return false, err
	}
	log.Debug("Artifactory response: ", resp.Status)
	err = errorutils.CheckResponseStatus(resp, http.StatusOK)
	if err != nil {
		return false, err
	}
	equal, err := fileutils.IsEqualToLocalFile(filepath.Join(pluginsDir, pluginName), details.Checksum.Md5, details.Checksum.Sha1)
	return !equal, err
}

// Returns the path of the plugin's executable in registry, corresponding to the local architecture.
func getRequiredPluginRtPath(pluginName, version string) (string, error) {
	arc, err := commandsUtils.GetLocalArchitecture()
	if err != nil {
		return "", err
	}
	return commandsUtils.GetPluginPathInArtifactory(pluginName, version, arc), nil
}

func createPluginsDir(pluginsDir string) error {
	return os.MkdirAll(pluginsDir, 0777)
}

func downloadPlugin(pluginsDir, pluginName, downloadUrl string, httpDetails httputils.HttpClientDetails) error {
	exeName := commandsUtils.GetLocalPluginExecutableName(pluginName)
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

	resp, err := client.DownloadFileWithProgress(downloadDetails, "", httpDetails, false, progressMgr)
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
		return split[0], commandsUtils.LatestVersionName, nil
	}
	if len(split) > 2 {
		return "", "", errorutils.CheckError(errors.New("unexpected number of '@' separators in provided argument"))
	}
	return split[0], split[1], nil
}
