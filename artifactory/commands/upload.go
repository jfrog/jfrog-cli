package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"errors"
	"runtime"
	"time"
	"path/filepath"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/log"
	"bytes"
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
	producerConsumer := utils.NewProducerConsumer(flags.Threads, true)
	errorsQueue := utils.NewErrorsQueue(1)
	prepareUploadTasks(producerConsumer, uploadSpec, artifactHandlerFunc, errorsQueue, flags)
	return performUploadTasks(producerConsumer, &uploadSummery, errorsQueue)
}

func prepareUploadTasks(producer utils.Producer, uploadSpec *utils.SpecFiles, artifactHandlerFunc artifactContext, errorsQueue *utils.ErrorsQueue, flags *UploadFlags)  {
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

func performUploadTasks(consumer utils.Consumer, uploadSummery *uploadResult, errorsQueue *utils.ErrorsQueue) (buildInfoArtifacts [][]utils.ArtifactsBuildInfo, totalUploaded, totalFailed int, err error) {
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

func addBuildProps(uploadFiles *utils.Files, flags *UploadFlags) (err error) {
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
	uploadFiles.Props = addProps(uploadFiles.Props, props)
	return
}

func getSingleFileToUpload(rootPath, targetPath string, flat bool) cliutils.Artifact {
	var uploadPath string
	if !strings.HasSuffix(targetPath, "/") {
		uploadPath = targetPath
	} else {
		if flat {
			uploadPath, _ = ioutils.GetFileAndDirFromPath(rootPath)
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
			sha1, err := ioutils.CalcSha1(artifact.LocalPath)
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

func collectFilesForUpload(uploadSpec *utils.SpecFiles, flags *UploadFlags, producer utils.Producer, artifactHandlerFunc artifactContext, errorsQueue *utils.ErrorsQueue) {
	defer producer.Close()
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
		if !uploadMetaData.IsDir || (flags.Symlink && ioutils.IsPathSymlink(uploadFile.Pattern)) {
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
	if !ioutils.IsPathExists(rootPath) {
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
	if ioutils.IsFileSymlink(fileInfo) {
		symlinkPath, e = os.Readlink(filePath)
		if cliutils.CheckError(e) != nil {
			return "", e
		}
	}
	return symlinkPath, nil
}

func getUploadPaths(isRecursiveString, rootPath string, flags *UploadFlags) ([]string, error) {
	var paths []string
	isRecursive, err := cliutils.StringToBool(isRecursiveString, true)
	if err != nil {
		return paths, err
	}
	if isRecursive {
		paths, err = ioutils.ListFilesRecursiveWalkIntoDirSymlink(rootPath, !flags.Symlink)
	} else {
		paths, err = ioutils.ListFiles(rootPath)
	}
	if err != nil {
		return paths, err
	}
	return paths, nil
}

func collectPatternMatchingFiles(uploadFile utils.Files, uploadMetaData uploadDescriptor, producer utils.Producer, artifactHandlerFunc artifactContext, errorsQueue *utils.ErrorsQueue, flags *UploadFlags) error {
	r, err := regexp.Compile(uploadFile.Pattern)
	if cliutils.CheckError(err) != nil {
		return err
	}

	paths, err := getUploadPaths(uploadFile.Recursive, uploadMetaData.RootPath, flags)
	if err != nil {
		return err
	}
	for _, path := range paths {
		dir, err := ioutils.IsDir(path)
		if err != nil {
			return err
		}
		if dir && (!flags.Symlink || !ioutils.IsPathSymlink(path)) {
			continue
		}
		groups := r.FindStringSubmatch(path)
		size := len(groups)
		target := uploadFile.Target
		if size > 0 {
			for i := 1; i < size; i++ {
				group := strings.Replace(groups[i], "\\", "/", -1)
				target = strings.Replace(target, "{" + strconv.Itoa(i) + "}", group, -1)
			}
			if strings.HasSuffix(target, "/") {
				if uploadMetaData.IsFlat {
					fileName, _ := ioutils.GetFileAndDirFromPath(path)
					target += fileName
				} else {
					uploadPath := cliutils.TrimPath(path)
					target += uploadPath
				}
			}
			symlinkPath, e := getFileSymlinkPath(path)
			if e != nil {
				return e
			}
			artifact := cliutils.Artifact{LocalPath:path, TargetPath:target, Symlink:symlinkPath}
			props, e := addSymlinkProps(uploadFile.Props, artifact, flags)
			if e != nil {
				return e
			}
			uploadData := UploadData{Artifact:artifact, Props:props}
			task := artifactHandlerFunc(uploadData)
			producer.AddTaskWithError(task, errorsQueue.AddError)
		}
	}
	return nil
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
	fileName, _ := ioutils.GetFileAndDirFromPath(targetPath)
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
	var details *ioutils.FileDetails
	var body []byte
	httpClientsDetails := utils.GetArtifactoryHttpClientDetails(flags.ArtDetails)
	fileStat, err := os.Lstat(localPath)
	if cliutils.CheckError(err) != nil {
		return utils.ArtifactsBuildInfo{}, false, err
	}
	if flags.Symlink && ioutils.IsFileSymlink(fileStat) {
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
	artifact := createBuildArtifactItem(fileName, details)
	return artifact, (flags.DryRun || checksumDeployed || resp.StatusCode == 201 || resp.StatusCode == 200), nil
}

func uploadSymlink(targetPath string, httpClientsDetails ioutils.HttpClientDetails, flags *UploadFlags) (resp *http.Response, details *ioutils.FileDetails, body []byte, err error) {
	details = createSymlinkFileDetails()
	resp, body, err = utils.UploadFile(nil, targetPath, flags.ArtDetails, details, httpClientsDetails)
	return
}

func doUpload(file *os.File, localPath, targetPath, logMsgPrefix string, httpClientsDetails ioutils.HttpClientDetails, fileInfo os.FileInfo, minChecksumDeploySize int64, flags *UploadFlags) (*http.Response, *ioutils.FileDetails, []byte, error) {
	var details *ioutils.FileDetails
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
		details, err = ioutils.GetFileDetails(localPath)
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
	if (details == nil) {
		details, err = ioutils.GetFileDetails(localPath)
	}
	return resp, details, body, err
}

func logUploadResponse(logMsgPrefix string, resp *http.Response, body []byte, checksumDeployed bool, flags *UploadFlags) {
	if resp == nil {
		log.Error(logMsgPrefix + "Artifactory response is not accessible." )
		return
	}
	if resp.StatusCode != 201 && resp.StatusCode != 200 {
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
func createSymlinkFileDetails() *ioutils.FileDetails {
	const symlinkBody = ""
	details := new(ioutils.FileDetails)
	details.Md5, _ = ioutils.GetMd5(bytes.NewBuffer([]byte(ioutils.SYMLINK_FILE_CONTENT)))
	details.Sha1, _ = ioutils.GetSha1(bytes.NewBuffer([]byte(ioutils.SYMLINK_FILE_CONTENT)))
	details.Size = int64(0)
	return details
}

func createBuildArtifactItem(fileName string, details *ioutils.FileDetails) utils.ArtifactsBuildInfo {
	return utils.ArtifactsBuildInfo{
		Name: fileName,
		BuildInfoCommon : &utils.BuildInfoCommon{
			Sha1: details.Sha1,
			Md5: details.Md5,
		},
	}
}

func addExplodeHeader(httpClientsDetails *ioutils.HttpClientDetails, isExplode bool) {
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

func extractOnlyFileNameFromPath(file string) string {
	if runtime.GOOS == "windows" {
		splitedPath := strings.Split(file, "\\")
		return splitedPath[len(splitedPath) - 1]
	} else {
		_, fileName := filepath.Split(file)
		return fileName
	}
}

func tryChecksumDeploy(filePath, targetPath string, flags *UploadFlags,
httpClientsDetails ioutils.HttpClientDetails) (resp *http.Response, details *ioutils.FileDetails, body []byte, err error) {

	details, err = ioutils.GetFileDetails(filePath)
	if err != nil {
		return
	}
	headers := make(map[string]string)
	headers["X-Checksum-Deploy"] = "true"
	headers["X-Checksum-Sha1"] = details.Sha1
	headers["X-Checksum-Md5"] = details.Md5
	requestClientDetails := httpClientsDetails.Clone()
	cliutils.MergeMaps(headers, requestClientDetails.Headers)
	if flags.DryRun {
		return
	}
	utils.AddAuthHeaders(headers, flags.ArtDetails)
	cliutils.MergeMaps(headers, requestClientDetails.Headers)
	resp, body, err = ioutils.SendPut(targetPath, nil, *requestClientDetails)
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
		p.IsDir, p.Err = ioutils.IsDir(p.RootPath)
	}
}

type uploadResult struct {
	UploadCount           []int
	TotalCount            []int
	BuildInfoArtifacts    [][]utils.ArtifactsBuildInfo
}

type artifactContext func(UploadData) utils.Task

func createArtifactHandlerFunc(s *uploadResult, minChecksumDeploySize int64, flags *UploadFlags) artifactContext {
	return func(artifact UploadData) utils.Task {
		return func(threadId int) (e error) {
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
