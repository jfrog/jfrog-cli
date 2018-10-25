package generic

import (
	"errors"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/spec"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	clientutils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"strconv"
)

func Download(downloadSpec *spec.SpecFiles, configuration *DownloadConfiguration) (successCount, failCount int, err error) {
	servicesManager, err := createDownloadServiceManager(configuration.ArtDetails, configuration)
	if err != nil {
		return 0, 0, err
	}
	isCollectBuildInfo := len(configuration.BuildName) > 0 && len(configuration.BuildNumber) > 0
	if isCollectBuildInfo && !configuration.DryRun {
		if err = utils.SaveBuildGeneralDetails(configuration.BuildName, configuration.BuildNumber); err != nil {
			return 0, 0, err
		}
	}
	if !configuration.DryRun {
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

		explode, err := downloadSpec.Get(i).IsExplode(false)
		if err != nil {
			errorOccurred = true
			log.Error(err)
			continue
		}

		currentBuildDependencies, expected, err := servicesManager.DownloadFiles(&services.DownloadParamsImpl{ArtifactoryCommonParams: params, ValidateSymlink: configuration.ValidateSymlink, Symlink: configuration.Symlink, Flat: flat, Explode: explode, Retries: configuration.Retries})
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
	if configuration.DryRun {
		return totalExpected, 0, err
	}
	log.Debug("Downloaded", strconv.Itoa(len(filesInfo)), "artifacts.")
	buildDependencies := convertFileInfoToBuildDependencies(filesInfo)
	if isCollectBuildInfo {
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
	serviceConfig, err := artifactory.NewConfigBuilder().
		SetArtDetails(artAuth).
		SetDryRun(flags.DryRun).
		SetCertificatesPath(certPath).
		SetSplitCount(flags.SplitCount).
		SetMinSplitSize(flags.MinSplitSize).
		SetThreads(flags.Threads).
		SetLogger(log.Logger).
		Build()
	if err != nil {
		return nil, err
	}
	return artifactory.New(serviceConfig)
}
