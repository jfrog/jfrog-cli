package transfer

import (
	"encoding/json"
	"fmt"
	"github.com/jfrog/build-info-go/entities"
	"github.com/jfrog/gofrog/parallel"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	coreCommonCommands "github.com/jfrog/jfrog-cli-core/v2/common/commands"
	coreConfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	artifactoryUtils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/http/httpclient"
	clientUtils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/urfave/cli"
	"io/ioutil"
	"net/http"
	"path"
	"path/filepath"
	"strconv"
	"time"
)

const (
	fileTransferTimestampProperty = "jf.cli.transfer.timestamp"
	iso8601TimeFormat             = "2006-01-02T15:04:05Z0700"
	minChecksumDeploySize         = 1024
	tasksMaxCapacity              = 500000
)

var startTime time.Time
var tempDir string

func RunTransfer(c *cli.Context) (err error) {
	if c.NArg() != 3 {
		return cliutils.PrintHelpAndReturnError("wrong number of arguments.", c)
	}

	tc, err := getCommandConfig(c)
	if err != nil {
		return err
	}

	log.Output("Running with", tc.threads, "threads")

	tempDir, err = fileutils.CreateTempDir()
	log.Info("Created temp directory at:", tempDir)
	if err != nil {
		return err
	}
	defer func() {
		e := fileutils.RemoveTempDir(tempDir)
		if err == nil {
			err = e
		}
	}()

	startTime = time.Now()

	totalFolderTasks, totalFileTasks, err := tc.handleRepository(tc.repository)
	if err != nil {
		log.Output("Error:", err.Error())
	}
	summaryErr := tc.printSummary(tc.repository, totalFolderTasks, totalFileTasks, time.Since(startTime))
	if summaryErr != nil {
		if err == nil {
			return summaryErr
		}
		log.Error(summaryErr)
	}
	return
}

type transferCommandConfig struct {
	sourceRtDetails        *coreConfig.ServerDetails
	destRtDetails          *coreConfig.ServerDetails
	repository             string
	threads                int
	retries                int
	retryWaitTimeMilliSecs int
}

func (tc *transferCommandConfig) handleRepository(repoName string) (totalFolderTasks, totalFileTasks int, err error) {
	producerConsumer := parallel.NewRunner(tc.threads, tasksMaxCapacity, false)
	errorsQueue := clientUtils.NewErrorsQueue(1)
	expectedChan := make(chan int, 1)
	folderTasksCounters := make([]int, tc.threads)
	fileTasksCounters := make([]int, tc.threads)

	go func() {
		folderHandler := tc.createFolderHandlerFunc(producerConsumer, expectedChan, errorsQueue, folderTasksCounters, fileTasksCounters)
		_, _ = producerConsumer.AddTaskWithError(folderHandler(folderParams{repoName: repoName, relativePath: "."}), errorsQueue.AddError)
	}()

	var runnerErr error
	go func() {
		runnerErr = producerConsumer.DoneWhenAllIdle(15)
	}()
	// Blocked until finish consuming
	producerConsumer.Run()

	if runnerErr != nil {
		return 0, 0, runnerErr
	}

	// Don't count repository root.
	totalFolderTasks = -1
	for _, v := range folderTasksCounters {
		totalFolderTasks += v
	}
	for _, v := range fileTasksCounters {
		totalFileTasks += v
	}
	return totalFolderTasks, totalFileTasks, errorsQueue.GetError()
}

type folderHandlerFunc func(params folderParams) parallel.TaskFunc
type fileHandlerFunc func(file artifactoryUtils.ResultItem) parallel.TaskFunc

type folderParams struct {
	repoName     string
	relativePath string
}

func (tc *transferCommandConfig) createFolderHandlerFunc(producerConsumer parallel.Runner, expectedChan chan int, errorsQueue *clientUtils.ErrorsQueue, folderTasksCounters, fileTasksCounters []int) folderHandlerFunc {
	return func(params folderParams) parallel.TaskFunc {
		return func(threadId int) error {
			logMsgPrefix := clientUtils.GetLogMsgPrefix(threadId, false)
			err := tc.handleFolderAql(params, producerConsumer, expectedChan, errorsQueue, folderTasksCounters, fileTasksCounters, logMsgPrefix)
			if err != nil {
				return err
			}
			folderTasksCounters[threadId]++
			return nil
		}
	}
}

func (tc *transferCommandConfig) createFileHandlerFunc(fileTasksCounters []int) fileHandlerFunc {
	return func(file artifactoryUtils.ResultItem) parallel.TaskFunc {
		return func(threadId int) error {
			logMsgPrefix := clientUtils.GetLogMsgPrefix(threadId, false)
			err := tc.handleFile(file, logMsgPrefix)
			if err != nil {
				return err
			}
			fileTasksCounters[threadId]++
			return nil
		}
	}
}

func (tc *transferCommandConfig) handleFolderAql(params folderParams, producerConsumer parallel.Runner, expectedChan chan int, errorsQueue *clientUtils.ErrorsQueue, folderTasksCounters, fileTasksCounters []int, logMsgPrefix string) (err error) {
	log.Info(logMsgPrefix+"Visited folder:", path.Join(params.repoName, params.relativePath))

	result, err := tc.getDirectoryContentsAql(params.repoName, params.relativePath)
	if err != nil {
		return err
	}

	for _, item := range result.Results {
		if item.Name == "." {
			continue
		}
		if item.Type == "folder" {
			newRelativePath := item.Name
			if params.relativePath != "." {
				newRelativePath = path.Join(params.relativePath, newRelativePath)
			}
			folderHandler := tc.createFolderHandlerFunc(producerConsumer, expectedChan, errorsQueue, folderTasksCounters, fileTasksCounters)
			_, _ = producerConsumer.AddTaskWithError(folderHandler(folderParams{repoName: params.repoName, relativePath: newRelativePath}), errorsQueue.AddError)
		} else {
			fileHandler := tc.createFileHandlerFunc(fileTasksCounters)
			_, _ = producerConsumer.AddTaskWithError(fileHandler(item), errorsQueue.AddError)
		}
	}
	return nil
}

func (tc *transferCommandConfig) handleFile(item artifactoryUtils.ResultItem, logMsgPrefix string) error {
	filePath := path.Join(item.GetItemRelativePath())
	log.Info(logMsgPrefix+"Handling file:", filePath)

	required, err := isTransferRequired(item)
	if err != nil {
		return err
	}
	if !required {
		log.Info(logMsgPrefix+"File doesn't require transfer:", filePath)
		return nil
	}

	// Set the transfer property on the artifact in source.
	propsService, err := tc.createSourcePropsServiceManager()
	if err != nil {
		return err
	}

	transferProp := fileTransferTimestampProperty + "=" + strconv.FormatInt(time.Now().Unix(), 10)
	err = propsService.ModifyEncodedProperty(logMsgPrefix, item.GetItemRelativePath(), transferProp, false)
	if err != nil {
		return err
	}

	uploadService, err := tc.createDestUploadServiceManager()
	if err != nil {
		return err
	}

	_, targetPathWithProps, err := services.BuildUploadUrls(uploadService.ArtDetails.GetUrl(), item.GetItemRelativePath(), "", "", item.GetPropertiesAsMap())
	if err != nil {
		return err
	}

	fileDetails := &fileutils.FileDetails{Checksum: entities.Checksum{Sha1: item.Actual_Sha1, Md5: item.Actual_Md5, Sha256: item.Sha256}, Size: item.Size}

	if shouldTryChecksumDeploy(item.Size) {
		resp, _, err := uploadService.TryChecksumDeploy(fileDetails, targetPathWithProps, uploadService.ArtDetails.CreateHttpClientDetails())
		if err != nil {
			return err
		}
		// Checksum deploy successful.
		if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
			log.Info(logMsgPrefix+"Checksum deploy successful:", filePath)
			return nil
		}
	}

	downloadService, err := tc.createSourceDownloadServiceManager()
	if err != nil {
		return err
	}

	localRelativePath := filepath.Join(tempDir, item.Repo, item.Path)
	downloadFileDetails := &httpclient.DownloadFileDetails{
		FileName:      item.Name,
		DownloadPath:  clientUtils.AddTrailingSlashIfNeeded(downloadService.GetArtifactoryDetails().GetUrl()) + item.GetItemRelativePath(),
		RelativePath:  item.GetItemRelativePath(),
		LocalPath:     localRelativePath,
		LocalFileName: item.Name,
		Size:          item.Size,
		ExpectedSha1:  item.Actual_Sha1}
	downloadParams := services.NewDownloadParams()
	downloadParams.Flat = true
	log.Info(logMsgPrefix+"Downloading:", filePath)
	err = downloadService.DownloadFile(downloadFileDetails, logMsgPrefix, downloadParams)
	if err != nil {
		return err
	}

	uploadParams := services.NewUploadParams()
	uploadParams.Flat = true
	localPath := filepath.Join(localRelativePath, item.Name)
	log.Info(logMsgPrefix+"Uploading:", filePath)
	resp, body, err := artifactoryUtils.UploadFile(localPath, targetPathWithProps, logMsgPrefix, &uploadService.ArtDetails, fileDetails,
		uploadService.ArtDetails.CreateHttpClientDetails(), uploadService.GetJfrogHttpClient(), uploadParams.ChecksumsCalcEnabled, uploadService.Progress)
	if err != nil {
		return err
	}

	if !(resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK) {
		// Failed uploading.
		return errorutils.CheckErrorf(logMsgPrefix + "Artifactory response: " + resp.Status + "\n" + clientUtils.IndentJson(body))
	}
	log.Info(logMsgPrefix+"Done:", filePath)

	return nil
}

func (tc *transferCommandConfig) createSourcePropsServiceManager() (*services.PropsService, error) {
	serviceManager, err := utils.CreateServiceManager(tc.sourceRtDetails, 0, 0, false)
	if err != nil {
		return nil, err
	}
	propsService := services.NewPropsService(serviceManager.Client())
	propsService.ArtDetails = serviceManager.GetConfig().GetServiceDetails()
	return propsService, nil
}

func (tc *transferCommandConfig) createDestUploadServiceManager() (*services.UploadService, error) {
	serviceManager, err := utils.CreateServiceManager(tc.destRtDetails, 0, 0, false)
	if err != nil {
		return nil, err
	}
	uploadService := services.NewUploadService(serviceManager.Client())
	uploadService.ArtDetails = serviceManager.GetConfig().GetServiceDetails()
	uploadService.Threads = serviceManager.GetConfig().GetThreads()
	return uploadService, nil
}

func (tc *transferCommandConfig) createSourceDownloadServiceManager() (*services.DownloadService, error) {
	serviceManager, err := utils.CreateServiceManager(tc.sourceRtDetails, 0, 0, false)
	if err != nil {
		return nil, err
	}
	downloadService := services.NewDownloadService(serviceManager.GetConfig().GetServiceDetails(), serviceManager.Client())
	downloadService.Threads = serviceManager.GetConfig().GetThreads()
	return downloadService, nil
}

func shouldTryChecksumDeploy(fileSize int64) bool {
	return fileSize >= minChecksumDeploySize
}

// Transfer required if wasn't transferred yet, or if modified since transferred.
func isTransferRequired(item artifactoryUtils.ResultItem) (bool, error) {
	lastTransferValue := item.GetProperty(fileTransferTimestampProperty)
	// If the property was not set, the artifact wasn't transferred.
	if lastTransferValue == "" {
		return true, nil
	}

	// Transfer property set but modified is empty, no transfer needed.
	if item.Modified == "" {
		return false, nil
	}

	// Compare transfer and modified times.
	lastTransferTimestamp, err := strconv.ParseInt(lastTransferValue, 10, 64)
	if err != nil {
		return false, errorutils.CheckErrorf("failed parsing transfer timestamp")
	}

	lastModifiedTime, err := time.Parse(iso8601TimeFormat, item.Modified)
	if err != nil {
		return false, err
	}

	return lastModifiedTime.Unix() > lastTransferTimestamp, nil
}

func (tc *transferCommandConfig) getDirectoryContentsAql(repoName, relativePath string) (result *artifactoryUtils.AqlSearchResult, err error) {
	query := tc.generateAqlQuery(repoName, relativePath)
	serviceManager, err := utils.CreateServiceManager(tc.sourceRtDetails, -1, 0, false)
	if err != nil {
		return nil, err
	}
	reader, err := serviceManager.Aql(query)
	if err != nil {
		return nil, err
	}
	defer func() {
		if reader != nil {
			e := reader.Close()
			if err == nil {
				err = errorutils.CheckError(e)
			}
		}
	}()

	respBody, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, errorutils.CheckError(err)
	}

	result = &artifactoryUtils.AqlSearchResult{}
	err = json.Unmarshal(respBody, result)
	return result, errorutils.CheckError(err)
}

func (tc *transferCommandConfig) generateAqlQuery(repoName, relativePath string) string {
	return fmt.Sprintf(`items.find({"type":"any","$or":[{"$and":[{"repo":"%s","path":{"$match":"%s"},"name":{"$match":"*"}}]}]}).include("repo","path","name","created","modified","updated","created_by","modified_by","type","actual_md5","actual_sha1","sha256","size","property","stat")`, repoName, relativePath)
}

func (tc *transferCommandConfig) getStorageInfo() (*artifactoryUtils.StorageInfo, error) {
	serviceManager, err := utils.CreateServiceManager(tc.sourceRtDetails, -1, 0, false)
	if err != nil {
		return nil, err
	}
	return serviceManager.StorageInfo()
}

func (tc *transferCommandConfig) printSummary(sourceRepo string, totalFolderTasks, totalFileTasks int, timeElapsed time.Duration) error {
	log.Output("Done. Time elapsed:", timeElapsed)
	log.Output("")
	log.Output("Summary:")
	log.Output("total folders:", totalFolderTasks)
	log.Output("total files:", totalFileTasks)
	log.Output("total items:", totalFolderTasks+totalFileTasks)

	storageInfo, err := tc.getStorageInfo()
	if err != nil {
		return err
	}

	for _, repo := range storageInfo.RepositoriesSummaryList {
		if repo.RepoKey == sourceRepo {
			log.Output("")
			log.Output("Expected:")
			log.Output("total folders:", repo.FoldersCount)
			log.Output("total files:", repo.FilesCount)
			log.Output("total items:", repo.ItemsCount)
			log.Output("used space:", repo.UsedSpace)
			return nil
		}
	}
	return errorutils.CheckErrorf("could not find repo '%s' at storage info", sourceRepo)
}

func getCommandConfig(c *cli.Context) (tc transferCommandConfig, err error) {
	tc.sourceRtDetails, err = coreCommonCommands.GetConfig(c.Args().Get(0), false)
	if err != nil {
		return tc, err
	}

	tc.destRtDetails, err = coreCommonCommands.GetConfig(c.Args().Get(1), false)
	if err != nil {
		return tc, err
	}

	tc.repository = c.Args().Get(2)

	tc.threads, err = cliutils.GetThreadsCount(c, 16)
	if err != nil {
		return tc, err
	}

	tc.retries, err = cliutils.GetRetries(c)
	if err != nil {
		return tc, err
	}

	tc.retryWaitTimeMilliSecs, err = cliutils.GetRetryWaitTime(c)
	return
}
