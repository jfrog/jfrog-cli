package commands

import (
	"fmt"
	"github.com/JFrogDev/jfrog-cli-go/bintray/utils"
	"github.com/JFrogDev/jfrog-cli-go/cliutils"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

func Upload(versionDetails *utils.VersionDetails, localPath, uploadPath string,
	uploadFlags *UploadFlags) (totalUploaded, totalFailed int) {

	if uploadFlags.BintrayDetails.User == "" {
		uploadFlags.BintrayDetails.User = versionDetails.Subject
	}
	if !uploadFlags.DryRun {
        verifyPackageExists(versionDetails, uploadFlags)
        createVersionIfNeeded(versionDetails, uploadFlags)
	}
	// Get the list of artifacts to be uploaded to:
	artifacts := getFilesToUpload(localPath, uploadPath, versionDetails.Package, uploadFlags)

	baseUrl := uploadFlags.BintrayDetails.ApiUrl + "content/" + versionDetails.Subject + "/" +
		versionDetails.Repo + "/" + versionDetails.Package + "/" + versionDetails.Version + "/"

	totalUploaded, totalFailed = uploadFiles(artifacts, baseUrl, uploadFlags)
	return
}

func uploadFiles(artifacts []cliutils.Artifact, baseUrl string, flags *UploadFlags) (totalUploaded,
    totalFailed int) {

	size := len(artifacts)
	var wg sync.WaitGroup

	// Create an array of integers, to store the total file that were uploaded successfully.
	// Each array item is used by a single thread.
	uploadCount := make([]int, flags.Threads, flags.Threads)
    matrixParams := getMatrixParams(flags)
	for i := 0; i < flags.Threads; i++ {
		wg.Add(1)
		go func(threadId int) {
			logMsgPrefix := cliutils.GetLogMsgPrefix(threadId, flags.DryRun)
			for j := threadId; j < size; j += flags.Threads {
				url := baseUrl + artifacts[j].TargetPath + matrixParams
				if !flags.DryRun {
					if uploadFile(artifacts[j], url, logMsgPrefix, flags.BintrayDetails) {
						uploadCount[threadId]++
					}
				} else {
					fmt.Println("[Dry Run] Uploading artifact: " + url)
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
	fmt.Println("Uploaded " + strconv.Itoa(totalUploaded) + " artifacts to Bintray.")
	totalFailed = size - totalUploaded
	if totalFailed > 0 {
		fmt.Println("Failed uploading " + strconv.Itoa(totalFailed) + " artifacts to Bintray.")
	}
	return
}

func getMatrixParams(flags *UploadFlags) string {
    params := ""
    if flags.Publish {
        params += ";publish=1"
    }
    if flags.Override {
        params += ";override=1"
    }
    if flags.Explode {
        params += ";explode=1"
    }
    if flags.Deb != "" {
        params += getDebianMatrixParams(flags.Deb)
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
    return "pool/" + component + "/" + packageName[0:1] + "/" + packageName + "/"
}

func uploadFile(artifact cliutils.Artifact, url, logMsgPrefix string, bintrayDetails *cliutils.BintrayDetails) bool {
	fmt.Println(logMsgPrefix + " Uploading artifact to: " + url)

	f, err := os.Open(artifact.LocalPath)
	cliutils.CheckError(err)
	defer f.Close()

	resp := cliutils.UploadFile(f, url,
		bintrayDetails.User, bintrayDetails.Key, nil)

	fmt.Println(logMsgPrefix + " Bintray response: " + resp.Status)
	return resp.StatusCode == 201 || resp.StatusCode == 200
}

func verifyPackageExists(versionDetails *utils.VersionDetails, uploadFlags *UploadFlags) {
	fmt.Println("Verifying package " + versionDetails.Package + " exists...")
	resp := utils.HeadPackage(versionDetails, uploadFlags.BintrayDetails)
	if resp.StatusCode == 404 {
        promptPackageNotExist(versionDetails)
	} else if resp.StatusCode != 200 {
		cliutils.Exit(cliutils.ExitCodeError, "Bintray response: "+resp.Status)
	}
}

func promptPackageNotExist(versionDetails *utils.VersionDetails) {
    msg := "It looks like package '" + versionDetails.Package +
       "' does not exist in the '" + versionDetails.Repo + "' repository.\n" +
       "You can create the package by running the package-create command. For example:\n" +
       "jfrog bt pc " +
       versionDetails.Subject + "/" + versionDetails.Repo + "/" + versionDetails.Package +
       " --vcs-url=https://github.com/example"
    if cliutils.ReadBintrayConf().DefPackageLicenses == "" {
        msg += " --licenses=Apache-2.0-example"
    }
    cliutils.Exit(cliutils.ExitCodeError, msg)
}

func createVersionIfNeeded(versionDetails *utils.VersionDetails, uploadFlags *UploadFlags) {
	fmt.Println("Checking if version " + versionDetails.Version + " exists...")
	resp := utils.HeadVersion(versionDetails, uploadFlags.BintrayDetails)
	if resp.StatusCode == 404 {
		fmt.Println("Creating version " + versionDetails.Version + "...")
		resp, body := DoCreateVersion(versionDetails, uploadFlags.BintrayDetails)
		if resp.StatusCode != 201 {
			fmt.Println("Bintray response: " + resp.Status)
			cliutils.Exit(cliutils.ExitCodeError, cliutils.IndentJson(body))
		}
		fmt.Println("Bintray response: " + resp.Status)
	} else if resp.StatusCode != 200 {
		cliutils.Exit(cliutils.ExitCodeError, "Bintray response: "+resp.Status)
	}
}

func getFilesToUpload(localpath, targetPath, packageName string, flags *UploadFlags) []cliutils.Artifact {
    var debianDefaultPath string
    if targetPath == "" && flags.Deb != "" {
        debianDefaultPath = getDebianDefaultPath(flags.Deb, packageName)
    }

	rootPath := cliutils.GetRootPathForUpload(localpath, flags.UseRegExp)
	if !cliutils.IsPathExists(rootPath) {
		cliutils.Exit(cliutils.ExitCodeError, "Path does not exist: "+rootPath)
	}
	localpath = cliutils.PrepareLocalPathForUpload(localpath, flags.UseRegExp)

	artifacts := []cliutils.Artifact{}
	// If the path is a single file then return it
	if !cliutils.IsDir(rootPath) {
	    targetPath := cliutils.PrepareUploadPath(targetPath)
	    if targetPath == "" || strings.HasSuffix(targetPath, "/") {
            targetPath += debianDefaultPath + rootPath
	    }

		artifacts = append(artifacts, cliutils.Artifact{rootPath, targetPath})
		return artifacts
	}

	r, err := regexp.Compile(localpath)
	cliutils.CheckError(err)

	var paths []string
	if flags.Recursive {
		paths = cliutils.ListFilesRecursive(rootPath)
	} else {
		paths = cliutils.ListFiles(rootPath)
	}

	for _, path := range paths {
		if cliutils.IsDir(path) {
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

			if target == "" || strings.HasSuffix(target, "/") {
				if target == "" {
				    target = debianDefaultPath
				}
				if flags.Flat {
					fileName, _ := cliutils.GetFileAndDirFromPath(path)
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

type UploadFlags struct {
	BintrayDetails *cliutils.BintrayDetails
	Deb         string
	DryRun      bool
	Recursive   bool
	Flat        bool
	Publish     bool
	Override    bool
	Explode     bool
	Threads     int
	UseRegExp   bool
}
