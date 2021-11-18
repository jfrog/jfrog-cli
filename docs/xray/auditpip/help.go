package auditpip

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description = "Execute an audit Pip command, using the configured Xray details."

var Usage = []string{cliutils.CliExecutableName + " xr audit-pip [command options]"}
