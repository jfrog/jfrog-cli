package commands

import (
	"fmt"
	"github.com/jfrogdev/jfrog-cli-go/artifactory/utils"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/utils/ioutils"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

// Uploads the artifacts in the specified local path pattern to the specified target path.
// Returns the total number of artifacts successfully uploaded.
func Upload(localPath, targetPath string, flags *utils.Flags) (totalUploaded, totalFailed int) {
	if flags.ArtDetails.SshKeyPath != "" {
		utils.SshAuthentication(flags.ArtDetails)
	}
	if !flags.DryRun {
		utils.PingArtifactory(flags.ArtDetails)
	}
	minChecksumDeploySize := getMinChecksumDeploySize()

	// Get the list of artifacts to be uploaded to Artifactory:
	artifacts := getFilesToUpload(localPath, targetPath, flags)
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
				target := flags.ArtDetails.Url + artifacts[j].TargetPath
				if uploadFile(artifacts[j].LocalPath, target, flags,
				    minChecksumDeploySize, logMsgPrefix) {
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

	fmt.Println("Uploaded " + strconv.Itoa(totalUploaded) + " artifacts to Artifactory.")
	totalFailed = size - totalUploaded
	if totalFailed > 0 {
		fmt.Println("Failed uploading " + strconv.Itoa(totalFailed) + " artifacts to Artifactory.")
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
            uploadPath = cliutils.PrepareUploadPath(uploadPath)
        }
    }
    return cliutils.Artifact{rootPath, uploadPath}
}

func getFilesToUpload(localpath string, targetPath string, flags *utils.Flags) []cliutils.Artifact {
	if strings.Index(targetPath, "/") < 0 {
		targetPath += "/"
	}
	rootPath := cliutils.GetRootPathForUpload(localpath, flags.UseRegExp)
	if !ioutils.IsPathExists(rootPath) {
		cliutils.Exit(cliutils.ExitCodeError, "Path does not exist: "+rootPath)
	}
	localpath = cliutils.PrepareLocalPathForUpload(localpath, flags.UseRegExp)

	artifacts := []cliutils.Artifact{}
	// If the path is a single file then return it
	if !ioutils.IsDir(rootPath) {
        artifact := getSingleFileToUpload(rootPath, targetPath, flags.Flat)
        return append(artifacts, artifact)
	}

	r, err := regexp.Compile(localpath)
	cliutils.CheckError(err)

	var paths []string
	if flags.Recursive {
		paths = ioutils.ListFilesRecursive(rootPath)
	} else {
		paths = ioutils.ListFiles(rootPath)
	}

	for _, path := range paths {
		if ioutils.IsDir(path) {
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
                    uploadPath := cliutils.PrepareUploadPath(path)
                    target += uploadPath
                }
            }
			artifacts = append(artifacts, cliutils.Artifact{path, target})
		}
	}
	return artifacts
}

// Uploads the file in the specified local path to the specified target path.
// Returns true if the file was successfully uploaded.
func uploadFile(localPath string, targetPath string, flags *utils.Flags,
    minChecksumDeploySize int64, logMsgPrefix string) bool {
	if flags.Props != "" {
		targetPath += ";" + flags.Props
	}
	if flags.Deb != "" {
		targetPath += getDebianMatrixParams(flags.Deb)
	}

	fmt.Println(logMsgPrefix + "Uploading artifact: " + targetPath)
	file, err := os.Open(localPath)
	cliutils.CheckError(err)
	defer file.Close()
	fileInfo, err := file.Stat()
	cliutils.CheckError(err)

	var checksumDeployed bool = false
	var resp *http.Response
	var details *ioutils.FileDetails
	httpClientsDetails := utils.GetArtifactoryHttpClientDetails(flags.ArtDetails)
	if fileInfo.Size() >= minChecksumDeploySize {
		resp, details = tryChecksumDeploy(localPath, targetPath, flags, httpClientsDetails)
		checksumDeployed = !flags.DryRun && (resp.StatusCode == 201 || resp.StatusCode == 200)
	}
	if !flags.DryRun && !checksumDeployed {
		resp = utils.UploadFile(file, targetPath, flags.ArtDetails, details, httpClientsDetails)
	}
	if !flags.DryRun {
		var strChecksumDeployed string
		if checksumDeployed {
			strChecksumDeployed = " (Checksum deploy)"
		} else {
			strChecksumDeployed = ""
		}
		fmt.Println(logMsgPrefix + "Artifactory response: " + resp.Status + strChecksumDeployed)
	}

	return flags.DryRun || checksumDeployed || resp.StatusCode == 201 || resp.StatusCode == 200
}

func getMinChecksumDeploySize() int64 {
    minChecksumDeploySize := os.Getenv("JFROG_CLI_MIN_CHECKSUM_DEPLOY_SIZE_KB")
    if minChecksumDeploySize == "" {
        return 10240
    }
    minSize, err := strconv.ParseInt(minChecksumDeploySize, 10, 64)
    cliutils.CheckError(err)
    return minSize * 1000
}

func tryChecksumDeploy(filePath, targetPath string, flags *utils.Flags, httpClientsDetails ioutils.HttpClientDetails) (*http.Response, *ioutils.FileDetails) {
	details := ioutils.GetFileDetails(filePath)
	headers := make(map[string]string)
	headers["X-Checksum-Deploy"] = "true"
	headers["X-Checksum-Sha1"] = details.Sha1
	headers["X-Checksum-Md5"] = details.Md5
	requestClientDetails := httpClientsDetails.Clone()
	cliutils.MergeMaps(headers, requestClientDetails.Headers)
	if flags.DryRun {
		return nil, details
	}
	utils.AddAuthHeaders(headers, flags.ArtDetails)
	cliutils.MergeMaps(headers, requestClientDetails.Headers)
	resp, _ := ioutils.SendPut(targetPath, nil, *requestClientDetails)
	return resp, details
}

func getDebianMatrixParams(debianPropsStr string) string {
	debProps := strings.Split(debianPropsStr, "/")
	return ";deb.distribution=" + debProps[0] +
        ";deb.component=" + debProps[1] +
        ";deb.architecture=" + debProps[2]
}