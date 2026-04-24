package missioncontrol

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"text/tabwriter"

	commonCliUtils "github.com/jfrog/jfrog-cli-core/v2/common/cliutils"
	coreCommonCommands "github.com/jfrog/jfrog-cli-core/v2/common/commands"
	coreformat "github.com/jfrog/jfrog-cli-core/v2/common/format"
	corecommon "github.com/jfrog/jfrog-cli-core/v2/docs/common"
	"github.com/jfrog/jfrog-cli-core/v2/missioncontrol/commands"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-cli/docs/common"
	"github.com/jfrog/jfrog-cli/docs/missioncontrol/jpdadd"
	"github.com/jfrog/jfrog-cli/docs/missioncontrol/jpddelete"
	"github.com/jfrog/jfrog-cli/docs/missioncontrol/licenseacquire"
	"github.com/jfrog/jfrog-cli/docs/missioncontrol/licensedeploy"
	"github.com/jfrog/jfrog-cli/docs/missioncontrol/licenserelease"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/urfave/cli"
)

func GetCommands() []cli.Command {
	return cliutils.GetSortedCommands(cli.CommandsByName{
		{
			Name:         "license-acquire",
			Flags:        cliutils.GetCommandFlags(cliutils.LicenseAcquire),
			Usage:        licenseacquire.GetDescription(),
			HelpName:     corecommon.CreateUsage("mc license-acquire", licenseacquire.GetDescription(), licenseacquire.Usage),
			UsageText:    licenseacquire.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			Aliases:      []string{"la"},
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action:       licenseAcquire,
		},
		{
			Name:         "license-deploy",
			Flags:        cliutils.GetCommandFlags(cliutils.LicenseDeploy),
			Usage:        licensedeploy.GetDescription(),
			HelpName:     corecommon.CreateUsage("mc license-deploy", licensedeploy.GetDescription(), licensedeploy.Usage),
			UsageText:    licensedeploy.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			Aliases:      []string{"ld"},
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action:       licenseDeploy,
		},
		{
			Name:         "license-release",
			Flags:        cliutils.GetCommandFlags(cliutils.LicenseRelease),
			Usage:        licenserelease.GetDescription(),
			HelpName:     corecommon.CreateUsage("mc license-release", licenserelease.GetDescription(), licenserelease.Usage),
			UsageText:    licenserelease.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			Aliases:      []string{"lr"},
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action:       licenseRelease,
		},
		{
			Name:         "jpd-add",
			Flags:        cliutils.GetCommandFlags(cliutils.JpdAdd),
			Usage:        jpdadd.GetDescription(),
			HelpName:     corecommon.CreateUsage("mc jpd-add", jpdadd.GetDescription(), jpdadd.Usage),
			UsageText:    jpdadd.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			Aliases:      []string{"ja"},
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action:       jpdAdd,
		},
		{
			Name:         "jpd-delete",
			Flags:        cliutils.GetCommandFlags(cliutils.JpdDelete),
			Usage:        jpddelete.GetDescription(),
			HelpName:     corecommon.CreateUsage("mc jpd-delete", jpddelete.GetDescription(), jpddelete.Usage),
			UsageText:    jpddelete.GetArguments(),
			ArgsUsage:    common.CreateEnvVars(),
			Aliases:      []string{"jd"},
			BashComplete: corecommon.CreateBashCompletionFunc(),
			Action:       jpdDelete,
		},
	})
}

func jpdAdd(c *cli.Context) error {
	if len(c.Args()) != 1 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}
	outputFormat, err := getJpdAddOutputFormat(c)
	if err != nil {
		return err
	}
	jpdAddFlags, err := createJpdAddFlags(c)
	if err != nil {
		return err
	}
	body, err := commands.JpdAdd(jpdAddFlags)
	if err != nil {
		return err
	}
	return printJpdAddResponse(body, outputFormat, os.Stdout)
}

func getJpdAddOutputFormat(c *cli.Context) (coreformat.OutputFormat, error) {
	if !c.IsSet(cliutils.Format) {
		return coreformat.Json, nil
	}
	return coreformat.ParseOutputFormat(c.String(cliutils.Format), []coreformat.OutputFormat{coreformat.Json, coreformat.Table})
}

func printJpdAddResponse(body []byte, outputFormat coreformat.OutputFormat, w io.Writer) error {
	switch outputFormat {
	case coreformat.Json:
		log.Output(clientutils.IndentJson(body))
		return nil
	case coreformat.Table:
		return printJpdAddTable(body, w)
	default:
		return errorutils.CheckErrorf("unsupported format '%s' for jpd-add. Accepted values: table, json", outputFormat)
	}
}

func printJpdAddTable(body []byte, w io.Writer) error {
	var data map[string]any
	if err := json.Unmarshal(body, &data); err != nil {
		return errorutils.CheckErrorf("failed to parse jpd-add response: %s", err.Error())
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "FIELD\tVALUE")
	orderedKeys := []string{"id", "name", "url", "location", "licenses", "type"}
	printed := map[string]bool{}
	for _, k := range orderedKeys {
		if v, ok := data[k]; ok && v != nil {
			_, _ = fmt.Fprintf(tw, "%s\t%v\n", k, v)
			printed[k] = true
		}
	}
	remaining := []string{}
	for k := range data {
		if !printed[k] {
			remaining = append(remaining, k)
		}
	}
	sort.Strings(remaining)
	for _, k := range remaining {
		if data[k] != nil {
			_, _ = fmt.Fprintf(tw, "%s\t%v\n", k, data[k])
		}
	}
	return tw.Flush()
}

func jpdDelete(c *cli.Context) error {
	if len(c.Args()) != 1 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}
	mcDetails, err := createMissionControlDetails(c)
	if err != nil {
		return err
	}
	return commands.JpdDelete(c.Args()[0], mcDetails)
}

func licenseAcquire(c *cli.Context) error {
	size := len(c.Args())
	if size != 2 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}
	outputFormat, err := getLicenseAcquireOutputFormat(c)
	if err != nil {
		return err
	}
	mcDetails, err := createMissionControlDetails(c)
	if err != nil {
		return err
	}
	licenseKey, err := commands.LicenseAcquire(c.Args()[0], c.Args()[1], mcDetails)
	if err != nil {
		return err
	}
	return printLicenseAcquireResponse(licenseKey, outputFormat, os.Stdout)
}

// getLicenseAcquireOutputFormat defaults to table for a labeled output.
func getLicenseAcquireOutputFormat(c *cli.Context) (coreformat.OutputFormat, error) {
	if !c.IsSet(cliutils.Format) {
		return coreformat.Table, nil
	}
	return coreformat.ParseOutputFormat(c.String(cliutils.Format), []coreformat.OutputFormat{coreformat.Table, coreformat.Json})
}

func printLicenseAcquireResponse(licenseKey string, outputFormat coreformat.OutputFormat, w io.Writer) error {
	switch outputFormat {
	case coreformat.Json:
		log.Output(clientutils.IndentJson([]byte(`{"license_key":"` + licenseKey + `"}`)))
		return nil
	case coreformat.Table:
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintln(tw, "FIELD\tVALUE")
		_, _ = fmt.Fprintf(tw, "license_key\t%s\n", licenseKey)
		return tw.Flush()
	default:
		return errorutils.CheckErrorf("unsupported format '%s' for license-acquire. Accepted values: table, json", outputFormat)
	}
}

func licenseDeploy(c *cli.Context) error {
	size := len(c.Args())
	if size != 2 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}
	outputFormat, err := getLicenseDeployOutputFormat(c)
	if err != nil {
		return err
	}
	flags, err := createLicenseDeployFlags(c)
	if err != nil {
		return err
	}
	body, err := commands.LicenseDeploy(c.Args()[0], c.Args()[1], flags)
	if err != nil {
		return err
	}
	return printLicenseDeployResponse(body, outputFormat, os.Stdout)
}

func getLicenseDeployOutputFormat(c *cli.Context) (coreformat.OutputFormat, error) {
	if !c.IsSet(cliutils.Format) {
		return coreformat.Json, nil
	}
	return coreformat.ParseOutputFormat(c.String(cliutils.Format), []coreformat.OutputFormat{coreformat.Json, coreformat.Table})
}

func printLicenseDeployResponse(body []byte, outputFormat coreformat.OutputFormat, w io.Writer) error {
	switch outputFormat {
	case coreformat.Json:
		log.Output(clientutils.IndentJson(body))
		return nil
	case coreformat.Table:
		return printLicenseDeployTable(body, w)
	default:
		return errorutils.CheckErrorf("unsupported format '%s' for license-deploy. Accepted values: table, json", outputFormat)
	}
}

func printLicenseDeployTable(body []byte, w io.Writer) error {
	var data map[string]any
	if err := json.Unmarshal(body, &data); err != nil {
		return errorutils.CheckErrorf("failed to parse license-deploy response: %s", err.Error())
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "FIELD\tVALUE")
	orderedKeys := []string{"bucket_id", "jpd_id", "license_count", "status"}
	printed := map[string]bool{}
	for _, k := range orderedKeys {
		if v, ok := data[k]; ok && v != nil {
			_, _ = fmt.Fprintf(tw, "%s\t%v\n", k, v)
			printed[k] = true
		}
	}
	remaining := []string{}
	for k := range data {
		if !printed[k] {
			remaining = append(remaining, k)
		}
	}
	sort.Strings(remaining)
	for _, k := range remaining {
		if data[k] != nil {
			_, _ = fmt.Fprintf(tw, "%s\t%v\n", k, data[k])
		}
	}
	return tw.Flush()
}

func licenseRelease(c *cli.Context) error {
	size := len(c.Args())
	if size != 2 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}
	mcDetails, err := createMissionControlDetails(c)
	if err != nil {
		return err
	}
	return commands.LicenseRelease(c.Args()[0], c.Args()[1], mcDetails)
}

func offerConfig(c *cli.Context) (*config.ServerDetails, error) {
	confirmed, err := commonCliUtils.ShouldOfferConfig()
	if !confirmed || err != nil {
		return nil, err
	}
	details, err := createMCDetailsFromFlags(c)
	if err != nil {
		return nil, err
	}
	configCmd := coreCommonCommands.NewConfigCommand(coreCommonCommands.AddOrEdit, details.ServerId).SetDefaultDetails(details).SetInteractive(true)
	err = configCmd.Run()
	if err != nil {
		return nil, err
	}

	return configCmd.ServerDetails()
}

func createLicenseDeployFlags(c *cli.Context) (flags *commands.LicenseDeployFlags, err error) {
	flags = new(commands.LicenseDeployFlags)
	flags.ServerDetails, err = createMissionControlDetails(c)
	if err != nil {
		return
	}
	flags.LicenseCount = cliutils.DefaultLicenseCount
	if c.IsSet("license-count") {
		flags.LicenseCount, err = strconv.Atoi(c.String("license-count"))
		if err != nil {
			return nil, cliutils.PrintHelpAndReturnError("The '--license-count' option must have a numeric value. ", c)
		}
		if flags.LicenseCount < 1 {
			return nil, cliutils.PrintHelpAndReturnError("The --license-count option must be at least "+strconv.Itoa(cliutils.DefaultLicenseCount), c)
		}
	}
	return
}

func createJpdAddFlags(c *cli.Context) (flags *commands.JpdAddFlags, err error) {
	flags = new(commands.JpdAddFlags)
	flags.ServerDetails, err = createMissionControlDetails(c)
	if err != nil {
		return
	}
	flags.JpdConfig, err = fileutils.ReadFile(c.Args()[0])
	if errorutils.CheckError(err) != nil {
		return
	}
	return
}

func createMissionControlDetails(c *cli.Context) (*config.ServerDetails, error) {
	createdDetails, err := offerConfig(c)
	if err != nil {
		return nil, err
	}
	if createdDetails != nil {
		return createdDetails, nil
	}

	details, err := createMCDetailsFromFlags(c)
	if err != nil {
		return nil, err
	}
	// If urls or credentials were passed as options, use options as they are.
	// For security reasons, we'd like to avoid using part of the connection details from command options and the rest from the config.
	// Either use command options only or config only.
	if credentialsChanged(details) {
		return details, nil
	}

	// Else, use details from config for requested serverId, or for default server if empty.
	confDetails, err := coreCommonCommands.GetConfig(details.ServerId, true)
	if err != nil {
		return nil, err
	}

	confDetails.Url = clientutils.AddTrailingSlashIfNeeded(confDetails.MissionControlUrl)
	return confDetails, nil
}

func createMCDetailsFromFlags(c *cli.Context) (details *config.ServerDetails, err error) {
	details, err = cliutils.CreateServerDetailsFromFlags(c)
	if err != nil {
		return
	}
	details.MissionControlUrl = details.Url
	details.Url = ""
	return
}

func credentialsChanged(details *config.ServerDetails) bool {
	return details.MissionControlUrl != "" || details.User != "" || details.Password != "" || details.AccessToken != ""
}
