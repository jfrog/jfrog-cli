package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/io/httputils"
	"github.com/jfrogdev/jfrog-cli-go/utils/io/fileutils"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"errors"
	"time"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/gofrog/parallel"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/log"
	"bytes"
	"sort"
)

// Uploads the artifacts in the specified local path pattern to the specified target path.
// Returns the total number of artifacts successfully uploaded.
func Upload(uploadSpec *utils.SpecFiles, flags *UploadFlags) (totalUploaded, totalFailed int, err error) {
	err = utils.PreCommandSetup(flags)
	if err != nil {
		return 0, 0, err
	}

	isCollectBuildInfo := len(flags.BuildName) > 0 && len(flags.BuildNumber) > 0
	if isCollectBuildInfo && !flags.DryRun {
		if err := utils.SaveBuildGeneralDetails(flags.BuildName, flags.BuildNumber); err != nil {
			return 0, 0, err
		}
	}
	minChecksumDeploySize, err := getMinChecksumDeploySize()
	if err != nil {
		return 0, 0, err
	}
	buildArtifacts, totalUploaded, totalFailed, err := uploadFiles(minChecksumDeploySize, uploadSpec, flags)
	if err != nil {
		return 0, 0, err
	}
	if totalFailed > 0 {
		return
	}
	if isCollectBuildInfo && !flags.DryRun {
		populateFunc := func(tempWrapper *utils.ArtifactBuildInfoWrapper) {
			tempWrapper.Artifacts = toBuildInfoArtifacts(buildArtifacts)
		}
		err = utils.SavePartialBuildInfo(flags.BuildName, flags.BuildNumber, populateFunc)
	}
	return
}

func uploadFiles(minChecksumDeploySize int64, uploadSpec *utils.SpecFiles, flags *UploadFlags) ([][]utils.ArtifactsBuildInfo, int, int, error) {
	uploadSummery := uploadResult{
		UploadCount: make([]int, flags.Threads),
		TotalCount: make([]int, flags.Threads),
		BuildInfoArtifacts: make([][]utils.ArtifactsBuildInfo, flags.Threads),
	}
	artifactHandlerFunc := createArtifactHandlerFunc(&uploadSummery, minChecksumDeploySize, flags)
	producerConsumer := parallel.NewBounedRunner(flags.Threads, true)
	errorsQueue := utils.NewErrorsQueue(1)
	prepareUploadTasks(producerConsumer, uploadSpec, artifactHandlerFunc, errorsQueue, flags)
	return performUploadTasks(producerConsumer, &uploadSummery, errorsQueue)
}

func prepareUploadTasks(producer parallel.Runner, uploadSpec *utils.SpecFiles, artifactHandlerFunc artifactContext, errorsQueue *utils.ErrorsQueue, flags *UploadFlags)  {
	go func() {
		collectFilesForUpload(uploadSpec, flags, producer, artifactHandlerFunc, errorsQueue)
	}()
}

func toBuildInfoArtifacts(artifactsBuildInfo [][]utils.ArtifactsBuildInfo) []utils.ArtifactsBuildInfo {
	var buildInfo []utils.ArtifactsBuildInfo
	for _, v := range artifactsBuildInfo {
		buildInfo = append(buildInfo, v...)
	}
	return buildInfo
}

func performUploadTasks(consumer parallel.Runner, uploadSummery *uploadResult, errorsQueue *utils.ErrorsQueue) (buildInfoArtifacts [][]utils.ArtifactsBuildInfo, totalUploaded, totalFailed int, err error) {
	// Blocking until we finish consuming for some reason
	consumer.Run()
	if e := errorsQueue.GetError(); e != nil {
		err = e
		return
	}
	totalUploaded = sumIntArray(uploadSummery.UploadCount)
	totalUploadAttempted := sumIntArray(uploadSummery.TotalCount)

	log.Info("Uploaded", strconv.Itoa(totalUploaded), "artifacts.")
	totalFailed = totalUploadAttempted - totalUploaded
	if totalFailed > 0 {
		log.Error("Failed uploading", strconv.Itoa(totalFailed), "artifacts.")
	}
	buildInfoArtifacts = uploadSummery.BuildInfoArtifacts
	return
}

func sumIntArray(arr []int) int {
	sum := 0
	for _, i := range arr {
		sum += i
	}
	return sum
}

func addBuildProps(uploadFile *utils.File, flags *UploadFlags) (err error) {
	if flags.BuildName == "" || flags.BuildNumber == "" {
		return
	}
	props := "build.name=" + flags.BuildName
	props += ";build.number=" + flags.BuildNumber
	buildGeneralDetails, err := utils.ReadBuildInfoGeneralDetails(flags.BuildName, flags.BuildNumber)
	if err != nil {
		return
	}
	props += ";build.timestamp=" + strconv.FormatInt(buildGeneralDetails.Timestamp.UnixNano() / int64(time.Millisecond), 10)
	uploadFile.Props = addProps(uploadFile.Props, props)
	return
}

func getSingleFileToUpload(rootPath, targetPath string, flat bool) cliutils.Artifact {
	var uploadPath string
	if !strings.HasSuffix(targetPath, "/") {
		uploadPath = targetPath
	} else {
		if flat {
			uploadPath, _ = fileutils.GetFileAndDirFromPath(rootPath)
			uploadPath = targetPath + uploadPath
		} else {
			uploadPath = targetPath + rootPath
			uploadPath = cliutils.TrimPath(uploadPath)
		}
	}
	symlinkPath, e := getFileSymlinkPath(rootPath)
	if e != nil {
		return cliutils.Artifact{}
	}
	return cliutils.Artifact{LocalPath: rootPath, TargetPath: uploadPath, Symlink: symlinkPath}
}

func addProps(oldProps, additionalProps string) string {
	if len(oldProps) > 0 && !strings.HasSuffix(oldProps, ";")  && len(additionalProps) > 0 {
		oldProps += ";"
	}
	return oldProps + additionalProps
}

func addSymlinkProps(props string, artifact cliutils.Artifact, flags *UploadFlags) (string, error) {
	artifactProps := ""
	artifactSymlink := artifact.Symlink
	if flags.Symlink && len(artifactSymlink) > 0 {
		sha1Property := ""
		fileInfo, err := os.Stat(artifact.LocalPath)
		if err != nil {
			return "", err
		}
		if !fileInfo.IsDir() {
			sha1, err := fileutils.CalcSha1(artifact.LocalPath)
			if err != nil {
				return "", err
			}
			sha1Property = ";" + utils.SYMLINK_SHA1 + "=" + sha1
		}
		artifactProps += utils.ARTIFACTORY_SYMLINK + "=" + artifactSymlink + sha1Property
	}
	artifactProps = addProps(props, artifactProps)
	return artifactProps, nil
}

func collectFilesForUpload(uploadSpec *utils.SpecFiles, flags *UploadFlags, producer parallel.Runner, artifactHandlerFunc artifactContext, errorsQueue *utils.ErrorsQueue) {
	defer producer.Done()
	for _, uploadFile := range uploadSpec.Files {
		addBuildProps(&uploadFile, flags)
		if strings.Index(uploadFile.Target, "/") < 0 {
			uploadFile.Target += "/"
		}
		uploadMetaData := uploadDescriptor{}
		uploadFile.Pattern = cliutils.ReplaceTildeWithUserHome(uploadFile.Pattern)
		uploadMetaData.CreateUploadDescriptor(uploadFile.Regexp, uploadFile.Flat, uploadFile.Pattern)
		if uploadMetaData.Err != nil {
			errorsQueue.AddError(uploadMetaData.Err)
			return
		}
		// If the path is a single file then return it
		if !uploadMetaData.IsDir || (flags.Symlink && fileutils.IsPathSymlink(uploadFile.Pattern)) {
			artifact := getSingleFileToUpload(uploadMetaData.RootPath, uploadFile.Target, uploadMetaData.IsFlat)
			props, err := addSymlinkProps(uploadFile.Props, artifact, flags)
			if err != nil {
				errorsQueue.AddError(err)
				return
			}
			uploadData := UploadData{Artifact:artifact, Props:props}
			task := artifactHandlerFunc(uploadData)
			producer.AddTaskWithError(task, errorsQueue.AddError)
			continue
		}
		uploadFile.Pattern = cliutils.PrepareLocalPathForUpload(uploadFile.Pattern, uploadMetaData.IsRegexp)
		err := collectPatternMatchingFiles(uploadFile, uploadMetaData, producer, artifactHandlerFunc, errorsQueue, flags)
		if err != nil {
			errorsQueue.AddError(err)
			return
		}
	}
}

func getRootPath(pattern string, isRegexp bool) (string, error){
	rootPath := cliutils.GetRootPathForUpload(pattern, isRegexp)
	if !fileutils.IsPathExists(rootPath) {
		err := cliutils.CheckError(errors.New("Path does not exist: " + rootPath))
		if err != nil {
			return "", err
		}
	}
	return rootPath, nil
}

// If filePath is path to a symlink we should return the link content e.g where the link points
func getFileSymlinkPath(filePath string) (string, error){
	fileInfo, e := os.Lstat(filePath)
	if cliutils.CheckError(e) != nil {
		return "", e
	}
	var symlinkPath = ""
	if fileutils.IsFileSymlink(fileInfo) {
		symlinkPath, e = os.Readlink(filePath)
		if cliutils.CheckError(e) != nil {
			return "", e
		}
	}
	return symlinkPath, nil
}

func getUploadPaths(isRecursiveString, rootPath string, includeDirs bool, flags *UploadFlags) ([]string, error) {
	var paths []string
	isRecursive, err := cliutils.StringToBool(isRecursiveString, true)
	if err != nil {
		return paths, err
	}
	if isRecursive {
		paths, err = fileutils.ListFilesRecursiveWalkIntoDirSymlink(rootPath, !flags.Symlink)
	} else {
		paths, err = fileutils.ListFiles(rootPath, includeDirs)
	}
	if err != nil {
		return paths, err
	}
	return paths, nil
}

func collectPatternMatchingFiles(uploadFile utils.File, uploadMetaData uploadDescriptor, producer parallel.Runner, artifactHandlerFunc artifactContext, errorsQueue *utils.ErrorsQueue, flags *UploadFlags) error {
	r, err := regexp.Compile(uploadFile.Pattern)
	if cliutils.CheckError(err) != nil {
		return err
	}

	paths, err := getUploadPaths(uploadFile.Recursive, uploadMetaData.RootPath, uploadFile.IsIncludeDirs(), flags)
	if err != nil {
		return err
	}
	// Longest paths first
	sort.Sort(sort.Reverse(sort.StringSlice(paths)))
	// 'foldersPaths' is a subset of the 'paths' array. foldersPaths is in use only when we need to upload folders with flat=true.
	// 'foldersPaths' will contain only the directories paths which are in the 'paths' array.
	var foldersPaths[]string
	for index, path := range paths {
		isDir, err := fileutils.IsDir(path)
		if err != nil {
			return err
		}
		isSymlinkFlow := flags.Symlink && fileutils.IsPathSymlink(path)
		if isDir && !uploadFile.IsIncludeDirs() && !isSymlinkFlow {
			continue
		}
		groups := r.FindStringSubmatch(path)
		size := len(groups)
		target := uploadFile.Target
		if size > 0 {
			tempPaths := paths
			tempIndex := index
			// In case we need to upload directories with flat=true, we want to avoid the creation of unnecessary paths in Artifactory.
			// To achieve this, we need to take into consideration the directories which had already been uploaded, ignoring all files paths.
			// When flat=false we take into consideration folder paths which were created implicitly by file upload
			if uploadMetaData.IsFlat && uploadFile.IsIncludeDirs() && isDir {
				foldersPaths = append(foldersPaths, path)
				tempPaths = foldersPaths
				tempIndex = len(foldersPaths) - 1
			}
			taskData := &uploadTaskData{target: target, path: path, isDir: isDir, isSymlinkFlow: isSymlinkFlow, paths: tempPaths,
				groups:                     groups, index: tempIndex, size: size, uploadFile: uploadFile, uploadMetaData: uploadMetaData,
				producer:                   producer, artifactHandlerFunc: artifactHandlerFunc, errorsQueue: errorsQueue, flags: flags,
			}
			createUploadTask(taskData)
		}
	}
	return nil
}

type uploadTaskData struct {
	target              string
	path                string
	isDir               bool
	isSymlinkFlow       bool
	paths               []string
	groups              []string
	index               int
	size                int
	uploadFile          utils.File
	uploadMetaData      uploadDescriptor
	producer            parallel.Runner
	artifactHandlerFunc artifactContext
	errorsQueue         *utils.ErrorsQueue
	flags               *UploadFlags
}

func createUploadTask(taskData *uploadTaskData) error {
	for i := 1; i < taskData.size; i++ {
		group := strings.Replace(taskData.groups[i], "\\", "/", -1)
		taskData.target = strings.Replace(taskData.target, "{"+strconv.Itoa(i)+"}", group, -1)
	}
	var task parallel.TaskFunc
	taskData.target = getUploadTarget(taskData.uploadMetaData.IsFlat, taskData.path, taskData.target)
	// If case taskData.path is a symlink we get the symlink link path.
	symlinkPath, e := getFileSymlinkPath(taskData.path)
	if e != nil {
		return e
	}
	artifact := cliutils.Artifact{LocalPath: taskData.path, TargetPath: taskData.target, Symlink: symlinkPath}
	props, e := addSymlinkProps(taskData.uploadFile.Props, artifact, taskData.flags)
	if e != nil {
		return e
	}
	uploadData := UploadData{Artifact: artifact, Props: props}
	if taskData.isDir && taskData.uploadFile.IsIncludeDirs() && !taskData.isSymlinkFlow {
		if taskData.path != "." && (taskData.index == 0 || !utils.IsSubPath(taskData.paths, taskData.index, fileutils.GetFileSeperator())) {
			uploadData.IsDir = true
		} else {
			return nil
		}
	}
	task = taskData.artifactHandlerFunc(uploadData)
	taskData.producer.AddTaskWithError(task, taskData.errorsQueue.AddError)
	return nil
}

func getUploadTarget(isFlat bool, path, target string) string {
	if strings.HasSuffix(target, "/") {
		if isFlat {
			fileName, _ := fileutils.GetFileAndDirFromPath(path)
			target += fileName
		} else {
			target += cliutils.TrimPath(path)
		}
	}
	return target
}

func addPropsToTargetPath(targetPath, props string, flags *UploadFlags) (string, error) {
	if props != "" {
		encodedProp, err := utils.EncodeParams(props)
		if err != nil {
			return "", err
		}
		targetPath += ";" + encodedProp
	}
	if flags.Deb != "" {
		targetPath += getDebianMatrixParams(flags.Deb)
	}
	return targetPath, nil
}

func prepareUploadData(targetPath, localPath, props, logMsgPrefix string, flags *UploadFlags) (os.FileInfo, string, string, error) {
	fileName, _ := fileutils.GetFileAndDirFromPath(targetPath)
	targetPath, err := addPropsToTargetPath(targetPath, props, flags)
	if cliutils.CheckError(err) != nil {
		return nil, "", "", err
	}
	log.Info(logMsgPrefix + "Uploading artifact:", localPath)
	file, err := os.Open(localPath)
	defer file.Close()
	if cliutils.CheckError(err) != nil {
		return nil, "", "", err
	}
	fileInfo, err := file.Stat()
	if cliutils.CheckError(err) != nil {
		return nil, "", "", err
	}
	return fileInfo, targetPath, fileName, nil
}

// Uploads the file in the specified local path to the specified target path.
// Returns true if the file was successfully uploaded.
func uploadFile(localPath, targetPath, props string, flags *UploadFlags, minChecksumDeploySize int64, logMsgPrefix string) (utils.ArtifactsBuildInfo, bool, error) {
	fileInfo, targetPath, fileName, err := prepareUploadData(targetPath, localPath, props, logMsgPrefix, flags)
	if err != nil {
		return utils.ArtifactsBuildInfo{}, false, err
	}
	file, err := os.Open(localPath)
	defer file.Close()
	if cliutils.CheckError(err) != nil {
		return utils.ArtifactsBuildInfo{}, false, err
	}
	var checksumDeployed bool = false
	var resp *http.Response
	var details *fileutils.FileDetails
	var body []byte
	httpClientsDetails := utils.GetArtifactoryHttpClientDetails(flags.ArtDetails)
	fileStat, err := os.Lstat(localPath)
	if cliutils.CheckError(err) != nil {
		return utils.ArtifactsBuildInfo{}, false, err
	}
	if flags.Symlink && fileutils.IsFileSymlink(fileStat) {
		resp, details, body, err = uploadSymlink(targetPath, httpClientsDetails, flags)
		if err != nil {
			return utils.ArtifactsBuildInfo{}, false, err
		}
	} else {
		resp, details, body, err = doUpload(file, localPath, targetPath, logMsgPrefix, httpClientsDetails, fileInfo, minChecksumDeploySize, flags)
	}
	if err != nil {
		return utils.ArtifactsBuildInfo{}, false, err
	}
	logUploadResponse(logMsgPrefix, resp, body, checksumDeployed, flags)
	artifact := utils.CreateArtifactsBuildInfo(fileName, details)
	return artifact, flags.DryRun || checksumDeployed || resp.StatusCode == 201 || resp.StatusCode == 200, nil
}

func uploadSymlink(targetPath string, httpClientsDetails httputils.HttpClientDetails, flags *UploadFlags) (resp *http.Response, details *fileutils.FileDetails, body []byte, err error) {
	details = createSymlinkFileDetails()
	resp, body, err = utils.UploadFile(nil, targetPath, flags.ArtDetails, details, httpClientsDetails)
	return
}

func doUpload(file *os.File, localPath, targetPath, logMsgPrefix string, httpClientsDetails httputils.HttpClientDetails, fileInfo os.FileInfo, minChecksumDeploySize int64, flags *UploadFlags) (*http.Response, *fileutils.FileDetails, []byte, error) {
	var details *fileutils.FileDetails
	var checksumDeployed bool
	var resp *http.Response
	var body []byte
	var err error
	addExplodeHeader(&httpClientsDetails, flags.ExplodeArchive)
	if fileInfo.Size() >= minChecksumDeploySize && !flags.ExplodeArchive {
		resp, details, body, err = tryChecksumDeploy(localPath, targetPath, flags, httpClientsDetails)
		if err != nil {
			return resp, details, body, err
		}
		checksumDeployed = !flags.DryRun && (resp.StatusCode == 201 || resp.StatusCode == 200)
	}
	if !flags.DryRun && !checksumDeployed {
		var body []byte
		resp, body, err = utils.UploadFile(file, targetPath, flags.ArtDetails, details, httpClientsDetails)
		if err != nil {
			return resp, details, body, err
		}
		if resp.StatusCode != 201 && resp.StatusCode != 200 {
			log.Error(logMsgPrefix + "Artifactory response: " + resp.Status + "\n" + cliutils.IndentJson(body))
		}
	}
	if !flags.DryRun {
		var strChecksumDeployed string
		if checksumDeployed {
			strChecksumDeployed = " (Checksum deploy)"
		}
		log.Debug(logMsgPrefix, "Artifactory response:", resp.Status, strChecksumDeployed)
	}
	if details == nil {
		details, err = fileutils.GetFileDetails(localPath)
	}
	return resp, details, body, err
}

func logUploadResponse(logMsgPrefix string, resp *http.Response, body []byte, checksumDeployed bool, flags *UploadFlags) {
	if resp != nil && resp.StatusCode != 201 && resp.StatusCode != 200 {
		log.Error(logMsgPrefix + "Artifactory response: " + resp.Status + "\n" + cliutils.IndentJson(body))
	}
	if !flags.DryRun {
		var strChecksumDeployed string
		if checksumDeployed {
			strChecksumDeployed = " (Checksum deploy)"
		} else {
			strChecksumDeployed = ""
		}
		log.Debug(logMsgPrefix, "Artifactory response:", resp.Status, strChecksumDeployed)
	}
}

// When handling symlink we want to simulate the creation of  empty file
func createSymlinkFileDetails() *fileutils.FileDetails {
	details := new(fileutils.FileDetails)
	details.Checksum.Md5, _ = fileutils.GetMd5(bytes.NewBuffer([]byte(fileutils.SYMLINK_FILE_CONTENT)))
	details.Checksum.Sha1, _ = fileutils.GetSha1(bytes.NewBuffer([]byte(fileutils.SYMLINK_FILE_CONTENT)))
	details.Size = int64(0)
	return details
}

func addExplodeHeader(httpClientsDetails *httputils.HttpClientDetails, isExplode bool) {
	if isExplode {
		utils.AddHeader("X-Explode-Archive", "true", &httpClientsDetails.Headers)
	}
}

func getMinChecksumDeploySize() (int64, error) {
	minChecksumDeploySize := os.Getenv("JFROG_CLI_MIN_CHECKSUM_DEPLOY_SIZE_KB")
	if minChecksumDeploySize == "" {
		return 10240, nil
	}
	minSize, err := strconv.ParseInt(minChecksumDeploySize, 10, 64)
	err = cliutils.CheckError(err)
	if err != nil {
		return 0, err
	}
	return minSize * 1000, nil
}

func tryChecksumDeploy(filePath, targetPath string, flags *UploadFlags,
	httpClientsDetails httputils.HttpClientDetails) (resp *http.Response, details *fileutils.FileDetails, body []byte, err error) {

	details, err = fileutils.GetFileDetails(filePath)
	if err != nil {
		return
	}
	headers := make(map[string]string)
	headers["X-Checksum-Deploy"] = "true"
	headers["X-Checksum-Sha1"] = details.Checksum.Sha1
	headers["X-Checksum-Md5"] = details.Checksum.Md5
	requestClientDetails := httpClientsDetails.Clone()
	cliutils.MergeMaps(headers, requestClientDetails.Headers)
	if flags.DryRun {
		return
	}
	utils.AddAuthHeaders(headers, flags.ArtDetails)
	cliutils.MergeMaps(headers, requestClientDetails.Headers)
	resp, body, err = httputils.SendPut(targetPath, nil, *requestClientDetails)
	return
}

func getDebianMatrixParams(debianPropsStr string) string {
	debProps := strings.Split(debianPropsStr, "/")
	return ";deb.distribution=" + debProps[0] +
		";deb.component=" + debProps[1] +
		";deb.architecture=" + debProps[2]
}

type UploadFlags struct {
	ArtDetails     *config.ArtifactoryDetails
	DryRun         bool
	ExplodeArchive bool
	Deb            string
	Threads        int
	BuildName      string
	BuildNumber    string
	Symlink        bool
}

func (flags *UploadFlags) GetArtifactoryDetails() *config.ArtifactoryDetails {
	return flags.ArtDetails
}

func (flags *UploadFlags) IsDryRun() bool {
	return flags.DryRun
}

type UploadData struct {
	Artifact cliutils.Artifact
	Props    string
	IsDir    bool
}

type uploadDescriptor struct {
	IsFlat   bool
	IsRegexp bool
	IsDir    bool
	RootPath string
	Err      error
}

func (p *uploadDescriptor) CreateUploadDescriptor(isRegexp, isFlat, pattern string) {
	p.isRegexp(isRegexp)
	p.isFlat(isFlat)
	p.setRootPath(pattern)
	p.checkIfDir()
}

func (p *uploadDescriptor) isRegexp(isRegexpString string) {
	if p.Err == nil {
		p.IsRegexp, p.Err = cliutils.StringToBool(isRegexpString, false)
	}
}

func (p *uploadDescriptor) isFlat(isFlatString string) {
	if p.Err == nil {
		p.IsFlat, p.Err = cliutils.StringToBool(isFlatString, true)
	}
}

func (p *uploadDescriptor) setRootPath(pattern string) {
	if p.Err == nil {
		p.RootPath, p.Err = getRootPath(pattern, p.IsRegexp)
	}
}

func (p *uploadDescriptor) checkIfDir() {
	if p.Err == nil {
		p.IsDir, p.Err = fileutils.IsDir(p.RootPath)
	}
}

type uploadResult struct {
	UploadCount           []int
	TotalCount            []int
	BuildInfoArtifacts    [][]utils.ArtifactsBuildInfo
}

type artifactContext func(UploadData) parallel.TaskFunc

func createArtifactHandlerFunc(s *uploadResult, minChecksumDeploySize int64, flags *UploadFlags) artifactContext {
	return func(artifact UploadData) parallel.TaskFunc {
		return func(threadId int) (e error) {
			if artifact.IsDir {
				createFolderInArtifactory(artifact, flags)
				return
			}
			var uploaded bool
			var target string
			var buildInfoArtifact utils.ArtifactsBuildInfo
			logMsgPrefix := cliutils.GetLogMsgPrefix(threadId, flags.DryRun)
			target, e = utils.BuildArtifactoryUrl(flags.ArtDetails.Url, artifact.Artifact.TargetPath, make(map[string]string))
			if e != nil {
				return
			}
			buildInfoArtifact, uploaded, e = uploadFile(artifact.Artifact.LocalPath, target, artifact.Props, flags, minChecksumDeploySize, logMsgPrefix)
			if e != nil {
				return
			}
			if uploaded {
				s.UploadCount[threadId]++
				s.BuildInfoArtifacts[threadId] = append(s.BuildInfoArtifacts[threadId], buildInfoArtifact)
			}
			s.TotalCount[threadId]++
			return
		}
	}
}

func createFolderInArtifactory(artifact UploadData, flags *UploadFlags) error {
	url, err := utils.BuildArtifactoryUrl(flags.ArtDetails.Url, artifact.Artifact.TargetPath, make(map[string]string))
	url = cliutils.AddTrailingSlashIfNeeded(url)
	if err != nil {
		return err
	}
	content := make([]byte, 0)
	httpClientsDetails := utils.GetArtifactoryHttpClientDetails(flags.ArtDetails)
	resp, body, err := httputils.SendPut(url, content, httpClientsDetails)
	if err != nil {
		log.Debug(resp)
		return err
	}
	logUploadResponse("Uploaded folder :", resp, body, false, flags)
	return err
}