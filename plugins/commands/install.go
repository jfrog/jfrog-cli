package commands

import (
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	commandsUtils "github.com/jfrog/jfrog-cli/plugins/commands/utils"
	clientUtils "github.com/jfrog/jfrog-client-go/utils"

	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	logUtils "github.com/jfrog/jfrog-cli/utils/log"
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

	pluginRtDirPath, executableName, err := getRequiredPluginRtPathInformation(pluginName, version)
	if err != nil {
		return err
	}
	execDownloadUrl := clientUtils.AddTrailingSlashIfNeeded(url) + pluginRtDirPath + "/" + executableName

	should, err := shouldDownloadPlugin(pluginsDir, pluginName, execDownloadUrl, commandsUtils.CreatePluginsHttpDetails(&serverDetails))
	if err != nil {
		return err
	}
	if !should {
		return errorutils.CheckErrorf("the plugin with the requested version already exists locally")
	}

	return downloadPlugin(pluginsDir, pluginName, pluginRtDirPath, commandsUtils.CreatePluginsHttpDetails(&serverDetails))
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
	equal, err := fileutils.IsEqualToLocalFile(filepath.Join(pluginsDir, pluginName, "bin", pluginName), details.Checksum.Md5, details.Checksum.Sha1)
	return !equal, err
}

// Returns information of the plugin, corresponding to the local architecture.
// 	pluginDirRtPath - path of the plugin's directory in registry.
// 	executableName - name of the plugin's executable in registry.
func getRequiredPluginRtPathInformation(pluginName, version string) (pluginDirRtPath, executableName string, err error) {
	arc, err := commandsUtils.GetLocalArchitecture()
	if err != nil {
		return "", "", err
	}
	pluginDirRtPath, executableName = commandsUtils.GetPluginPathDetailsInArtifactory(pluginName, version, arc)
	return
}

func createPluginsDir(pluginsDir string) error {
	return os.MkdirAll(pluginsDir, 0777)
}

func downloadPlugin(pluginsDir, pluginName, downloadUrl string, httpDetails httputils.HttpClientDetails) (err error) {
	exeName := commandsUtils.GetLocalPluginExecutableName(pluginName)
	log.Debug("Downloading plugin from: ", downloadUrl)
	downloadDetails := &httpclient.DownloadFileDetails{
		FileName:      pluginName,
		DownloadPath:  downloadUrl + "?archiveType=zip",
		LocalPath:     pluginsDir,
		LocalFileName: exeName,
		RelativePath:  exeName,
	}

	client, err := httpclient.ClientBuilder().Build()
	if err != nil {
		return
	}
	// Init progress bar.
	progressMgr, logFile, err := progressbar.InitProgressBarIfPossible(false)
	if err != nil {
		return
	}
	if progressMgr != nil {
		progressMgr.InitProgressReaders()
		progressMgr.IncGeneralProgressTotalBy(1)
		defer func() {
			progressMgr.Quit()
			e := logUtils.CloseLogFile(logFile)
			if err == nil {
				err = e
			}
		}()
	}
	log.Info("Downloading plugin: " + pluginName)

	resp, err := client.DownloadFileWithProgress(downloadDetails, "", httpDetails, false, progressMgr)
	if err != nil {
		return
	}
	log.Debug("Artifactory response: ", resp.Status)
	err = errorutils.CheckResponseStatus(resp, http.StatusOK)
	if err != nil {
		return
	}
	log.Debug("Plugin downloaded successfully.")
	err = os.Chmod(filepath.Join(pluginsDir, exeName), 0777)
	return
}

func downloadPlugin2(pluginsDir, pluginName, pluginRtDirPath string, serverDetails *config.ServerDetails) (err error) {
	exeName := commandsUtils.GetLocalPluginExecutableName(pluginName)
	log.Debug("Downloading plugin from: ", pluginRtDirPath)

	// Init progress bar.
	progressMgr, logFile, err := progressbar.InitProgressBarIfPossible(false)
	if err != nil {
		return
	}
	if progressMgr != nil {
		progressMgr.InitProgressReaders()
		progressMgr.IncGeneralProgressTotalBy(1)
		defer func() {
			progressMgr.Quit()
			e := logUtils.CloseLogFile(logFile)
			if err == nil {
				err = e
			}
		}()
	}
	// Create Service Manager:
	servicesManager, err := utils.CreateDownloadServiceManager(serverDetails, 3, 0, 0, false, progressMgr)
	if err != nil {
		return err
	}
	log.Info("Downloading plugin: " + pluginName)

	downloadParams := services.NewDownloadParams()
	downloadParams.CommonParams.Pattern = pluginRtDirPath + "/*"
	downloadParams.CommonParams.Target = filepath.Join(pluginsDir, pluginName)
	summary, err := servicesManager.DownloadFilesWithSummary(downloadParams)
	if err != nil {
		log.Error(err)
	}
	if summary.TotalFailed != 0 {
		log.Info("Failed downloading plugin: " + pluginName)
		log.Debug("Total Succeeded: ", summary.TotalSucceeded)
		log.Debug("Total Failed: ", summary.TotalFailed)
	}
	log.Debug("Plugin downloaded successfully.")
	// TODO: chmod to directory
	err = os.Chmod(filepath.Join(pluginsDir, pluginName, "bin", exeName), 0777)
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
