package summary

import (
	"errors"
	"fmt"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/commandssummaries"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	commonCliUtils "github.com/jfrog/jfrog-cli-core/v2/common/cliutils"
	coreConfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	securityUtils "github.com/jfrog/jfrog-cli-security/utils"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/urfave/cli"
	"os"
	"path/filepath"
	"strings"
)

type MarkdownSection string

const (
	Security  MarkdownSection = "security"
	BuildInfo MarkdownSection = "build-info"
	Upload    MarkdownSection = "upload"
)

var markdownSections = []MarkdownSection{Security, BuildInfo, Upload}

func (ms MarkdownSection) String() string {
	return string(ms)
}

func GenerateSummaryMarkdown(c *cli.Context) (err error) {
	// Verify that data has been recorded.
	if os.Getenv(coreutils.OutputDirPathEnv) == "" {
		return fmt.Errorf("cannot generate command summary: the output directory for command recording is not defined."+
			" Please set the environment variable %s before executing your commands to view their summary", coreutils.OutputDirPathEnv)
	}
	// Get needed values to generate platform links.
	serverUrl, majorVersion, err := extractServerUrlAndVersion(c)
	if err != nil {
		return err
	}
	// Invoke each section's Markdown generation.
	for _, section := range markdownSections {
		if err = generateSectionMarkdown(section, serverUrl, majorVersion); err != nil {
			log.Warn("Failed to generate markdown for section %s: %v", section, err)
		}
	}
	// Combine all Markdown files into a single file.
	if err = combineMarkdownFiles(); err != nil {
		return fmt.Errorf("failed to combine markdown files: %v", err)
	}
	return
}

// Combines the Markdown content of each section into a single report.
// The combined report is saved in the root directory of the command summary.
func combineMarkdownFiles() (err error) {
	var finalMarkdown strings.Builder
	for _, section := range markdownSections {
		sectionContent, err := getSectionMarkdownContent(section)
		if err != nil {
			return err
		}
		if _, err = finalMarkdown.WriteString(sectionContent); err != nil {
			return err
		}
	}
	if finalMarkdown.Len() == 0 {
		log.Debug("no markdown content found")
		return nil
	}

	// Write the combined Markdown content to a file.
	basePath := filepath.Join(os.Getenv(coreutils.OutputDirPathEnv), "jfrog-command-summary")
	fd, err := os.Create(filepath.Join(basePath, "summary.md"))
	defer func() {
		err = fd.Close()
	}()
	_, err = fd.WriteString(finalMarkdown.String())
	if err != nil {
		return
	}
	return
}

func wrapCollapsableSection(section MarkdownSection, markdown string) (string, error) {
	var sectionTitle string
	switch section {
	case Upload:
		sectionTitle = "📁 Files uploaded to Artifactory by this workflow"
	case BuildInfo:
		sectionTitle = "📦 Artifacts published to Artifactory by this workflow"
	case Security:
		sectionTitle = "🔒 Security Summary"
	default:
		return "", fmt.Errorf("failed to get unknown section: %s, title", section)
	}
	return fmt.Sprintf("\n\n\n<details open>\n\n<summary>  %s </summary><p></p> \n\n %s \n\n</details>\n\n\n", sectionTitle, markdown), nil
}

func getSectionMarkdownContent(section MarkdownSection) (content string, err error) {
	basePath := os.Getenv(coreutils.OutputDirPathEnv)
	sectionFilepath := filepath.Join(basePath, "jfrog-command-summary", string(section), "markdown.md")

	// Check if the file exists
	if _, err := os.Stat(sectionFilepath); os.IsNotExist(err) {
		return "", nil
	}

	contentBytes, err := os.ReadFile(sectionFilepath)
	if err != nil {
		return
	}
	if len(contentBytes) == 0 {
		return "", nil
	}
	content, err = wrapCollapsableSection(section, string(contentBytes))
	if err != nil {
		return
	}
	return
}

func generateSectionMarkdown(section MarkdownSection, serverUrl string, majorVersion int) error {
	switch section {
	case Security:
		securitySummary, err := securityUtils.SecurityCommandsJobSummary()
		if err != nil {
			return err
		}
		return securitySummary.GenerateMarkdown()
	case BuildInfo:
		buildInfoSummary, err := commandssummaries.NewBuildInfoSummary(serverUrl, majorVersion)
		if err != nil {
			return err
		}
		return buildInfoSummary.GenerateMarkdown()
	case Upload:
		uploadSummary, err := commandssummaries.NewUploadSummary(serverUrl, majorVersion)
		if err != nil {
			return err
		}
		return uploadSummary.GenerateMarkdown()
	default:
		return fmt.Errorf("unknown section: %s", section)
	}
}

func createPlatformDetailsByFlags(c *cli.Context) (*coreConfig.ServerDetails, error) {
	platformDetails, err := cliutils.CreateServerDetailsWithConfigOffer(c, true, commonCliUtils.Platform)
	if err != nil {
		return nil, err
	}
	if platformDetails.Url == "" {
		return nil, errors.New("platform URL is mandatory for access token creation")
	}
	return platformDetails, nil
}

func extractServerUrlAndVersion(c *cli.Context) (string, int, error) {
	serverDetails, err := createPlatformDetailsByFlags(c)
	if err != nil {
		return "", 0, err
	}
	serverUrl := serverDetails.GetUrl()
	servicesManager, err := utils.CreateServiceManager(serverDetails, -1, 0, false)
	if err != nil {
		return "", 0, err
	}
	majorVersion, err := utils.GetRtMajorVersion(servicesManager)
	if err != nil {
		return "", 0, err
	}
	return serverUrl, majorVersion, nil
}
