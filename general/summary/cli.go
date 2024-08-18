package summary

import (
	"errors"
	"fmt"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"os"
	"path/filepath"
	"strings"

	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/commandssummaries"
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

func GenerateSummaryMarkdown(c *cli.Context) error {
	if !ShouldRecordSummary() {
		return fmt.Errorf("cannot generate command summary: the output directory for command recording is not defined. "+
			"Please set the environment variable %s before executing your commands to view their summary", coreutils.OutputDirPathEnv)
	}
	// Get URL and Version to generate summary links
	serverUrl, majorVersion, err := extractServerUrlAndVersion(c)
	if err != nil {
		log.Warn("Failed to get server URL or major version: %v. This means markdown URLs will be invalid!", err)
	}
	// Load each section content
	for _, section := range markdownSections {
		if err := generateSectionMarkdown(section, serverUrl, majorVersion); err != nil {
			log.Warn("Failed to generate markdown for section %s: %v", section, err)
		}
	}
	// Combine all sections into a single Markdown file
	if err := combineMarkdownFiles(); err != nil {
		return fmt.Errorf("failed to combine markdown files: %w", err)
	}
	return nil
}

func combineMarkdownFiles() error {
	var finalMarkdown strings.Builder

	for _, section := range markdownSections {
		sectionContent, err := getSectionMarkdownContent(section)
		if err != nil {
			return fmt.Errorf("error getting markdown content for section %s: %w", section, err)
		}
		if _, err := finalMarkdown.WriteString(sectionContent); err != nil {
			return fmt.Errorf("error writing markdown content for section %s: %w", section, err)
		}
	}

	if finalMarkdown.Len() == 0 {
		log.Debug("No markdown content found")
		return nil
	}

	basePath := filepath.Join(os.Getenv(coreutils.OutputDirPathEnv), JfrogCliSummaryDir)
	filePath := filepath.Join(basePath, MarkdownFileName)

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating markdown file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			err = fmt.Errorf("error closing file: %w", closeErr)
		}
	}()

	if _, err := file.WriteString(finalMarkdown.String()); err != nil {
		return fmt.Errorf("error writing to markdown file: %w", err)
	}
	return nil
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

func generateSectionMarkdown(section MarkdownSection, serverUrl string, majorVersion int) error {
	switch section {
	case Security:
		return generateSecurityMarkdown()
	case BuildInfo:
		return generateBuildInfoMarkdown(serverUrl, majorVersion)
	case Upload:
		return generateUploadMarkdown(serverUrl, majorVersion)
	default:
		return fmt.Errorf("unknown section: %s", section)
	}
}

func generateSecurityMarkdown() error {
	securitySummary, err := securityUtils.SecurityCommandsJobSummary()
	if err != nil {
		return fmt.Errorf("error generating security markdown: %w", err)
	}
	return securitySummary.GenerateMarkdown()
}

func generateBuildInfoMarkdown(serverUrl string, majorVersion int) error {
	buildInfoSummary, err := commandssummaries.NewBuildInfoSummary(serverUrl, majorVersion)
	if err != nil {
		return fmt.Errorf("error generating build-info markdown: %w", err)
	}
	return buildInfoSummary.GenerateMarkdown()
}

func generateUploadMarkdown(serverUrl string, majorVersion int) error {
	if should, err := shouldGenerateUploadSummary(); err != nil || !should {
		return err
	}
	uploadSummary, err := commandssummaries.NewUploadSummary(serverUrl, majorVersion)
	if err != nil {
		return fmt.Errorf("error generating upload markdown: %w", err)
	}
	return uploadSummary.GenerateMarkdown()
}

func shouldGenerateUploadSummary() (bool, error) {
	basePath := os.Getenv(coreutils.OutputDirPathEnv)
	buildInfoPath := filepath.Join(basePath, "jfrog-command-summary", string(BuildInfo))

	fileInfo, err := os.Stat(buildInfoPath)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, fmt.Errorf("error checking buildInfoPath: %w", err)
	}

	if !fileInfo.IsDir() {
		return false, nil
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

func ShouldRecordSummary() bool {
	return os.Getenv(coreutils.OutputDirPathEnv) != ""
}
