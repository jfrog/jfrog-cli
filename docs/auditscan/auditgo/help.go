package auditgo

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description = "Execute an audit Go command, using the configured Xray details."

var Usage = []string{cliutils.CliExecutableName + " audit-go [command options]"}
