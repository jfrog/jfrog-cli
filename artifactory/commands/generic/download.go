package generic

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jfrog/jfrog-cli/artifactory/spec"
	"github.com/jfrog/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-cli/utils/progressbar"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	clientutils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	ioUtils "github.com/jfrog/jfrog-client-go/utils/io"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

type DownloadCommand struct {
	buildConfiguration *utils.BuildConfiguration
	GenericCommand
	configuration *utils.DownloadConfiguration
	logFile       *os.File
}

func NewDownloadCommand() *DownloadCommand {
	return &DownloadCommand{GenericCommand: *NewGenericCommand()}
}

func (dc *DownloadCommand) LogFile() *os.File {
	return dc.logFile
}

func (dc *DownloadCommand) SetBuildConfiguration(buildConfiguration *utils.BuildConfiguration) *DownloadCommand {
	dc.buildConfiguration = buildConfiguration
	return dc
}

func (dc *DownloadCommand) Configuration() *utils.DownloadConfiguration {
	return dc.configuration
}

func (dc *DownloadCommand) SetConfiguration(configuration *utils.DownloadConfiguration) *DownloadCommand {
	dc.configuration = configuration
	return dc
}

func (dc *DownloadCommand) CommandName() string {
	return "rt_download"
}

func (dc *DownloadCommand) Run() error {
	if dc.SyncDeletesPath() != "" && !dc.Quiet() && !cliutils.InteractiveConfirm("Sync-deletes may delete some files in your local file system. Are you sure you want to continue?\n"+
		"You can avoid this confirmation message by adding --quiet to the command.") {
		return nil
	}
	// Initialize Progress bar, set logger to a log file
	var err error
	var progressBar ioUtils.Progress
	progressBar, dc.logFile, err = progressbar.InitProgressBarIfPossible()
	if err != nil {
		return err
	}
	if progressBar != nil {
		defer progressBar.Quit()
	}

	// Create Service Manager:
	servicesManager, err := utils.CreateDownloadServiceManager(dc.rtDetails, dc.configuration.Threads, dc.DryRun(), progressBar)
	if err != nil {
		return err
	}

	// Build Info Collection:
	isCollectBuildInfo := len(dc.buildConfiguration.BuildName) > 0 && len(dc.buildConfiguration.BuildNumber) > 0
	if isCollectBuildInfo && !dc.DryRun() {
		if err = utils.SaveBuildGeneralDetails(dc.buildConfiguration.BuildName, dc.buildConfiguration.BuildNumber); err != nil {
			return err
		}
	}

	var errorOccurred = false
	var downloadParamsArray []services.DownloadParams
	// Create DownloadParams for all File-Spec groups.
	for i := 0; i < len(dc.Spec().Files); i++ {
		downParams, err := getDownloadParams(dc.Spec().Get(i), dc.configuration)
		if err != nil {
			errorOccurred = true
			log.Error(err)
			continue
		}
		downloadParamsArray = append(downloadParamsArray, downParams)
	}
	// Perform download.
	filesInfo, totalExpected, err := servicesManager.DownloadFiles(downloadParamsArray...)
	if err != nil {
		errorOccurred = true
		log.Error(err)
	}

	dc.result.SetSuccessCount(len(filesInfo))
	dc.result.SetFailCount(totalExpected - len(filesInfo))
	// Check for errors.
	if errorOccurred {
		return errors.New("Download finished with errors, please review the logs.")
	}
	if dc.DryRun() {
		dc.result.SetSuccessCount(totalExpected)
		dc.result.SetFailCount(0)
		return err
	} else if dc.SyncDeletesPath() != "" {
		absSyncDeletesPath, err := filepath.Abs(dc.SyncDeletesPath())
		if err != nil {
			return errorutils.CheckError(err)
		}
		if _, err = os.Stat(absSyncDeletesPath); err == nil {
			walkFn := createSyncDeletesWalkFunction(filesInfo)
			err = fileutils.Walk(dc.SyncDeletesPath(), walkFn, false)
			if err != nil {
				return errorutils.CheckError(err)
			}
		} else if os.IsNotExist(err) {
			log.Info("Sync-deletes path", absSyncDeletesPath, "does not exists.")
		}
	}
	log.Debug("Downloaded", strconv.Itoa(len(filesInfo)), "artifacts.")

	// Build Info
	if isCollectBuildInfo {
		buildDependencies := convertFileInfoToBuildDependencies(filesInfo)
		populateFunc := func(partial *buildinfo.Partial) {
			partial.Dependencies = buildDependencies
			partial.ModuleId = dc.buildConfiguration.Module
		}
		err = utils.SavePartialBuildInfo(dc.buildConfiguration.BuildName, dc.buildConfiguration.BuildNumber, populateFunc)
	}

	return err
}

func convertFileInfoToBuildDependencies(filesInfo []clientutils.FileInfo) []buildinfo.Dependency {
	buildDependencies := make([]buildinfo.Dependency, len(filesInfo))
	for i, fileInfo := range filesInfo {
		dependency := buildinfo.Dependency{Checksum: &buildinfo.Checksum{}}
		dependency.Md5 = fileInfo.Md5
		dependency.Sha1 = fileInfo.Sha1
		// Artifact name in build info as the name in artifactory
		filename, _ := fileutils.GetFileAndDirFromPath(fileInfo.ArtifactoryPath)
		dependency.Id = filename
		buildDependencies[i] = dependency
	}
	return buildDependencies
}

func getDownloadParams(f *spec.File, configuration *utils.DownloadConfiguration) (downParams services.DownloadParams, err error) {
	downParams = services.NewDownloadParams()
	downParams.ArtifactoryCommonParams = f.ToArtifactoryCommonParams()
	downParams.Symlink = configuration.Symlink
	downParams.MinSplitSize = configuration.MinSplitSize
	downParams.SplitCount = configuration.SplitCount
	downParams.Retries = configuration.Retries

	downParams.Recursive, err = f.IsRecursive(true)
	if err != nil {
		return
	}

	downParams.IncludeDirs, err = f.IsIncludeDirs(false)
	if err != nil {
		return
	}

	downParams.Flat, err = f.IsFlat(false)
	if err != nil {
		return
	}

	downParams.Explode, err = f.IsExplode(false)
	if err != nil {
		return
	}

	downParams.ValidateSymlink, err = f.IsVlidateSymlinks(false)
	if err != nil {
		return
	}

	return
}

func createSyncDeletesWalkFunction(downloadedFiles []clientutils.FileInfo) fileutils.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		// Convert path to absolute path
		path, err = filepath.Abs(path)
		if errorutils.CheckError(err) != nil {
			return err
		}
		// Go over the downloaded files list
		for _, file := range downloadedFiles {
			// If the current path is a prefix of a downloaded file - we won't delete it.
			fileAbsPath, err := filepath.Abs(file.LocalPath)
			if errorutils.CheckError(err) != nil {
				return err
			}
			if strings.HasPrefix(fileAbsPath, path) {
				return nil
			}
		}
		// The current path is not a prefix of any downloaded file so it should be deleted
		log.Info("Deleting:", path)
		if info.IsDir() {
			// If current path is a dir - remove all content and return SkipDir to stop walking this path
			err = os.RemoveAll(path)
			if err == nil {
				return fileutils.SkipDir
			}
		} else {
			// Path is a file
			err = os.Remove(path)
		}

		return errorutils.CheckError(err)
	}
}
