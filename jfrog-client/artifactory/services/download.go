package services

import (
	"errors"
	"github.com/jfrogdev/gofrog/parallel"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/errors/httperrors"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/httpclient"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/artifactory/services/utils/auth"
	clientutils "github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/fileutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/fileutils/checksum"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
	"os"
	"path/filepath"
	"sort"
	"net/http"
)

type DownloadService struct {
	client       *httpclient.HttpClient
	ArtDetails   *auth.ArtifactoryDetails
	DryRun       bool
	Threads      int
	MinSplitSize int64
	SplitCount   int
	Retries      int
}

func NewDownloadService(client *httpclient.HttpClient) *DownloadService {
	return &DownloadService{client: client}
}

func (ds *DownloadService) GetArtifactoryDetails() *auth.ArtifactoryDetails {
	return ds.ArtDetails
}

func (ds *DownloadService) SetArtifactoryDetails(rt *auth.ArtifactoryDetails) {
	ds.ArtDetails = rt
}

func (ds *DownloadService) IsDryRun() bool {
	return ds.DryRun
}

func (ds *DownloadService) GetJfrogHttpClient() *httpclient.HttpClient {
	return ds.client
}

func (ds *DownloadService) GetThreads() int {
	return ds.Threads
}

func (ds *DownloadService) SetThreads(threads int) {
	ds.Threads = threads
}

func (ds *DownloadService) SetArtDetails(artDetails *auth.ArtifactoryDetails) {
	ds.ArtDetails = artDetails
}

func (ds *DownloadService) SetDryRun(isDryRun bool) {
	ds.DryRun = isDryRun
}

func (ds *DownloadService) setMinSplitSize(minSplitSize int64) {
	ds.MinSplitSize = minSplitSize
}

func (ds *DownloadService) DownloadFiles(downloadParams DownloadParams) ([]utils.FileInfo, int, error) {
	buildDependencies := make([][]utils.FileInfo, ds.GetThreads())
	producerConsumer := parallel.NewBounedRunner(ds.GetThreads(), false)
	errorsQueue := utils.NewErrorsQueue(1)
	fileHandlerFunc := ds.createFileHandlerFunc(buildDependencies, downloadParams)
	log.Info("Searching items to download...")
	expectedChan := make(chan int, 1)
	ds.prepareTasks(producerConsumer, fileHandlerFunc, expectedChan, errorsQueue, downloadParams)
	err := performTasks(producerConsumer, errorsQueue)
	return utils.StripThreadId(buildDependencies), <-expectedChan, err
}

func (ds *DownloadService) prepareTasks(producer parallel.Runner, fileContextHandler fileHandlerFunc, expectedChan chan int, errorsQueue *utils.ErrorsQueue, downloadParams DownloadParams) {
	go func() {
		defer producer.Done()
		var err error
		var resultItems []utils.ResultItem
		switch downloadParams.GetSpecType() {
		case utils.WILDCARD, utils.SIMPLE:
			resultItems, err = ds.collectFilesUsingWildcardPattern(downloadParams)
		case utils.AQL:
			resultItems, err = utils.AqlSearchBySpec(downloadParams.GetFile(), ds)
		}

		if err != nil {
			errorsQueue.AddError(err)
			return
		}
		tasks, err := produceTasks(resultItems, downloadParams, producer, fileContextHandler, errorsQueue)
		if err != nil {
			errorsQueue.AddError(err)
			return
		}
		expectedChan <- tasks
	}()
}

func (ds *DownloadService) collectFilesUsingWildcardPattern(downloadParams DownloadParams) ([]utils.ResultItem, error) {
	return utils.AqlSearchDefaultReturnFields(downloadParams.GetFile(), ds)
}

func produceTasks(items []utils.ResultItem, downloadParams DownloadParams, producer parallel.Runner, fileHandler fileHandlerFunc, errorsQueue *utils.ErrorsQueue) (int, error) {
	flat := downloadParams.IsFlat()
	// Collect all folders path which might be needed to create.
	// key = folder path, value = the necessary data for producing create folder task.
	directoriesData := make(map[string]DownloadData)
	// Store all the paths which was created implicitly due to file upload.
	alreadyCreatedDirs := make(map[string]bool)
	// Store all the keys of directoriesData as an array.
	var directoriesDataKeys []string
	// Task counter
	var tasksCount int
	for _, v := range items {
		tempData := DownloadData{
			Dependency:   v,
			DownloadPath: downloadParams.GetPattern(),
			Target:       downloadParams.GetTarget(),
			Flat:         flat,
		}
		if v.Type != "folder" {
			// Add a task, task is a function of type TaskFunc which later on will be executed by other go routine, the communication is done using channels.
			// The second argument is a error handling func in case the taskFunc return an error.
			tasksCount++
			producer.AddTaskWithError(fileHandler(tempData), errorsQueue.AddError)
			// We don't want to create directories which are created explicitly by download files when the --include-dirs flag is used.
			alreadyCreatedDirs[v.Path] = true
		} else {
			directoriesData, directoriesDataKeys = collectDirPathsToCreate(v, directoriesData, tempData, directoriesDataKeys)
		}
	}

	addCreateDirsTasks(directoriesDataKeys, alreadyCreatedDirs, producer, fileHandler, directoriesData, errorsQueue, flat)
	return tasksCount, nil
}

// Extract for the aqlResultItem the directory path, store the path the directoriesDataKeys and in the directoriesData map.
// In addition directoriesData holds the correlate DownloadData for each key, later on this DownloadData will be used to create a create dir tasks if needed.
// This function append the new data to directoriesDataKeys and to directoriesData and return the new map and the new []string
// We are storing all the keys of directoriesData in additional array(directoriesDataKeys) so we could sort the keys and access the maps in the sorted order.
func collectDirPathsToCreate(aqlResultItem utils.ResultItem, directoriesData map[string]DownloadData, tempData DownloadData, directoriesDataKeys []string) (map[string]DownloadData, []string) {
	key := aqlResultItem.Name
	if aqlResultItem.Path != "." {
		key = aqlResultItem.Path + "/" + aqlResultItem.Name
	}
	directoriesData[key] = tempData
	directoriesDataKeys = append(directoriesDataKeys, key)
	return directoriesData, directoriesDataKeys
}

func addCreateDirsTasks(directoriesDataKeys []string, alreadyCreatedDirs map[string]bool, producer parallel.Runner, fileHandler fileHandlerFunc, directoriesData map[string]DownloadData, errorsQueue *utils.ErrorsQueue, isFlat bool) {
	// Longest path first
	// We are going to create the longest path first by doing so all sub paths of the longest path will be created implicitly.
	sort.Sort(sort.Reverse(sort.StringSlice(directoriesDataKeys)))
	for index, v := range directoriesDataKeys {
		// In order to avoid duplication we need to check the path wasn't already created by the previous action.
		if v != "." && // For some files the returned path can be the root path, ".", in that case we doing need to create any directory.
			(index == 0 || !utils.IsSubPath(directoriesDataKeys, index, "/")) { // directoriesDataKeys store all the path which might needed to be created, that's include duplicated paths.
			// By sorting the directoriesDataKeys we can assure that the longest path was created and therefore no need to create all it's sub paths.

			// Some directories were created due to file download when we aren't in flat download flow.
			if isFlat {
				producer.AddTaskWithError(fileHandler(directoriesData[v]), errorsQueue.AddError)
			} else if !alreadyCreatedDirs[v] {
				producer.AddTaskWithError(fileHandler(directoriesData[v]), errorsQueue.AddError)
			}
		}
	}
	return
}

func performTasks(consumer parallel.Runner, errorsQueue *utils.ErrorsQueue) error {
	// Blocked until finish consuming
	consumer.Run()
	return errorsQueue.GetError()
}

func createDependencyFileInfo(resultItem utils.ResultItem, localPath, localFileName string) utils.FileInfo {
	fileInfo := utils.FileInfo{
		ArtifactoryPath: resultItem.GetItemRelativePath(),
		FileHashes: &utils.FileHashes{
			Sha1: resultItem.Actual_Sha1,
			Md5:  resultItem.Actual_Md5,
		},
	}
	fileInfo.LocalPath = filepath.Join(localPath, localFileName)
	return fileInfo
}

func createDownloadFileDetails(downloadPath, localPath, localFileName string, downloadData DownloadData) (details *httpclient.DownloadFileDetails) {
	details = &httpclient.DownloadFileDetails{
		FileName:      downloadData.Dependency.Name,
		DownloadPath:  downloadPath,
		LocalPath:     localPath,
		LocalFileName: localFileName,
		Size:          downloadData.Dependency.Size}
	return
}

func (ds *DownloadService) downloadFile(downloadFileDetails *httpclient.DownloadFileDetails, logMsgPrefix string, downloadParams DownloadParams) error {
	httpClientsDetails := ds.ArtDetails.CreateArtifactoryHttpClientDetails()
	bulkDownload := ds.SplitCount == 0 || ds.MinSplitSize < 0 || ds.MinSplitSize*1000 > downloadFileDetails.Size
	if !bulkDownload {
		acceptRange, err := ds.isFileAcceptRange(downloadFileDetails)
		if err != nil {
			return err
		}
		bulkDownload = !acceptRange
	}
	if bulkDownload {
		var resp *http.Response
		resp, err := ds.client.DownloadFile(downloadFileDetails, logMsgPrefix, httpClientsDetails, downloadParams.GetRetries(), downloadParams.IsExplode())
		if resp != nil {
			log.Debug(logMsgPrefix, "Artifactory response:", resp.Status)
		}
		return err
	}

	concurrentDownloadFlags := httpclient.ConcurrentDownloadFlags{
		FileName:      downloadFileDetails.FileName,
		DownloadPath:  downloadFileDetails.DownloadPath,
		LocalFileName: downloadFileDetails.LocalFileName,
		LocalPath:     downloadFileDetails.LocalPath,
		FileSize:      downloadFileDetails.Size,
		SplitCount:    ds.SplitCount,
		Explode:       downloadParams.IsExplode(),
		Retries:       downloadParams.GetRetries()}

	return ds.client.DownloadFileConcurrently(concurrentDownloadFlags, logMsgPrefix, httpClientsDetails)
}

func (ds *DownloadService) isFileAcceptRange(downloadFileDetails *httpclient.DownloadFileDetails) (bool, error) {
	httpClientsDetails := ds.ArtDetails.CreateArtifactoryHttpClientDetails()
	return ds.client.IsAcceptRanges(downloadFileDetails.DownloadPath, httpClientsDetails)
}

func shouldDownloadFile(localFilePath, md5, sha1 string) (bool, error) {
	exists, err := fileutils.IsFileExists(localFilePath)
	if err != nil {
		return false, err
	}
	if !exists {
		return true, nil
	}
	localFileDetails, err := fileutils.GetFileDetails(localFilePath)
	if err != nil {
		return false, err
	}
	return localFileDetails.Checksum.Md5 != md5 || localFileDetails.Checksum.Sha1 != sha1, nil
}

func removeIfSymlink(localSymlinkPath string) error {
	if fileutils.IsPathSymlink(localSymlinkPath) {
		if err := os.Remove(localSymlinkPath); errorutils.CheckError(err) != nil {
			return err
		}
	}
	return nil
}

func createLocalSymlink(localPath, localFileName, symlinkArtifact string, symlinkChecksum bool, symlinkContentChecksum string, logMsgPrefix string) error {
	if symlinkChecksum && symlinkContentChecksum != "" {
		if !fileutils.IsPathExists(symlinkArtifact) {
			return errorutils.CheckError(errors.New("Symlink validation failed, target doesn't exist: " + symlinkArtifact))
		}
		file, err := os.Open(symlinkArtifact)
		errorutils.CheckError(err)
		if err != nil {
			return err
		}
		defer file.Close()
		checksumInfo, err := checksum.Calc(file, checksum.SHA1)
		if err != nil {
			return err
		}
		sha1 := checksumInfo[checksum.SHA1]
		if sha1 != symlinkContentChecksum {
			return errorutils.CheckError(errors.New("Symlink validation failed for target: " + symlinkArtifact))
		}
	}
	localSymlinkPath := filepath.Join(localPath, localFileName)
	isFileExists, err := fileutils.IsFileExists(localSymlinkPath)
	if err != nil {
		return err
	}
	// We can't create symlink in case a file with the same name already exist, we must remove the file before creating the symlink
	if isFileExists {
		if err := os.Remove(localSymlinkPath); err != nil {
			return err
		}
	}
	// Need to prepare the directories hierarchy
	_, err = fileutils.CreateFilePath(localPath, localFileName)
	if err != nil {
		return err
	}
	err = os.Symlink(symlinkArtifact, localSymlinkPath)
	if errorutils.CheckError(err) != nil {
		return err
	}
	log.Debug(logMsgPrefix, "Creating symlink file.")
	return nil
}

func getArtifactPropertyByKey(properties []utils.Property, key string) string {
	for _, v := range properties {
		if v.Key == key {
			return v.Value
		}
	}
	return ""
}

func getArtifactSymlinkPath(properties []utils.Property) string {
	return getArtifactPropertyByKey(properties, utils.ARTIFACTORY_SYMLINK)
}

func getArtifactSymlinkChecksum(properties []utils.Property) string {
	return getArtifactPropertyByKey(properties, utils.SYMLINK_SHA1)
}

type fileHandlerFunc func(DownloadData) parallel.TaskFunc

func (ds *DownloadService) createFileHandlerFunc(buildDependencies [][]utils.FileInfo, downloadParams DownloadParams) fileHandlerFunc {
	return func(downloadData DownloadData) parallel.TaskFunc {
		return func(threadId int) error {
			logMsgPrefix := clientutils.GetLogMsgPrefix(threadId, ds.DryRun)
			downloadPath, e := utils.BuildArtifactoryUrl(ds.ArtDetails.Url, downloadData.Dependency.GetItemRelativePath(), make(map[string]string))
			if e != nil {
				return e
			}
			log.Info(logMsgPrefix+"Downloading", downloadData.Dependency.GetItemRelativePath())
			if ds.DryRun {
				return nil
			}

			regexpPattern := clientutils.PathToRegExp(downloadData.DownloadPath)
			placeHolderTarget, e := clientutils.ReformatRegexp(regexpPattern, downloadData.Dependency.GetItemRelativePath(), downloadData.Target)
			if e != nil {
				return e
			}
			localPath, localFileName := fileutils.GetLocalPathAndFile(downloadData.Dependency.Name, downloadData.Dependency.Path, placeHolderTarget, downloadData.Flat)
			if downloadData.Dependency.Type == "folder" {
				return createDir(localPath, localFileName, logMsgPrefix)
			}
			removeIfSymlink(filepath.Join(localPath, localFileName))
			if downloadParams.IsSymlink() {
				if isSymlink, e := createSymlinkIfNeeded(localPath, localFileName, logMsgPrefix, downloadData, buildDependencies, threadId, downloadParams); isSymlink {
					return e
				}
			}
			dependency := createDependencyFileInfo(downloadData.Dependency, localPath, localFileName)
			e = ds.downloadFileIfNeeded(downloadPath, localPath, localFileName, logMsgPrefix, downloadData, downloadParams)

			if e == nil {
				buildDependencies[threadId] = append(buildDependencies[threadId], dependency)
			} else if !httperrors.IsResponseStatusError(e) {
				// Ignore response status errors to continue downloading
				log.Error(logMsgPrefix,"Received an error: "+ e.Error())
				return e
			}
			return nil
		}
	}
}

func (ds *DownloadService) downloadFileIfNeeded(downloadPath, localPath, localFileName, logMsgPrefix string, downloadData DownloadData, downloadParams DownloadParams) error {
	shouldDownload, e := shouldDownloadFile(filepath.Join(localPath, localFileName), downloadData.Dependency.Actual_Md5, downloadData.Dependency.Actual_Sha1)
	if e != nil {
		return e
	}
	if !shouldDownload {
		log.Debug(logMsgPrefix, "File already exists locally.")
		return nil
	}
	downloadFileDetails := createDownloadFileDetails(downloadPath, localPath, localFileName, downloadData)
	return ds.downloadFile(downloadFileDetails, logMsgPrefix, downloadParams)
}

func createDir(localPath, localFileName, logMsgPrefix string) error {
	folderPath := filepath.Join(localPath, localFileName)
	e := fileutils.CreateDirIfNotExist(folderPath)
	if e != nil {
		return e
	}
	log.Info(logMsgPrefix + "Creating folder: " + folderPath)
	return nil
}

func createSymlinkIfNeeded(localPath, localFileName, logMsgPrefix string, downloadData DownloadData, buildDependencies [][]utils.FileInfo, threadId int, downloadParams DownloadParams) (bool, error) {
	symlinkArtifact := getArtifactSymlinkPath(downloadData.Dependency.Properties)
	isSymlink := len(symlinkArtifact) > 0
	if isSymlink {
		symlinkChecksum := getArtifactSymlinkChecksum(downloadData.Dependency.Properties)
		if e := createLocalSymlink(localPath, localFileName, symlinkArtifact, downloadParams.ValidateSymlinks(), symlinkChecksum, logMsgPrefix); e != nil {
			return isSymlink, e
		}
		dependency := createDependencyFileInfo(downloadData.Dependency, localPath, localFileName)
		buildDependencies[threadId] = append(buildDependencies[threadId], dependency)
		return isSymlink, nil
	}
	return isSymlink, nil
}

type DownloadData struct {
	Dependency   utils.ResultItem
	DownloadPath string
	Target       string
	Flat         bool
}

type DownloadParams interface {
	utils.FileGetter
	IsSymlink() bool
	ValidateSymlinks() bool
	GetFile() *utils.ArtifactoryCommonParams
	IsFlat() bool
	GetRetries() int
}

type DownloadParamsImpl struct {
	*utils.ArtifactoryCommonParams
	Symlink         bool
	ValidateSymlink bool
	Flat            bool
	Explode         bool
	Retries         int
}

func (ds *DownloadParamsImpl) IsFlat() bool {
	return ds.Flat
}

func (ds *DownloadParamsImpl) IsExplode() bool {
	return ds.Explode
}

func (ds *DownloadParamsImpl) GetFile() *utils.ArtifactoryCommonParams {
	return ds.ArtifactoryCommonParams
}

func (ds *DownloadParamsImpl) IsSymlink() bool {
	return ds.Symlink
}

func (ds *DownloadParamsImpl) ValidateSymlinks() bool {
	return ds.ValidateSymlink
}

func (ds *DownloadParamsImpl) GetRetries() int {
	return ds.Retries
}
