package summary

import (
	"errors"
	"fmt"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils/commandsummary"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
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
	markdownFileName   = "markdown.md"
	finalSarifFileName = "final.sarif"
)

var markdownSections = []MarkdownSection{Security, BuildInfo, Upload}

func (ms MarkdownSection) String() string {
	return string(ms)
}

func FinalizeCommandSummaries(c *cli.Context) error {
	if !shouldGenerateSummary() {
		return fmt.Errorf("unable to generate the command summary because the output directory is not specified."+
			" Please ensure that the environment variable '%s' is set before running your commands to enable summary generation", coreutils.SummaryOutputDirPathEnv)
	}

	if err := generateSummaryMarkdown(c); err != nil {
		return err
	}

	return aggregatedCodeScanningSarifs()
}

// generateSummaryMarkdown creates a summary of recorded CLI commands in Markdown format.
func generateSummaryMarkdown(c *cli.Context) error {
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
			log.Warn("Failed to generate markdown for section:", section, err)
		}
	}

	// Combine all sections into a single Markdown file
	finalMarkdown, err := mergeMarkdownFiles()
	if err != nil {
		return fmt.Errorf("error combining markdown files: %w", err)
	}

	// Saves the final Markdown to the root directory of the command summaries
	return saveMarkdownToFileSystem(finalMarkdown)
}

func aggregatedCodeScanningSarifs() error {
	files, err := getSarifFiles()
	if err != nil {
		return err
	}
	if len(files) == 0 {
		log.Debug("No sarif reports were found")
		return nil
	}
	finalSarif, err := securityUtils.CombineSarifOutputFiles(files)
	if err != nil {
		return err
	}
	return saveFinalSarifToFileSystem(string(finalSarif))
}

func getSarifReportsDir() string {
	return filepath.Join(os.Getenv(coreutils.SummaryOutputDirPathEnv), commandsummary.OutputDirName, string(Security), string(commandsummary.SarifReport))
}

// The CLI generates summaries in sections, with each section as a separate Markdown file.
// This function merges all sections into a single Markdown file and saves it in the root of the
// command summary output directory.
func mergeMarkdownFiles() (string, error) {
	var combinedMarkdown strings.Builder
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

// saveMarkdownToFileSystem saves markdown content in the specified directory.
func saveMarkdownToFileSystem(finalMarkdown string) (err error) {
	if finalMarkdown == "" {
		return nil
	}
	filePath := filepath.Join(os.Getenv(coreutils.SummaryOutputDirPathEnv), commandsummary.OutputDirName, markdownFileName)
	return saveFile(finalMarkdown, filePath)
}

func saveFile(content, filePath string) (err error) {
	if content == "" {
		return nil
	}
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, file.Close())
	}()
	if _, err = file.WriteString(content); err != nil {
		return err
	}
	return nil
}

func getSectionMarkdownContent(section MarkdownSection) (string, error) {
	sectionFilepath := filepath.Join(os.Getenv(coreutils.SummaryOutputDirPathEnv), commandsummary.OutputDirName, string(section), markdownFileName)
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

func getSarifFiles() (files []string, err error) {
	sarifsDir := getSarifReportsDir()
	exists, err := fileutils.IsDirExists(sarifsDir, false)
	if err != nil || !exists {
		return
	}
	entries, err := os.ReadDir(sarifsDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		files = append(files, filepath.Join(sarifsDir, entry.Name()))
	}
	return
}

func saveFinalSarifToFileSystem(finalSarif string) (err error) {
	filePath := filepath.Join(getSarifReportsDir(), finalSarifFileName)
	return saveFile(finalSarif, filePath)
}

// Initiate the desired command summary implementation and invoke its Markdown generation.
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

// mapScanResults maps the scan results saved during runtime into scan components.
func mapScanResults(commandSummary *commandsummary.CommandSummary) (err error) {
	// Gets the saved scan results file paths.
	indexedFiles, err := commandSummary.GetIndexedDataFilesPaths()
	if err != nil {
		return err
	}
	securityJobSummary := &securityUtils.SecurityJobSummary{}
	// Init scan result map
	scanResultsMap := make(map[string]commandsummary.ScanResult)
	// Set default not scanned component view
	scanResultsMap[commandsummary.NonScannedResult] = securityJobSummary.GetNonScannedResult()
	commandsummary.StaticMarkdownConfig.SetScanResultsMapping(scanResultsMap)
	// Process each scan result file by its type and append to map
	for index, keyValue := range indexedFiles {
		for scannedEntityName, scanResultDataFilePath := range keyValue {
			scanResultsMap, err = processScan(index, scanResultDataFilePath, scannedEntityName, securityJobSummary, scanResultsMap)
			if err != nil {
				return
			}
		}
	}
	return
}

// Each scan result should be processed according to its index.
// To generate custom view for each scan type.
func processScan(index commandsummary.Index, filePath string, scannedName string, sec *securityUtils.SecurityJobSummary, scanResultsMap map[string]commandsummary.ScanResult) (map[string]commandsummary.ScanResult, error) {
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
	scanResultsMap[scannedName] = res
	if err != nil {
		return nil, err
	}
	return scanResultsMap, nil
}

// shouldGenerateUploadSummary checks if upload summary should be generated.
func shouldGenerateUploadSummary() (bool, error) {
	buildInfoPath := filepath.Join(os.Getenv(coreutils.SummaryOutputDirPathEnv), commandsummary.OutputDirName, string(BuildInfo))
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

// shouldGenerateSummary checks if the summary should be generated.
func shouldGenerateSummary() bool {
	return os.Getenv(coreutils.SummaryOutputDirPathEnv) != ""
}
