package cliutils

import "time"

const (
	// General CLI constants
	CliVersion  = "2.71.5"
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
	CmdPipelines      = "pl"

	// Download
	DownloadMinSplitKb    = 5120
	DownloadSplitCount    = 3
	DownloadMaxSplitCount = 15

	// Upload
	UploadMinSplitMb    = 200
	UploadSplitCount    = 5
	UploadMaxSplitCount = 100
	UploadChunkSizeMb   = 20

	// Common
	Retries                       = 3
	RetryWaitMilliSecs            = 0
	ArtifactoryTokenExpiry        = 3600
	DefaultLicenseCount           = 1
	LatestCliVersionCheckInterval = time.Hour * 6

	// Env
	BuildUrl                       = "JFROG_CLI_BUILD_URL"
	EnvExclude                     = "JFROG_CLI_ENV_EXCLUDE"
	UserAgent                      = "JFROG_CLI_USER_AGENT"
	JfrogCliAvoidNewVersionWarning = "JFROG_CLI_AVOID_NEW_VERSION_WARNING"
)
