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
	"sync"
	"errors"
	"runtime"
	"time"
	"path/filepath"
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/logger"
)

// Uploads the artifacts in the specified local path pattern to the specified target path.
// Returns the total number of artifacts successfully uploaded.
func Upload(uploadSpec *utils.SpecFiles, flags *UploadFlags) (totalUploaded, totalFailed int, err error) {
	utils.PreCommandSetup(flags)
	isCollectBuildInfo := len(flags.BuildName) > 0 && len(flags.BuildNumber) > 0
	if isCollectBuildInfo && !flags.DryRun {
		if err := utils.SaveBuildGeneralDetails(flags.BuildName, flags.BuildNumber); err != nil {
			return 0, 0, err
		}
	}
	uploadData, err := buildUploadData(uploadSpec, flags)
	if err != nil {
		return 0, 0, err
	}

	buildArtifacts, totalUploaded, totalFailed, err := uploadWildcard(uploadData, flags)
	if err != nil {
		return 0, 0, err
	}
	if isCollectBuildInfo && !flags.DryRun {
		populateFunc := func(tempWrapper *utils.ArtifactBuildInfoWrapper) {
			tempWrapper.Artifacts = toBuildInfoArtifacts(buildArtifacts)
		}
		err = utils.PrepareBuildInfoForSave(flags.BuildName, flags.BuildNumber, populateFunc)
	}
	return
}

func buildUploadData(uploadSpec *utils.SpecFiles, flags *UploadFlags) ([]UploadData, error) {
	var result []UploadData
	for _, v := range uploadSpec.Files {
		artifacts, err := getFilesToUpload(&v)
		if err != nil {
			return nil, err
		}
		addBuildProps(&v, flags)
		for _, artifact := range artifacts {
			result = append(result, UploadData{
				Artifact: artifact,
				Props: v.Props,
			})
		}
	}
	return result, nil
}

func toBuildInfoArtifacts(artifactsBuildInfo [][]utils.ArtifactBuildInfo) []utils.ArtifactBuildInfo {
	var buildInfo []utils.ArtifactBuildInfo
	for _, v := range artifactsBuildInfo {
		buildInfo = append(buildInfo, v...)
	}
	return buildInfo
}

func uploadWildcard(artifacts []UploadData, flags *UploadFlags) (buildInfoArtifacts [][]utils.ArtifactBuildInfo, totalUploaded, totalFailed int, err error) {
	minChecksumDeploySize, e := getMinChecksumDeploySize()
	if e != nil {
		err = e
		return
	}

	size := len(artifacts)
	var wg sync.WaitGroup

	// Create an array of integers, to store the total file that were uploaded successfully.
	// Each array item is used by a single thread.
	uploadCount := make([]int, flags.Threads, flags.Threads)
	buildInfoArtifacts = make([][]utils.ArtifactBuildInfo, flags.Threads)
	for i := 0; i < flags.Threads; i++ {
		wg.Add(1)
		go func(threadId int) {
			logMsgPrefix := cliutils.GetLogMsgPrefix(threadId, flags.DryRun)
			for j := threadId; j < size && err == nil; j += flags.Threads {
				var e error
				var uploaded bool
				var target string
				var buildInfoArtifact utils.ArtifactBuildInfo
				target, e = utils.BuildArtifactoryUrl(flags.ArtDetails.Url, artifacts[j].Artifact.TargetPath, make(map[string]string))
				if e != nil {
					err = e
					break
				}
				buildInfoArtifact, uploaded, e = uploadFile(artifacts[j].Artifact.LocalPath, target, artifacts[j].Props, flags, minChecksumDeploySize, logMsgPrefix)
				if e != nil {
					err = e
					break
				}
				if uploaded {
					uploadCount[threadId]++
					buildInfoArtifacts[threadId] = append(buildInfoArtifacts[threadId], buildInfoArtifact)
				}
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	if err != nil {
		return
	}
	totalUploaded = 0
	for _, i := range uploadCount {
		totalUploaded += i
	}

	logger.Logger.Info("Uploaded " + strconv.Itoa(totalUploaded) + " artifacts to Artifactory.")
	totalFailed = size - totalUploaded
	if totalFailed > 0 {
		logger.Logger.Info("Failed uploading " + strconv.Itoa(totalFailed) + " artifacts to Artifactory.")
	}
	return
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

func getFilesToUpload(uploadFiles *utils.Files) ([]cliutils.Artifact, error) {
	if strings.Index(uploadFiles.Target, "/") < 0 {
		uploadFiles.Target += "/"
	}
	isRegexpe, err := cliutils.StringToBool(uploadFiles.Regexp, false)
	if err != nil {
		return nil, err
	}
	rootPath := cliutils.GetRootPathForUpload(uploadFiles.Pattern, isRegexpe)
	if !ioutils.IsPathExists(rootPath) {
		err := cliutils.CheckError(errors.New("Path does not exist: " + rootPath))
		if err != nil {
			return nil, err
		}
	}
	uploadFiles.Pattern = cliutils.PrepareLocalPathForUpload(uploadFiles.Pattern, isRegexpe)
	artifacts := []cliutils.Artifact{}
	// If the path is a single file then return it
	dir, err := ioutils.IsDir(rootPath)
	if err != nil {
		return nil, err
	}
	isFlat, err := cliutils.StringToBool(uploadFiles.Flat, true)
	if err != nil {
		return nil, err
	}
	if !dir {
		artifact := getSingleFileToUpload(rootPath, uploadFiles.Target, isFlat)
		return append(artifacts, artifact), nil
	}

	r, err := regexp.Compile(uploadFiles.Pattern)
	cliutils.CheckError(err)
	if err != nil {
		return nil, err
	}

	var paths []string
	isRecursive, err := cliutils.StringToBool(uploadFiles.Recursive, true)
	if err != nil {
		return nil, err
	}
	if isRecursive {
		paths, err = ioutils.ListFilesRecursive(rootPath)
	} else {
		paths, err = ioutils.ListFiles(rootPath)
	}
	if err != nil {
		return nil, err
	}

	for _, path := range paths {
		dir, err := ioutils.IsDir(path)
		if err != nil {
			return nil, err
		}
		if dir {
			continue
		}
		groups := r.FindStringSubmatch(path)
		size := len(groups)
		target := uploadFiles.Target
		if size > 0 {
			for i := 1; i < size; i++ {
				group := strings.Replace(groups[i], "\\", "/", -1)
				target = strings.Replace(target, "{" + strconv.Itoa(i) + "}", group, -1)
			}
			if strings.HasSuffix(target, "/") {
				if isFlat {
					fileName, _ := ioutils.GetFileAndDirFromPath(path)
					target += fileName
				} else {
					uploadPath := cliutils.TrimPath(path)
					target += uploadPath
				}
			}
			artifacts = append(artifacts, cliutils.Artifact{path, target})
		}
	}
	return artifacts, nil
}

// Uploads the file in the specified local path to the specified target path.
// Returns true if the file was successfully uploaded.
func uploadFile(localPath, targetPath, props string, flags *UploadFlags, minChecksumDeploySize int64, logMsgPrefix string) (utils.ArtifactBuildInfo, bool, error) {
	if props != "" {
		encodedProp, err := utils.EncodeParams(props)
		if err != nil {
			return utils.ArtifactBuildInfo{}, false, err
		}
		targetPath += ";" + encodedProp
	}
	if flags.Deb != "" {
		targetPath += getDebianMatrixParams(flags.Deb)
	}

	logger.Logger.Info(logMsgPrefix + "Uploading artifact: " + targetPath)
	file, err := os.Open(localPath)
	err = cliutils.CheckError(err)
	if err != nil {
		return utils.ArtifactBuildInfo{}, false, err
	}
	defer file.Close()
	fileInfo, err := file.Stat()
	err = cliutils.CheckError(err)
	if err != nil {
		return utils.ArtifactBuildInfo{}, false, err
	}

	var checksumDeployed bool = false
	var resp *http.Response
	var details *ioutils.FileDetails
	httpClientsDetails := utils.GetArtifactoryHttpClientDetails(flags.ArtDetails)
	if fileInfo.Size() >= minChecksumDeploySize {
		resp, details, err = tryChecksumDeploy(localPath, targetPath, flags, httpClientsDetails)
		if err != nil {
			return utils.ArtifactBuildInfo{}, false, err
		}
		checksumDeployed = !flags.DryRun && (resp.StatusCode == 201 || resp.StatusCode == 200)
	}
	if !flags.DryRun && !checksumDeployed {
		resp, err = utils.UploadFile(file, targetPath, flags.ArtDetails, details, httpClientsDetails)
		if err != nil {
			return utils.ArtifactBuildInfo{}, false, err
		}
	}
	if !flags.DryRun {
		var strChecksumDeployed string
		if checksumDeployed {
			strChecksumDeployed = " (Checksum deploy)"
		} else {
			strChecksumDeployed = ""
		}
		logger.Logger.Info(logMsgPrefix + "Artifactory response: " + resp.Status + strChecksumDeployed)
	}
	if (details == nil) {
		details, err = ioutils.GetFileDetails(localPath)
	}
	artifact := createBuildArtifactItem(targetPath, details)
	return artifact, (flags.DryRun || checksumDeployed || resp.StatusCode == 201 || resp.StatusCode == 200), nil
}

func createBuildArtifactItem(targetPath string, details *ioutils.FileDetails) utils.ArtifactBuildInfo {
	fileName, _ := ioutils.GetFileAndDirFromPath(targetPath)
	fileName = strings.Split(fileName, ";")[0]
	return utils.ArtifactBuildInfo{
		Name: fileName,
		BuildInfoCommon : &utils.BuildInfoCommon{
			Sha1: details.Sha1,
			Md5: details.Md5,
		},
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
	ArtDetails  *config.ArtifactoryDetails
	DryRun      bool
	Deb         string
	Threads     int
	BuildName   string
	BuildNumber string
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