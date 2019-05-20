package cliutils

const (
	// General CLI constants
	CliVersion           = "1.25.0"
	ClientAgent          = "jfrog-cli-go"
	OnErrorPanic OnError = "panic"

	// CLI base commands constants:
	CmdArtifactory    = "rt"
	CmdBintray        = "bt"
	CmdMissionControl = "mc"
	CmdXray           = "xr"

	// Download
	DownloadMinSplitKb    = 5120
	DownloadSplitCount    = 3
	DownloadMaxSplitCount = 15

	// Common
	Retries = 3

	// Env
	ReportUsage     = "JFROG_CLI_REPORT_USAGE"
	LogLevel        = "JFROG_CLI_LOG_LEVEL"
	OfferConfig     = "JFROG_CLI_OFFER_CONFIG"
	JfrogHomeDirEnv = "JFROG_CLI_HOME_DIR"
	// Deprecated:
	JfrogHomeEnv = "JFROG_CLI_HOME"
)
