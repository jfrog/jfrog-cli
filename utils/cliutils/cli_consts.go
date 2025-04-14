package cliutils

import "time"

const (
	// General CLI constants
	CliVersion  = "2.74.1"
	ClientAgent = "jfrog-cli-go"

	// CLI base commands constants:
	CmdArtifactory    = "rt"
	CmdMissionControl = "mc"
	CmdCompletion     = "completion"
	CmdPlugin         = "plugin"
	CmdConfig         = "config"
	CmdOptions        = "options"
	CmdPipelines      = "pl"

	// Common
	Retries                       = 3
	ArtifactoryTokenExpiry        = 3600
	DefaultLicenseCount           = 1
	LatestCliVersionCheckInterval = time.Hour * 6

	// Env
	UserAgent                      = "JFROG_CLI_USER_AGENT"
	JfrogCliAvoidNewVersionWarning = "JFROG_CLI_AVOID_NEW_VERSION_WARNING"
)
