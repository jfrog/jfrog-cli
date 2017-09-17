package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/artifactory/utils"
	clientutils "github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/fileutils"
	"strconv"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/cliutils"
)

func Download(downloadSpec *utils.SpecFiles, flags *DownloadConfiguration) error {
	servicesManager, err := createDownloadServiceManager(flags.ArtDetails, flags)
	if err != nil {
		return err
	}
	isCollectBuildInfo := len(flags.BuildName) > 0 && len(flags.BuildNumber) > 0
	if isCollectBuildInfo && !flags.DryRun {
		if err = utils.SaveBuildGeneralDetails(flags.BuildName, flags.BuildNumber); err != nil {
			return err
		}
	}
	if !flags.DryRun {
		err = fileutils.CreateTempDirPath()
		if err != nil {
			return err
		}
		defer fileutils.RemoveTempDir()
	}
	var filesInfo []clientutils.FileInfo
	for i := 0; i < len(downloadSpec.Files); i++ {
		params, err := downloadSpec.Get(i).ToArtifatoryDownloadParams()
		if err != nil {
			return err
		}
		flat, err := downloadSpec.Get(i).IsFlat(false)
		if err != nil {
			return err
		}
		currentBuildDependencies, err := servicesManager.DownloadFiles(&services.DownloadParamsImpl{ArtifactoryCommonParams: params, ValidateSymlink: flags.ValidateSymlink, Symlink: flags.Symlink, Flat:flat, Retries: flags.Retries})
		if err != nil {
			cliutils.CliLogger.Info("Downloaded", strconv.Itoa(len(filesInfo)), "artifacts.")
			return err
		}
		filesInfo = append(filesInfo, currentBuildDependencies...)
	}
	cliutils.CliLogger.Info("Downloaded", strconv.Itoa(len(filesInfo)), "artifacts.")
	buildDependencies := convertFileInfoToBuildDependencies(filesInfo)
	if isCollectBuildInfo && !flags.DryRun {
		populateFunc := func(tempWrapper *utils.ArtifactBuildInfoWrapper) {
			tempWrapper.Dependencies = buildDependencies
		}
		err = utils.SavePartialBuildInfo(flags.BuildName, flags.BuildNumber, populateFunc)
	}
	return err
}

func convertFileInfoToBuildDependencies(filesInfo []clientutils.FileInfo) []utils.DependenciesBuildInfo {
	buildDependecies := make([]utils.DependenciesBuildInfo, len(filesInfo))
	for i, fileInfo := range filesInfo {
		dependency := new(utils.DependenciesBuildInfo)
		dependency.FileHashes = fileInfo.FileHashes
		filename, _ := fileutils.GetFileAndDirFromPath(fileInfo.ArtifactoryPath)
		dependency.Id = filename
		buildDependecies[i] = *dependency
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
	serviceConfig, err := (&artifactory.ArtifactoryServicesConfigBuilder{}).
		SetArtDetails(artDetails.CreateArtAuthConfig()).
		SetDryRun(flags.DryRun).
		SetCertificatesPath(certPath).
		SetSplitCount(flags.SplitCount).
		SetMinSplitSize(flags.MinSplitSize).
		SetNumOfThreadPerOperation(flags.Threads).
		SetLogger(cliutils.CliLogger).
		Build()
	if err != nil {
		return nil, err
	}
	return artifactory.NewArtifactoryService(serviceConfig)
}