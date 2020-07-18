package cliutils

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	serviceutils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/utils/io/content"
	"github.com/jfrog/jfrog-client-go/utils/prompt"

	"github.com/codegangsta/cli"
	"github.com/jfrog/jfrog-cli/utils/summary"
	"github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/pkg/errors"

	clientutils "github.com/jfrog/jfrog-client-go/utils"
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
var ExitCodeVulnerableBuild = ExitCode{3}

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

type detailedSummaryRecord struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

// Print summary report.
// a given non-nil error will pass through and be returned as is if no other errors are raised.
// In case of a nil error, the current function error will be returned.
func summaryPrintError(summaryError, originalError error) error {
	if originalError != nil {
		if summaryError != nil {
			log.Error(summaryError)
		}
		return originalError
	}
	return summaryError
}

// If a resultReader is provided, we will iterate over the result and print a detailed summary including the affected files.
func PrintSummaryReport(success, failed int, resultReader *content.ContentReader, rtUrl string, originalErr error) error {
	basicSummary, mErr := CreateSummaryReportString(success, failed, originalErr)
	if mErr != nil {
		return summaryPrintError(mErr, originalErr)
	}
	// A reader wasn't provided, prints the basic summary json and return.
	if resultReader == nil {
		log.Output(basicSummary)
		return summaryPrintError(mErr, originalErr)
	}
	resultReader.Reset()
	defer resultReader.Close()
	writer, mErr := content.NewContentWriter("files", false, true)
	if mErr != nil {
		log.Output(basicSummary)
		return summaryPrintError(mErr, originalErr)
	}
	// We remove the closing curly bracket in order to append the affected files array using a responseWriter to write directly to stdout.
	basicSummary = strings.TrimSuffix(basicSummary, "\n}") + ","
	log.Output(basicSummary)
	defer log.Output("}")
	var file serviceutils.FileInfo
	for e := resultReader.NextRecord(&file); e == nil; e = resultReader.NextRecord(&file) {
		record := detailedSummaryRecord{
			Source: rtUrl + file.ArtifactoryPath,
			Target: file.LocalPath,
		}
		writer.Write(record)
	}
	mErr = writer.Close()
	return summaryPrintError(mErr, originalErr)
}

func CreateSummaryReportString(success, failed int, err error) (string, error) {
	summaryReport := summary.New(err)
	summaryReport.Totals.Success = success
	summaryReport.Totals.Failure = failed
	if err == nil && summaryReport.Totals.Failure != 0 {
		summaryReport.Status = summary.Failure
	}
	content, mErr := summaryReport.Marshal()
	if errorutils.CheckError(mErr) != nil {
		return "", mErr
	}
	return utils.IndentJson(content), mErr
}

func PrintHelpAndReturnError(msg string, context *cli.Context) error {
	log.Error(msg + " " + GetDocumentationMessage())
	cli.ShowCommandHelp(context, context.Command.Name)
	return errors.New(msg)
}

// This function indicates whether the command should be executed without
// confirmation warning or not.
// If the --quiet option was sent, it is used to determine whether to prompt the confirmation or not.
// If not, the command will prompt the confirmation, unless the CI environment variable was set to true.
func GetQuietValue(c *cli.Context) bool {
	if c.IsSet("quiet") {
		return c.Bool("quiet")
	}

	return getCiValue()
}

// This function indicates whether the command should be executed in
// an interactive mode.
// If the --interactive option was sent, it is used to determine the mode.
// If not, the mode will be interactive, unless the CI environment variable was set to true.
func GetInteractiveValue(c *cli.Context) bool {
	if c.IsSet("interactive") {
		return c.BoolT("interactive")
	}

	return !getCiValue()
}

// Return the true if the CI environment variable was set to true.
func getCiValue() bool {
	var ci bool
	var err error
	if ci, err = clientutils.GetBoolEnvValue(CI, false); err != nil {
		return false
	}
	return ci
}

func InteractiveConfirm(message string, defaultValue bool) bool {
	var confirm string
	defStr := "[n]"
	if defaultValue {
		defStr = "[y]"
	}
	fmt.Print(message + " (y/n) " + defStr + ": ")
	fmt.Scanln(&confirm)
	return confirmAnswer(confirm, defaultValue)
}

func confirmAnswer(answer string, defaultValue bool) bool {
	answer = strings.ToLower(strings.TrimSpace(answer))
	if answer == "" {
		return defaultValue
	}
	return answer == "y" || answer == "yes"
}

func GetVersion() string {
	return CliVersion
}

func GetConfigVersion() string {
	return "3"
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

func IsLinux() bool {
	return runtime.GOOS == "linux"
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

func ReplaceVars(content []byte, specVars map[string]string) []byte {
	log.Debug("Replacing variables in the provided content: \n" + string(content))
	for key, val := range specVars {
		key = "${" + key + "}"
		log.Debug(fmt.Sprintf("Replacing '%s' with '%s'", key, val))
		content = bytes.Replace(content, []byte(key), []byte(val), -1)
	}
	log.Debug("The reformatted content is: \n" + string(content))
	return content
}

func GetJfrogHomeDir() (string, error) {
	// The JfrogHomeEnv environment variable has been deprecated and replaced with HomeDir
	if os.Getenv(HomeDir) != "" {
		return os.Getenv(HomeDir), nil
	} else if os.Getenv(JfrogHomeEnv) != "" {
		return path.Join(os.Getenv(JfrogHomeEnv), ".jfrog"), nil
	}

	userHomeDir := fileutils.GetHomeDir()
	if userHomeDir == "" {
		err := errorutils.CheckError(errors.New("couldn't find home directory. Make sure your HOME environment variable is set"))
		if err != nil {
			return "", err
		}
	}
	return filepath.Join(userHomeDir, ".jfrog"), nil
}

func CreateDirInJfrogHome(dirName string) (string, error) {
	homeDir, err := GetJfrogHomeDir()
	if err != nil {
		return "", err
	}
	folderName := filepath.Join(homeDir, dirName)
	err = fileutils.CreateDirIfNotExist(folderName)
	return folderName, err
}

func GetJfrogSecurityDir() (string, error) {
	homeDir, err := GetJfrogHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, JfrogSecurityDirName), nil
}

func GetJfrogCertsDir() (string, error) {
	securityDir, err := GetJfrogSecurityDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(securityDir, JfrogCertsDirName), nil
}

func GetJfrogSecurityConfFilePath() (string, error) {
	securityDir, err := GetJfrogSecurityDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(securityDir, JfrogSecurityConfFile), nil
}

func GetJfrogBackupDir() (string, error) {
	homeDir, err := GetJfrogHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, JfrogBackupDirName), nil
}

func AskYesNo(message string, defaultStr string, label string) (bool, error) {
	question := &prompt.YesNo{
		Msg:     message,
		Default: defaultStr,
		Label:   label,
	}
	if err := question.Read(); err != nil {
		return false, errorutils.CheckError(err)
	}
	return question.Result.GetBool(label), nil
}
