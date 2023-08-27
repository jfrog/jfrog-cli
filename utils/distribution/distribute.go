package distribution

import (
	"github.com/jfrog/jfrog-cli-core/v2/common/spec"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	distributionUtils "github.com/jfrog/jfrog-client-go/utils/distribution"
	"github.com/urfave/cli"
)

func CreateDefaultDistributionRules(c *cli.Context) *spec.DistributionRules {
	return &spec.DistributionRules{
		DistributionRules: []spec.DistributionRule{{
			SiteName:     c.String("site"),
			CityName:     c.String("city"),
			CountryCodes: cliutils.GetStringsArrFlagValue(c, "country-codes"),
		}},
	}
}

func ValidateReleaseBundleDistributeCmd(c *cli.Context) error {
	if c.NArg() != 2 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}
	if c.IsSet("max-wait-minutes") && !c.IsSet("sync") {
		return cliutils.PrintHelpAndReturnError("The --max-wait-minutes option can't be used without --sync", c)
	}

	if c.IsSet("dist-rules") && (c.IsSet("site") || c.IsSet("city") || c.IsSet("country-code")) {
		return cliutils.PrintHelpAndReturnError("The --dist-rules option can't be used with --site, --city or --country-code", c)
	}

	return nil
}

func InitReleaseBundleDistributeCmd(c *cli.Context) (distributionRules *spec.DistributionRules, maxWaitMinutes int, params distributionUtils.DistributionParams, err error) {
	if c.IsSet("dist-rules") {
		distributionRules, err = spec.CreateDistributionRulesFromFile(c.String("dist-rules"))
		if err != nil {
			return
		}
	} else {
		distributionRules = CreateDefaultDistributionRules(c)
	}

	maxWaitMinutes, err = cliutils.GetIntFlagValue(c, "max-wait-minutes", 60)
	if err != nil {
		return
	}

	params = distributionUtils.NewDistributeReleaseBundleParams(c.Args().Get(0), c.Args().Get(1))
	return
}
