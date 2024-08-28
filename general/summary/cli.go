package summary

import (
	"errors"
	"fmt"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils/commandsummary"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"os"
	"path/filepath"
	"strings"

	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	commonCliUtils "github.com/jfrog/jfrog-cli-core/v2/common/cliutils"
	coreConfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	securityUtils "github.com/jfrog/jfrog-cli-security/utils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/urfave/cli"
)

type MarkdownSection string

const (
	Security  MarkdownSection = "security"
	BuildInfo MarkdownSection = "build-info"
	Upload    MarkdownSection = "upload"
)

const (
	JfrogCliSummaryDir = "jfrog-command-summary"
	MarkdownFileName   = "markdown.md"
)

var markdownSections = []MarkdownSection{Security, BuildInfo, Upload}

func (ms MarkdownSection) String() string {
	return string(ms)
}

// Creates a summary of recorded CLI commands that were executed on the current machine.
// The summary is generated in Markdown format
// and saved in the directory stored in the JFROG_CLI_COMMAND_SUMMARY_OUTPUT_DIR environment variable.
func GenerateSummaryMarkdown(c *cli.Context) error {
	if !ShouldGenerateSummary() {
		return fmt.Errorf("unable to generate the command summary because the output directory is not specified."+
			" Please ensure that the environment variable '%s' is set before running your commands to enable summary generation", coreutils.SummaryOutputDirPathEnv)
	}
	// Get URL and Version to generate summary links
	serverUrl, majorVersion, err := extractServerUrlAndVersion(c)
	if err != nil {
		return fmt.Errorf("failed to get server URL or major version: %v. This means markdown URLs will be invalid", err)
	}

	if err = commandsummary.InitMarkdownGenerationValues(serverUrl, majorVersion); err != nil {
		return fmt.Errorf("failed to initialize command summary values: %w", err)
	}

	// Invoke each section's markdown generation function
	for _, section := range markdownSections {
		if err := invokeSectionMarkdownGeneration(section); err != nil {
			log.Warn("Failed to generate markdown for section %s: %v", section, err)
		}
	}

	// Combine all sections into a single Markdown file
	finalMarkdown, err := combineMarkdownFiles()
	if err != nil {
		return fmt.Errorf("error combining markdown files: %w", err)
	}

	return saveMarkdownToFileSystem(finalMarkdown)
}

func combineMarkdownFiles() (string, error) {
	var combinedMarkdown strings.Builder
	// Read each section content and append it to the final Markdown
	for _, section := range markdownSections {
		sectionContent, err := getSectionMarkdownContent(section)
		if err != nil {
			return "", fmt.Errorf("error getting markdown content for section %s: %w", section, err)
		}
		if _, err := combinedMarkdown.WriteString(sectionContent); err != nil {
			return "", fmt.Errorf("error writing markdown content for section %s: %w", section, err)
		}
	}
	return combinedMarkdown.String(), nil
}

// Saves markdown content in the directory stored in the JFROG_CLI_COMMAND_SUMMARY_OUTPUT_DIR environment variable.
func saveMarkdownToFileSystem(finalMarkdown string) (err error) {
	if finalMarkdown == "" {
		return nil
	}
	filePath := filepath.Join(os.Getenv(coreutils.SummaryOutputDirPathEnv), JfrogCliSummaryDir, MarkdownFileName)
	file, err := os.Create(filePath)
	defer func() {
		err = file.Close()
	}()
	if err != nil {
		return fmt.Errorf("error creating markdown file: %w", err)
	}
	// Write to file
	if _, err := file.WriteString(finalMarkdown); err != nil {
		return fmt.Errorf("error writing to markdown file: %w", err)
	}
	return
}

func getSectionMarkdownContent(section MarkdownSection) (string, error) {
	sectionFilepath := filepath.Join(os.Getenv(coreutils.SummaryOutputDirPathEnv), JfrogCliSummaryDir, string(section), MarkdownFileName)
	if _, err := os.Stat(sectionFilepath); os.IsNotExist(err) {
		return "", nil
	}

	contentBytes, err := os.ReadFile(sectionFilepath)
	if err != nil {
		return "", fmt.Errorf("error reading markdown file for section %s: %w", section, err)
	}
	if len(contentBytes) == 0 {
		return "", nil
	}
	return string(contentBytes), nil
}

func invokeSectionMarkdownGeneration(section MarkdownSection) error {
	switch section {
	case Security:
		return generateSecurityMarkdown()
	case BuildInfo:
		return generateBuildInfoMarkdown()
	case Upload:
		return generateUploadMarkdown()
	default:
		return fmt.Errorf("unknown section: %s", section)
	}
}

func generateSecurityMarkdown() error {
	securitySummary, err := securityUtils.NewSecurityJobSummary()
	if err != nil {
		return fmt.Errorf("error generating security markdown: %w", err)
	}
	return securitySummary.GenerateMarkdown()
}

func generateBuildInfoMarkdown() error {
	buildInfoSummary, err := commandsummary.NewBuildInfoSummary()
	if err != nil {
		return fmt.Errorf("error generating build-info markdown: %w", err)
	}
	if err = mapScanResults(buildInfoSummary); err != nil {
		return fmt.Errorf("error mapping scan results: %w", err)
	}
	return buildInfoSummary.GenerateMarkdown()
}

// Maps the scan results saved during runtime into scan components
// which can be used to extract Markdown content
func mapScanResults(buildInfoSummary *commandsummary.CommandSummary) error {
	// Gets the saved scan results file paths.
	indexedFiles, err := buildInfoSummary.GetIndexedDataFilesPaths()
	if err != nil {
		return err
	}
	myMappedResults := make(map[string]commandsummary.ScanResult, 1)
	sec := &securityUtils.SecurityJobSummary{}
	for index, keyValue := range indexedFiles {
		for scannedName, filePath := range keyValue {
			myMappedResults = processScan(index, filePath, scannedName, sec, myMappedResults)
		}
	}
	myMappedResults[commandsummary.NonScannedResult], _ = sec.GetNonScannedResult()
	commandsummary.ScanResultsMapping = myMappedResults
	return nil
}

func processScan(index commandsummary.Index, filePath string, scannedName string, sec *securityUtils.SecurityJobSummary, myMappedResults map[string]commandsummary.ScanResult) map[string]commandsummary.ScanResult {
	var res commandsummary.ScanResult
	var err error
	switch index {
	case commandsummary.DockerScan:
		res, err = sec.DockerScan([]string{filePath})
	case commandsummary.BuildScan:
		res, err = sec.BuildScan([]string{filePath})
	case commandsummary.BinariesScan:
		res, err = sec.BinaryScan([]string{filePath})
	}

	myMappedResults[scannedName] = res
	if err != nil {
		log.Warn("Failed to generate scan result for %s: %v", scannedName, err)
	}
	return myMappedResults
}

func generateUploadMarkdown() error {
	if should, err := shouldGenerateUploadSummary(); err != nil || !should {
		log.Debug("Skipping upload summary generation due build-info data to avoid duplications...")
		return err
	}
	uploadSummary, err := commandsummary.NewUploadSummary()
	if err != nil {
		return fmt.Errorf("error generating upload markdown: %w", err)
	}
	return uploadSummary.GenerateMarkdown()
}

// Upload summary should be generated only if the no build-info data exists
func shouldGenerateUploadSummary() (bool, error) {
	buildInfoPath := filepath.Join(os.Getenv(coreutils.SummaryOutputDirPathEnv), JfrogCliSummaryDir, string(BuildInfo))
	if _, err := os.Stat(buildInfoPath); os.IsNotExist(err) {
		return true, nil
	}
	dirEntries, err := os.ReadDir(buildInfoPath)
	if err != nil {
		return false, fmt.Errorf("error reading directory: %w", err)
	}
	return len(dirEntries) == 0, nil
}

func createPlatformDetailsByFlags(c *cli.Context) (*coreConfig.ServerDetails, error) {
	platformDetails, err := cliutils.CreateServerDetailsWithConfigOffer(c, true, commonCliUtils.Platform)
	if err != nil {
		return nil, fmt.Errorf("error creating platform details: %w", err)
	}
	if platformDetails.Url == "" {
		return nil, errors.New("platform URL is mandatory for access token creation")
	}
	return platformDetails, nil
}

func extractServerUrlAndVersion(c *cli.Context) (platformUrl string, platformMajorVersion int, err error) {
	serverDetails, err := createPlatformDetailsByFlags(c)
	if err != nil {
		return "", 0, fmt.Errorf("error extracting server details: %w", err)
	}
	platformUrl = serverDetails.Url

	servicesManager, err := utils.CreateServiceManager(serverDetails, -1, 0, false)
	if err != nil {
		return "", 0, fmt.Errorf("error creating services manager: %w", err)
	}
	if platformMajorVersion, err = utils.GetRtMajorVersion(servicesManager); err != nil {
		return "", 0, fmt.Errorf("error getting Artifactory major platformMajorVersion: %w", err)
	}
	return
}

// Summary should be generated only when the output directory is defined
func ShouldGenerateSummary() bool {
	return os.Getenv(coreutils.SummaryOutputDirPathEnv) != ""
}
