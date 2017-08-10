package utils

import (
	"bytes"
	"encoding/json"
	"strings"
	"regexp"
	"strconv"
	"runtime"
	"os"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
)

func MergeMaps(src map[string]string, dst map[string]string) {
	for k, v := range src {
		dst[k] = v
	}
}

func IndentJson(jsonStr []byte) string {
	var content bytes.Buffer
	err := json.Indent(&content, jsonStr, "", "  ")
	if err == nil {
		return content.String()
	}
	return string(jsonStr)
}

// Remove all chars from the given string.
func StripChars(str, chr string) string {
	return strings.Map(func(r rune) rune {
		if strings.IndexRune(chr, r) < 0 {
			return r
		}
		return -1
	}, str)
}

func PathToRegExp(localPath string) string {
	var SPECIAL_CHARS = []string{".", "+"}
	for _, char := range SPECIAL_CHARS {
		localPath = strings.Replace(localPath, char, "\\" + char, -1)
	}
	var wildcard = ".*"
	localPath = strings.Replace(localPath, "*", wildcard, -1)
	if strings.HasSuffix(localPath, "/") || strings.HasSuffix(localPath, "\\") {
		localPath += wildcard
	}
	localPath = "^" + localPath + "$"
	return localPath
}


// Replaces matched regular expression from sourceString to corresponding {i} at destString.
// For example:
//      regexpString = "1(.*)234" ; sourceString = "1hello234" ; destString = "{1}"
//      returns "hello"
func ReformatRegexp(regexpString, sourceString, destString string) (string, error) {
	r, err := regexp.Compile(regexpString)
	err = errorutils.CheckError(err)
	if err != nil {
		return "", err
	}

	groups := r.FindStringSubmatch(sourceString)
	size := len(groups)
	target := destString
	if size > 0 {
		for i := 1; i < size; i++ {
			group := strings.Replace(groups[i], "\\", "/", -1)
			target = strings.Replace(target, "{" + strconv.Itoa(i) + "}", group, -1)
		}
	}
	return target, nil
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

func
ReplaceTildeWithUserHome(path string) string {
	if len(path) > 1 && path[0:1] == "~" {
		return GetUserHomeDir() + path[1:len(path)]
	}
	return path
}

func GetUserHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}

type Artifact struct {
	LocalPath  string
	TargetPath string
	Symlink    string
}
