package deleteprops

import "github.com/jfrog/jfrog-cli/docs/common"

var Usage = []string{"rt delp [command options] <files pattern> <properties list>",
	"rt delp <properties list> --spec=<File Spec path> [command options]"}

const EnvVar string = common.JfrogCliFailNoOp

func GetDescription() string {
	return "Delete properties on existing files in Artifactory."
}

func GetArguments() string {
	return `	files pattern
		Properties of artifacts that match this pattern will be removed.
		in the following format: <repository name>/<repository path>. You can use wildcards to specify multiple artifacts.	

	properties list
		List of comma-separated properties, in the form of key1,key2,..., to be removed from the matching files.`
}
