package commands

import (
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/bintray/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/cliutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/config"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/fileutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/httputils"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"errors"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
	"time"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
	clientutils "github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils"
)

func Upload(versionDetails *utils.VersionDetails, localPath, uploadPath string,
	uploadFlags *UploadFlags) (totalUploaded, totalFailed int, err error) {

	if uploadFlags.BintrayDetails.User == "" {
		uploadFlags.BintrayDetails.User = versionDetails.Subject
	}
	if !uploadFlags.DryRun {
		verifyRepoExists(versionDetails, uploadFlags)
		err = verifyPackageExists(versionDetails, uploadFlags)
		if err != nil {
			return
		}
		createVersionIfNeeded(versionDetails, uploadFlags)
	}
	// Get the list of artifacts to be uploaded to:
	var artifacts []clientutils.Artifact
	artifacts, err = getFilesToUpload(localPath, uploadPath, versionDetails.Package, uploadFlags)
	if err != nil {
		return
	}

	baseUrl := uploadFlags.BintrayDetails.ApiUrl + "content/" + versionDetails.Subject + "/" +
		versionDetails.Repo + "/" + versionDetails.Package + "/" + versionDetails.Version + "/"

	totalUploaded, totalFailed, err = uploadFiles(artifacts, baseUrl, uploadFlags)
	return
}

func uploadFiles(artifacts []clientutils.Artifact, baseUrl string, flags *UploadFlags) (totalUploaded,
totalFailed int, err error) {

	size := len(artifacts)
	var wg sync.WaitGroup

	// Create an array of integers, to store the total file that were uploaded successfully.
	// Each array item is used by a single thread.
	uploadCount := make([]int, flags.Threads, flags.Threads)
	matrixParams := getMatrixParams(flags)
	for i := 0; i < flags.Threads; i++ {
		wg.Add(1)
		go func(threadId int) {
			logMsgPrefix := clientutils.GetLogMsgPrefix(threadId, flags.DryRun)
			for j := threadId; j < size; j += flags.Threads {
				if err != nil {
					break
				}
				url := baseUrl + artifacts[j].TargetPath + matrixParams
				if !flags.DryRun {
					uploaded, e := uploadFile(artifacts[j], url, logMsgPrefix, flags.BintrayDetails)
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

func uploadFile(artifact clientutils.Artifact, url, logMsgPrefix string, bintrayDetails *config.BintrayDetails) (bool, error) {
	log.Info(logMsgPrefix+"Uploading artifact:", artifact.LocalPath)

	f, err := os.Open(artifact.LocalPath)
	err = errorutils.CheckError(err)
	if err != nil {
		return false, err
	}
	defer f.Close()
	httpClientsDetails := utils.GetBintrayHttpClientDetails(bintrayDetails)
	resp, body, err := httputils.UploadFile(f, url, httpClientsDetails)
	if err != nil {
		return false, err
	}
	log.Debug(logMsgPrefix+"Bintray response:", resp.Status)
	if resp.StatusCode != 201 && resp.StatusCode != 200 {
		log.Error(logMsgPrefix + "Bintray response: " + resp.Status + "\n" + clientutils.IndentJson(body))
	}

	return resp.StatusCode == 201 || resp.StatusCode == 200, nil
}

func verifyPackageExists(versionDetails *utils.VersionDetails, uploadFlags *UploadFlags) error {
	log.Info("Verifying package", versionDetails.Package, "exists...")
	resp, _, err := utils.HeadPackage(versionDetails, uploadFlags.BintrayDetails)
	if err != nil {
		return err
	}
	if resp.StatusCode == 404 {
		err = promptPackageNotExist(versionDetails)
		if err != nil {
			return err
		}
	} else if resp.StatusCode != 200 {
		err = errorutils.CheckError(errors.New("Bintray response: " + resp.Status))
	}
	return err
}

func verifyRepoExists(versionDetails *utils.VersionDetails, uploadFlags *UploadFlags) error {
	log.Info("Verifying repository", versionDetails.Repo, "exists...")
	resp, _, err := utils.HeadRepo(versionDetails, uploadFlags.BintrayDetails)
	if err != nil {
		return err
	}
	if resp.StatusCode == 404 {
		err = promptRepoNotExist(versionDetails)
	} else if resp.StatusCode != 200 {
		err = errorutils.CheckError(errors.New("Bintray response: " + resp.Status))
	}
	return err
}

func promptPackageNotExist(versionDetails *utils.VersionDetails) error {
	msg := "It looks like package '" + versionDetails.Package +
		"' does not exist in the '" + versionDetails.Repo + "' repository.\n" +
		"You can create the package by running the package-create command. For example:\n" +
		"jfrog bt pc " +
		versionDetails.Subject + "/" + versionDetails.Repo + "/" + versionDetails.Package +
		" --vcs-url=https://github.com/example"

	conf, err := config.ReadBintrayConf()
	if err != nil {
		return err
	}
	if conf.DefPackageLicenses == "" {
		msg += " --licenses=Apache-2.0-example"
	}
	err = errorutils.CheckError(errors.New(msg))
	return err
}

func promptRepoNotExist(versionDetails *utils.VersionDetails) error {
	msg := "It looks like repository '" + versionDetails.Repo + "' does not exist.\n"
	return errorutils.CheckError(errors.New(msg))
}

func createVersionIfNeeded(versionDetails *utils.VersionDetails, uploadFlags *UploadFlags) error {
	log.Info("Verifying version", versionDetails.Version, "exists...")
	resp, _, err := utils.HeadVersion(versionDetails, uploadFlags.BintrayDetails)
	if err != nil {
		return err
	}
	if resp.StatusCode == 404 {
		log.Info("Creating version...")
		resp, body, err := DoCreateVersion(versionDetails, uploadFlags.BintrayDetails)
		if err != nil {
			return err
		}
		if resp.StatusCode != 201 {
			return errorutils.CheckError(errors.New("Bintray response: " + resp.Status + "\n" + clientutils.IndentJson(body)))
		}
		log.Debug("Bintray response:", resp.Status)
		log.Info("Created version", versionDetails.Version+".")
	} else if resp.StatusCode != 200 {
		err = errorutils.CheckError(errors.New("Bintray response: " + resp.Status))
	}
	return err
}

func getSingleFileToUpload(rootPath, targetPath, debianDefaultPath string, flat bool) clientutils.Artifact {
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

func getFilesToUpload(localPath, targetPath, packageName string, flags *UploadFlags) ([]clientutils.Artifact, error) {
	var debianDefaultPath string
	if targetPath == "" && flags.Deb != "" {
		debianDefaultPath = getDebianDefaultPath(flags.Deb, packageName)
	}

	rootPath := cliutils.GetRootPathForUpload(localPath, flags.UseRegExp)
	if !fileutils.IsPathExists(rootPath) {
		err := errorutils.CheckError(errors.New("Path does not exist: " + rootPath))
		if err != nil {
			return nil, err
		}
	}
	localPath = clientutils.ReplaceTildeWithUserHome(localPath)
	localPath = cliutils.PrepareLocalPathForUpload(localPath, flags.UseRegExp)

	artifacts := []clientutils.Artifact{}
	// If the path is a single file then return it
	dir, err := fileutils.IsDir(rootPath)
	if err != nil {
		return nil, err
	}

	if !dir {
		artifact := getSingleFileToUpload(rootPath, targetPath, debianDefaultPath, flags.Flat)
		return append(artifacts, artifact), nil
	}

	r, err := regexp.Compile(localPath)
	err = errorutils.CheckError(err)
	if err != nil {
		return nil, err
	}

	spinner := cliutils.NewSpinner("[Info] Collecting files for upload:", time.Second)
	spinner.Start()
	paths, err := listFiles(flags, err, rootPath)
	if err != nil {
		return nil, err
	}

	for _, path := range paths {
		dir, err := fileutils.IsDir(path)
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

			if target == "" || strings.HasSuffix(target, "/") {
				if target == "" {
					target = debianDefaultPath
				}
				if flags.Flat {
					fileName, _ := fileutils.GetFileAndDirFromPath(path)
					target += fileName
				} else {
					uploadPath := clientutils.TrimPath(path)
					target += uploadPath
				}
			}

			artifacts = append(artifacts, clientutils.Artifact{path, target, ""})
		}
	}
	spinner.Stop()
	return artifacts, nil
}
func listFiles(flags *UploadFlags, err error, rootPath string) ([]string, error) {
	var paths []string
	if flags.Recursive {
		paths, err = fileutils.ListFilesRecursiveWalkIntoDirSymlink(rootPath, false)
	} else {
		paths, err = fileutils.ListFiles(rootPath, false)
	}
	return paths, err
}

type UploadFlags struct {
	BintrayDetails *config.BintrayDetails
	Deb            string
	DryRun         bool
	Recursive      bool
	Flat           bool
	Publish        bool
	Override       bool
	Explode        bool
	Threads        int
	UseRegExp      bool
}
