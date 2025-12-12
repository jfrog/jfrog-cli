package cliutils

import "time"

const (
	// General CLI constants
	CliVersion  = "2.87.0"
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
	//#nosec G101
	JfrogCliGithubToken = "JFROG_CLI_GITHUB_TOKEN"
	JfrogCliHideSurvey  = "JFROG_CLI_HIDE_SURVEY"
)
