package commands

import (
	"os"
	"fmt"
	"sync"
	"regexp"
	"strings"
	"strconv"
	"github.com/jFrogdev/jfrog-cli-go/cliutils"
    "github.com/jFrogdev/jfrog-cli-go/bintray/utils"
)

func Upload(versionDetails *utils.VersionDetails, localPath, uploadPath string,
    uploadFlags *UploadFlags, packageFlags *utils.PackageFlags, versionFlags *utils.VersionFlags) {

    if uploadFlags.BintrayDetails.User == "" {
        uploadFlags.BintrayDetails.User = versionDetails.Subject
    }
    createPackageIfNeeded(versionDetails, uploadFlags, packageFlags)
    createVersionIfNeeded(versionDetails, uploadFlags, versionFlags)

    // Get the list of artifacts to be uploaded to:
    artifacts := getFilesToUpload(localPath, uploadPath, uploadFlags)

    baseUrl := uploadFlags.BintrayDetails.ApiUrl + "content/" + versionDetails.Subject + "/" +
           versionDetails.Repo + "/" + versionDetails.Package + "/" + versionDetails.Version + "/";

    uploadFiles(artifacts, baseUrl, uploadFlags)
}

func uploadFiles(artifacts []Artifact, baseUrl string, flags *UploadFlags) (totalUploaded, totalFailed int) {
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
                url := baseUrl + artifacts[j].TargetPath
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
    totalFailed = size-totalUploaded
    if totalFailed > 0 {
        fmt.Println("Failed uploading " + strconv.Itoa(totalFailed) + " artifacts to Bintray.")
    }
    return
}

func uploadFile(artifact Artifact, url, logMsgPrefix string, bintrayDetails *utils.BintrayDetails) bool {
    fmt.Println(logMsgPrefix + " Uploading artifact to: " + url)

    f, err := os.Open(artifact.LocalPath)
    cliutils.CheckError(err)
    defer f.Close()

    resp := cliutils.UploadFile(f, url,
        bintrayDetails.User, bintrayDetails.Key, nil)

    fmt.Println(logMsgPrefix + " Bintray response: " + resp.Status)
    return resp.StatusCode == 201 || resp.StatusCode == 200
}

func createPackageIfNeeded(versionDetails *utils.VersionDetails, uploadFlags *UploadFlags,
    packageFlags *utils.PackageFlags) {

    fmt.Println("Checking if package " + versionDetails.Package + " exists...")
    resp := utils.HeadPackage(versionDetails, uploadFlags.BintrayDetails)
    if resp.StatusCode == 404 {
        fmt.Println("Creating package " + versionDetails.Package + "...")
        resp, body := DoCreatePackage(versionDetails, packageFlags)
        if resp.StatusCode != 201 {
            fmt.Println("Bintray response: " + resp.Status)
            cliutils.Exit(cliutils.ExitCodeError, cliutils.IndentJson(body))
        }
        fmt.Println("Bintray response: " + resp.Status)
    } else
    if resp.StatusCode != 200 {
        cliutils.Exit(cliutils.ExitCodeError, "Bintray response: " + resp.Status)
    }
}

func createVersionIfNeeded(versionDetails *utils.VersionDetails, uploadFlags *UploadFlags,
    versionFlags *utils.VersionFlags) {

    fmt.Println("Checking if version " + versionDetails.Version + " exists...")
    resp := utils.HeadVersion(versionDetails, uploadFlags.BintrayDetails)
    if resp.StatusCode == 404 {
        fmt.Println("Creating version " + versionDetails.Version + "...")
        resp, body := DoCreateVersion(versionDetails, versionFlags)
        if resp.StatusCode != 201 {
            fmt.Println("Bintray response: " + resp.Status)
            cliutils.Exit(cliutils.ExitCodeError, cliutils.IndentJson(body))
        }
        fmt.Println("Bintray response: " + resp.Status)
    } else
    if resp.StatusCode != 200 {
        cliutils.Exit(cliutils.ExitCodeError, "Bintray response: " + resp.Status)
    }
}

func getFilesToUpload(localpath string, targetPath string, flags *UploadFlags) []Artifact {
    rootPath := getRootPath(localpath, flags.UseRegExp)
    if !cliutils.IsPathExists(rootPath) {
        cliutils.Exit(cliutils.ExitCodeError, "Path does not exist: " + rootPath)
    }
    localpath = prepareLocalPath(localpath, flags.UseRegExp)

    artifacts := []Artifact{}
    // If the path is a single file then return it
    if !cliutils.IsDir(rootPath) {
        targetPath := prepareUploadPath(targetPath + rootPath)
        artifacts = append(artifacts, Artifact{rootPath, targetPath})
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
        if (size > 0) {
            for i := 1; i < size; i++ {
                group := strings.Replace(groups[i], "\\", "/", -1)
                target = strings.Replace(target, "{" + strconv.Itoa(i) + "}", group, -1)
            }

            if target == "" || strings.HasSuffix(target, "/") {
                if flags.Flat {
                    target += cliutils.GetFileNameFromPath(path)
                } else {
                    uploadPath := prepareUploadPath(path)
                    target += uploadPath
                }
            }

            artifacts = append(artifacts, Artifact{path, target})
        }
    }
    return artifacts
}

// Get the local root path, from which to start collecting artifacts to be uploaded.
func getRootPath(path string, useRegExp bool) string {
    // The first step is to split the local path pattern into sections, by the file seperator.
    seperator := "/"
    sections := strings.Split(path, seperator)
    if len(sections) == 1 {
        seperator = "\\"
        sections = strings.Split(path, seperator)
    }

    // Now we start building the root path, making sure to leave out the sub-directory that includes the pattern.
    rootPath := ""
    for _, section := range sections {
        if section == "" {
            continue
        }
        if useRegExp {
            if strings.Index(section, "(") != -1 {
                break
            }
        } else {
            if strings.Index(section, "*") != -1 {
                break
            }
        }
        if rootPath != "" {
            rootPath += seperator
        }
        rootPath += section
    }
    if len(sections) > 0 && sections[0] == "" {
        rootPath = seperator + rootPath
    }
    if rootPath == "" {
        return "."
    }
    return rootPath
}

func prepareUploadPath(path string) string {
    path = strings.Replace(path, "\\", "/", -1)
    path = strings.Replace(path, "../", "", -1)
    path = strings.Replace(path, "./", "", -1)
    return path
}

func prepareLocalPath(localpath string, useRegExp bool) string {
    if localpath == "./" || localpath == ".\\" {
        return "^.*$"
    }
    if strings.HasPrefix(localpath, "./") {
        localpath = localpath[2:]
    } else
    if strings.HasPrefix(localpath, ".\\") {
        localpath = localpath[3:]
    }
    if !useRegExp {
        localpath = localPathToRegExp(localpath)
    }
    return localpath
}

func localPathToRegExp(localpath string) string {
    var wildcard = ".*"

    localpath = strings.Replace(localpath, ".", "\\.", -1)
    localpath = strings.Replace(localpath, "*", wildcard, -1)
    if strings.HasSuffix(localpath, "/") {
        localpath += wildcard
    } else
    if strings.HasSuffix(localpath, "\\") {
       size := len(localpath)
       if size > 1 && localpath[size-2 : size-1] != "\\" {
            localpath += "\\"
       }
       localpath += wildcard
    }
    localpath = "^" + localpath + "$"
    return localpath
}

type Artifact struct {
    LocalPath string
    TargetPath string
}

type UploadFlags struct {
    BintrayDetails *utils.BintrayDetails
    DryRun bool
    Recursive bool
    Flat bool
    Threads int
    UseRegExp bool
}