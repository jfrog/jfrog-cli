package cliutils

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/jfrog/jfrog-cli-go/utils/summary"
	"github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/pkg/errors"
)

// Error modes (how should the application behave when the CheckError function is invoked):
type OnError string

var cliTempDir string
var cliUserAgent string

func init() {
	// Initialize error handling.
	if os.Getenv(ErrorHandling) == string(OnErrorPanic) {
		errorutils.CheckError = PanicOnError
	}

	// Initialize the temp base-dir path of the CLI executions.
	cliTempDir = os.Getenv(TempDir)
	if cliTempDir == "" {
		cliTempDir = os.TempDir()
	}
	fileutils.SetTempDirBase(cliTempDir)

	// Initialize agent name and version.
	cliUserAgent = os.Getenv(UserAgent)
	if cliUserAgent == "" {
		cliUserAgent = ClientAgent + "/" + CliVersion
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

type CliError struct {
	ExitCode
	ErrorMsg string
}

func (err CliError) Error() string {
	return err.ErrorMsg
}

func PanicOnError(err error) error {
	if err != nil {
		panic(err)
	}
	return err
}

func ExitOnErr(err error) {
	if err, ok := err.(CliError); ok {
		traceExit(err.ExitCode, err)
	}
	if exitCode := GetExitCode(err, 0, 0, false); exitCode != ExitCodeNoError {
		traceExit(exitCode, err)
	}
}

func GetCliError(err error, success, failed int, failNoOp bool) error {
	switch GetExitCode(err, success, failed, failNoOp) {
	case ExitCodeError:
		{
			var errorMessage string
			if err != nil {
				errorMessage = err.Error()
			}
			return CliError{ExitCodeError, errorMessage}
		}
	case ExitCodeFailNoOp:
		return CliError{ExitCodeFailNoOp, "No errors, but also no files affected (fail-no-op flag)."}
	default:
		return nil
	}
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

func PrintHelpAndReturnError(msg string, context *cli.Context) error {
	log.Error(msg + " " + GetDocumentationMessage())
	cli.ShowCommandHelp(context, context.Command.Name)
	return errors.New(msg)
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
	return CliVersion
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

// Return the path of CLI temp dir.
// This path should be persistent, meaning - should not be cleared at the end of a CLI run.
func GetCliPersistentTempDirPath() string {
	return cliTempDir
}

func GetUserAgent() string {
	return cliUserAgent
}

type Credentials interface {
	SetUser(string)
	SetPassword(string)
	GetUser() string
	GetPassword() string
}
