package cliutils

const (
	// General CLI constants
	CliVersion           = "1.39.4"
	ClientAgent          = "jfrog-cli-go"
	OnErrorPanic OnError = "panic"

	// CLI base commands constants:
	CmdArtifactory    = "rt"
	CmdBintray        = "bt"
	CmdMissionControl = "mc"
	CmdXray           = "xr"
	CmdCompletion     = "completion"

	// Download
	DownloadMinSplitKb    = 5120
	DownloadSplitCount    = 3
	DownloadMaxSplitCount = 15

	// Common
	Retries             = 3
	TokenExpiry         = 3600
	DefaultLicenseCount = 1

	// Env
	OfferConfig   = "JFROG_CLI_OFFER_CONFIG"
	ErrorHandling = "JFROG_CLI_ERROR_HANDLING"
	TempDir       = "JFROG_CLI_TEMP_DIR"
	BuildUrl      = "JFROG_CLI_BUILD_URL"
	EnvExclude    = "JFROG_CLI_ENV_EXCLUDE"
	UserAgent     = "JFROG_CLI_USER_AGENT"
)
