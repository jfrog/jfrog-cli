package cliutils

const (
	// General CLI constants
	CliVersion           = "1.19.1"
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
)
