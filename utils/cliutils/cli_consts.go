package cliutils

const (
	// General CLI constants
	CliVersion  = "2.26.0"
	ClientAgent = "jfrog-cli-go"

	// CLI base commands constants:
	CmdArtifactory    = "rt"
	CmdMissionControl = "mc"
	CmdXray           = "xr"
	CmdDistribution   = "ds"
	CmdCompletion     = "completion"
	CmdPlugin         = "plugin"
	CmdConfig         = "config"
	CmdOptions        = "options"
	CmdProject        = "project"

	// Download
	DownloadMinSplitKb    = 5120
	DownloadSplitCount    = 3
	DownloadMaxSplitCount = 15

	// Common
	Retries             = 3
	RetryWaitMilliSecs  = 0
	Threads             = 3
	TokenExpiry         = 3600
	DefaultLicenseCount = 1

	// Env
	BuildUrl                         = "JFROG_CLI_BUILD_URL"
	EnvExclude                       = "JFROG_CLI_ENV_EXCLUDE"
	UserAgent                        = "JFROG_CLI_USER_AGENT"
	JfrogCliAvoidDeprecationWarnings = "JFROG_CLI_AVOID_DEPRECATION_WARNINGS"
)
