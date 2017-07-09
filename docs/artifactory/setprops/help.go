package setprops

const Description = "Set properties."

var Usage = []string{"jfrog rt sp [command options] <artifacts pattern> <artifact properties>",
	"jfrog rt sp <artifact properties> --spec=<File Spec path> [command options]"}

const Arguments string =
`	artifacts pattern
		Specifies the artifacts path in Artifactory, which their properties are going to be update.

	artifact properties
		List of properties in the form of key1=value1;key2=value2,... to be set on the matched artifacts.
`