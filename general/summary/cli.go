package summary

import (
	"errors"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/commands/commandssummaries"
	"github.com/jfrog/jfrog-cli-core/v2/artifactory/utils"
	"github.com/jfrog/jfrog-cli-core/v2/commandsummary"
	commonCliUtils "github.com/jfrog/jfrog-cli-core/v2/common/cliutils"
	coreConfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/urfave/cli"
)

var markdownSections = []string{"security", "build-info", "upload"}

func CreateSummaryMarkdown(c *cli.Context) (err error) {

	serverUrl, majorVersion, err := extractServerUrlAndVersion(c)
	if err != nil {
		return err
	}

	for _, section := range markdownSections {
		switch section {
		case "security":
		case "build-info":
			buildInfoSummary, _ := commandsummary.New(commandssummaries.NewBuildInfoWithUrl(serverUrl, majorVersion), "build-info")
			_ = buildInfoSummary.GenerateMarkdown()
		case "upload":
			uploadSummary, _ := commandsummary.New(commandssummaries.NewUploadSummary(serverUrl, majorVersion), "upload")
			_ = uploadSummary.GenerateMarkdown()
		default:
			log.Warn("unknown section")
		}
	}
	return
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
