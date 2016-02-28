package cliutils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var ExitCodeError ExitCode = ExitCode{1}
var ExitCodeWarning ExitCode = ExitCode{2}

type ExitCode struct {
	Code int
}

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}

func Exit(exitCode ExitCode, msg string) {
	if msg != "" {
		fmt.Println(msg)
	}
	os.Exit(exitCode.Code)
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

// Creates a string in the form of ["item-1","item-2","item-3"...] from an input
// in the form of item-1,item-1,item-1...
func BuildListString(listStr string) string {
	if listStr == "" {
		return ""
	}
	split := strings.Split(listStr, ",")
	size := len(split)
	str := "[\""
	for i := 0; i < size; i++ {
		str += split[i]
		if i+1 < size {
			str += "\",\""
		}
	}
	str += "\"]"
	return str
}

func MapToJson(m map[string]string) string {
	first := true
	json := "{"

	for key := range m {
		val := m[key]
		if val != "" {
			if !first {
				json += ","
			}
			first = false
			if !strings.HasPrefix(val, "[") || !strings.HasSuffix(val, "]") {
				val = "\"" + val + "\""
			}
			json += "\"" + key + "\": " + val
		}
	}
	if first {
		return ""
	}
	json += "}"
	return json
}

func ConfirmAnswer(answer string) bool {
	answer = strings.ToLower(answer)
	return answer == "y" || answer == "yes"
}

func GetLogMsgPrefix(threadId int, dryRun bool) string {
	var strDryRun string
	if dryRun {
		strDryRun = " [Dry run]"
	} else {
		strDryRun = ""
	}
	return "[Thread " + strconv.Itoa(threadId) + "]" + strDryRun
}

func GetVersion() string {
	return "1.0.0"
}

// Get the local root path, from which to start collecting artifacts to be uploaded to Artifactory.
func GetRootPathForUpload(path string, useRegExp bool) string {
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

func PrepareLocalPathForUpload(localpath string, useRegExp bool) string {
	if localpath == "./" || localpath == ".\\" {
		return "^.*$"
	}
	if strings.HasPrefix(localpath, "./") {
		localpath = localpath[2:]
	} else if strings.HasPrefix(localpath, ".\\") {
		localpath = localpath[3:]
	}
	if !useRegExp {
		localpath = localPathToRegExp(localpath)
	}
	return localpath
}

func PrepareUploadPath(path string) string {
	path = strings.Replace(path, "\\", "/", -1)
	path = strings.Replace(path, "../", "", -1)
	path = strings.Replace(path, "./", "", -1)
	return path
}

func localPathToRegExp(localpath string) string {
	var wildcard = ".*"

	localpath = strings.Replace(localpath, ".", "\\.", -1)
	localpath = strings.Replace(localpath, "*", wildcard, -1)
	if strings.HasSuffix(localpath, "/") {
		localpath += wildcard
	} else if strings.HasSuffix(localpath, "\\") {
		size := len(localpath)
		if size > 1 && localpath[size-2:size-1] != "\\" {
			localpath += "\\"
		}
		localpath += wildcard
	}
	localpath = "^" + localpath + "$"
	return localpath
}

type Artifact struct {
	LocalPath  string
	TargetPath string
}