package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/types"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	"strconv"
	"errors"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/log"
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
	buildDependencies, err := downloadFiles(downloadSpec, flags)
	if err != nil {
		return
	}
	logDownloadTotals(buildDependencies)
	if isCollectBuildInfo && !flags.DryRun {
		populateFunc := func(tempWrapper *utils.ArtifactBuildInfoWrapper) {
			tempWrapper.Dependencies = stripThreadIdFromBuildInfoDependencies(buildDependencies)
		}
		err = utils.SavePartialBuildInfo(flags.BuildName, flags.BuildNumber, populateFunc)
	}
	return
}

func downloadFiles(downloadSpec *utils.SpecFiles, flags *DownloadFlags) ([][]utils.DependenciesBuildInfo, error) {
	buildDependencies := make([][]utils.DependenciesBuildInfo, flags.Threads)
	producerConsumer := utils.NewProducerConsumer(flags.Threads)
	dependencyHandlerFunc := createDependencyHandlerFunc(buildDependencies, flags)
	log.Info("Searching artifacts...")
	getDependenciesForDownload(producerConsumer, downloadSpec, dependencyHandlerFunc, flags)
	err := performDownloadTasks(producerConsumer)
	return buildDependencies, err
}

func getDependenciesForDownload(producer utils.Producer, downloadSpec *utils.SpecFiles, dependencyContextHandler dependencyHandlerFunc, flags *DownloadFlags) {
	go func() {
		defer producer.Finish()
		var err error
		for i := 0; i < len(downloadSpec.Files); i++ {
			switch downloadSpec.Get(i).GetSpecType() {
			case utils.WILDCARD, utils.SIMPLE:
				err = collectWildcardDependencies(downloadSpec.Get(i), producer, dependencyContextHandler, flags)
			case utils.AQL:
				err = collectAqlDependencies(downloadSpec.Get(i), producer, dependencyContextHandler, flags)
			}
			if err != nil {
				producer.SetError(err)
				return
			}
		}
	}()
}

func stripThreadIdFromBuildInfoDependencies(dependenciesBuildInfo [][]utils.DependenciesBuildInfo) []utils.DependenciesBuildInfo {
	var buildInfo []utils.DependenciesBuildInfo
	for _, v := range dependenciesBuildInfo {
		buildInfo = append(buildInfo, v...)
	}
	return buildInfo
}

func collectWildcardDependencies(fileSpec *utils.Files, producer  utils.Producer, dependencyHandler dependencyHandlerFunc, flags *DownloadFlags) error {
	isRecursive, err := cliutils.StringToBool(fileSpec.Recursive, true)
	if err != nil {
		return err
	}
	resultItems, err := utils.AqlSearchDefaultReturnFields(fileSpec.Pattern, isRecursive, fileSpec.Props, flags)
	if err != nil {
		return err
	}
	return produceDownloadTasks(resultItems, fileSpec, producer, dependencyHandler)
}

func collectAqlDependencies(fileSpec *utils.Files, producer utils.Producer, dependencyHandler dependencyHandlerFunc, flags *DownloadFlags) error {
	resultItems, err := utils.AqlSearchBySpec(fileSpec.Aql, flags)
	if err != nil {
		return err
	}
	return produceDownloadTasks(resultItems, fileSpec, producer, dependencyHandler)
}

func produceDownloadTasks(items []utils.AqlSearchResultItem, fileSpec *utils.Files, producer utils.Producer, dependencyHandler dependencyHandlerFunc) error {
	flat, err := cliutils.StringToBool(fileSpec.Flat, false)
	if err != nil {
		return err
	}
	for _, v := range items {
		tempData := DownloadData{
			Dependency: v,
			DownloadPath: fileSpec.Pattern,
			Target: fileSpec.Target,
			Flat: flat,
		}
		producer.Produce(dependencyHandler(tempData))
	}
	return nil
}

func performDownloadTasks(producerConsumer utils.Consumer) (err error) {
	// Blocks until finish consuming
	producerConsumer.Consume()
	if e := producerConsumer.GetError(); e != nil {
		err = e
	}
	return
}

func logDownloadTotals(buildDependencies [][]utils.DependenciesBuildInfo) {
	var totalDownload int
	for _, v := range buildDependencies {
		totalDownload += len(v)
	}
	log.Info("Downloaded", strconv.Itoa(totalDownload), "artifacts.")
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
		log.Debug(logMsgPrefix, "Artifactory response:", resp.Status)
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

type dependencyHandlerFunc func(DownloadData) utils.Task
func createDependencyHandlerFunc(buildDependencies [][]utils.DependenciesBuildInfo, flags *DownloadFlags) dependencyHandlerFunc {
	return func(downloadData DownloadData) utils.Task {
		return func(threadId int) error {
			logMsgPrefix := cliutils.GetLogMsgPrefix(threadId, flags.DryRun)
			downloadPath, e := utils.BuildArtifactoryUrl(flags.ArtDetails.Url, downloadData.Dependency.GetFullUrl(), make(map[string]string))
			if e != nil {
				return e
			}
			log.Info(logMsgPrefix + "Downloading", downloadData.Dependency.GetFullUrl())
			if flags.DryRun {
				return nil
			}

			regexpPattern := cliutils.PathToRegExp(downloadData.DownloadPath)
			placeHolderTarget, e := cliutils.ReformatRegexp(regexpPattern, downloadData.Dependency.GetFullUrl(), downloadData.Target)
			if e != nil {
				return e
			}
			localPath, localFileName := ioutils.GetLocalPathAndFile(downloadData.Dependency.Name, downloadData.Dependency.Path, placeHolderTarget, downloadData.Flat)
			shouldDownload, e := shouldDownloadFile(path.Join(localPath, downloadData.Dependency.Name), downloadData.Dependency.Actual_Md5, downloadData.Dependency.Actual_Sha1)
			if e != nil {
				return e
			}
			dependency := createBuildDependencyItem(downloadData.Dependency)
			if !shouldDownload {
				buildDependencies[threadId] = append(buildDependencies[threadId], dependency)
				log.Debug(logMsgPrefix, "File already exists locally.")
				return nil
			}

			downloadFileDetails := createDownloadFileDetails(downloadPath, localPath, localFileName, nil, downloadData.Dependency.Size)
			e = downloadFile(downloadFileDetails, logMsgPrefix, flags)
			if e != nil {
				return e
			}
			buildDependencies[threadId] = append(buildDependencies[threadId], dependency)
			return nil
		}
	}
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