package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/types"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	"strconv"
	"sync"
	"strings"
	"errors"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/logger"
	"path"
)

func Download(downloadSpec *utils.SpecFiles, flags *DownloadFlags) (err error) {
	utils.PreCommandSetup(flags)
	isCollectBuildInfo := len(flags.BuildName) > 0  && len(flags.BuildNumber) > 0
	if isCollectBuildInfo {
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

	buildDependecies := make(map[int][]utils.DependenciesBuildInfo)
	for i := 0; i < len(downloadSpec.Files); i++ {
		var tempBuildDependecies map[int][]utils.DependenciesBuildInfo
		switch downloadSpec.Get(i).GetSpecType() {
		case utils.WILDCARD:
			tempBuildDependecies, err = downloadWildcard(downloadSpec.Get(i), flags)
		case utils.SIMPLE:
			tempBuildDependecies, err = downloadSimple(downloadSpec.Get(i), flags)
		case utils.AQL:
			tempBuildDependecies, err = downloadAql(downloadSpec.Get(i), flags)
		}
		if err != nil {
			return
		}
		for k, v := range tempBuildDependecies {
			buildDependecies[k] = append(buildDependecies[k], v...)
		}
	}
	if isCollectBuildInfo && err == nil {
		populateFunc := func(tempWrapper *utils.ArtifactBuildInfoWrapper) {
			tempWrapper.Dependencies = stripThreadIdFromBuildInfoDependencies(buildDependecies)
		}
		err = utils.PrepareBuildInfoForSave(flags.BuildName, flags.BuildNumber, populateFunc)
	}
	return
}

func stripThreadIdFromBuildInfoDependencies(dependenciesBuildInfo map[int][]utils.DependenciesBuildInfo) []utils.DependenciesBuildInfo {
	var buildInfo []utils.DependenciesBuildInfo
	for _, v := range dependenciesBuildInfo {
		buildInfo = append(buildInfo, v...)
	}
	return buildInfo
}

func downloadAql(fileSpec *utils.Files, flags *DownloadFlags) (map[int][]utils.DependenciesBuildInfo, error) {
	resultItems, err := utils.AqlSearchBySpec(fileSpec.Aql, flags)
	if err != nil {
		return nil, err
	}
	buildDependencies, err := downloadAqlResult("", resultItems, fileSpec.Target, fileSpec.Flat, flags)
	if err != nil {
		return nil, err
	}

	logger.Logger.Info("Downloaded " + strconv.Itoa(len(resultItems)) + " artifacts from Artifactory.")
	return buildDependencies, nil
}

func downloadWildcard(fileSpec *utils.Files, flags *DownloadFlags) (map[int][]utils.DependenciesBuildInfo, error) {
	resultItems, err := utils.AqlSearchDefaultReturnFields(fileSpec.Pattern, fileSpec.Recursive, fileSpec.Props, flags)
	if err != nil {
		return nil, err
	}
	buildDependencies, err := downloadAqlResult(fileSpec.Pattern, resultItems, fileSpec.Target, fileSpec.Flat, flags)
	if err != nil {
		return nil, err
	}

	logger.Logger.Info("Downloaded " + strconv.Itoa(len(resultItems)) + " artifacts from Artifactory.")
	return buildDependencies, nil
}

func downloadSimple(fileSpec *utils.Files, flags *DownloadFlags) (map[int][]utils.DependenciesBuildInfo, error) {
	props := "";
	if fileSpec.Props != "" {
		props = ";" + utils.EncodeParams(fileSpec.Props)
	}
	cleanPattern := cliutils.StripChars(fileSpec.Pattern, "()")
	downloadPath, err := utils.BuildArtifactoryUrl(flags.ArtDetails.Url, cleanPattern + props, make(map[string]string))
	if err != nil {
		return nil, err
	}
	logMsgPrefix := cliutils.GetLogMsgPrefix(0, flags.DryRun)
	logger.Logger.Info(logMsgPrefix + "Downloading " + downloadPath)
	if flags.DryRun {
		return nil, nil
	}

	regexpPattern := cliutils.PathToRegExp(fileSpec.Pattern)
	placeHolderTarget, err := cliutils.ReformatRegexp(regexpPattern, cleanPattern, fileSpec.Target)
	if err != nil {
		return nil, err
	}

	patternFileName, patternFilePath := ioutils.GetFileAndDirFromPath(trimRepo(fileSpec.Pattern))
	localPath, localFileName := getLocalPathAndFile(patternFileName, patternFilePath, placeHolderTarget, fileSpec.Flat)

	details, err := getFileRemoteDetails(downloadPath, flags)
	if err != nil {
		return nil, err
	}
	shouldDownload, err := shouldDownloadFile(path.Join(localPath, localFileName), details.Md5, details.Sha1)
	dependency := utils.DependenciesBuildInfo{
		Id: localFileName,
		BuildInfoCommon : &utils.BuildInfoCommon{
			Sha1: details.Sha1,
			Md5: details.Md5,
		},
	}
	if err != nil {
		return nil, err
	}
	if shouldDownload {
		downloadFileDetails := createDownloadFileDetails(downloadPath, localPath, localFileName,
			details.AcceptRanges, details.Size, flags)
		err = downloadFile(downloadFileDetails, logMsgPrefix + ": ", flags)
		if err != nil {
			return nil, err
		}
	} else {
		logger.Logger.Info(logMsgPrefix + "File already exists locally.")
	}
	buildDependencies := make(map[int][]utils.DependenciesBuildInfo)
	buildDependencies[0] = append(buildDependencies[0], dependency)
	return buildDependencies, nil
}

func trimRepo(path string) string {
	index := strings.Index(path, "/")
	if (index != -1) {
		return path[index + 1:]
	}
	return path
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

func downloadAqlResult(downloadPattern string, resultItems []utils.AqlSearchResultItem, targetPath string, flat bool, flags *DownloadFlags) (buildDependencies map[int][]utils.DependenciesBuildInfo, err error) {
	size := len(resultItems)
	buildDependencies = make(map[int][]utils.DependenciesBuildInfo)
	var wg sync.WaitGroup
	for i := 0; i < flags.Threads; i++ {
		wg.Add(1)
		go func(threadId int) {
			logMsgPrefix := cliutils.GetLogMsgPrefix(threadId, flags.DryRun)
			for j := threadId; j < size && err == nil; j += flags.Threads {
				downloadPath, e := utils.BuildArtifactoryUrl(flags.ArtDetails.Url, resultItems[j].GetFullUrl(), make(map[string]string))
				if e != nil {
					err = e
					break
				}
				logger.Logger.Info(logMsgPrefix + "Downloading " + downloadPath)
				if flags.DryRun {
					continue
				}

				regexpPattern := cliutils.PathToRegExp(downloadPattern)
				placeHolderTarget, e := cliutils.ReformatRegexp(regexpPattern, resultItems[j].GetFullUrl(), targetPath)
				if e != nil {
					err = e
					break
				}
				localPath, localFileName := getLocalPathAndFile(resultItems[j].Name, resultItems[j].Path, placeHolderTarget, flat)
				shouldDownload, e := shouldDownloadFile(path.Join(localPath, resultItems[j].Name), resultItems[j].Actual_Md5, resultItems[j].Actual_Sha1)
				if e != nil {
					err = e
					break
				}
				dependency := createBuildDependencyItem(resultItems[j])
				if !shouldDownload {
					buildDependencies[threadId] = append(buildDependencies[threadId], dependency)
					logger.Logger.Info(logMsgPrefix + "File already exists locally.")
					continue
				}

				downloadFileDetails := createDownloadFileDetails(downloadPath, localPath, localFileName, nil, resultItems[j].Size, flags)
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
	return
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

func createDownloadFileDetails(downloadPath, localPath, localFileName string, acceptRanges *types.BoolEnum, size int64, flags *DownloadFlags) (details *DownloadFileDetails) {
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

func (flags *DownloadFlags) GetArtifactoryDetails() *config.ArtifactoryDetails {
	return flags.ArtDetails
}

func (flags *DownloadFlags) IsDryRun() bool {
	return flags.DryRun
}