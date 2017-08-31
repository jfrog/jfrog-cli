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
	"github.com/jfrogdev/jfrog-cli-go/utils/cliutils/log"
	"fmt"
)

const CliAgent = "jfrog-cli-go"

// CLI base commands constants:
const CmdArtifactory = "rt"
const CmdBintray = "bt"
const CmdMissionControl = "mc"
const CmdXray = "xr"

// Error modes (how should the application behave when the CheckError function is invoked):
type OnError string

const (
	OnErrorPanic OnError = "panic"
	OnErrorReturnError OnError = "return"
)

var onError OnError

func init() {
	onError = OnErrorReturnError
	if os.Getenv("JFROG_CLI_ERROR_HANDLING") == string(OnErrorPanic) {
		onError = OnErrorPanic
	}
}

// Exit codes:
type ExitCode struct {
	Code int
}

var ExitCodeError ExitCode = ExitCode{1}
var ExitCodeWarning ExitCode = ExitCode{2}

func CheckError(err error) error {
	if err != nil {
		if onError == OnErrorPanic {
			panic(err)
		}
	}
	return err
}

func CheckErrorWithMessage(err error, message string) error {
	if err != nil {
		log.Error(message)
		err = CheckError(err)
	}
	return err
}

func ExitOnErr(err error) {
	if err != nil {
		Exit(ExitCodeError, err.Error())
	}
}

func Exit(exitCode ExitCode, msg string) {
	if msg != "" {
		log.Error(msg)
	}
	os.Exit(exitCode.Code)
}

func PrintHelpAndExitWithError(msg string, context *cli.Context) {
	log.Error(msg + " " + GetDocumentationMessage())
	cli.ShowCommandHelp(context, context.Command.Name)
	os.Exit(ExitCodeError.Code)
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

func InteractiveConfirm(message string) bool {
	var confirm string
	fmt.Print(message + " (y/n): ")
	fmt.Scanln(&confirm)
	return confirmAnswer(confirm)
}

func confirmAnswer(answer string) bool {
	answer = strings.ToLower(answer)
	return answer == "y" || answer == "yes"
}

func GetLogMsgPrefix(threadId int, dryRun bool) string {
	var strDryRun string
	if dryRun {
		strDryRun = "[Dry run] "
	}
	return "[Thread " + strconv.Itoa(threadId) + "] " + strDryRun
}

func GetVersion() string {
	return "1.10.3"
}

func GetConfigVersion() string {
	return "1"
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
		localPath = PathToRegExp(localPath)
	}
	return localPath
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

func StringToBool(boolVal string, defaultValue bool) (bool, error) {
	if len(boolVal) > 0 {
		result, err := strconv.ParseBool(boolVal)
		CheckError(err)
		return result, err
	}
	return defaultValue, nil
}

func GetDocumentationMessage() string {
	return "You can read the documentation at https://www.jfrog.com/confluence/display/CLI/JFrog+CLI"
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

func SpecVarsStringToMap(rawVars string) map[string]string {
	if len(rawVars) == 0 {
		return nil
	}
	varCandidates := strings.Split(rawVars, ";")
	varsList := []string{}
	for _, v := range varCandidates {
		if len(varsList) > 0 && isEndsWithEscapeChar(varsList[len(varsList) - 1]) {
			currentLastVar := varsList[len(varsList) - 1]
			varsList[len(varsList) - 1] = strings.TrimSuffix(currentLastVar, "\\") + ";" + v
			continue
		}
		varsList = append(varsList, v)
	}
	return varsAsMap(varsList)
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

func isEndsWithEscapeChar(lastVar string) bool {
	return strings.HasSuffix(lastVar, "\\")
}

func varsAsMap(vars []string) map[string]string {
	result := map[string]string{}
	for _, v := range vars {
		keyVal := strings.SplitN(v, "=", 2)
		result[keyVal[0]] = keyVal[1]
	}
	return result
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
	Symlink    string
}
