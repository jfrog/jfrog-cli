package commands

import (
  "os"
  "fmt"
  "sync"
  "strings"
  "regexp"
  "strconv"
  "net/http"
  "github.com/jfrogdev/jfrog-cli-go/cliutils"
  "github.com/jfrogdev/jfrog-cli-go/artifactory/utils"
)

// Uploads the artifacts in the specified local path pattern to the specified target path.
// Returns the total number of artifacts successfully uploaded.
func Upload(localPath, targetPath string, flags *utils.Flags) (totalUploaded, totalFailed int) {
    if flags.ArtDetails.SshKeyPath != "" {
        utils.SshAuthentication(flags.ArtDetails)
    }

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
                if uploadFile(artifacts[j].LocalPath, target, flags, logMsgPrefix) {
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
    totalFailed = size-totalUploaded
    if totalFailed > 0 {
        fmt.Println("Failed uploading " + strconv.Itoa(totalFailed) + " artifacts to Artifactory.")
    }
    return
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

func getFilesToUpload(localpath string, targetPath string, flags *utils.Flags) []Artifact {
    if strings.Index(targetPath, "/") < 0 {
        targetPath += "/"
    }
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
            if strings.HasSuffix(target, "/") {
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

// Get the local root path, from which to start collecting artifacts to be uploaded to Artifactory.
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

func getDebianProps(debianPropsStr string) (debianProperties string) {
    debProps := strings.Split(debianPropsStr, "/")
    debianProperties =
        ";deb.distribution=" + debProps[0] +
        ";deb.component=" + debProps[1] +
        ";deb.architecture=" + debProps[2]

    return 
}

// Uploads the file in the specified local path to the specified target path.
// Returns true if the file was successfully uploaded.
func uploadFile(localPath string, targetPath string, flags *utils.Flags, logMsgPrefix string) bool {
    if (flags.Props != "") {
        targetPath += ";" + flags.Props
    }
    if flags.Deb != "" {
        targetPath += getDebianProps(flags.Deb)
    }

    fmt.Println(logMsgPrefix + " Uploading artifact: " + targetPath)
    file, err := os.Open(localPath)
    cliutils.CheckError(err)
    defer file.Close()
    fileInfo, err := file.Stat()
    cliutils.CheckError(err)

    var checksumDeployed bool = false
    var resp *http.Response
    var details *utils.FileDetails
    if fileInfo.Size() >= 10240 {
        resp, details = tryChecksumDeploy(localPath, targetPath, flags)
        checksumDeployed = !flags.DryRun && (resp.StatusCode == 201 || resp.StatusCode == 200)
    }
    if !flags.DryRun && !checksumDeployed {
        resp = utils.UploadFile(file, targetPath, flags.ArtDetails, details)
    }
    if !flags.DryRun {
        var strChecksumDeployed string
        if checksumDeployed {
            strChecksumDeployed = " (Checksum deploy)"
        } else {
            strChecksumDeployed = ""
        }
        fmt.Println(logMsgPrefix + " Artifactory response: " + resp.Status + strChecksumDeployed)
    }

    return flags.DryRun || checksumDeployed || resp.StatusCode == 201 || resp.StatusCode == 200
}

func tryChecksumDeploy(filePath, targetPath string, flags *utils.Flags) (*http.Response, *utils.FileDetails) {
    details := utils.GetFileDetails(filePath)
    headers := make(map[string]string)
    headers["X-Checksum-Deploy"] = "true"
    headers["X-Checksum-Sha1"] = details.Sha1
    headers["X-Checksum-Md5"] = details.Md5

    if flags.DryRun {
        return nil, details
    }
    utils.AddAuthHeaders(headers, flags.ArtDetails)
    resp, _ := cliutils.SendPut(targetPath, nil, headers, flags.ArtDetails.User, flags.ArtDetails.Password)
    return resp, details
}

type Artifact struct {
    LocalPath string
    TargetPath string
}