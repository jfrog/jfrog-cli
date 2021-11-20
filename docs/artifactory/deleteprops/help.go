package deleteprops

var Usage = []string{"rt delp [command options] <artifacts pattern> <artifact properties>",
	"rt delp <artifact properties> --spec=<File Spec path> [command options]"}

func GetDescription() string {
	return "Delete properties on existing files in Artifactory."
}

func GetArguments() string {
	return `	artifacts pattern
		Properties of artifacts that match this pattern will be removed.

	artifact properties
		The list of properties, in the form of key1,key2,..., to be removed from the matching artifacts.`
}
