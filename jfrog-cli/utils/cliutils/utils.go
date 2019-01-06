package cliutils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/summary"
	"github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

// Error modes (how should the application behave when the CheckError function is invoked):
type OnError string

func init() {
	if os.Getenv("JFROG_CLI_ERROR_HANDLING") == string(OnErrorPanic) {
		errorutils.CheckError = PanicOnError
	}
}

// Exit codes:
type ExitCode struct {
	Code int
}

var ExitCodeNoError = ExitCode{0}
var ExitCodeError = ExitCode{1}
var ExitCodeFailNoOp = ExitCode{2}
var ExitCodeBuildScan = ExitCode{3}

func PanicOnError(err error) error {
	if err != nil {
		panic(err)
	}
	return err
}

func ExitOnErr(err error) {
	if exitCode := GetExitCode(err, 0, 0, false); exitCode != ExitCodeNoError {
		traceExit(exitCode, err)
	}
}

func FailNoOp(err error, success, failed int, failNoOp bool) {
	if exitCode := GetExitCode(err, success, failed, failNoOp); exitCode != ExitCodeNoError {
		traceExit(exitCode, err)
	}
}

func ExitBuildScan(failBuild bool, err error) {
	if failBuild {
		traceExit(ExitCodeBuildScan, err)
	}
	ExitOnErr(err)
}

func GetExitCode(err error, success, failed int, failNoOp bool) ExitCode {
	// Error occurred - Return 1
	if err != nil || failed > 0 {
		return ExitCodeError
	}
	// No errors, but also no files affected - Return 2 if failNoOp
	if success == 0 && failNoOp {
		return ExitCodeFailNoOp
	}
	// Otherwise - Return 0
	return ExitCodeNoError
}

func traceExit(exitCode ExitCode, err error) {
	if err != nil && len(err.Error()) > 0 {
		log.Error(err)
	}
	os.Exit(exitCode.Code)
}

// Print summary report.
// The given error will pass through and be returned as is if no other errors are raised.
func PrintSummaryReport(success, failed int, err error) error {
	summaryReport := summary.New(err)
	summaryReport.Totals.Success = success
	summaryReport.Totals.Failure = failed
	if err == nil && summaryReport.Totals.Failure != 0 {
		summaryReport.Status = summary.Failure
	}
	content, mErr := summaryReport.Marshal()
	if errorutils.CheckError(mErr) != nil {
		log.Error(mErr)
		return err
	}
	log.Output(utils.IndentJson(content))
	return err
}

func PrintHelpAndExitWithError(msg string, context *cli.Context) {
	log.Error(msg + " " + GetDocumentationMessage())
	cli.ShowCommandHelp(context, context.Command.Name)
	os.Exit(ExitCodeError.Code)
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
	if ok, _ := utils.GetBoolEnvValue("JFROG_CLI_SHOW_UPDATE", true); ok {
		latestRelease, err := getLatestRelease()
		if err != nil {
			return CliVersion
		}

		if CliVersion == latestRelease {
			return fmt.Sprintf("%s (you're running the latest version)", CliVersion)
		}
		return fmt.Sprintf("%s (there is a newer version available [v%s])", CliVersion, latestRelease)
	}
	return CliVersion
}

func getLatestRelease() (string, error) {
	req, err := http.NewRequest(http.MethodGet, "https://registry.npmjs.org/jfrog-cli-go/latest", nil)
	if err != nil {
		log.Output("Error creating HTTP request: %s", err.Error())
		return "", err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Output("Error sending HTTP request: %s", err.Error())
		return "", err
	}

	if res.StatusCode != 200 {
		log.Output("Error receiving HTTP response (HTTP/[%d]): %s", res.StatusCode, err.Error())
		return "", err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	var data map[string]interface{}

	if err := json.Unmarshal(body, &data); err != nil {
		log.Output("Error unmarshalling HTTP response: %s", err.Error())
		return "", err
	}

	return data["version"].(string), nil
}

func GetConfigVersion() string {
	return "1"
}

func GetDocumentationMessage() string {
	return "You can read the documentation at https://www.jfrog.com/confluence/display/CLI/JFrog+CLI"
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
		if len(varsList) > 0 && isEndsWithEscapeChar(varsList[len(varsList)-1]) {
			currentLastVar := varsList[len(varsList)-1]
			varsList[len(varsList)-1] = strings.TrimSuffix(currentLastVar, "\\") + ";" + v
			continue
		}
		varsList = append(varsList, v)
	}
	return varsAsMap(varsList)
}

func isEndsWithEscapeChar(lastVar string) bool {
	return strings.HasSuffix(lastVar, "\\")
}

func varsAsMap(vars []string) map[string]string {
	result := map[string]string{}
	for _, v := range vars {
		keyVal := strings.SplitN(v, "=", 2)
		if len(keyVal) != 2 {
			continue
		}
		result[keyVal[0]] = keyVal[1]
	}
	return result
}

func IsWindows() bool {
	return runtime.GOOS == "windows"
}

func GetTempDir() string {
	tempDirPath := os.Getenv("JFROG_CLI_TEMP_DIR")
	if tempDirPath != "" {
		return tempDirPath
	}
	return os.TempDir()
}

type Credentials interface {
	SetUser(string)
	SetPassword(string)
	GetUser() string
	GetPassword() string
}
