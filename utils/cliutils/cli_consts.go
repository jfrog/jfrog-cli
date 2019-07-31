package cliutils

const (
	// General CLI constants
	CliVersion           = "1.26.3"
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
	Retries = 3

	// Env
	ReportUsage           = "JFROG_CLI_REPORT_USAGE"
	LogLevel              = "JFROG_CLI_LOG_LEVEL"
	OfferConfig           = "JFROG_CLI_OFFER_CONFIG"
	JfrogHomeDirEnv       = "JFROG_CLI_HOME_DIR"
	JFrogCliErrorHandling = "JFROG_CLI_ERROR_HANDLING"
	JFrogCliTempDir       = "JFROG_CLI_TEMP_DIR"
	CI                    = "CI"
	// Deprecated:
	JfrogHomeEnv = "JFROG_CLI_HOME"
)
