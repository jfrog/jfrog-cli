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
	"github.com/jfrogdev/jfrog-cli-go/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/logger"
)

// Uploads the artifacts in the specified local path pattern to the specified target path.
// Returns the total number of artifacts successfully uploaded.
func Upload(localPath, targetPath string, flags *UploadFlags) (totalUploaded, totalFailed int, err error) {
	utils.PreCommandSetup(flags)
	minChecksumDeploySize, err := getMinChecksumDeploySize()
    if err != nil {
        return
    }

	// Get the list of artifacts to be uploaded to Artifactory:
	var artifacts []cliutils.Artifact
	artifacts, err = getFilesToUpload(localPath, targetPath, flags)
	if err != nil {
	    return
	}
	size := len(artifacts)

	var wg sync.WaitGroup

	// Create an array of integers, to store the total file that were uploaded successfully.
	// Each array item is used by a single thread.
	uploadCount := make([]int, flags.Threads, flags.Threads)

	for i := 0; i < flags.Threads; i++ {
		wg.Add(1)
		go func(threadId int) {
			logMsgPrefix := cliutils.GetLogMsgPrefix(threadId, flags.DryRun)
			for j := threadId; j < size; j += flags.Threads {
                if err != nil {
                    break;
                }
                var e error
                var uploaded bool
                var target string
				target, e = utils.BuildArtifactoryUrl(flags.ArtDetails.Url, artifacts[j].TargetPath, make(map[string]string))
                if e != nil {
                    err = e
                    break
                }
				uploaded, e = uploadFile(artifacts[j].LocalPath, target, flags,
                    minChecksumDeploySize, logMsgPrefix)
                if e != nil {
                    err = e
                    break
                }
				if uploaded {
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

	logger.Logger.Info("Uploaded " + strconv.Itoa(totalUploaded) + " artifacts to Artifactory.")
	totalFailed = size - totalUploaded
	if totalFailed > 0 {
		logger.Logger.Info("Failed uploading " + strconv.Itoa(totalFailed) + " artifacts to Artifactory.")
	}
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
    return cliutils.Artifact{rootPath, uploadPath}
}

func getFilesToUpload(localpath string, targetPath string, flags *UploadFlags) ([]cliutils.Artifact, error) {
	if strings.Index(targetPath, "/") < 0 {
		targetPath += "/"
	}
	rootPath := cliutils.GetRootPathForUpload(localpath, flags.UseRegExp)
	if !ioutils.IsPathExists(rootPath) {
		err := cliutils.CheckError(errors.New("Path does not exist: " + rootPath))
        if err != nil {
            return nil, err
        }
	}
	localpath = cliutils.PrepareLocalPathForUpload(localpath, flags.UseRegExp)

	artifacts := []cliutils.Artifact{}
	// If the path is a single file then return it
	dir, err := ioutils.IsDir(rootPath)
	if err != nil {
	    return nil, err
	}
	if !dir {
        artifact := getSingleFileToUpload(rootPath, targetPath, flags.Flat)
        return append(artifacts, artifact), nil
	}

	r, err := regexp.Compile(localpath)
	err = cliutils.CheckError(err)
	if err != nil {
	    return nil, err
	}

	var paths []string
	if flags.Recursive {
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
		target := targetPath
		if size > 0 {
			for i := 1; i < size; i++ {
				group := strings.Replace(groups[i], "\\", "/", -1)
				target = strings.Replace(target, "{"+strconv.Itoa(i)+"}", group, -1)
			}
			if strings.HasSuffix(target, "/") {
                if flags.Flat {
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
func uploadFile(localPath string, targetPath string, flags *UploadFlags,
    minChecksumDeploySize int64, logMsgPrefix string) (bool, error) {
	if flags.Props != "" {
		targetPath += ";" + flags.Props
	}
	if flags.Deb != "" {
		targetPath += getDebianMatrixParams(flags.Deb)
	}

	logger.Logger.Info(logMsgPrefix + "Uploading artifact: " + targetPath)
	file, err := os.Open(localPath)
	err = cliutils.CheckError(err)
	if err != nil {
	    return false, err
	}
	defer file.Close()
	fileInfo, err := file.Stat()
	err = cliutils.CheckError(err)
	if err != nil {
	    return false, err
	}

	var checksumDeployed bool = false
	var resp *http.Response
	var details *ioutils.FileDetails
	httpClientsDetails := utils.GetArtifactoryHttpClientDetails(flags.ArtDetails)
	if fileInfo.Size() >= minChecksumDeploySize {
		resp, details, err = tryChecksumDeploy(localPath, targetPath, flags, httpClientsDetails)
        if err != nil {
            return false, err
        }
		checksumDeployed = !flags.DryRun && (resp.StatusCode == 201 || resp.StatusCode == 200)
	}
	if !flags.DryRun && !checksumDeployed {
		resp, err = utils.UploadFile(file, targetPath, flags.ArtDetails, details, httpClientsDetails)
        if err != nil {
            return false, err
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

	return (flags.DryRun || checksumDeployed || resp.StatusCode == 201 || resp.StatusCode == 200), nil
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
	ArtDetails   *config.ArtifactoryDetails
	DryRun       bool
	Props        string
	Deb          string
	Recursive    bool
	Flat         bool
	UseRegExp    bool
	Threads      int
}

func (flags *UploadFlags) GetArtifactoryDetails() *config.ArtifactoryDetails {
	return flags.ArtDetails
}

func (flags *UploadFlags) IsRecursive() bool {
	return flags.Recursive
}

func (flags *UploadFlags) GetProps() string {
	return flags.Props
}

func (flags *UploadFlags) IsDryRun() bool {
	return flags.DryRun
}