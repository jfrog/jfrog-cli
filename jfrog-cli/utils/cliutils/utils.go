package cliutils

import (
	"github.com/codegangsta/cli"
	"os"
	"strconv"
	"strings"
	"runtime"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
	"fmt"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/io/httputils"
)

const CliAgent = "jfrog-cli-go"

// CLI base commands constants:
const CmdArtifactory = "rt"
const CmdBintray = "bt"
const CmdMissionControl = "mc"
const CmdXray = "xr"

// Error modes (how should the application behave when the CheckError function is invoked):
type OnError string
const OnErrorPanic OnError = "panic"

func init() {
	httputils.UserAgent = CliAgent + "/" + GetVersion()
	if os.Getenv("JFROG_CLI_ERROR_HANDLING") == string(OnErrorPanic) {
		errorutils.CheckError = PanicOnError
	}
}

// Exit codes:
type ExitCode struct {
	Code int
}

var ExitCodeError ExitCode = ExitCode{1}
var ExitCodeWarning ExitCode = ExitCode{2}

func PanicOnError(err error) error {
	if err != nil {
		panic(err)
	}
	return err
}

func CheckErrorWithMessage(err error, message string) error {
	if err != nil {
		log.Error(message)
		err = errorutils.CheckError(err)
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

func GetVersion() string {
	return "1.11.1"
}

func GetConfigVersion() string {
	return "1"
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
			rootPath += utils.GetUserHomeDir()
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
		localPath = utils.PathToRegExp(localPath)
	}
	return localPath
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
		errorutils.CheckError(err)
		return result, err
	}
	return defaultValue, nil
}

func GetDocumentationMessage() string {
	return "You can read the documentation at https://www.jfrog.com/confluence/display/CLI/JFrog+CLI"
}

func GetTestsFileSeperator() string {
	if runtime.GOOS == "windows" {
		return "\\\\"
	}
	return "/"
}

func SumTrueValues(boolArr []bool) int {
	counter := 0
	for _, val := range boolArr {
		counter += utils.Bool2Int(val)
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
