package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/types"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	"sync"
	"strconv"
	"errors"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/logger"
	"path"
)

func Download(downloadSpec *utils.SpecFiles, flags *DownloadFlags) (err error) {
	err = utils.PreCommandSetup(flags)
	if err != nil {
		return
	}

	isCollectBuildInfo := len(flags.BuildName) > 0  && len(flags.BuildNumber) > 0
	if isCollectBuildInfo && !flags.DryRun {
		if err = utils.SaveBuildGeneralDetails(flags.BuildName, flags.BuildNumber); err != nil {
			return
		}
	}
	if !flags.DryRun {
		err = ioutils.CreateTempDirPath()
		if err != nil {
			return
		}
		defer ioutils.RemoveTempDir()
	}

	buildDependecies := make([][]utils.DependenciesBuildInfo, flags.Threads)
	var downloadData []DownloadData
	for i := 0; i < len(downloadSpec.Files); i++ {
		var partialDownloadData []DownloadData
		switch downloadSpec.Get(i).GetSpecType() {
		case utils.WILDCARD, utils.SIMPLE:
			partialDownloadData, err = collectWildcardDependecies(downloadSpec.Get(i), flags)
		case utils.AQL:
			partialDownloadData, err = collectAqlDependecies(downloadSpec.Get(i), flags)
		}
		if err != nil {
			return
		}
		downloadData = append(downloadData, partialDownloadData...)
	}
	buildDependecies, err = downloadAqlResult(downloadData, flags)
	if err != nil {
		return
	}
	if isCollectBuildInfo && !flags.DryRun {
		populateFunc := func(tempWrapper *utils.ArtifactBuildInfoWrapper) {
			tempWrapper.Dependencies = stripThreadIdFromBuildInfoDependencies(buildDependecies)
		}
		err = utils.PrepareBuildInfoForSave(flags.BuildName, flags.BuildNumber, populateFunc)
	}
	return
}

func stripThreadIdFromBuildInfoDependencies(dependenciesBuildInfo [][]utils.DependenciesBuildInfo) []utils.DependenciesBuildInfo {
	var buildInfo []utils.DependenciesBuildInfo
	for _, v := range dependenciesBuildInfo {
		buildInfo = append(buildInfo, v...)
	}
	return buildInfo
}

func collectWildcardDependecies(fileSpec *utils.Files, flags *DownloadFlags) ([]DownloadData, error) {
	isRecursive, err := cliutils.StringToBool(fileSpec.Recursive, true)
	if err != nil {
		return nil, err
	}

	resultItems, err := utils.AqlSearchDefaultReturnFields(fileSpec.Pattern, isRecursive, fileSpec.Props, flags)
	if err != nil {
		return nil, err
	}

	return createDownloadDataList(resultItems, fileSpec)
}

func collectAqlDependecies(fileSpec *utils.Files, flags *DownloadFlags) ([]DownloadData, error) {
	resultItems, err := utils.AqlSearchBySpec(fileSpec.Aql, flags)
	if err != nil {
		return nil, err
	}
	return createDownloadDataList(resultItems, fileSpec)
}

func createDownloadDataList(items []utils.AqlSearchResultItem, fileSpec *utils.Files) ([]DownloadData, error) {
	var dependencies []DownloadData
	for _, v := range items {
		flat, err := cliutils.StringToBool(fileSpec.Flat, false)
		if err != nil {
			return dependencies, err
		}
		dependencies = append(dependencies, DownloadData{
			Dependency: v,
			DownloadPath: fileSpec.Pattern,
			Target: fileSpec.Target,
			Flat: flat,
		})
	}
	return dependencies, nil
}

func getLocalPathAndFile(originalFileName, relativePath, targetPath string, flat bool) (localTargetPath, fileName string) {
	targetFileName, targetDirPath := ioutils.GetFileAndDirFromPath(targetPath)

	localTargetPath = targetDirPath
	if !flat {
		localTargetPath = path.Join(targetDirPath, relativePath)
	}

	fileName = originalFileName
	if targetFileName != "" {
		fileName = targetFileName
	}
	return
}

func downloadAqlResult(dependecies []DownloadData, flags *DownloadFlags) (buildDependencies [][]utils.DependenciesBuildInfo, err error) {
	size := len(dependecies)
	buildDependencies = make([][]utils.DependenciesBuildInfo, flags.Threads)
	var wg sync.WaitGroup
	for i := 0; i < flags.Threads; i++ {
		wg.Add(1)
		go func(threadId int) {
			logMsgPrefix := cliutils.GetLogMsgPrefix(threadId, flags.DryRun)
			for j := threadId; j < size && err == nil; j += flags.Threads {
				downloadPath, e := utils.BuildArtifactoryUrl(flags.ArtDetails.Url, dependecies[j].Dependency.GetFullUrl(), make(map[string]string))
				if e != nil {
					err = e
					break
				}
				logger.Logger.Info(logMsgPrefix + "Downloading " + dependecies[j].Dependency.GetFullUrl())
				if flags.DryRun {
					continue
				}

				regexpPattern := cliutils.PathToRegExp(dependecies[j].DownloadPath)
				placeHolderTarget, e := cliutils.ReformatRegexp(regexpPattern, dependecies[j].Dependency.GetFullUrl(), dependecies[j].Target)
				if e != nil {
					err = e
					break
				}
				localPath, localFileName := getLocalPathAndFile(dependecies[j].Dependency.Name, dependecies[j].Dependency.Path, placeHolderTarget, dependecies[j].Flat)
				shouldDownload, e := shouldDownloadFile(path.Join(localPath, dependecies[j].Dependency.Name), dependecies[j].Dependency.Actual_Md5, dependecies[j].Dependency.Actual_Sha1)
				if e != nil {
					err = e
					break
				}
				dependency := createBuildDependencyItem(dependecies[j].Dependency)
				if !shouldDownload {
					buildDependencies[threadId] = append(buildDependencies[threadId], dependency)
					logger.Logger.Info(logMsgPrefix + "File already exists locally.")
					continue
				}

				downloadFileDetails := createDownloadFileDetails(downloadPath, localPath, localFileName, nil, dependecies[j].Dependency.Size)
				e = downloadFile(downloadFileDetails, logMsgPrefix, flags)
				if e != nil {
					err = e
					break
				}
				buildDependencies[threadId] = append(buildDependencies[threadId], dependency)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	logDownloadTotals(buildDependencies)
	return
}

func logDownloadTotals(buildDependencies [][]utils.DependenciesBuildInfo) {
	var totalDownloded int
	for _, v := range buildDependencies {
		totalDownloded += len(v)
	}
	logger.Logger.Info("Downloaded " + strconv.Itoa(totalDownloded) + " artifacts from Artifactory.")
}

func createBuildDependencyItem(resultItem utils.AqlSearchResultItem) utils.DependenciesBuildInfo {
	return utils.DependenciesBuildInfo{
		Id: resultItem.Name,
		BuildInfoCommon : &utils.BuildInfoCommon{
			Sha1: resultItem.Actual_Sha1,
			Md5: resultItem.Actual_Md5,
		},
	}
}

func createDownloadFileDetails(downloadPath, localPath, localFileName string, acceptRanges *types.BoolEnum, size int64) (details *DownloadFileDetails) {
	details = &DownloadFileDetails{
		DownloadPath: downloadPath,
		LocalPath: localPath,
		LocalFileName: localFileName,
		AcceptRanges: acceptRanges,
		Size: size}
	return
}

func getFileRemoteDetails(downloadPath string, flags *DownloadFlags) (*ioutils.FileDetails, error) {
	httpClientsDetails := utils.GetArtifactoryHttpClientDetails(flags.ArtDetails)
	details, err := ioutils.GetRemoteFileDetails(downloadPath, httpClientsDetails)
	if err != nil {
		err = cliutils.CheckError(errors.New("Artifactory " + err.Error()))
		if err != nil {
			return details, err
		}
	}
	return details, nil
}

func downloadFile(downloadFileDetails *DownloadFileDetails, logMsgPrefix string, flags *DownloadFlags) error {
	httpClientsDetails := utils.GetArtifactoryHttpClientDetails(flags.ArtDetails)
	bulkDownload := flags.SplitCount == 0 || flags.MinSplitSize < 0 || flags.MinSplitSize * 1000 > downloadFileDetails.Size
	if !bulkDownload {
		acceptRange, err := isFileAcceptRange(downloadFileDetails, flags)
		if err != nil {
			return err
		}
		bulkDownload = !acceptRange
	}
	if bulkDownload {
		resp, err := ioutils.DownloadFile(downloadFileDetails.DownloadPath, downloadFileDetails.LocalPath, downloadFileDetails.LocalFileName, httpClientsDetails)
		if err != nil {
			return err
		}
		logger.Logger.Info(logMsgPrefix + "Artifactory response:", resp.Status)
	} else {
		concurrentDownloadFlags := ioutils.ConcurrentDownloadFlags{
			DownloadPath: downloadFileDetails.DownloadPath,
			FileName:     downloadFileDetails.LocalFileName,
			LocalPath:    downloadFileDetails.LocalPath,
			FileSize:     downloadFileDetails.Size,
			SplitCount:   flags.SplitCount}

		ioutils.DownloadFileConcurrently(concurrentDownloadFlags, logMsgPrefix, httpClientsDetails)
	}
	return nil
}

func isFileAcceptRange(downloadFileDetails *DownloadFileDetails, flags *DownloadFlags) (bool, error) {
	if downloadFileDetails.AcceptRanges == nil {
		details, err := getFileRemoteDetails(downloadFileDetails.DownloadPath, flags)
		if err != nil {
			return false, err
		}
		return details.AcceptRanges.GetValue(), nil
	}
	return downloadFileDetails.AcceptRanges.GetValue(), nil
}

func shouldDownloadFile(localFilePath, md5, sha1 string) (bool, error) {
	exists, err := ioutils.IsFileExists(localFilePath)
	if err != nil {
		return false, err
	}
	if !exists {
		return true, nil
	}
	localFileDetails, err := ioutils.GetFileDetails(localFilePath)
	if err != nil {
		return false, err
	}
	if localFileDetails.Md5 != md5 || localFileDetails.Sha1 != sha1 {
		return true, nil
	}
	return false, nil
}

type DownloadFileDetails struct {
	DownloadPath  string          `json:"DownloadPath,omitempty"`
	LocalPath     string          `json:"LocalPath,omitempty"`
	LocalFileName string          `json:"LocalFileName,omitempty"`
	AcceptRanges  *types.BoolEnum `json:"AcceptRanges,omitempty"`
	Size          int64           `json:"Size,omitempty"`
}

type DownloadFlags struct {
	ArtDetails   *config.ArtifactoryDetails
	DryRun       bool
	Threads      int
	MinSplitSize int64
	SplitCount   int
	BuildName    string
	BuildNumber  string
}

type DownloadData struct {
	Dependency   utils.AqlSearchResultItem
	DownloadPath string
	Target       string
	Flat         bool
}

func (flags *DownloadFlags) GetArtifactoryDetails() *config.ArtifactoryDetails {
	return flags.ArtDetails
}

func (flags *DownloadFlags) IsDryRun() bool {
	return flags.DryRun
}