package cliutils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/jfrog/gofrog/version"
	"github.com/jfrog/jfrog-client-go/utils/log"

	corecontainercmds "github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/container"
	commandUtils "github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/utils"
	artifactoryUtils "github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	containerutils "github.com/jfrog/jfrog-cli-core/v2/artifactory/utils/container"
	buildUtils "github.com/jfrog/jfrog-cli-core/v2/common/build"
	"github.com/jfrog/jfrog-cli-core/v2/common/cliutils"
	commonCliUtils "github.com/jfrog/jfrog-cli-core/v2/common/cliutils"
	commonCommands "github.com/jfrog/jfrog-cli-core/v2/common/commands"
	"github.com/jfrog/jfrog-cli-core/v2/common/project"
	speccore "github.com/jfrog/jfrog-cli-core/v2/common/spec"
	coreConfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-cli/utils/summary"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/content"
	"github.com/urfave/cli"
)

// Error modes (how should the application behave when the CheckError function is invoked):
type OnError string

type githubResponse struct {
	TagName string `json:"tag_name,omitempty"`
}

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
	Source string `json:"source,omitempty"`
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

func PrintBriefSummaryReport(success, failed int, failNoOp bool, originalErr error) error {
	basicSummary, mErr := CreateSummaryReportString(success, failed, failNoOp, originalErr)
	if mErr == nil {
		log.Output(basicSummary)
	}
	return summaryPrintError(mErr, originalErr)
}

// Print a file tree based on the items' path in the reader's list.
func PrintDeploymentView(reader *content.ContentReader) error {
	tree := artifactoryUtils.NewFileTree()
	for transferDetails := new(clientutils.FileTransferDetails); reader.NextRecord(transferDetails) == nil; transferDetails = new(clientutils.FileTransferDetails) {
		tree.AddFile(transferDetails.TargetPath)
	}
	if err := reader.GetError(); err != nil {
		return err
	}
	reader.Reset()
	output := tree.String()
	if len(output) > 0 {
		log.Info("These files were uploaded:\n\n" + output)
	}
	return nil
}

// Prints a summary report.
// If a resultReader is provided, we will iterate over the result and print a detailed summary including the affected files.
func PrintDetailedSummaryReport(basicSummary string, reader *content.ContentReader, uploaded bool, originalErr error) error {
	// A reader wasn't provided, prints the basic summary json and return.
	if reader == nil {
		log.Output(basicSummary)
		return nil
	}
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
		log.Output("  \"files\": []")
	} else {
		for transferDetails := new(clientutils.FileTransferDetails); reader.NextRecord(transferDetails) == nil; transferDetails = new(clientutils.FileTransferDetails) {
			writer.Write(getDetailedSummaryRecord(transferDetails, uploaded))
		}
		reader.Reset()
	}
	mErr = writer.Close()
	if mErr != nil {
		return summaryPrintError(mErr, originalErr)
	}
	rErr := reader.GetError()
	if rErr != nil {
		return summaryPrintError(rErr, originalErr)
	}
	return summaryPrintError(reader.GetError(), originalErr)
}

// Get the detailed summary record.
// For uploads, we need to print the sha256 of the uploaded file along with the source and target, and prefix the target with the Artifactory URL.
func getDetailedSummaryRecord(transferDetails *clientutils.FileTransferDetails, uploaded bool) interface{} {
	record := DetailedSummaryRecord{
		Source: transferDetails.SourcePath,
		Target: transferDetails.TargetPath,
	}
	if uploaded {
		record.Target = transferDetails.RtUrl + record.Target
		extendedRecord := ExtendedDetailedSummaryRecord{
			DetailedSummaryRecord: record,
			Sha256:                transferDetails.Sha256,
		}
		return extendedRecord
	}
	record.Source = transferDetails.RtUrl + record.Source
	return record
}

func PrintBuildInfoSummaryReport(succeeded bool, sha256 string, originalErr error) error {
	success, failed := 1, 0
	if !succeeded {
		success, failed = 0, 1
	}
	buildInfoSummary, mErr := CreateBuildInfoSummaryReportString(success, failed, sha256, originalErr)
	if mErr != nil {
		return summaryPrintError(mErr, originalErr)
	}
	log.Output(buildInfoSummary)
	return summaryPrintError(mErr, originalErr)
}

func PrintCommandSummary(result *commandUtils.Result, detailedSummary, printDeploymentView, failNoOp bool, originalErr error) (err error) {
	// We would like to print a basic summary of total failures/successes in the case of an error.
	err = originalErr
	if result == nil {
		// We don't have a total of failures/successes artifacts, so we are done.
		return
	}
	defer func() {
		err = GetCliError(err, result.SuccessCount(), result.FailCount(), failNoOp)
	}()
	basicSummary, err := CreateSummaryReportString(result.SuccessCount(), result.FailCount(), failNoOp, err)
	if err != nil {
		// Print the basic summary and return the original error.
		log.Output(basicSummary)
		return
	}
	if detailedSummary {
		err = PrintDetailedSummaryReport(basicSummary, result.Reader(), true, err)
	} else {
		if printDeploymentView {
			err = PrintDeploymentView(result.Reader())
		}
		log.Output(basicSummary)
	}
	return
}

func CreateSummaryReportString(success, failed int, failNoOp bool, err error) (string, error) {
	summaryReport := summary.GetSummaryReport(success, failed, failNoOp, err)
	summaryContent, mErr := summaryReport.Marshal()
	if errorutils.CheckError(mErr) != nil {
		// Don't swallow the original error. Log the marshal error and return the original error.
		return "", summaryPrintError(mErr, err)
	}
	return clientutils.IndentJson(summaryContent), err
}

func CreateBuildInfoSummaryReportString(success, failed int, sha256 string, err error) (string, error) {
	buildInfoSummary := summary.NewBuildInfoSummary(success, failed, sha256, err)
	buildInfoSummaryContent, mErr := buildInfoSummary.Marshal()
	if errorutils.CheckError(mErr) != nil {
		return "", mErr
	}
	return clientutils.IndentJson(buildInfoSummaryContent), mErr
}

func PrintHelpAndReturnError(msg string, context *cli.Context) error {
	return commonCliUtils.PrintHelpAndReturnError(msg, GetPrintCurrentCmdHelp(context))
}

func WrongNumberOfArgumentsHandler(context *cli.Context) error {
	return commonCliUtils.WrongNumberOfArgumentsHandler(context.NArg(), GetPrintCurrentCmdHelp(context))
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
	return "You can read the documentation at " + coreutils.JFrogHelpUrl + "jfrog-cli"
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

func CreateServerDetailsFromFlags(c *cli.Context) (details *coreConfig.ServerDetails, err error) {
	details = new(coreConfig.ServerDetails)
	details.Url = clientutils.AddTrailingSlashIfNeeded(c.String(url))
	details.ArtifactoryUrl = clientutils.AddTrailingSlashIfNeeded(c.String(configRtUrl))
	details.DistributionUrl = clientutils.AddTrailingSlashIfNeeded(c.String(configDistUrl))
	details.XrayUrl = clientutils.AddTrailingSlashIfNeeded(c.String(configXrUrl))
	details.MissionControlUrl = clientutils.AddTrailingSlashIfNeeded(c.String(configMcUrl))
	details.PipelinesUrl = clientutils.AddTrailingSlashIfNeeded(c.String(configPlUrl))
	details.User = c.String(user)
	details.Password, err = handleSecretInput(c, password, passwordStdin)
	if err != nil {
		return
	}
	details.AccessToken, err = handleSecretInput(c, accessToken, accessTokenStdin)
	if err != nil {
		return
	}
	details.SshKeyPath = c.String(sshKeyPath)
	details.SshPassphrase = c.String(sshPassphrase)
	details.ClientCertPath = c.String(ClientCertPath)
	details.ClientCertKeyPath = c.String(ClientCertKeyPath)
	details.ServerId = c.String(serverId)
	if details.ServerId == "" {
		details.ServerId = os.Getenv(coreutils.ServerID)
	}
	details.InsecureTls = c.Bool(InsecureTls)
	return
}

func handleSecretInput(c *cli.Context, stringFlag, stdinFlag string) (secret string, err error) {
	return commonCliUtils.HandleSecretInput(stringFlag, c.String(stringFlag), stdinFlag, c.Bool(stdinFlag))
}

func GetSpec(c *cli.Context, isDownload bool) (specFiles *speccore.SpecFiles, err error) {
	specFiles, err = speccore.CreateSpecFromFile(c.String("spec"), coreutils.SpecVarsStringToMap(c.String("spec-vars")))
	if err != nil {
		return nil, err
	}
	// Override spec with CLI options
	for i := 0; i < len(specFiles.Files); i++ {
		if isDownload {
			specFiles.Get(i).Pattern = strings.TrimPrefix(specFiles.Get(i).Pattern, "/")
		}
		OverrideFieldsIfSet(specFiles.Get(i), c)
	}
	return
}

func GetFileSystemSpec(c *cli.Context) (fsSpec *speccore.SpecFiles, err error) {
	fsSpec, err = speccore.CreateSpecFromFile(c.String("spec"), coreutils.SpecVarsStringToMap(c.String("spec-vars")))
	if err != nil {
		return
	}
	// Override spec with CLI options
	for i := 0; i < len(fsSpec.Files); i++ {
		fsSpec.Get(i).Target = strings.TrimPrefix(fsSpec.Get(i).Target, "/")
		OverrideFieldsIfSet(fsSpec.Get(i), c)
	}
	return
}

// If `fieldName` exist in the cli args, read it to `field` as a string.
func overrideStringIfSet(field *string, c *cli.Context, fieldName string) {
	if c.IsSet(fieldName) {
		*field = c.String(fieldName)
	}
}

// If `fieldName` exist in the cli args, read it to `field` as an array split by `;`.
func overrideArrayIfSet(field *[]string, c *cli.Context, fieldName string) {
	if c.IsSet(fieldName) {
		*field = append([]string{}, strings.Split(c.String(fieldName), ";")...)
	}
}

// If `fieldName` exist in the cli args, read it to `field` as a int.
func overrideIntIfSet(field *int, c *cli.Context, fieldName string) {
	if c.IsSet(fieldName) {
		*field = c.Int(fieldName)
	}
}

// Exclude refreshable tokens parameter should be true when working with external tools (build tools, curl, etc)
// or when sending requests not via ArtifactoryHttpClient.
func CreateServerDetailsWithConfigOffer(c *cli.Context, excludeRefreshableTokens bool, domain cliutils.CommandDomain) (*coreConfig.ServerDetails, error) {
	return cliutils.CreateServerDetailsWithConfigOffer(func() (*coreConfig.ServerDetails, error) { return createServerDetailsFromFlags(c, domain) }, excludeRefreshableTokens)
}

func createServerDetailsFromFlags(c *cli.Context, domain cliutils.CommandDomain) (details *coreConfig.ServerDetails, err error) {
	details, err = CreateServerDetailsFromFlags(c)
	if err != nil {
		return
	}
	switch domain {
	case cliutils.Rt:
		details.ArtifactoryUrl = details.Url
	case cliutils.Xr:
		details.XrayUrl = details.Url
	case cliutils.Ds:
		details.DistributionUrl = details.Url
	case cliutils.Platform:
		return
	}
	details.Url = ""

	return
}

func GetPrintCurrentCmdHelp(c *cli.Context) func() error {
	return func() error {
		return cli.ShowCommandHelp(c, c.Command.Name)
	}
}

// This function checks whether the command received --help as a single option.
// If it did, the command's help is shown and true is returned.
// This function should be used iff the SkipFlagParsing option is used.
func ShowCmdHelpIfNeeded(c *cli.Context, args []string) (bool, error) {
	return commonCliUtils.ShowCmdHelpIfNeeded(args, GetPrintCurrentCmdHelp(c))
}

// This function checks whether the command received --help as a single option.
// This function should be used iff the SkipFlagParsing option is used.
// Generic commands such as docker, don't have dedicated subcommands. As a workaround, printing the help of their subcommands,
// we use a dummy command with no logic but the help message. to trigger the print of those dummy commands,
// each generic command must decide what cmdName it needs to pass to this function.
// For example, 'jf docker scan --help' passes cmdName='dockerscanhelp' to print our help and not the origin from docker client/cli.
func ShowGenericCmdHelpIfNeeded(c *cli.Context, args []string, cmdName string) (bool, error) {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			err := cli.ShowCommandHelp(c, cmdName)
			return true, err
		}
	}
	return false, nil
}

func OverrideFieldsIfSet(spec *speccore.File, c *cli.Context) {
	overrideArrayIfSet(&spec.Exclusions, c, "exclusions")
	overrideArrayIfSet(&spec.SortBy, c, "sort-by")
	overrideIntIfSet(&spec.Offset, c, "offset")
	overrideIntIfSet(&spec.Limit, c, "limit")
	overrideStringIfSet(&spec.SortOrder, c, "sort-order")
	overrideStringIfSet(&spec.Props, c, "props")
	overrideStringIfSet(&spec.TargetProps, c, "target-props")
	overrideStringIfSet(&spec.ExcludeProps, c, "exclude-props")
	overrideStringIfSet(&spec.Build, c, "build")
	overrideStringIfSet(&spec.Project, c, "project")
	overrideStringIfSet(&spec.ExcludeArtifacts, c, "exclude-artifacts")
	overrideStringIfSet(&spec.IncludeDeps, c, "include-deps")
	overrideStringIfSet(&spec.Bundle, c, "bundle")
	overrideStringIfSet(&spec.Recursive, c, "recursive")
	overrideStringIfSet(&spec.Flat, c, "flat")
	overrideStringIfSet(&spec.Explode, c, "explode")
	overrideStringIfSet(&spec.BypassArchiveInspection, c, "bypass-archive-inspection")
	overrideStringIfSet(&spec.Regexp, c, "regexp")
	overrideStringIfSet(&spec.IncludeDirs, c, "include-dirs")
	overrideStringIfSet(&spec.ValidateSymlinks, c, "validate-symlinks")
	overrideStringIfSet(&spec.Symlinks, c, "symlinks")
	overrideStringIfSet(&spec.Transitive, c, "transitive")
	overrideStringIfSet(&spec.PublicGpgKey, c, "gpg-key")
}

func FixWinPathsForFileSystemSourcedCmds(uploadSpec *speccore.SpecFiles, c *cli.Context) {
	commonCliUtils.FixWinPathsForFileSystemSourcedCmds(uploadSpec, c.IsSet("spec"), c.IsSet("exclusions"))
}

func CreateConfigCmd(c *cli.Context, confType project.ProjectType) error {
	if c.NArg() != 0 {
		return WrongNumberOfArgumentsHandler(c)
	}
	return commonCommands.CreateBuildConfig(c, confType)
}

func RunNativeCmdWithDeprecationWarning(cmdName string, projectType project.ProjectType, c *cli.Context, cmd func(c *cli.Context) error) error {
	if cliutils.ShouldLogWarning() {
		LogNativeCommandDeprecation(cmdName, projectType.String())
	}
	return cmd(c)
}

func ShowDockerDeprecationMessageIfNeeded(containerManagerType containerutils.ContainerManagerType, isGetRepoSupported func() (bool, error)) error {
	if containerManagerType == containerutils.DockerClient {
		// Show a deprecation message for this command, if Artifactory supports fetching the physical docker repository name.
		supported, err := isGetRepoSupported()
		if err != nil {
			return err
		}
		if supported {
			LogNativeCommandDeprecation("docker", "Docker")
		}
	}
	return nil
}

func LogNativeCommandDeprecation(cmdName, projectType string) {
	log.Warn(
		`You are using a deprecated syntax of the command.
	The new command syntax is quite similar to the syntax used by the native ` + projectType + ` client.
	All you need to do is to add '` + coreutils.GetCliExecutableName() + `' as a prefix to the command.
	For example:
	$ ` + coreutils.GetCliExecutableName() + ` ` + cmdName + ` ...
	The --build-name and --build-number options are still supported.`)
}

func NotSupportedNativeDockerCommand(oldCmdName string) error {
	return errorutils.CheckErrorf(
		`This command requires Artifactory version %s or higher.
		 With your current Artifactory version, you can use the old and deprecated command instead:
		 %s rt %s <image> <repository name>`, corecontainercmds.MinRtVersionForRepoFetching, coreutils.GetCliExecutableName(), oldCmdName)
}

func RunConfigCmdWithDeprecationWarning(cmdName, oldSubcommand string, confType project.ProjectType, c *cli.Context,
	cmd func(c *cli.Context, confType project.ProjectType) error) error {
	commonCliUtils.LogNonNativeCommandDeprecation(cmdName, oldSubcommand)
	return cmd(c, confType)
}

func RunCmdWithDeprecationWarning(cmdName, oldSubcommand string, c *cli.Context,
	cmd func(c *cli.Context) error) error {
	commonCliUtils.LogNonNativeCommandDeprecation(cmdName, oldSubcommand)
	return cmd(c)
}

func SetCliExecutableName(executablePath string) {
	coreutils.SetCliExecutableName(filepath.Base(executablePath))
}

// Returns build configuration struct using the args (build name/number) and options (project) provided by the user.
// Any empty configuration could be later overridden by environment variables if set.
func CreateBuildConfiguration(c *cli.Context) *buildUtils.BuildConfiguration {
	buildConfiguration := new(buildUtils.BuildConfiguration)
	buildNameArg, buildNumberArg := c.Args().Get(0), c.Args().Get(1)
	if buildNameArg == "" || buildNumberArg == "" {
		buildNameArg = ""
		buildNumberArg = ""
	}
	buildConfiguration.SetBuildName(buildNameArg).SetBuildNumber(buildNumberArg).SetProject(c.String("project")).SetModule(c.String("module"))
	return buildConfiguration
}

// Returns build configuration struct using the options provided by the user.
// Any empty configuration could be later overridden by environment variables if set.
func CreateBuildConfigurationWithModule(c *cli.Context) (buildConfigConfiguration *buildUtils.BuildConfiguration, err error) {
	buildConfigConfiguration = new(buildUtils.BuildConfiguration)
	err = buildConfigConfiguration.SetBuildName(c.String("build-name")).SetBuildNumber(c.String("build-number")).
		SetProject(c.String("project")).SetModule(c.String("module")).ValidateBuildAndModuleParams()
	return
}

func CreateArtifactoryDetailsByFlags(c *cli.Context) (*coreConfig.ServerDetails, error) {
	artDetails, err := CreateServerDetailsWithConfigOffer(c, false, cliutils.Rt)
	if err != nil {
		return nil, err
	}
	if artDetails.ArtifactoryUrl == "" {
		return nil, errors.New("the --url option is mandatory")
	}
	return artDetails, nil
}

func IsFailNoOp(context *cli.Context) bool {
	if isContextFailNoOp(context) {
		return true
	}
	return isEnvFailNoOp()
}

func isContextFailNoOp(context *cli.Context) bool {
	if context == nil {
		return false
	}
	return context.Bool("fail-no-op")
}

func isEnvFailNoOp() bool {
	return strings.ToLower(os.Getenv(coreutils.FailNoOp)) == "true"
}

func CleanupResult(result *commandUtils.Result, err *error) {
	if result != nil && result.Reader() != nil {
		*err = errors.Join(*err, result.Reader().Close())
	}
}

// Checks if the requested plugin exists in registry and does not exist locally.
func CheckNewCliVersionAvailable(currentVersion string) (warningMessage string, err error) {
	shouldCheck, err := shouldCheckLatestCliVersion()
	if err != nil || !shouldCheck {
		return
	}
	githubVersionInfo, err := getLatestCliVersionFromGithubAPI()
	if err != nil {
		return
	}
	latestVersion := strings.TrimPrefix(githubVersionInfo.TagName, "v")
	if version.NewVersion(latestVersion).Compare(currentVersion) < 0 {
		warningMessage = strings.Join([]string{
			coreutils.PrintComment(
				fmt.Sprintf("You are using JFrog CLI version %s, however version ", currentVersion)) +
				coreutils.PrintTitle(latestVersion) +
				coreutils.PrintComment(" is available."),
			coreutils.PrintComment("To install the latest version, visit: ") + coreutils.PrintLink(coreutils.JFrogComUrl+"getcli"),
			coreutils.PrintComment("To see the release notes, visit: ") + coreutils.PrintLink("https://github.com/jfrog/jfrog-cli/releases"),
			coreutils.PrintComment(fmt.Sprintf("To avoid this message, set the %s variable to TRUE", JfrogCliAvoidNewVersionWarning)),
		},
			"\n")
	}
	return
}

func shouldCheckLatestCliVersion() (shouldCheck bool, err error) {
	if strings.ToLower(os.Getenv(JfrogCliAvoidNewVersionWarning)) == "true" {
		return
	}
	homeDir, err := coreutils.GetJfrogHomeDir()
	if err != nil {
		return
	}
	indicatorFile := path.Join(homeDir, "Latest_Cli_Version_Check_Indicator")
	fileInfo, err := os.Stat(indicatorFile)
	if err != nil && !os.IsNotExist(err) {
		err = fmt.Errorf("couldn't get indicator file %s info: %s", indicatorFile, err.Error())
		return
	}
	if err == nil && (time.Now().UnixMilli()-fileInfo.ModTime().UnixMilli()) < LatestCliVersionCheckInterval.Milliseconds() {
		// Timestamp file exists and updated less than 6 hours ago, therefor no need to check version again
		return
	}
	return true, os.WriteFile(indicatorFile, []byte{}, 0666)
}

func getLatestCliVersionFromGithubAPI() (githubVersionInfo githubResponse, err error) {
	client := &http.Client{Timeout: time.Second * 2}
	req, err := http.NewRequest(http.MethodGet, "https://api.github.com/repos/jfrog/jfrog-cli/releases/latest", nil)
	if errorutils.CheckError(err) != nil {
		return
	}
	resp, body, err := doHttpRequest(client, req)
	if err != nil {
		err = errors.New("couldn't get latest JFrog CLI latest version info from GitHub API: " + err.Error())
		return
	}
	err = errorutils.CheckResponseStatusWithBody(resp, body, http.StatusOK)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &githubVersionInfo)
	return
}

func doHttpRequest(client *http.Client, req *http.Request) (resp *http.Response, body []byte, err error) {
	req.Close = true
	resp, err = client.Do(req)
	if errorutils.CheckError(err) != nil {
		return
	}
	defer func() {
		if resp != nil && resp.Body != nil {
			e := errorutils.CheckError(resp.Body.Close())
			err = errors.Join(err, e)
		}
	}()
	body, err = io.ReadAll(resp.Body)
	return resp, body, errorutils.CheckError(err)
}

// Get project key from flag or environment variable
func GetProject(c *cli.Context) string {
	projectKey := c.String("project")
	return getOrDefaultEnv(projectKey, coreutils.Project)
}
