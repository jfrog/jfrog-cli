package releasebundleupdate

import "github.com/jfrog/jfrog-cli/utils/cliutils"

const Description = "Updates an existing unsigned release bundle version."

var Usage = []string{cliutils.CliExecutableName + " ds rbu [command options] <release bundle name> <release bundle version> <pattern>",
	cliutils.CliExecutableName + " ds rbu --spec=<File Spec path> [command options] <release bundle name> <release bundle version>"}

const Arguments string = `	release bundle name
		The name of the release bundle.

	release bundle version
		The release bundle version.

	pattern
		Specifies the source path in Artifactory, from which the artifacts should be bundled,
		in the following format: <repository name>/<repository path>. You can use wildcards to specify multiple artifacts.`
