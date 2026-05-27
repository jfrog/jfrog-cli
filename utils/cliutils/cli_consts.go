package cliutils

import "time"

const (
	// General CLI constants
	CliVersion  = "2.105.0"
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
	// JfrogCliErrorOutputFormat controls how HTTP response errors are surfaced.
	// Set to "json" to emit the structured response (status code + body) as JSON
	// on stderr instead of the default human-readable text. Unset or "text" keeps
	// the legacy behavior. Applies uniformly to all commands, including OIDC
	// token-exchange failures.
	JfrogCliErrorOutputFormat = "JFROG_CLI_ERROR_OUTPUT_FORMAT"
)

// ErrorFormatJSON is the env-var value that switches HTTP error reporting to JSON-on-stderr.
const ErrorFormatJSON = "json"
