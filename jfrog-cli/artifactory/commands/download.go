package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils"
	clientutils "github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/fileutils"
	"strconv"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils/buildinfo"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils/spec"
	"errors"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
)

func Download(downloadSpec *spec.SpecFiles, flags *DownloadConfiguration) (int, int, error) {
	servicesManager, err := createDownloadServiceManager(flags.ArtDetails, flags)
	if err != nil {
		return 0, 0, err
	}
	isCollectBuildInfo := len(flags.BuildName) > 0 && len(flags.BuildNumber) > 0
	if isCollectBuildInfo && !flags.DryRun {
		if err = utils.SaveBuildGeneralDetails(flags.BuildName, flags.BuildNumber); err != nil {
			return 0, 0, err
		}
	}
	if !flags.DryRun {
		err = fileutils.CreateTempDirPath()
		if err != nil {
			return 0, 0, err
		}
		defer fileutils.RemoveTempDir()
	}
	var filesInfo []clientutils.FileInfo
	var totalExpected int
	var errorOccurred = false
	for i := 0; i < len(downloadSpec.Files); i++ {
		params, err := downloadSpec.Get(i).ToArtifatoryDownloadParams()
		if err != nil {
			errorOccurred = true
			log.Error(err)
			continue
		}
		flat, err := downloadSpec.Get(i).IsFlat(false)
		if err != nil {
			errorOccurred = true
			log.Error(err)
			continue
		}
		currentBuildDependencies, expected, err := servicesManager.DownloadFiles(&services.DownloadParamsImpl{ArtifactoryCommonParams: params, ValidateSymlink: flags.ValidateSymlink, Symlink: flags.Symlink, Flat: flat, Retries: flags.Retries})
		totalExpected += expected
		filesInfo = append(filesInfo, currentBuildDependencies...)
		if err != nil {
			errorOccurred = true
			log.Error(err)
			continue
		}
	}
	if errorOccurred {
		return len(filesInfo), totalExpected - len(filesInfo), errors.New("Download finished with errors. Please review the logs")
	}
	log.Debug("Downloaded", strconv.Itoa(len(filesInfo)), "artifacts.")
	buildDependencies := convertFileInfoToBuildDependencies(filesInfo)
	if isCollectBuildInfo && !flags.DryRun {
		populateFunc := func(partial *buildinfo.Partial) {
			partial.Dependencies = buildDependencies
		}
		err = utils.SavePartialBuildInfo(flags.BuildName, flags.BuildNumber, populateFunc)
	}

	return len(filesInfo), totalExpected - len(filesInfo), err
}

func convertFileInfoToBuildDependencies(filesInfo []clientutils.FileInfo) []buildinfo.Dependencies {
	buildDependecies := make([]buildinfo.Dependencies, len(filesInfo))
	for i, fileInfo := range filesInfo {
		dependency := buildinfo.Dependencies{Checksum: &buildinfo.Checksum{}}
		dependency.Md5 = fileInfo.Md5
		dependency.Sha1 = fileInfo.Sha1
		filename, _ := fileutils.GetFileAndDirFromPath(fileInfo.ArtifactoryPath)
		dependency.Id = filename
		buildDependecies[i] = dependency
	}
	return buildDependecies
}

type DownloadConfiguration struct {
	Threads         int
	SplitCount      int
	MinSplitSize    int64
	BuildName       string
	BuildNumber     string
	DryRun          bool
	Symlink         bool
	ValidateSymlink bool
	ArtDetails      *config.ArtifactoryDetails
	Retries         int
}

func createDownloadServiceManager(artDetails *config.ArtifactoryDetails, flags *DownloadConfiguration) (*artifactory.ArtifactoryServicesManager, error) {
	certPath, err := utils.GetJfrogSecurityDir()
	if err != nil {
		return nil, err
	}
	artAuth, err := artDetails.CreateArtAuthConfig()
	if err != nil {
		return nil, err
	}
	serviceConfig, err := (&artifactory.ArtifactoryServicesConfigBuilder{}).
		SetArtDetails(artAuth).
		SetDryRun(flags.DryRun).
		SetCertificatesPath(certPath).
		SetSplitCount(flags.SplitCount).
		SetMinSplitSize(flags.MinSplitSize).
		SetNumOfThreadPerOperation(flags.Threads).
		SetLogger(log.Logger).
		Build()
	if err != nil {
		return nil, err
	}
	return artifactory.NewArtifactoryService(serviceConfig)
}