package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/jfrog/jfrog-cli-go/jfrog-client/utils/errorutils"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

var userAgent = getDefaultUserAgent()

func getVersion() string {
	return "0.1.0"
}

func GetUserAgent() string {
	return userAgent
}

func SetUserAgent(newUserAgent string) {
	userAgent = newUserAgent
}

func getDefaultUserAgent() string {
	return fmt.Sprintf("jfrog-client-go/%s", getVersion())
}

// Get the local root path, from which to start collecting artifacts to be used for:
// 1. Uploaded to Artifactory,
// 2. Adding to the local build-info, to be later published to Artifactory.
func GetRootPath(path string, useRegExp bool) string {
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
		if section == "~" {
			rootPath += GetUserHomeDir()
		} else {
			rootPath += section
		}
	}
	if len(sections) > 0 && sections[0] == "" {
		rootPath = seperator + rootPath
	}
	if rootPath == "" {
		return "."
	}
	return rootPath
}

func StringToBool(boolVal string, defaultValue bool) (bool, error) {
	if len(boolVal) > 0 {
		result, err := strconv.ParseBool(boolVal)
		errorutils.CheckError(err)
		return result, err
	}
	return defaultValue, nil
}

func AddTrailingSlashIfNeeded(url string) string {
	if url != "" && !strings.HasSuffix(url, "/") {
		url += "/"
	}
	return url
}

func IndentJson(jsonStr []byte) string {
	var content bytes.Buffer
	err := json.Indent(&content, jsonStr, "", "  ")
	if err == nil {
		return content.String()
	}
	return string(jsonStr)
}

func MergeMaps(src map[string]string, dst map[string]string) {
	for k, v := range src {
		dst[k] = v
	}
}

func CopyMap(src map[string]string) (dst map[string]string) {
	if dst == nil {
		dst = make(map[string]string)
	}
	for k, v := range src {
		dst[k] = v
	}
	return
}

func PrepareLocalPathForUpload(localPath string, useRegExp bool) string {
	if localPath == "./" || localPath == ".\\" {
		return "^.*$"
	}
	if strings.HasPrefix(localPath, "./") {
		localPath = localPath[2:]
	} else if strings.HasPrefix(localPath, ".\\") {
		localPath = localPath[3:]
	}
	if !useRegExp {
		localPath = pathToRegExp(localPath)
	}
	return localPath
}

func pathToRegExp(localPath string) string {
	var SPECIAL_CHARS = []string{".", "^", "$", "+"}
	for _, char := range SPECIAL_CHARS {
		localPath = strings.Replace(localPath, char, "\\"+char, -1)
	}
	var wildcard = ".*"
	localPath = strings.Replace(localPath, "*", wildcard, -1)
	if strings.HasSuffix(localPath, "/") || strings.HasSuffix(localPath, "\\") {
		localPath += wildcard
	}
	localPath = "^" + localPath + "$"
	return localPath
}

// Replaces matched regular expression from foundPath to corresponding {i} at destString.
// Example 1:
//      providedPath = "repoA/1(.*)234" ; foundPath = "repoA/1hello234" ; destination = "{1}" ; ignoreRepo = false
//      returns "hello"
// Example 2:
//      providedPath = "repoA/1(.*)234" ; foundPath = "repoB/1hello234" ; destination = "{1}" ; ignoreRepo = true
//      returns "hello"
func ReformatDestByPaths(providedPath, foundPath, destination string, ignoreRepo bool) (string, error) {
	if ignoreRepo {
		providedPath = cleanRepoFromPath(providedPath)
		foundPath = cleanRepoFromPath(foundPath)
	}
	providedPath = pathToRegExp(providedPath)
	r, err := regexp.Compile(providedPath)
	err = errorutils.CheckError(err)
	if err != nil {
		return "", err
	}

	groups := r.FindStringSubmatch(foundPath)
	size := len(groups)
	target := destination
	if size > 0 {
		for i := 1; i < size; i++ {
			group := strings.Replace(groups[i], "\\", "/", -1)
			target = strings.Replace(target, "{"+strconv.Itoa(i)+"}", group, -1)
		}
	}
	return target, nil
}

func cleanRepoFromPath(path string) string {
	if idx := strings.Index(path, "/"); idx != -1 {
		return path[idx:]
	}
	return path
}

func GetLogMsgPrefix(threadId int, dryRun bool) string {
	var strDryRun string
	if dryRun {
		strDryRun = "[Dry run] "
	}
	return "[Thread " + strconv.Itoa(threadId) + "] " + strDryRun
}

func TrimPath(path string) string {
	path = strings.Replace(path, "\\", "/", -1)
	path = strings.Replace(path, "//", "/", -1)
	path = strings.Replace(path, "../", "", -1)
	path = strings.Replace(path, "./", "", -1)
	return path
}

func Bool2Int(b bool) int {
	if b {
		return 1
	}
	return 0
}

func ReplaceTildeWithUserHome(path string) string {
	if len(path) > 1 && path[0:1] == "~" {
		return GetUserHomeDir() + path[1:]
	}
	return path
}

func GetUserHomeDir() string {
	if IsWindows() {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return strings.Replace(home, "\\", "\\\\", -1)
	}
	return os.Getenv("HOME")
}

func GetMapFromStringSlice(slice []string, sep string) map[string]string {
	mapFromSlice := make(map[string]string)
	for _, value := range slice {
		splitted := strings.Split(value, sep)
		if len(splitted) == 2 {
			mapFromSlice[splitted[0]] = splitted[1]
		}
		if len(splitted) == 1 {
			mapFromSlice[splitted[0]] = ""
		}
	}
	return mapFromSlice
}

// Split str by the provided separator, escaping the separator if it is prefixed by a back-slash.
func SplitWithEscape(str string, separator rune) []string {
	var parts []string
	var current bytes.Buffer
	escaped := false
	for _, char := range str {
		if char == '\\' {
			if escaped {
				current.WriteRune(char)
			}
			escaped = true
		} else if char == separator && !escaped {
			parts = append(parts, current.String())
			current.Reset()
		} else {
			escaped = false
			current.WriteRune(char)
		}
	}
	parts = append(parts, current.String())
	return parts
}

func IsWindows() bool {
	return runtime.GOOS == "windows"
}

type Artifact struct {
	LocalPath  string
	TargetPath string
	Symlink    string
}
