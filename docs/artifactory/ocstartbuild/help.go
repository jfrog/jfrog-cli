package ocstartbuild

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description = "Run OpenShift CLI (oc) start-build command."

var Usage = []string{cliutils.CliExecutableName + " rt oc start-build <build config name | --from-build=<build name>> --repo=<target repository> [command options]"}
