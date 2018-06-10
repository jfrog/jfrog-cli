package services

import (
	"errors"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/bintray/auth"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/bintray/services/versions"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/httpclient"
	clientutils "github.com/jfrog/jfrog-cli-go/jfrog-client/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/errorutils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/io/fileutils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/io/httputils"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/log"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

func NewUploadService(client *httpclient.HttpClient) *UploadService {
	us := &UploadService{client: client}
	return us
}

func NewUploadParams() *UploadParams {
	return &UploadParams{Params: &versions.Params{}}
}

type UploadService struct {
	client         *httpclient.HttpClient
	BintrayDetails auth.BintrayDetails
	DryRun         bool
	Threads        int
}

type UploadParams struct {
	// Files pattern to be uploaded
	Pattern string

	// Target version details
	*versions.Params

	// Target local path
	TargetPath string

	UseRegExp bool
	Flat      bool
	Recursive bool
	Explode   bool
	Override  bool
	Publish   bool
	Deb       string
}

func (us *UploadService) Upload(uploadDetails *UploadParams) (totalUploaded, totalFailed int, err error) {
	if us.BintrayDetails.GetUser() == "" {
		us.BintrayDetails.SetUser(uploadDetails.Subject)
	}

	// Get the list of artifacts to be uploaded to:
	var artifacts []clientutils.Artifact
	artifacts, err = us.getFilesToUpload(uploadDetails)
	if err != nil {
		return
	}

	baseUrl := us.BintrayDetails.GetApiUrl() + path.Join("content", uploadDetails.Subject, uploadDetails.Repo, uploadDetails.Package, uploadDetails.Version)
	totalUploaded, totalFailed, err = us.uploadFiles(uploadDetails, artifacts, baseUrl)
	return
}

func (us *UploadService) uploadFiles(uploadDetails *UploadParams, artifacts []clientutils.Artifact, baseUrl string) (totalUploaded, totalFailed int, err error) {
	size := len(artifacts)
	var wg sync.WaitGroup

	// Create an array of integers, to store the total file that were uploaded successfully.
	// Each array item is used by a single thread.
	uploadCount := make([]int, us.Threads, us.Threads)
	matrixParams := getMatrixParams(uploadDetails)
	for i := 0; i < us.Threads; i++ {
		wg.Add(1)
		go func(threadId int) {
			logMsgPrefix := clientutils.GetLogMsgPrefix(threadId, us.DryRun)
			for j := threadId; j < size; j += us.Threads {
				if err != nil {
					break
				}
				url := baseUrl + "/" + artifacts[j].TargetPath + matrixParams
				if !us.DryRun {
					uploaded, e := uploadFile(artifacts[j], url, logMsgPrefix, us.BintrayDetails)
					if e != nil {
						err = e
						break
					}
					if uploaded {
						uploadCount[threadId]++
					}
				} else {
					log.Info("[Dry Run] Uploading artifact:", artifacts[j].LocalPath)
					uploadCount[threadId]++
				}
			}
			wg.Done()
		}(i)
	}
	wg.Wait()

	totalUploaded = 0
	for _, i := range uploadCount {
		totalUploaded += i
	}
	log.Debug("Uploaded", strconv.Itoa(totalUploaded), "artifacts.")
	totalFailed = size - totalUploaded
	if totalFailed > 0 {
		log.Error("Failed uploading", strconv.Itoa(totalFailed), "artifacts.")
	}
	return
}

func getMatrixParams(uploadDetails *UploadParams) string {
	params := ""
	if uploadDetails.Publish {
		params += ";publish=1"
	}
	if uploadDetails.Override {
		params += ";override=1"
	}
	if uploadDetails.Explode {
		params += ";explode=1"
	}
	if uploadDetails.Deb != "" {
		params += getDebianMatrixParams(uploadDetails.Deb)
	}
	return params
}

func getDebianMatrixParams(debianPropsStr string) string {
	debProps := strings.Split(debianPropsStr, "/")
	return ";deb_distribution=" + debProps[0] +
		";deb_component=" + debProps[1] +
		";deb_architecture=" + debProps[2]
}

func getDebianDefaultPath(debianPropsStr, packageName string) string {
	debProps := strings.Split(debianPropsStr, "/")
	component := strings.Split(debProps[1], ",")[0]
	return path.Join("pool", component, packageName[0:1], packageName) + "/"
}

func uploadFile(artifact clientutils.Artifact, url, logMsgPrefix string, bintrayDetails auth.BintrayDetails) (bool, error) {
	log.Info(logMsgPrefix+"Uploading artifact:", artifact.LocalPath)

	f, err := os.Open(artifact.LocalPath)
	err = errorutils.CheckError(err)
	if err != nil {
		return false, err
	}
	defer f.Close()
	httpClientsDetails := bintrayDetails.CreateHttpClientDetails()
	resp, body, err := httputils.UploadFile(f, url, httpClientsDetails)
	if err != nil {
		return false, err
	}
	log.Debug(logMsgPrefix+"Bintray response:", resp.Status)
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		log.Error(logMsgPrefix + "Bintray response: " + resp.Status + "\n" + clientutils.IndentJson(body))
	}

	return resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK, nil
}

func getSingleFileToUpload(rootPath, targetPath string, flat bool) clientutils.Artifact {
	var uploadPath string
	rootPathOrig := rootPath
	if targetPath != "" && !strings.HasSuffix(targetPath, "/") {
		rootPath = targetPath
		targetPath = ""
	}
	if flat {
		uploadPath, _ = fileutils.GetFileAndDirFromPath(rootPath)
		uploadPath = targetPath + uploadPath
	} else {
		uploadPath = targetPath + rootPath
		uploadPath = clientutils.TrimPath(uploadPath)
	}
	return clientutils.Artifact{LocalPath: rootPathOrig, TargetPath: uploadPath}
}

func (us *UploadService) getFilesToUpload(uploadDetails *UploadParams) ([]clientutils.Artifact, error) {
	var debianDefaultPath string
	if uploadDetails.TargetPath == "" && uploadDetails.Deb != "" {
		debianDefaultPath = getDebianDefaultPath(uploadDetails.Deb, uploadDetails.Package)
	}

	rootPath := clientutils.GetRootPath(uploadDetails.Pattern, uploadDetails.UseRegExp)
	if !fileutils.IsPathExists(rootPath) {
		err := errorutils.CheckError(errors.New("Path does not exist: " + rootPath))
		if err != nil {
			return nil, err
		}
	}
	localPath := clientutils.ReplaceTildeWithUserHome(uploadDetails.Pattern)
	localPath = clientutils.PrepareLocalPathForUpload(localPath, uploadDetails.UseRegExp)

	artifacts := []clientutils.Artifact{}
	// If the path is a single file then return it
	dir, err := fileutils.IsDir(rootPath)
	if err != nil {
		return nil, err
	}

	if !dir {
		artifact := getSingleFileToUpload(rootPath, uploadDetails.TargetPath, uploadDetails.Flat)
		return append(artifacts, artifact), nil
	}

	r, err := regexp.Compile(localPath)
	err = errorutils.CheckError(err)
	if err != nil {
		return nil, err
	}

	log.Info("Collecting files for upload...")
	paths, err := us.listFiles(uploadDetails.Recursive, rootPath)
	if err != nil {
		return nil, err
	}

	for _, filePath := range paths {
		dir, err := fileutils.IsDir(filePath)
		if err != nil {
			return nil, err
		}
		if dir {
			continue
		}

		groups := r.FindStringSubmatch(filePath)
		size := len(groups)
		target := uploadDetails.TargetPath

		if size > 0 {
			for i := 1; i < size; i++ {
				group := strings.Replace(groups[i], "\\", "/", -1)
				target = strings.Replace(target, "{"+strconv.Itoa(i)+"}", group, -1)
			}

			if target == "" || strings.HasSuffix(target, "/") {
				if target == "" {
					target = debianDefaultPath
				}
				if uploadDetails.Flat {
					fileName, _ := fileutils.GetFileAndDirFromPath(filePath)
					target += fileName
				} else {
					uploadPath := clientutils.TrimPath(filePath)
					target += uploadPath
				}
			}

			artifacts = append(artifacts, clientutils.Artifact{LocalPath: filePath, TargetPath: target, Symlink: ""})
		}
	}
	return artifacts, nil
}

func (us *UploadService) listFiles(recursive bool, rootPath string) ([]string, error) {
	if recursive {
		return fileutils.ListFilesRecursiveWalkIntoDirSymlink(rootPath, false)
	}
	return fileutils.ListFiles(rootPath, false)
}
