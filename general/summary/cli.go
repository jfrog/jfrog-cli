package summary

import (
	"errors"
	"fmt"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/commandssummaries"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/v2/commandsummary"
	commonCliUtils "github.com/jfrog/jfrog-cli-core/v2/common/cliutils"
	coreConfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	securityUtils "github.com/jfrog/jfrog-cli-security/utils"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/urfave/cli"
	"os"
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

func CreateSummaryMarkdown(c *cli.Context) (err error) {
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
	return
}

func generateSectionMarkdown(section MarkdownSection, serverUrl string, majorVersion int) error {
	switch section {
	case Security:
		// Handle security section
		securitySummary, err := securityUtils.SecurityCommandsJobSummary()
		if err != nil {
			return err
		}
		return securitySummary.GenerateMarkdown()
	case BuildInfo:
		buildInfoSummary, _ := commandsummary.New(commandssummaries.NewBuildInfoWithUrl(serverUrl, majorVersion), string(BuildInfo))
		return buildInfoSummary.GenerateMarkdown()
	case Upload:
		uploadSummary, _ := commandsummary.New(commandssummaries.NewUploadSummary(serverUrl, majorVersion), string(Upload))
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
