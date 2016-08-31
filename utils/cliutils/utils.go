package cliutils

import (
	"bytes"
	"github.com/codegangsta/cli"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"runtime"
	"regexp"
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/logger"
)

// CLI base commands constants:
const CmdArtifactory = "rt"
const CmdBintray = "bt"
const CmdMissionControl = "mc"

// Error modes (how should the application behave when the CheckError function is invoked):
type OnError string
const (
    OnErrorPanic OnError = "panic"
    OnErrorReturnError OnError = "return"
)
var onError = OnErrorReturnError

// Exit codes:
type ExitCode struct {
	Code int
}
var ExitCodeError ExitCode = ExitCode{1}
var ExitCodeWarning ExitCode = ExitCode{2}

func CheckError(err error) error {
    return handleError(err, ExitCodeError)
}

func CheckWarning(err error) error {
    return handleError(err, ExitCodeWarning)
}

func handleError(err error, exitCode ExitCode) error {
    if err != nil {
        if onError == OnErrorPanic {
            panic(err)
        }
        logger.Logger.Error(err.Error())
    }
    return err
}

func CheckErrorWithMessage(err error, message string) error {
	if err != nil {
		logger.Logger.Error(message)
		err = CheckError(err)
	}
	return err
}

func Exit(exitCode ExitCode, msg string) {
	if msg != "" {
		logger.Logger.Error(msg)
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
		if i + 1 < size {
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
		strDryRun = " [Dry run] "
	} else {
		strDryRun = " "
	}
	return "[Thread " + strconv.Itoa(threadId) + "]" + strDryRun
}

func GetVersion() string {
	return "1.5.0-dev"
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

func ReplaceTildeWithUserHome(path string) string {
	if len(path) > 1 && path[0:1] == "~" {
		return GetUserHomeDir() + path[1:len(path)]
	}
	return path
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
		localpath = PathToRegExp(localpath)
	}
	return localpath
}

func TrimPath(path string) string {
	path = strings.Replace(path, "\\", "/", -1)
	path = strings.Replace(path, "//", "/", -1)
	path = strings.Replace(path, "../", "", -1)
	path = strings.Replace(path, "./", "", -1)
	return path
}

func GetBoolFlagValue(c *cli.Context, flagName string, defValue bool) bool {
	if c.IsSet(flagName) {
		return c.Bool(flagName)
	}
	return defValue
}

func GetBoolEnvValue(flagName string, defValue bool) (bool, error) {
	envVarValue := os.Getenv(flagName)
	if envVarValue == "" {
		return defValue, nil
	}
	val, err := strconv.ParseBool(envVarValue)
	err = CheckErrorWithMessage(err, "can't parse environment variable " + flagName)
    return val, err
}

func GetDocumentationMessage() string {
	return "You can read the documentation at https://github.com/jfrogdev/jfrog-cli-go/blob/master/README.md"
}

func PathToRegExp(localpath string) string {
	var wildcard = ".*"

	localpath = strings.Replace(localpath, ".", "\\.", -1)
	localpath = strings.Replace(localpath, "*", wildcard, -1)
	if strings.HasSuffix(localpath, "/") || strings.HasSuffix(localpath, "\\") {
		localpath += wildcard
	}
	localpath = "^" + localpath + "$"
	return localpath
}

// Replaces matched regular expression from sourceString to corresponding {i} at destString.
// For example:
//      regexpString = "1(.*)234" ; sourceString = "1hello234" ; destString = "{1}"
//      returns "hello"
func ReformatRegexp(regexpString, sourceString, destString string) (string, error) {
	r, err := regexp.Compile(regexpString)
	err = CheckError(err)
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

func GetTestsFileSeperator() string {
	if runtime.GOOS == "windows" {
		return "\\\\"
	}
	return "/"
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

func MergeMaps(src map[string]string, dst map[string]string) {
	for k, v := range src {
		dst[k] = v
	}
}

func Bool2Int(b bool) int {
	if b {
		return 1
	}
	return 0
}

func SumTrueValues(boolArr []bool) int {
	counter := 0
	for _, val := range boolArr {
		counter += Bool2Int(val)
	}
	return counter
}

type Credentials interface {
	SetUser(string)
	SetPassword(string)
	GetUser() string
	GetPassword() string
}

type Artifact struct {
	LocalPath  string
	TargetPath string
}

