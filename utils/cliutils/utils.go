package cliutils

import (
	"fmt"
	"os"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/jfrog/jfrog-cli-core/utils/config"
	"github.com/jfrog/jfrog-cli-core/utils/coreutils"
	"github.com/jfrog/jfrog-cli/utils/summary"
	serviceutils "github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/utils"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/content"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/pkg/errors"
)

// Error modes (how should the application behave when the CheckError function is invoked):
type OnError string

func init() {
	// Initialize cli-core values.
	cliUserAgent := os.Getenv(UserAgent)
	if cliUserAgent != "" {
		cliUserAgentName, cliUserAgentVersion := splitAgentNameAndVersion(cliUserAgent)
		coreutils.SetCliUserAgentName(cliUserAgentName)
		coreutils.SetCliUserAgentVersion(cliUserAgentVersion)
	} else {
		coreutils.SetCliUserAgentName(ClientAgent)
		coreutils.SetCliUserAgentVersion(CliVersion)
	}
	coreutils.SetClientAgentName(ClientAgent)
	coreutils.SetClientAgentVersion(CliVersion)
}

// Splits the full agent name to its name and version.
// The full agent name needs to be the agent name and version separated by a slash ('/').
// If the full agent name doesn't include a version, then it's returned as the agent name and an empty string is returned as the agent version.
func splitAgentNameAndVersion(fullAgentName string) (string, string) {
	var agentName, agentVersion string
	lastSlashIndex := strings.LastIndex(fullAgentName, "/")
	if lastSlashIndex == -1 {
		agentName = fullAgentName
	} else {
		agentName = fullAgentName[:lastSlashIndex]
		agentVersion = fullAgentName[lastSlashIndex+1:]
	}

	return agentName, agentVersion
}

func GetCliError(err error, success, failed int, failNoOp bool) error {
	switch coreutils.GetExitCode(err, success, failed, failNoOp) {
	case coreutils.ExitCodeError:
		{
			var errorMessage string
			if err != nil {
				errorMessage = err.Error()
			}
			return coreutils.CliError{ExitCode: coreutils.ExitCodeError, ErrorMsg: errorMessage}
		}
	case coreutils.ExitCodeFailNoOp:
		return coreutils.CliError{ExitCode: coreutils.ExitCodeFailNoOp, ErrorMsg: "No errors, but also no files affected (fail-no-op flag)."}
	default:
		return nil
	}
}

type DetailedSummaryRecord struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

type ExtendedDetailedSummaryRecord struct {
	DetailedSummaryRecord
	Sha256 string `json:"sha256"`
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

func PrintSummaryReport(success, failed int, originalErr error) error {
	basicSummary, mErr := CreateSummaryReportString(success, failed, originalErr)
	if mErr != nil {
		return summaryPrintError(mErr, originalErr)
	}
	log.Output(basicSummary)
	return summaryPrintError(mErr, originalErr)
}

// Prints a summary report.
// If a resultReader is provided, we will iterate over the result and print a detailed summary including the affected files.
func PrintDetailedSummaryReport(success, failed int, reader *content.ContentReader, printExtendedDetails bool, originalErr error) error {
	basicSummary, mErr := CreateSummaryReportString(success, failed, originalErr)
	if mErr != nil {
		return summaryPrintError(mErr, originalErr)
	}
	// A reader wasn't provided, prints the basic summary json and return.
	if reader == nil {
		log.Output(basicSummary)
		return summaryPrintError(mErr, originalErr)
	}
	reader.Reset()
	defer reader.Close()
	writer, mErr := content.NewContentWriter("files", false, true)
	if mErr != nil {
		log.Output(basicSummary)
		return summaryPrintError(mErr, originalErr)
	}
	// We remove the closing curly bracket in order to append the affected files array using a responseWriter to write directly to stdout.
	basicSummary = strings.TrimSuffix(basicSummary, "\n}") + ","
	log.Output(basicSummary)
	defer log.Output("}")
	readerLength, _ := reader.Length()
	// If the reader is empty we will print an empty array.
	if readerLength == 0 {
		log.Output("  files: []")
	} else {
		for transferDetails := new(serviceutils.FileTransferDetails); reader.NextRecord(transferDetails) == nil; transferDetails = new(serviceutils.FileTransferDetails) {
			writer.Write(getDetailedSummaryRecord(transferDetails, printExtendedDetails))
		}
	}
	mErr = writer.Close()
	return summaryPrintError(mErr, originalErr)
}

// Get the detailed summary record.
// In case of an upload/publish commands we want to print sha256 of the uploaded file in addition to the source and the target.
func getDetailedSummaryRecord(transferDetails *serviceutils.FileTransferDetails, extendDetailedSummary bool) interface{} {
	record := DetailedSummaryRecord{
		Source: transferDetails.SourcePath,
		Target: transferDetails.TargetPath,
	}
	if extendDetailedSummary {
		extendedRecord := ExtendedDetailedSummaryRecord{
			DetailedSummaryRecord: record,
			Sha256:                transferDetails.Sha256,
		}
		return extendedRecord
	}
	return record
}

func PrintBuildInfoSummaryReport(succeeded bool, sha256 string, originalErr error) error {
	success, failed := 1, 0
	if !succeeded {
		success, failed = 0, 1
	}
	summary, mErr := CreateBuildInfoSummaryReportString(success, failed, sha256, originalErr)
	if mErr != nil {
		return summaryPrintError(mErr, originalErr)
	}
	log.Output(summary)
	return summaryPrintError(mErr, originalErr)
}

func CreateSummaryReportString(success, failed int, err error) (string, error) {
	summaryReport := summary.GetSummaryReport(success, failed, err)
	content, mErr := summaryReport.Marshal()
	if errorutils.CheckError(mErr) != nil {
		return "", mErr
	}
	return utils.IndentJson(content), mErr
}

func CreateBuildInfoSummaryReportString(success, failed int, sha256 string, err error) (string, error) {
	buildInfoSummary := summary.NewBuildInfoSummary(success, failed, sha256, err)
	content, mErr := buildInfoSummary.Marshal()
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

// Return true if the CI environment variable was set to true.
func getCiValue() bool {
	var ci bool
	var err error
	if ci, err = clientutils.GetBoolEnvValue(coreutils.CI, false); err != nil {
		return false
	}
	return ci
}

func GetVersion() string {
	return CliVersion
}

func GetDocumentationMessage() string {
	return "You can read the documentation at https://www.jfrog.com/confluence/display/CLI/JFrog+CLI"
}

func GetBuildName(buildName string) string {
	return getOrDefaultEnv(buildName, coreutils.BuildName)
}

func GetBuildUrl(buildUrl string) string {
	return getOrDefaultEnv(buildUrl, BuildUrl)
}

func GetEnvExclude(envExclude string) string {
	return getOrDefaultEnv(envExclude, EnvExclude)
}

// Return argument if not empty or retrieve from environment variable
func getOrDefaultEnv(arg, envKey string) string {
	if arg != "" {
		return arg
	}
	return os.Getenv(envKey)
}

func ShouldOfferConfig() (bool, error) {
	exists, err := config.IsServerConfExists()
	if err != nil || exists {
		return false, err
	}

	var ci bool
	if ci, err = clientutils.GetBoolEnvValue(coreutils.CI, false); err != nil {
		return false, err
	}
	var offerConfig bool
	if offerConfig, err = clientutils.GetBoolEnvValue(OfferConfig, !ci); err != nil {
		return false, err
	}
	if !offerConfig {
		config.SaveServersConf(make([]*config.ServerDetails, 0))
		return false, nil
	}

	msg := fmt.Sprintf("To avoid this message in the future, set the %s environment variable to false.\n"+
		"The CLI commands require the URL and authentication details\n"+
		"Configuring JFrog CLI with these parameters now will save you having to include them as command options.\n"+
		"You can also configure these parameters later using the 'jfrog c' command.\n"+
		"Configure now?", OfferConfig)
	confirmed := coreutils.AskYesNo(msg, false)
	if !confirmed {
		config.SaveServersConf(make([]*config.ServerDetails, 0))
		return false, nil
	}
	return true, nil
}

func CreateServerDetailsFromFlags(c *cli.Context) (details *config.ServerDetails) {
	details = new(config.ServerDetails)
	details.Url = clientutils.AddTrailingSlashIfNeeded(c.String(url))
	details.ArtifactoryUrl = clientutils.AddTrailingSlashIfNeeded(c.String(configRtUrl))
	details.DistributionUrl = clientutils.AddTrailingSlashIfNeeded(c.String(distUrl))
	details.XrayUrl = clientutils.AddTrailingSlashIfNeeded(c.String(configXrUrl))
	details.MissionControlUrl = clientutils.AddTrailingSlashIfNeeded(c.String(configMcUrl))
	details.PipelinesUrl = clientutils.AddTrailingSlashIfNeeded(c.String(configPlUrl))
	details.ApiKey = c.String(apikey)
	details.User = c.String(user)
	details.Password = c.String(password)
	details.SshKeyPath = c.String(sshKeyPath)
	details.SshPassphrase = c.String(sshPassPhrase)
	details.AccessToken = c.String(accessToken)
	details.ClientCertPath = c.String(clientCertPath)
	details.ClientCertKeyPath = c.String(clientCertKeyPath)
	details.ServerId = c.String(serverId)
	details.InsecureTls = c.Bool(insecureTls)
	if details.ApiKey != "" && details.User != "" && details.Password == "" {
		// The API Key is deprecated, use password option instead.
		details.Password = details.ApiKey
		details.ApiKey = ""
	}
	return
}

func IsLegacyGoPublish(c *cli.Context) bool {
	return c.Command.Name == "go-publish" && c.NArg() > 1
}
