package cliutils

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-cli/utils/summary"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/errorutils"
	"github.com/jfrogdev/jfrog-cli-go/jfrog-client/utils/log"
	"os"
	"strconv"
	"strings"
)

// CLI base commands constants:
const CmdArtifactory = "rt"
const CmdBintray = "bt"
const CmdMissionControl = "mc"
const CmdXray = "xr"

// Error modes (how should the application behave when the CheckError function is invoked):
type OnError string

const OnErrorPanic OnError = "panic"

func init() {
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
		return mErr
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
	return "1.13.1"
}

func GetConfigVersion() string {
	return "1"
}

func GetBoolEnvValue(flagName string, defValue bool) (bool, error) {
	envVarValue := os.Getenv(flagName)
	if envVarValue == "" {
		return defValue, nil
	}
	val, err := strconv.ParseBool(envVarValue)
	err = CheckErrorWithMessage(err, "can't parse environment variable "+flagName)
	return val, err
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
