package commands

import (
	"fmt"
	"regexp"
	"strings"
	"strconv"
    "github.com/JFrogDev/bintray-cli-go/utils"
)

func Upload(versionDetails *utils.VersionDetails, localPath, uploadPath string, flags *UploadFlags) {

    // Get the list of artifacts to be uploaded to:
    artifacts := getFilesToUpload(localPath, uploadPath, flags)

    baseUrl := flags.BintrayDetails.ApiUrl + "content/" + versionDetails.Subject + "/" +
           versionDetails.Repo + "/" + versionDetails.Package + "/" + versionDetails.Version + "/";

    for _, artifact := range artifacts {
        url := baseUrl + artifact.TargetPath
        if !flags.DryRun {
            fmt.Println("Uploading artifact: " + url)
            resp := utils.UploadFile(artifact.LocalPath, url, flags.BintrayDetails.User, flags.BintrayDetails.Key)
            fmt.Println("Bintray response: " + resp.Status)
        } else {
            fmt.Println("[Dry Run] Uploading artifact: " + url)
        }
    }
}

func getFilesToUpload(localpath string, targetPath string, flags *UploadFlags) []Artifact {
    rootPath := getRootPath(localpath, flags.UseRegExp)
    if !utils.IsPathExists(rootPath) {
        utils.Exit("Path does not exist: " + rootPath)
    }
    localpath = prepareLocalPath(localpath, flags.UseRegExp)

    artifacts := []Artifact{}
    // If the path is a single file then return it
    if !utils.IsDir(rootPath) {
        targetPath := prepareUploadPath(targetPath + rootPath)
        artifacts = append(artifacts, Artifact{rootPath, targetPath})
        return artifacts
    }

    r, err := regexp.Compile(localpath)
    utils.CheckError(err)

    var paths []string
    if flags.Recursive {
        paths = utils.ListFilesRecursive(rootPath)
    } else {
        paths = utils.ListFiles(rootPath)
    }

    for _, path := range paths {
        if utils.IsDir(path) {
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
                    target += utils.GetFileNameFromPath(path)
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
    UseRegExp bool
}