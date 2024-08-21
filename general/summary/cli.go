package summary

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils/commandsummary"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"net/http"
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

// Creates a final summary of recorded CLI commands that were executed on the current machine.
// The summary is generated in Markdown format and saved in the root directory of JFROG_CLI_COMMAND_SUMMARY_OUTPUT_DIR.
func GenerateSummaryMarkdown(c *cli.Context) error {
	if !ShouldGenerateSummary() {
		return fmt.Errorf("cannot generate command summary: the output directory for command recording is not defined. "+
			"Please set the environment variable %s before executing your commands to view their summary", coreutils.OutputDirPathEnv)
	}
	// Get URL and Version to generate summary links
	serverUrl, majorVersion, err := extractServerUrlAndVersion(c)
	if err != nil {
		log.Warn("Failed to get server URL or major version: %v. This means markdown URLs will be invalid!", err)
	}

	// Check full summary entitlement
	extendedSummary, err := isEntitledForExtendedSummary(serverUrl)
	if err != nil {
		return fmt.Errorf("error checking extended summary entitlement: %w", err)
	}
	// Invoke each section's markdown generation function
	for _, section := range markdownSections {
		if err := invokeSectionMarkdownGeneration(section, serverUrl, majorVersion, extendedSummary); err != nil {
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

func combineMarkdownFiles() (content string, err error) {
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

	return combinedMarkdown.String(), err
}

// Save the final Markdown to a file in the root directory of the COMMAND_SUMMARY_DIR/markdown.md
func saveMarkdownToFileSystem(finalMarkdown string) (err error) {
	if finalMarkdown == "" {
		return nil
	}
	filePath := filepath.Join(os.Getenv(coreutils.OutputDirPathEnv), JfrogCliSummaryDir, MarkdownFileName)
	// Creates the file
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

func wrapCollapsibleSection(section MarkdownSection, markdown string) (string, error) {
	sectionTitle, err := getSectionTitle(section)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("\n\n\n<details open>\n\n<summary>  %s </summary><p></p> \n\n %s \n\n</details>\n\n\n", sectionTitle, markdown), nil
}

func getSectionTitle(section MarkdownSection) (string, error) {
	switch section {
	case Upload:
		return "📁 Files uploaded to Artifactory by this workflow", nil
	case BuildInfo:
		return "📦 Artifacts published to Artifactory by this workflow", nil
	case Security:
		return "🔒 Security Summary", nil
	default:
		return "", fmt.Errorf("unknown section: %s", section)
	}
}

func getSectionMarkdownContent(section MarkdownSection) (string, error) {
	basePath := os.Getenv(coreutils.OutputDirPathEnv)
	sectionFilepath := filepath.Join(basePath, JfrogCliSummaryDir, string(section), MarkdownFileName)

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

	content, err := wrapCollapsibleSection(section, string(contentBytes))
	if err != nil {
		return "", fmt.Errorf("error wrapping section %s: %w", section, err)
	}
	return content, nil
}

func invokeSectionMarkdownGeneration(section MarkdownSection, serverUrl string, majorVersion int, extendedSummary bool) error {
	switch section {
	case Security:
		return generateSecurityMarkdown(extendedSummary)
	case BuildInfo:
		return generateBuildInfoMarkdown(serverUrl, majorVersion, extendedSummary)
	case Upload:
		return generateUploadMarkdown(serverUrl, majorVersion, extendedSummary)
	default:
		return fmt.Errorf("unknown section: %s", section)
	}
}

func generateSecurityMarkdown(extendedSummary bool) error {
	securitySummary, err := securityUtils.SecurityCommandsJobSummary()
	if err != nil {
		return fmt.Errorf("error generating security markdown: %w", err)
	}
	return securitySummary.GenerateMarkdown(extendedSummary)
}

func generateBuildInfoMarkdown(serverUrl string, majorVersion int, extendedSummary bool) error {
	buildInfoSummary, err := commandsummary.NewBuildInfoSummary(serverUrl, majorVersion)
	if err != nil {
		return fmt.Errorf("error generating build-info markdown: %w", err)
	}
	return buildInfoSummary.GenerateMarkdown(extendedSummary)
}

func generateUploadMarkdown(serverUrl string, majorVersion int, extendedSummary bool) error {
	if should, err := shouldGenerateUploadSummary(); err != nil || !should {
		log.Debug("Skipping upload summary generation due build-info data to avoid duplications...")
		return err
	}
	uploadSummary, err := commandsummary.NewUploadSummary(serverUrl, majorVersion)
	if err != nil {
		return fmt.Errorf("error generating upload markdown: %w", err)
	}
	return uploadSummary.GenerateMarkdown(extendedSummary)
}

// Upload summary should be generated only if the no build-info data exists
func shouldGenerateUploadSummary() (bool, error) {
	buildInfoPath := filepath.Join(os.Getenv(coreutils.OutputDirPathEnv), JfrogCliSummaryDir, string(BuildInfo))
	_, err := os.Stat(buildInfoPath)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
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

func extractServerUrlAndVersion(c *cli.Context) (string, int, error) {
	serverDetails, err := createPlatformDetailsByFlags(c)
	if err != nil {
		return "", 0, fmt.Errorf("error extracting server details: %w", err)
	}

	serverUrl := serverDetails.GetUrl()
	servicesManager, err := utils.CreateServiceManager(serverDetails, -1, 0, false)
	if err != nil {
		return "", 0, fmt.Errorf("error creating services manager: %w", err)
	}

	majorVersion, err := utils.GetRtMajorVersion(servicesManager)
	if err != nil {
		return "", 0, fmt.Errorf("error getting Artifactory major version: %w", err)
	}
	return serverUrl, majorVersion, nil
}

func isEntitledForExtendedSummary(serverUrl string) (extendedSummary bool, err error) {
	url := fmt.Sprintf("%sui/api/v1/system/auth/screen/footer", serverUrl)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error making HTTP request:", err)
		return
	}
	defer func() {
		err = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Non-OK HTTP status:", resp.StatusCode)
		return
	}

	var result struct {
		PlatformId string `json:"platformId"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Println("Error decoding JSON response:", err)
		return false, err
	}
	entitled := strings.Contains(strings.ToLower(result.PlatformId), "enterprise")
	log.Debug("Entitled for full command summary: ", entitled)
	return entitled, nil
}

// Summary should be generated only when the output directory is defined
func ShouldGenerateSummary() bool {
	return os.Getenv(coreutils.OutputDirPathEnv) != ""
}
