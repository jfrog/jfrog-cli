package cliutils

const (
	// General CLI constants
	CliVersion  = "1.53.4"
	ClientAgent = "jfrog-cli-go"

	// CLI base commands constants:
	CmdArtifactory    = "rt"
	CmdBintray        = "bt"
	CmdMissionControl = "mc"
	CmdXray           = "xr"
	CmdCompletion     = "completion"
	CmdPlugin         = "plugin"
	CmdConfig         = "config"

	// Download
	DownloadMinSplitKb    = 5120
	DownloadSplitCount    = 3
	DownloadMaxSplitCount = 15

	// Common
	Retries             = 3
	Threads             = 3
	TokenExpiry         = 3600
	DefaultLicenseCount = 1

	// Env
	OfferConfig = "JFROG_CLI_OFFER_CONFIG"
	BuildUrl    = "JFROG_CLI_BUILD_URL"
	EnvExclude  = "JFROG_CLI_ENV_EXCLUDE"
	UserAgent   = "JFROG_CLI_USER_AGENT"
)
