package setprops

import "github.com/jfrog/jfrog-cli/docs/common"

var Usage = []string{"rt sp [command options] <files pattern> <file properties>",
	"rt sp <file properties> --spec=<File Spec path> [command options]"}

const EnvVar string = common.JfrogCliFailNoOp

func GetDescription() string {
	return "Set properties on existing files in Artifactory."
}

func GetArguments() string {
	return `	files pattern
		Artifacts that match the pattern will be set with the specified properties.

	file properties
		List of semicolon-separated key-value properties, in the form of "key1=value1;key2=value2;..." to be set on the matching artifacts.`
}
