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
	producerConsumer := utils.NewProducerConsumer(flags.Threads)
	prepareUploadTasks(producerConsumer, uploadSpec, artifactHandlerFunc, flags)
	return performUploadTasks(producerConsumer, &uploadSummery)
}

func prepareUploadTasks(producer utils.Producer, uploadSpec *utils.SpecFiles, artifactHandlerFunc artifactContext, flags *UploadFlags)  {
	go func() {
		collectFilesForUpload(uploadSpec, flags, producer, artifactHandlerFunc)
	}()
}

func toBuildInfoArtifacts(artifactsBuildInfo [][]utils.ArtifactsBuildInfo) []utils.ArtifactsBuildInfo {
	var buildInfo []utils.ArtifactsBuildInfo
	for _, v := range artifactsBuildInfo {
		buildInfo = append(buildInfo, v...)
	}
	return buildInfo
}

func performUploadTasks(consumer utils.Consumer, uploadSummery *uploadResult) (buildInfoArtifacts [][]utils.ArtifactsBuildInfo, totalUploaded, totalFailed int, err error) {
	// Blocking until we finish consuming for some reason
	consumer.Consume()
	if e := consumer.GetError(); e != nil {
		err = e
		return
	}
	if err != nil {
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
	props := uploadFiles.Props
	if (props != "") {
		props += ";"
	}
	props += "build.name=" + flags.BuildName
	props += ";build.number=" + flags.BuildNumber
	buildGeneralDetails, err := utils.ReadBuildInfoGeneralDetails(flags.BuildName, flags.BuildNumber)
	if err != nil {
		return
	}
	props += ";build.timestamp=" + strconv.FormatInt(buildGeneralDetails.Timestamp.UnixNano() / int64(time.Millisecond), 10)
	uploadFiles.Props = props
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
	return cliutils.Artifact{LocalPath: rootPath, TargetPath: uploadPath}
}

func collectFilesForUpload(uploadSpec *utils.SpecFiles, flags *UploadFlags, producer utils.Producer, artifactHandlerFunc artifactContext) {
	defer producer.Finish()
	for _, uploadFile := range uploadSpec.Files {
		addBuildProps(&uploadFile, flags)
		if strings.Index(uploadFile.Target, "/") < 0 {
			uploadFile.Target += "/"
		}
		uploadMetaData := uploadDescriptor{}
		uploadFile.Pattern = cliutils.ReplaceTildeWithUserHome(uploadFile.Pattern)
		uploadMetaData.CreateUploadDescriptor(uploadFile.Regexp, uploadFile.Flat, uploadFile.Pattern)
		if uploadMetaData.Err != nil {
			producer.SetError(uploadMetaData.Err)
			return
		}
		// If the path is a single file then return it
		if !uploadMetaData.IsDir {
			artifact := getSingleFileToUpload(uploadMetaData.RootPath, uploadFile.Target, uploadMetaData.IsFlat)
			uploadData := UploadData{artifact, uploadFile.Props}
			task := artifactHandlerFunc(uploadData)
			producer.Produce(task)
			continue
		}
		uploadFile.Pattern = cliutils.PrepareLocalPathForUpload(uploadFile.Pattern, uploadMetaData.IsRegexp)
		err := collectPatternMatchingFiles(uploadFile, uploadMetaData, producer, artifactHandlerFunc)
		if err != nil {
			producer.SetError(err)
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

func getUploadPaths(isRecursiveString, rootPath string) ([]string, error) {
	var paths []string
	isRecursive, err := cliutils.StringToBool(isRecursiveString, true)
	if err != nil {
		return paths, err
	}
	if isRecursive {
		paths, err = ioutils.ListFilesRecursive(rootPath)
	} else {
		paths, err = ioutils.ListFiles(rootPath)
	}
	if err != nil {
		return paths, err
	}
	return paths, nil
}

func collectPatternMatchingFiles(uploadFile utils.Files, uploadMetaData uploadDescriptor, producer utils.Producer, artifactHandlerFunc artifactContext) error {
	r, err := regexp.Compile(uploadFile.Pattern)
	if cliutils.CheckError(err) != nil {
		return err
	}

	paths, err := getUploadPaths(uploadFile.Recursive, uploadMetaData.RootPath)
	if err != nil {
		return err
	}
	for _, path := range paths {
		dir, err := ioutils.IsDir(path)
		if err != nil {
			return err
		}
		if dir {
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
			uploadData := UploadData{cliutils.Artifact{path, target}, uploadFile.Props}
			task := artifactHandlerFunc(uploadData)
			producer.Produce(task)
		}
	}
	return nil
}
// Uploads the file in the specified local path to the specified target path.
// Returns true if the file was successfully uploaded.
func uploadFile(localPath, targetPath, props string, flags *UploadFlags, minChecksumDeploySize int64, logMsgPrefix string) (utils.ArtifactsBuildInfo, bool, error) {
	fileName, _ := ioutils.GetFileAndDirFromPath(targetPath)
	if props != "" {
		encodedProp, err := utils.EncodeParams(props)
		if err != nil {
			return utils.ArtifactsBuildInfo{}, false, err
		}
		targetPath += ";" + encodedProp
	}
	if flags.Deb != "" {
		targetPath += getDebianMatrixParams(flags.Deb)
	}

	log.Info(logMsgPrefix + "Uploading artifact:", localPath)
	file, err := os.Open(localPath)
	err = cliutils.CheckError(err)
	if err != nil {
		return utils.ArtifactsBuildInfo{}, false, err
	}
	defer file.Close()
	fileInfo, err := file.Stat()
	err = cliutils.CheckError(err)
	if err != nil {
		return utils.ArtifactsBuildInfo{}, false, err
	}

	var checksumDeployed bool = false
	var resp *http.Response
	var details *ioutils.FileDetails
	httpClientsDetails := utils.GetArtifactoryHttpClientDetails(flags.ArtDetails)
	addExplodeHeader(&httpClientsDetails, flags.ExplodeArchive)
	if fileInfo.Size() >= minChecksumDeploySize && !flags.ExplodeArchive {
		resp, details, err = tryChecksumDeploy(localPath, targetPath, flags, httpClientsDetails)
		if err != nil {
			return utils.ArtifactsBuildInfo{}, false, err
		}
		checksumDeployed = !flags.DryRun && (resp.StatusCode == 201 || resp.StatusCode == 200)
	}
	if !flags.DryRun && !checksumDeployed {
		var body []byte
		resp, body, err = utils.UploadFile(file, targetPath, flags.ArtDetails, details, httpClientsDetails)
		if err != nil {
			return utils.ArtifactsBuildInfo{}, false, err
		}
		if resp.StatusCode != 201 && resp.StatusCode != 200 {
			log.Error(logMsgPrefix + "Artifactory response: " + resp.Status + "\n" + cliutils.IndentJson(body))
		}
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
	if (details == nil) {
		details, err = ioutils.GetFileDetails(localPath)
	}
	artifact := createBuildArtifactItem(fileName, details)
	return artifact, (flags.DryRun || checksumDeployed || resp.StatusCode == 201 || resp.StatusCode == 200), nil
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
httpClientsDetails ioutils.HttpClientDetails) (resp *http.Response, details *ioutils.FileDetails, err error) {

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
	resp, _, err = ioutils.SendPut(targetPath, nil, *requestClientDetails)
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
