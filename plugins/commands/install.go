package commands

import (
	ioutils "github.com/jfrog/jfrog-client-go/utils/io"
	"github.com/mholt/archiver/v3"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/plugins"
	commandsUtils "github.com/jfrog/jfrog-cli/plugins/commands/utils"
	clientUtils "github.com/jfrog/jfrog-client-go/utils"

	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-cli/utils/progressbar"
	"github.com/jfrog/jfrog-client-go/http/httpclient"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/io/httputils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/urfave/cli"
)

func InstallCmd(c *cli.Context) error {
	if c.NArg() != 1 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}
	err := assertValidEnv(c)
	if err != nil {
		return err
	}
	err = plugins.CheckPluginsVersionAndConvertIfNeeded()
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

	url, serverDetails, err := getServerDetails()
	if err != nil {
		return err
	}

	pluginRtDirPath, err := getRequiredPluginRtDirPath(pluginName, version)
	if err != nil {
		return err
	}
	execDownloadUrl := clientUtils.AddTrailingSlashIfNeeded(url) + pluginRtDirPath + "/"

	should, err := shouldDownloadPlugin(pluginsDir, pluginName, execDownloadUrl, commandsUtils.CreatePluginsHttpDetails(&serverDetails))
	if err != nil {
		return err
	}
	if !should {
		return errorutils.CheckErrorf("the plugin with the requested version already exists locally")
	}

	return downloadPlugin(pluginsDir, pluginName, execDownloadUrl, commandsUtils.CreatePluginsHttpDetails(&serverDetails))
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
func getServerDetails() (string, config.ServerDetails, error) {
	serverId := os.Getenv(commandsUtils.PluginsServerEnv)
	if serverId == "" {
		return commandsUtils.PluginsOfficialRegistryUrl, config.ServerDetails{ArtifactoryUrl: commandsUtils.PluginsOfficialRegistryUrl}, nil
	}

	rtDetails, err := config.GetSpecificConfig(serverId, false, true)
	if err != nil {
		return "", config.ServerDetails{}, err
	}
	return rtDetails.ArtifactoryUrl, *rtDetails, nil
}

// Checks if the requested plugin exists in registry and does not exists locally.
func shouldDownloadPlugin(pluginsDir, pluginName, downloadUrl string, httpDetails httputils.HttpClientDetails) (bool, error) {
	exists, err := fileutils.IsDirExists(filepath.Join(pluginsDir, pluginName), false)
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
	equal, err := fileutils.IsEqualToLocalFile(filepath.Join(pluginsDir, pluginName, coreutils.PluginsExecDirName, plugins.GetLocalPluginExecutableName(pluginName)), details.Checksum.Md5, details.Checksum.Sha1)
	return !equal, err
}

// Returns the path of the JFrog CLI plugin's directory in registry, corresponding to the local architecture.
func getRequiredPluginRtDirPath(pluginName, version string) (pluginDirRtPath string, err error) {
	arc, err := commandsUtils.GetLocalArchitecture()
	if err != nil {
		return
	}
	pluginDirRtPath = commandsUtils.GetPluginDirPath(pluginName, version, arc)
	return
}

func createPluginsDir(pluginsDir string) error {
	err := os.MkdirAll(pluginsDir, 0777)
	if err != nil {
		return errorutils.CheckError(err)
	}
	_, err = plugins.CreatePluginsConfigFile()
	return err
}

func downloadPlugin(pluginsDir, pluginName, downloadUrl string, httpDetails httputils.HttpClientDetails) (err error) {
	// Init progress bar.
	progressMgr, err := progressbar.InitFilesProgressBarIfPossible(true)
	if err != nil {
		return
	}
	if progressMgr != nil {
		progressMgr.InitProgressReaders()
		progressMgr.IncGeneralProgressTotalBy(1)
		defer func() {
			e := progressMgr.Quit()
			if err == nil {
				err = e
			}
		}()
	}

	err = downloadPluginExec(downloadUrl, pluginName, pluginsDir, httpDetails, progressMgr)
	if err != nil {
		return
	}
	err = downloadPluginsResources(downloadUrl, pluginName, pluginsDir, httpDetails, progressMgr)
	if err != nil {
		return
	}
	log.Info("Plugin downloaded successfully.")
	return
}

func getNameAndVersion(requested string) (name, version string, err error) {
	split := strings.Split(requested, "@")
	if len(split) == 1 || (len(split) == 2 && split[1] == "") {
		return split[0], commandsUtils.LatestVersionName, nil
	}
	if len(split) > 2 {
		return "", "", errorutils.CheckErrorf("unexpected number of '@' separators in provided argument")
	}
	return split[0], split[1], nil
}

func downloadPluginExec(downloadUrl, pluginName, pluginsDir string, httpDetails httputils.HttpClientDetails, progressMgr ioutils.ProgressMgr) (err error) {
	exeName := plugins.GetLocalPluginExecutableName(pluginName)
	downloadDetails := &httpclient.DownloadFileDetails{
		FileName:      pluginName,
		DownloadPath:  clientUtils.AddTrailingSlashIfNeeded(downloadUrl) + exeName,
		LocalPath:     filepath.Join(pluginsDir, pluginName, coreutils.PluginsExecDirName),
		LocalFileName: exeName,
		RelativePath:  exeName,
	}
	log.Debug("Downloading plugin's executable from: ", downloadDetails.DownloadPath)
	_, err = downloadFromArtifactory(downloadDetails, httpDetails, progressMgr)
	if err != nil {
		return
	}
	err = os.Chmod(filepath.Join(downloadDetails.LocalPath, downloadDetails.LocalFileName), 0777)
	if errorutils.CheckError(err) != nil {
		return
	}
	log.Debug("Plugin's executable downloaded successfully.")
	return
}

func downloadPluginsResources(downloadUrl, pluginName, pluginsDir string, httpDetails httputils.HttpClientDetails, progressMgr ioutils.ProgressMgr) (err error) {
	downloadDetails := &httpclient.DownloadFileDetails{
		FileName:      pluginName,
		DownloadPath:  clientUtils.AddTrailingSlashIfNeeded(downloadUrl) + coreutils.PluginsResourcesDirName + ".zip",
		LocalPath:     filepath.Join(pluginsDir, pluginName),
		LocalFileName: coreutils.PluginsResourcesDirName + ".zip",
		RelativePath:  coreutils.PluginsResourcesDirName + ".zip",
	}
	log.Debug("Downloading plugin's resources from: ", downloadDetails.DownloadPath)
	statusCode, err := downloadFromArtifactory(downloadDetails, httpDetails, progressMgr)
	if err != nil {
		return
	}
	if statusCode == http.StatusNotFound {
		log.Debug("No resources were downloaded.")
		return nil
	}
	err = archiver.Unarchive(filepath.Join(downloadDetails.LocalPath, downloadDetails.LocalFileName), filepath.Join(downloadDetails.LocalPath, coreutils.PluginsResourcesDirName)+string(os.PathSeparator))
	if errorutils.CheckError(err) != nil {
		return
	}
	err = os.Remove(filepath.Join(downloadDetails.LocalPath, downloadDetails.LocalFileName))
	if err != nil {
		return
	}
	err = coreutils.ChmodPluginsDirectoryContent()
	if errorutils.CheckError(err) != nil {
		return
	}
	log.Debug("Plugin's resources downloaded successfully.")
	return
}

func downloadFromArtifactory(downloadDetails *httpclient.DownloadFileDetails, httpDetails httputils.HttpClientDetails, progressMgr ioutils.ProgressMgr) (statusCode int, err error) {
	client, err := httpclient.ClientBuilder().Build()
	if err != nil {
		return
	}
	log.Info("Downloading: " + downloadDetails.FileName)
	resp, err := client.DownloadFileWithProgress(downloadDetails, "", httpDetails, false, progressMgr)
	if err != nil {
		return
	}
	statusCode = resp.StatusCode
	log.Debug("Artifactory response: ", statusCode)
	err = errorutils.CheckResponseStatus(resp, http.StatusOK, http.StatusNotFound)
	return
}
