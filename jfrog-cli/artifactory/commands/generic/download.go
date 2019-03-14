package generic

import (
	"errors"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/spec"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	clientutils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"strconv"
)

func Download(downloadSpec *spec.SpecFiles, configuration *utils.DownloadConfiguration) (successCount, failCount int, err error) {

	// Create Service Manager:
	servicesManager, err := utils.CreateDownloadServiceManager(configuration.ArtDetails, configuration)
	if err != nil {
		return 0, 0, err
	}

	// Build Info Collection:
	isCollectBuildInfo := len(configuration.BuildName) > 0 && len(configuration.BuildNumber) > 0
	if isCollectBuildInfo && !configuration.DryRun {
		if err = utils.SaveBuildGeneralDetails(configuration.BuildName, configuration.BuildNumber); err != nil {
			return 0, 0, err
		}
	}

	// Temp Directory:
	if !configuration.DryRun {
		err = fileutils.CreateTempDirPath()
		if err != nil {
			return 0, 0, err
		}
		defer fileutils.RemoveTempDir()
	}

	var errorOccurred = false
	var downloadParamsArray []services.DownloadParams
	// Create DownloadParams for all File-Spec groups.
	for i := 0; i < len(downloadSpec.Files); i++ {
		downParams, err := getDownloadParams(downloadSpec.Get(i), configuration)
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

	// Check for errors.
	if errorOccurred {
		return len(filesInfo), totalExpected - len(filesInfo), errors.New("Download finished with errors, please review the logs.")
	}
	if configuration.DryRun {
		return totalExpected, 0, err
	}
	log.Debug("Downloaded", strconv.Itoa(len(filesInfo)), "artifacts.")

	// Build Info
	if isCollectBuildInfo {
		buildDependencies := convertFileInfoToBuildDependencies(filesInfo)
		populateFunc := func(partial *buildinfo.Partial) {
			partial.Dependencies = buildDependencies
		}
		err = utils.SavePartialBuildInfo(configuration.BuildName, configuration.BuildNumber, populateFunc)
	}

	return len(filesInfo), totalExpected - len(filesInfo), err
}

func convertFileInfoToBuildDependencies(filesInfo []clientutils.FileInfo) []buildinfo.Dependency {
	buildDependecies := make([]buildinfo.Dependency, len(filesInfo))
	for i, fileInfo := range filesInfo {
		dependency := buildinfo.Dependency{Checksum: &buildinfo.Checksum{}}
		dependency.Md5 = fileInfo.Md5
		dependency.Sha1 = fileInfo.Sha1
		// Artifact name in build info as the name in artifactory
		filename, _ := fileutils.GetFileAndDirFromPath(fileInfo.ArtifactoryPath)
		dependency.Id = filename
		buildDependecies[i] = dependency
	}
	return buildDependecies
}

func getDownloadParams(f *spec.File, configuration *utils.DownloadConfiguration) (downParams services.DownloadParams, err error) {
	downParams = services.NewDownloadParams()
	downParams.ArtifactoryCommonParams = f.ToArtifactoryCommonParams()
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

	downParams.Symlink = configuration.Symlink
	downParams.ValidateSymlink = configuration.ValidateSymlink
	downParams.Retries = configuration.Retries
	return
}
