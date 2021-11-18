package auditgradle

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description = "Execute an audit Gradle command, using the configured Xray details."

var Usage = []string{cliutils.CliExecutableName + " audit-gradle [command options]"}
