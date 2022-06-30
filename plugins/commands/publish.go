package commands

import (
	buildinfoutils "github.com/jfrog/build-info-go/utils"
	"github.com/jfrog/gofrog/io"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/generic"
	commandsutils "github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/utils"
	rtutils "github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/v2/common/spec"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-cli/plugins/commands/utils"
	pluginsutils "github.com/jfrog/jfrog-cli/plugins/utils"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-client-go/http/httpclient"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/urfave/cli"
	"net/http"
	"os"
	"path"
	"path/filepath"
)

const pluginVersionCommandName = "-v"

func PublishCmd(c *cli.Context) error {
	if c.NArg() != 2 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}

	rtDetails, err := getRtDetails(c)
	if err != nil {
		return err
	}

	return runPublishCmd(c.Args().Get(0), c.Args().Get(1), rtDetails)
}

func runPublishCmd(pluginName, pluginVersion string, rtDetails *config.ServerDetails) error {
	err := verifyUniqueVersion(pluginName, pluginVersion, rtDetails)
	if err != nil {
		return err
	}

	return doPublish(pluginName, pluginVersion, rtDetails)
}

// Build and upload the plugin for every supported architecture.
func doPublish(pluginName, pluginVersion string, rtDetails *config.ServerDetails) error {
	tmpDir, err := fileutils.CreateTempDir()
	if err != nil {
		return err
	}

	localArc, err := utils.GetLocalArchitecture()
	if err != nil {
		return err
	}

	arcs, err := getOrderedArchitectures(localArc)
	if err != nil {
		return err
	}

	// Build and upload the plugin for all architectures.
	// Start with the local architecture, to assert versions match before uploading.
	for _, arc := range arcs {
		pluginPath, err := buildPlugin(pluginName, tmpDir, utils.ArchitecturesMap[arc])
		if err != nil {
			return err
		}
		if arc == localArc {
			err = verifyMatchingVersion(pluginPath, pluginVersion)
			if err != nil {
				return err
			}
		}
		err = uploadPlugin(pluginPath, pluginName, pluginVersion, arc, rtDetails)
		if err != nil {
			return err
		}
	}

	return copyToLatestDir(pluginName, pluginVersion, rtDetails)
}

// Returns a slice of all supported architectures names, starting with the local architecture.
// If the local architecture is not supported, abort command.
func getOrderedArchitectures(localArc string) ([]string, error) {
	isLocalArcSupported := false
	orderedSlice := []string{localArc}

	for arc := range utils.ArchitecturesMap {
		if arc == localArc {
			isLocalArcSupported = true
			continue
		}
		orderedSlice = append(orderedSlice, arc)
	}
	if !isLocalArcSupported {
		return nil, errorutils.CheckErrorf("local architecture is not supported. Please run again on a supported machine. Aborting")
	}
	return orderedSlice, nil
}

func verifyMatchingVersion(pluginFullPath, pluginVersion string) error {
	log.Info("Verifying versions matching...")
	err := os.Chmod(pluginFullPath, 0777)
	if err != nil {
		return err
	}
	pluginCmd := pluginsutils.PluginExecCmd{
		ExecPath: pluginFullPath,
		Command:  []string{pluginVersionCommandName},
	}
	output, err := io.RunCmdOutput(&pluginCmd)
	if err != nil {
		return err
	}
	return utils.AssertPluginVersion(output, pluginVersion)
}

func buildPlugin(pluginName, tmpDir string, arc utils.Architecture) (string, error) {
	log.Info("Building plugin for: " + arc.Goos + "-" + arc.Goarch + "...")
	outputPath := filepath.Join(tmpDir, pluginName+arc.FileExtension)
	buildCmd := utils.PluginBuildCmd{
		OutputFullPath: outputPath,
		Env: map[string]string{
			"GOOS":   arc.Goos,
			"GOARCH": arc.Goarch,
		},
	}
	err := io.RunCmd(&buildCmd)
	if err != nil {
		return "", errorutils.CheckError(err)
	}
	return outputPath, nil
}

// Get the Artifactory details corresponding to the server ID provided by env.
func getRtDetails(c *cli.Context) (*config.ServerDetails, error) {
	serverId := os.Getenv(utils.PluginsServerEnv)
	if serverId == "" {
		return nil, cliutils.PrintHelpAndReturnError("the "+utils.PluginsServerEnv+" env var is mandatory for the 'publish' command", c)
	}

	confDetails, err := config.GetSpecificConfig(serverId, false, true)
	if err != nil {
		return nil, err
	}

	confDetails.ArtifactoryUrl = clientutils.AddTrailingSlashIfNeeded(confDetails.ArtifactoryUrl)
	return confDetails, nil
}

// Verify the plugin's provided version does not exist on the plugins server.
func verifyUniqueVersion(pluginName, pluginVersion string, rtDetails *config.ServerDetails) error {
	log.Info("Verifying version uniqueness...")
	client, err := httpclient.ClientBuilder().Build()
	if err != nil {
		return err
	}
	url := clientutils.AddTrailingSlashIfNeeded(rtDetails.ArtifactoryUrl) + utils.GetPluginVersionDirInArtifactory(pluginName, pluginVersion)
	httpDetails := utils.CreatePluginsHttpDetails(rtDetails)

	resp, _, err := client.SendHead(url, httpDetails, "")
	if err != nil {
		return err
	}
	log.Debug("Artifactory response: ", resp.Status)
	if resp.StatusCode == http.StatusOK {
		return errorutils.CheckErrorf("plugin version already exists on server")
	}
	return errorutils.CheckResponseStatus(resp, http.StatusUnauthorized, http.StatusNotFound)
}

func uploadPlugin(pluginLocalPath, pluginName, pluginVersion, arc string, rtDetails *config.ServerDetails) error {
	pluginDirRtPath := utils.GetPluginDirPath(pluginName, pluginVersion, arc)
	log.Info("Upload plugin to: " + pluginDirRtPath + "...")
	// First uploading resources directory (this is the complex part). If the upload is successful, upload the executable file.
	// Upload plugin's resources directory if exists
	exists, err := buildinfoutils.IsDirExists(coreutils.PluginsResourcesDirName, true)
	if err != nil {
		return errorutils.CheckError(err)
	}
	if exists {
		empty, err := fileutils.IsDirEmpty(coreutils.PluginsResourcesDirName)
		if err != nil {
			return err
		}
		if !empty {
			resourcesPattern := filepath.Join(coreutils.PluginsResourcesDirName, "(*)")
			resourcesTargetPath := path.Join(pluginDirRtPath, coreutils.PluginsResourcesDirName+".zip")
			err = uploadPluginsResources(resourcesPattern, resourcesTargetPath, rtDetails)
			if err != nil {
				return err
			}
		}
	}
	// Upload plugin's executable
	execTargetPath := path.Join(pluginDirRtPath, utils.GetPluginExecutableName(pluginName, arc))
	err = uploadPluginsExec(pluginLocalPath, execTargetPath, rtDetails)
	if err != nil {
		return err
	}
	return nil
}

func uploadPluginsExec(pattern, target string, rtDetails *config.ServerDetails) error {
	log.Debug("Upload plugin's executable to: " + target + "...")
	result, err := createAndRunPluginsExecUploadCommand(pattern, target, rtDetails)
	if err != nil {
		return err
	}
	if result.SuccessCount() == 0 {
		return errorutils.CheckErrorf("plugin's executable upload failed as no files were affected. Verify source path is valid")
	}
	if result.SuccessCount() > 1 {
		return errorutils.CheckErrorf("while uploading plugin's executable more than one file was uploaded. Unexpected behaviour, aborting")
	}
	return nil
}

func uploadPluginsResources(pattern, target string, rtDetails *config.ServerDetails) error {
	log.Debug("Upload plugin's resources to: " + target + "...")
	result, err := createAndRunPluginsResourcesUploadCommand(pattern, target, rtDetails)
	if err != nil {
		return err
	}
	if result.SuccessCount() == 0 {
		return errorutils.CheckErrorf("plugin's resources upload failed as no files were affected. Verify source path is valid")
	}
	if result.SuccessCount() > 1 {
		return errorutils.CheckErrorf("while zipping and uploading plugin's resources directory more than one file was uploaded. Unexpected behaviour, aborting")
	}
	return nil
}

func createAndRunPluginsExecUploadCommand(pattern, target string, rtDetails *config.ServerDetails) (*commandsutils.Result, error) {
	uploadCmd := generic.NewUploadCommand()
	uploadCmd.SetUploadConfiguration(createUploadConfiguration()).
		SetServerDetails(rtDetails).
		SetSpec(createExecUploadSpec(pattern, target))
	err := uploadCmd.Run()
	if err != nil {
		return nil, err
	}
	return uploadCmd.Result(), nil
}

func createAndRunPluginsResourcesUploadCommand(pattern, target string, rtDetails *config.ServerDetails) (*commandsutils.Result, error) {
	uploadCmd := generic.NewUploadCommand()
	uploadCmd.SetUploadConfiguration(createUploadConfiguration()).
		SetServerDetails(rtDetails).
		SetSpec(createResourcesUploadSpec(pattern, target))
	err := uploadCmd.Run()
	if err != nil {
		return nil, err
	}
	return uploadCmd.Result(), nil
}

// Copy the uploaded version to override latest dir.
func copyToLatestDir(pluginName, pluginVersion string, rtDetails *config.ServerDetails) error {
	log.Info("Copying version to latest dir...")

	copyCmd := generic.NewCopyCommand()
	copyCmd.SetServerDetails(rtDetails).SetSpec(createCopySpec(pluginName, pluginVersion))
	return copyCmd.Run()
}

func createCopySpec(pluginName, pluginVersion string) *spec.SpecFiles {
	pluginsRepo := utils.GetPluginsRepo()
	return spec.NewBuilder().
		Pattern(path.Join(pluginsRepo, pluginName, pluginVersion, "(*)")).
		Target(path.Join(pluginsRepo, pluginName, utils.LatestVersionName, "{1}")).
		Flat(true).
		Recursive(true).
		IncludeDirs(true).
		BuildSpec()
}

func createExecUploadSpec(source, target string) *spec.SpecFiles {
	return spec.NewBuilder().
		Pattern(source).
		Target(target).
		BuildSpec()
}

// Resources directory is being uploaded to Artifactory in a zip file.
func createResourcesUploadSpec(source, target string) *spec.SpecFiles {
	return spec.NewBuilder().
		Pattern(source).
		Target(target).
		Archive("zip").
		Recursive(true).
		TargetPathInArchive("{1}").
		BuildSpec()
}

func createUploadConfiguration() *rtutils.UploadConfiguration {
	uploadConfiguration := new(rtutils.UploadConfiguration)
	uploadConfiguration.Threads = cliutils.Threads
	return uploadConfiguration
}
